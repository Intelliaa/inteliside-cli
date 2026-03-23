# Inteliside CLI

Instalador y configurador del [Marketplace de Plugins Inteliside](https://github.com/Intelliaa/marketplace-plugins-inteliside) para Claude Code.

Automatiza todo el setup que normalmente harías a mano: instalar dependencias, configurar servidores MCP, generar archivos de configuración y preparar tu proyecto para trabajar con los plugins.

---

## Requisitos previos

Antes de instalar el CLI, asegúrate de tener:

- **Claude Code** instalado (`npm install -g @anthropic-ai/claude-code`)
- **Node.js 18+** instalado ([nodejs.org](https://nodejs.org/))
- **Git** instalado
- **GitHub CLI** instalado (`brew install gh` en macOS)

---

## Guía de instalación paso a paso

### Paso 1 — Instalar el CLI

Elige una de estas opciones:

**macOS / Linux (Homebrew — recomendado):**

```bash
brew install Intelliaa/tap/inteliside
```

**macOS / Linux (script):**

```bash
curl -fsSL https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.sh | bash
```

**Windows:**

```powershell
irm https://raw.githubusercontent.com/Intelliaa/inteliside-cli/main/scripts/install.ps1 | iex
```

Verifica que se instaló correctamente:

```bash
inteliside version
```

---

### Paso 2 — Configurar tu máquina (una sola vez)

Este paso instala las dependencias globales que los plugins necesitan: servidores MCP, Engram (memoria persistente), y autenticación de GitHub.

Se ejecuta **una sola vez** por máquina, no por proyecto.

**Opción A — TUI interactiva (recomendado para primera vez):**

```bash
inteliside
```

Se abre un menú visual donde seleccionas tu rol y el CLI hace todo automáticamente.

**Opción B — Directo por rol:**

```bash
inteliside install --preset dev
```

Los roles disponibles son:

| Rol | Comando | Qué instala |
|-----|---------|-------------|
| Product Manager | `inteliside install --preset pm` | GitHub CLI autenticado |
| Diseñador UI/UX | `inteliside install --preset designer` | GitHub CLI + Figma Console MCP |
| Desarrollador | `inteliside install --preset dev` | GitHub CLI + Engram |
| Todo el equipo | `inteliside install --preset fullstack` | Todo lo anterior + n8n MCP |
| Automatización | `inteliside install --preset automation` | n8n MCP |

Durante la instalación el CLI te pedirá las API keys necesarias (Figma, etc.). Puedes pegarlas directamente desde el portapapeles.

Para ver qué haría sin ejecutar nada:

```bash
inteliside install --preset dev --dry-run
```

---

### Paso 3 — Preparar tu proyecto (una vez por proyecto)

Este paso genera los archivos de configuración que cada plugin necesita **dentro de tu proyecto**: archivos `CLAUDE.md`, reglas del pipeline, labels de GitHub, y servidores MCP específicos del proyecto.

```bash
cd mi-proyecto
inteliside init --preset dev
```

El CLI te preguntará datos del proyecto:

- Nombre del proyecto
- GitHub owner y repositorio
- Proyecto de Engram (para memoria compartida)
- URL de Figma (si eres designer)
- Google Cloud Project ID (si usas Stitch para diseño)

Estos datos se guardan en los archivos generados, no se envían a ningún servidor.

**Archivos que genera según el rol:**

| Rol | Archivos |
|-----|----------|
| PM | `docs/CLAUDE.md` |
| Designer | `docs/CLAUDE.md` + `docs/ux-ui/CLAUDE.md` + `.claude/settings.json` (Stitch MCP) |
| Dev | `CLAUDE.md` + `docs/CLAUDE.md` + `.claude/rules/` (5 archivos) + GitHub labels |
| Full Stack | Todos los anteriores combinados |
| Automation | `CLAUDE.md` (configuración n8n) |

La estructura resultante en tu proyecto:

```
mi-proyecto/
├── CLAUDE.md                      ← Configuración del Dev
├── .claude/
│   ├── settings.json              ← MCP servers del proyecto (Stitch)
│   └── rules/
│       ├── atl-workflow.md        ← Fases del pipeline
│       ├── engram-protocol.md     ← Memoria compartida
│       ├── subagent-architecture.md
│       ├── context-monitoring.md
│       └── team-rules.md
└── docs/
    ├── CLAUDE.md                  ← Configuración del PM
    └── ux-ui/
        └── CLAUDE.md              ← Configuración del Designer
```

Si un archivo ya existe, el CLI lo salta — nunca sobreescribe tu trabajo.

---

### Paso 4 — Instalar los plugins en Claude Code

Abre Claude Code y ejecuta estos comandos para agregar los plugins del marketplace:

```bash
claude
```

Dentro de Claude Code:

```
/plugin marketplace add Intelliaa/marketplace-plugins-inteliside
```

Luego instala los plugins que necesites según tu rol:

```
/plugin install sdd-wizards@marketplace-plugins-inteliside
/plugin install ux-studio@marketplace-plugins-inteliside
/plugin install atl-inteliside@marketplace-plugins-inteliside
/plugin install n8n-studio@marketplace-plugins-inteliside
```

---

### Paso 5 — Verificar que todo funciona

```bash
inteliside doctor
```

Muestra el estado de cada dependencia. Si algo falta, te dice exactamente cómo arreglarlo.

```bash
inteliside verify
```

Verifica que los servidores MCP están conectados y las configuraciones son correctas.

---

### Paso 6 — Empezar a trabajar

Cada rol abre Claude Code desde su directorio:

```bash
# PM — levantamiento de requerimientos
cd docs
claude
/sdd-wizards:prd-wizard

# Designer — investigación y diseño UI/UX
cd docs/ux-ui
claude
/ux-studio:ux-orchestrator

# Dev — implementación de features
claude
/atl-inteliside:orchestrador feat-login

# Automation — workflows n8n
claude
/n8n-studio:automation-wizard
```

---

## Ejemplo completo: proyecto nuevo desde cero

```bash
# 1. Instalar CLI (una sola vez)
brew install Intelliaa/tap/inteliside

# 2. Configurar tu máquina (una sola vez)
inteliside install --preset fullstack

# 3. Crear proyecto
mkdir mi-app && cd mi-app && git init

# 4. Preparar el proyecto
inteliside init --preset fullstack

# 5. Verificar
inteliside doctor

# 6. Instalar plugins en Claude Code
claude
/plugin marketplace add Intelliaa/marketplace-plugins-inteliside
/plugin install sdd-wizards@marketplace-plugins-inteliside
/plugin install ux-studio@marketplace-plugins-inteliside
/plugin install atl-inteliside@marketplace-plugins-inteliside
/plugin install n8n-studio@marketplace-plugins-inteliside
```

---

## Referencia de comandos

### Instalación y configuración

| Comando | Descripción |
|---------|-------------|
| `inteliside` | Abre la interfaz interactiva guiada |
| `inteliside install --preset <rol>` | Instala dependencias globales según el rol |
| `inteliside install --plugin <nombre>` | Instala dependencias de plugins específicos |
| `inteliside init --preset <rol>` | Prepara un proyecto con archivos de configuración |
| `inteliside init --plugin <nombre>` | Prepara un proyecto para plugins específicos |

### Diagnóstico y mantenimiento

| Comando | Descripción |
|---------|-------------|
| `inteliside doctor` | Revisa dependencias y sugiere cómo arreglar problemas |
| `inteliside verify` | Verifica que MCP servers y configuraciones funcionan |
| `inteliside sync` | Busca actualizaciones del marketplace |
| `inteliside setup <plugin>` | Re-configura un plugin individual |

### Información

| Comando | Descripción |
|---------|-------------|
| `inteliside list` | Muestra plugins y roles disponibles |
| `inteliside version` | Muestra la versión instalada |
| `inteliside config show` | Muestra la configuración actual |

### Backup y recuperación

| Comando | Descripción |
|---------|-------------|
| `inteliside backup list` | Lista los backups disponibles |
| `inteliside backup restore <id>` | Restaura configuración desde un backup |
| `inteliside config reset <plugin>` | Elimina la configuración de un plugin |

### Banderas disponibles

Estas banderas funcionan con `install` e `init`:

| Bandera | Descripción |
|---------|-------------|
| `--preset <rol>` | Selecciona rol: `pm`, `designer`, `dev`, `fullstack`, `automation`, `legacy` |
| `--plugin <nombre>` | Selecciona plugins específicos separados por coma |
| `--dry-run` | Muestra qué haría sin ejecutar nada |
| `--yes` o `-y` | Acepta todos los valores por defecto sin preguntar |
| `--verbose` o `-v` | Muestra información detallada |
| `--project-dir <ruta>` | Especifica el directorio del proyecto (default: directorio actual) |

---

## Roles y plugins

| Rol | Preset | Plugins incluidos |
|-----|--------|-------------------|
| Product Manager | `pm` | SDD-Wizards |
| Diseñador UI/UX | `designer` | UX Studio, SDD-Wizards |
| Desarrollador | `dev` | ATL Inteliside, SDD-Wizards |
| Todo el equipo | `fullstack` | Los 6 plugins |
| Automatización | `automation` | n8n Studio |
| Legacy onboarding | `legacy` | SDD-Legacy, ATL Inteliside |
| Personalizado | `custom` | Selección manual en la TUI |

---

## Seguridad

- Se crea un **backup automático** antes de cada operación
- `--dry-run` permite ver los cambios antes de aplicarlos
- Nunca se sobreescriben archivos existentes del usuario
- Las API keys se guardan solo donde Claude Code las necesita
- Los backups se pueden restaurar con `inteliside backup restore <id>`

---

## Actualizar el CLI

```bash
brew upgrade inteliside
```

---

## Desarrollo

```bash
git clone https://github.com/Intelliaa/inteliside-cli.git
cd inteliside-cli
go build ./cmd/inteliside/
go test ./...
./inteliside
```

---

## Licencia

MIT — [Inteliside](https://github.com/Intelliaa)
