#!/bin/bash
# Inteliside CLI — Install Script (macOS/Linux)
# Usage: curl -fsSL https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.sh | bash

set -euo pipefail

REPO="Intelliaa/inteliside-cli"
BINARY="inteliside"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
PURPLE='\033[0;35m'
NC='\033[0m'

echo ""
echo -e "${PURPLE}  ╔══════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}  ║        Inteliside CLI — Installer        ║${NC}"
echo -e "${PURPLE}  ╚══════════════════════════════════════════╝${NC}"
echo ""

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    *)
        echo -e "${RED}  Error: OS no soportado: $OS${NC}"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo -e "${RED}  Error: Arquitectura no soportada: $ARCH${NC}"
        exit 1
        ;;
esac

echo "  Detectado: ${OS}/${ARCH}"

# Get latest release tag
echo "  Obteniendo última versión..."
if command -v gh &> /dev/null; then
    TAG=$(gh release view --repo "$REPO" --json tagName -q .tagName 2>/dev/null || echo "")
fi

if [ -z "${TAG:-}" ]; then
    TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
fi

if [ -z "${TAG:-}" ]; then
    echo -e "${RED}  Error: No se pudo obtener la última versión${NC}"
    exit 1
fi

VERSION="${TAG#v}"
echo "  Versión: ${TAG}"

# Download binary
FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"

echo "  Descargando ${FILENAME}..."
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

curl -fsSL "${URL}" -o "${TMP_DIR}/${FILENAME}"

# Extract
echo "  Extrayendo..."
tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"

# Install
echo "  Instalando en ${INSTALL_DIR}..."
if [ -w "${INSTALL_DIR}" ]; then
    mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
    sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

# Verify
if command -v inteliside &> /dev/null; then
    echo ""
    echo -e "${GREEN}  ✓ Instalado exitosamente: $(inteliside version)${NC}"
    echo ""
    echo "  Ejecuta 'inteliside' para comenzar."
    echo ""
else
    echo ""
    echo -e "${RED}  ⚠ Instalado pero ${INSTALL_DIR} no está en tu PATH${NC}"
    echo "  Agrega a tu shell config:"
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
fi
