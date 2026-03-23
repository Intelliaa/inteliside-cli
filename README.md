# Inteliside CLI

CLI del [Marketplace de Plugins Inteliside](https://github.com/Intelliaa/marketplace-plugins-inteliside) para Claude Code. Automatiza la instalación y post-configuración de plugins con un solo comando.

**Un comando. Cualquier rol. Cualquier plataforma.**

## Instalación

### macOS / Linux

```bash
# Homebrew (recomendado)
brew install Intelliaa/tap/inteliside

# curl
curl -fsSL https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.sh | bash

# Go
go install github.com/Intelliaa/inteliside-cli/cmd/inteliside@latest
```

### Windows

```powershell
irm https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.ps1 | iex
```

## Uso rápido

```bash
# TUI interactiva (recomendado)
inteliside

# Instalación directa por rol
inteliside install --preset dev        # Desarrollador
inteliside install --preset pm         # Product Manager
inteliside install --preset designer   # Diseñador UI/UX
inteliside install --preset fullstack  # Todo el equipo
inteliside install --preset automation # Automatización n8n

# Plugins específicos
inteliside install --plugin atl-inteliside,n8n-studio

# Preview sin cambios
inteliside install --preset dev --dry-run
```

## Comandos

| Comando | Descripción |
|---------|-------------|
| `inteliside` | TUI interactiva guiada |
| `inteliside install` | Instala plugins + dependencias + post-config |
| `inteliside setup <plugin>` | Re-configura un plugin individual |
| `inteliside doctor` | Diagnostica dependencias con fixes sugeridos |
| `inteliside verify` | Health check post-instalación |
| `inteliside sync` | Verifica actualizaciones del marketplace |
| `inteliside list` | Lista plugins y presets disponibles |
| `inteliside config show` | Muestra configuración actual |
| `inteliside config reset <plugin>` | Elimina config de un plugin |
| `inteliside backup list` | Lista backups disponibles |
| `inteliside backup restore <id>` | Restaura desde un backup |
| `inteliside version` | Versión del CLI |

## Presets por rol

| Preset | Plugins | Qué configura |
|--------|---------|---------------|
| `pm` | SDD-Wizards | GitHub CLI auth |
| `designer` | UX Studio, SDD-Wizards | Figma Console MCP + token |
| `dev` | ATL Inteliside, SDD-Wizards | Engram + GitHub labels + rules + CLAUDE.md |
| `fullstack` | Los 6 plugins | Todo combinado |
| `automation` | n8n Studio | n8n MCP server |
| `legacy` | SDD-Legacy, ATL Inteliside | Engram + prereq checks |
| `custom` | Selección en TUI | Lo que elijas |

## Qué hace el CLI

El CLI automatiza toda la post-configuración que cada plugin necesita:

- **GitHub CLI**: Verifica autenticación y scopes
- **Engram**: Instala binary + conecta como plugin de Claude Code
- **MCP Servers**: Registra Figma Console y n8n MCP en settings.json
- **GitHub Labels**: Crea los 9+ labels que ATL necesita en tu repo
- **Claude Rules**: Copia los 5 archivos de reglas a `.claude/rules/`
- **CLAUDE.md**: Genera la sección de configuración ATL con marcadores

Todo con **additive-only merge** — nunca sobreescribe tu configuración existente.

## Seguridad

- Backup automático antes de cada instalación (`~/.inteliside/backups/`)
- `--dry-run` para preview sin cambios
- Restauración con `inteliside backup restore <id>`
- Tokens se almacenan solo donde Claude Code los espera (settings.json)

## Desarrollo

```bash
# Clonar
git clone https://github.com/Intelliaa/inteliside-cli.git
cd inteliside-cli

# Build
go build ./cmd/inteliside/

# Tests
go test ./...

# Run
./inteliside
```

## Licencia

MIT — [Inteliside](https://github.com/Intelliaa)
