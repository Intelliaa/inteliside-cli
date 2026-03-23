# Inteliside CLI — Install Script (Windows)
# Usage: irm https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "Intelliaa/inteliside-cli"
$Binary = "inteliside.exe"
$InstallDir = "$env:LOCALAPPDATA\inteliside\bin"

Write-Host ""
Write-Host "  ╔══════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "  ║        Inteliside CLI — Installer        ║" -ForegroundColor Magenta
Write-Host "  ╚══════════════════════════════════════════╝" -ForegroundColor Magenta
Write-Host ""

# Detect architecture
$Arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "  Error: Se requiere sistema de 64 bits" -ForegroundColor Red
    exit 1
}

Write-Host "  Detectado: windows/${Arch}"

# Get latest release
Write-Host "  Obteniendo ultima version..."
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/${Repo}/releases/latest"
$Tag = $Release.tag_name
$Version = $Tag.TrimStart("v")

Write-Host "  Version: ${Tag}"

# Download
$Filename = "inteliside_${Version}_windows_${Arch}.zip"
$Url = "https://github.com/${Repo}/releases/download/${Tag}/${Filename}"

Write-Host "  Descargando ${Filename}..."
$TmpDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "inteliside-install")
$ZipPath = Join-Path $TmpDir $Filename

Invoke-WebRequest -Uri $Url -OutFile $ZipPath

# Extract
Write-Host "  Extrayendo..."
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

# Install
Write-Host "  Instalando en ${InstallDir}..."
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item -Path (Join-Path $TmpDir $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

# Add to PATH if needed
$CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$CurrentPath;$InstallDir", "User")
    Write-Host "  Agregado ${InstallDir} al PATH del usuario"
}

# Cleanup
Remove-Item -Recurse -Force $TmpDir

# Verify
Write-Host ""
& (Join-Path $InstallDir $Binary) version
Write-Host ""
Write-Host "  Instalado exitosamente." -ForegroundColor Green
Write-Host "  Reinicia tu terminal y ejecuta 'inteliside' para comenzar."
Write-Host ""
