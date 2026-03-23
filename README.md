# Inteliside CLI

CLI del [Marketplace de Plugins Inteliside](https://github.com/Intelliaa/marketplace-plugins-inteliside) para Claude Code. Automatiza la instalación y post-configuración de plugins con un solo comando.

**Un comando. Cualquier rol. Cualquier plataforma.**

## Instalación del CLI

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

## Flujo de uso

El CLI separa dos momentos: **install** (una vez por máquina) e **init** (una vez por proyecto).

### Paso 1: Instalar dependencias globales

Ejecutar una sola vez en tu máquina. Configura MCP servers, Engram, GitHub auth y todo lo que vive fuera del proyecto.

```bash
# TUI interactiva (recomendado)
inteliside

# O directo por rol
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

**Qué hace `install` por preset:**

| Preset | Dependencias que configura |
|--------|---------------------------|
| `pm` | GitHub CLI auth + scope `repo` |
| `designer` | GitHub CLI + Figma Console MCP + Google Stitch MCP + tokens |
| `dev` | GitHub CLI + Engram (binary + plugin de Claude Code) |
| `fullstack` | Todo lo anterior + n8n MCP |
| `automation` | n8n MCP server (default compartido de Inteliside) |
| `legacy` | GitHub CLI + Engram |

### Paso 2: Inicializar cada proyecto

Ejecutar en la raíz de cada repositorio nuevo. Genera los `CLAUDE.md` que cada plugin necesita en la ubicación correcta, más rules y GitHub labels.

```bash
cd mi-proyecto

# Interactivo — pregunta nombre, GitHub owner, Engram project, etc.
inteliside init --preset dev

# Sin preguntas, usa defaults
inteliside init --preset fullstack --yes

# Solo un plugin
inteliside init --plugin ux-studio

# Preview
inteliside init --preset dev --dry-run
```

**Qué genera `init` por preset:**

| Preset | Archivos generados |
|--------|-------------------|
| `pm` | `docs/CLAUDE.md` — espacio de trabajo del PM |
| `designer` | `docs/CLAUDE.md` + `docs/ux-ui/CLAUDE.md` — PM + Designer |
| `dev` | `CLAUDE.md` (raíz) + `docs/CLAUDE.md` + `.claude/rules/` (5 archivos) + GitHub labels (9) |
| `fullstack` | Los 3 CLAUDE.md + `docs/ux-ui/CLAUDE.md` + rules + labels |
| `automation` | `CLAUDE.md` (raíz) — config n8n Studio |
| `legacy` | `CLAUDE.md` (raíz) + `docs/CLAUDE.md` + rules + labels |

**Ubicación de cada CLAUDE.md:**

```
mi-proyecto/
├── CLAUDE.md                    ← ATL Inteliside (Dev) o n8n Studio
├── docs/
│   ├── CLAUDE.md                ← SDD-Wizards (PM)
│   └── ux-ui/
│       └── CLAUDE.md            ← UX Studio (Designer)
└── .claude/
    └── rules/
        ├── atl-workflow.md      ← Pipeline ATL
        ├── engram-protocol.md   ← Memoria compartida
        ├── subagent-architecture.md
        ├── context-monitoring.md
        └── team-rules.md
```

Cada rol ejecuta Claude Code desde su directorio para que cargue el CLAUDE.md correcto:

```bash
# PM trabaja desde docs/
cd docs && claude
/sdd-wizards:prd-wizard

# Designer trabaja desde docs/ux-ui/
cd docs/ux-ui && claude
/ux-studio:ux-orchestrator

# Dev trabaja desde la raíz
cd mi-proyecto && claude
/atl-inteliside:orchestrador feat-login
```

`init` es idempotente — nunca sobreescribe archivos existentes. Si un CLAUDE.md ya existe, lo salta.

### Paso 3: Instalar los plugins en Claude Code

Después de `install` + `init`, instala los plugins del marketplace:

```bash
claude
/plugin marketplace add Intelliaa/marketplace-plugins-inteliside
/plugin install sdd-wizards@marketplace-plugins-inteliside
/plugin install atl-inteliside@marketplace-plugins-inteliside
/plugin install ux-studio@marketplace-plugins-inteliside
```

## Ejemplo completo: nuevo proyecto fullstack

```bash
# 1. Instalar deps globales (una sola vez)
inteliside install --preset fullstack

# 2. Crear proyecto
mkdir mi-app && cd mi-app && git init

# 3. Inicializar estructura
inteliside init --preset fullstack
# → Pregunta: nombre del proyecto, GitHub owner/repo, Engram project, Figma URL

# 4. Verificar que todo está OK
inteliside doctor
inteliside verify

# 5. Instalar plugins en Claude Code
claude
/plugin marketplace add Intelliaa/marketplace-plugins-inteliside
/plugin install sdd-wizards@marketplace-plugins-inteliside
/plugin install ux-studio@marketplace-plugins-inteliside
/plugin install atl-inteliside@marketplace-plugins-inteliside

# 6. Trabajar
cd docs && claude                    # PM: /sdd-wizards:prd-wizard
cd docs/ux-ui && claude              # Designer: /ux-studio:ux-orchestrator
cd mi-app && claude                  # Dev: /atl-inteliside:orchestrador feat-login
```

## Todos los comandos

| Comando | Tipo | Descripción |
|---------|------|-------------|
| `inteliside` | Global | TUI interactiva guiada |
| `inteliside install --preset <rol>` | Global | Instala deps globales (MCP servers, Engram, gh) |
| `inteliside init --preset <rol>` | Proyecto | Genera CLAUDE.md, rules y labels en el proyecto |
| `inteliside setup <plugin>` | Global | Re-configura un plugin individual |
| `inteliside doctor` | Global | Diagnostica dependencias con fixes sugeridos |
| `inteliside verify` | Global | Health check de MCP servers y deps |
| `inteliside sync` | Global | Verifica actualizaciones del marketplace |
| `inteliside list` | Info | Lista plugins y presets disponibles |
| `inteliside config show` | Info | Muestra configuración actual |
| `inteliside config reset <plugin>` | Global | Elimina config de un plugin |
| `inteliside backup list` | Info | Lista backups disponibles |
| `inteliside backup restore <id>` | Global | Restaura desde un backup |
| `inteliside version` | Info | Versión del CLI |

## Presets por rol

| Preset | Plugins incluidos |
|--------|-------------------|
| `pm` | SDD-Wizards |
| `designer` | UX Studio, SDD-Wizards |
| `dev` | ATL Inteliside, SDD-Wizards |
| `fullstack` | Los 6 plugins |
| `automation` | n8n Studio |
| `legacy` | SDD-Legacy, ATL Inteliside |
| `custom` | Selección manual en TUI |

## Seguridad

- **Backup automático** antes de cada operación (`~/.inteliside/backups/`)
- **`--dry-run`** en todos los comandos para preview sin cambios
- **Additive-only merge** — nunca sobreescribe configuración existente del usuario
- **Idempotente** — ejecutar dos veces no duplica ni rompe nada
- **Tokens** se almacenan solo donde Claude Code los espera (settings.json)
- **Restauración** con `inteliside backup restore <id>`

## Desarrollo

```bash
git clone https://github.com/Intelliaa/inteliside-cli.git
cd inteliside-cli

go build ./cmd/inteliside/    # Build
go test ./...                 # Tests (22 unit tests)
go vet ./...                  # Linting
./inteliside                  # Run
```

## Licencia

MIT — [Inteliside](https://github.com/Intelliaa)
