package cli

import (
	"bufio"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/backup"
	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/spf13/cobra"
)

// templateTarget maps a plugin to its CLAUDE.md destination relative to project root
type templateTarget struct {
	pluginID    string
	relPath     string // e.g. "docs/CLAUDE.md"
	description string
	getTemplate func(vars map[string]string) string
}

var allTemplates = []templateTarget{
	{
		pluginID:    "atl-inteliside",
		relPath:     "CLAUDE.md",
		description: "CLAUDE.md raíz (Dev — ATL Inteliside)",
		getTemplate: templateATL,
	},
	{
		pluginID:    "sdd-wizards",
		relPath:     "docs/CLAUDE.md",
		description: "docs/CLAUDE.md (PM — SDD-Wizards)",
		getTemplate: templatePM,
	},
	{
		pluginID:    "ux-studio",
		relPath:     "docs/ux-ui/CLAUDE.md",
		description: "docs/ux-ui/CLAUDE.md (Designer — UX Studio)",
		getTemplate: templateDesigner,
	},
	{
		pluginID:    "n8n-studio",
		relPath:     "CLAUDE.md",
		description: "CLAUDE.md raíz (Automation — n8n Studio)",
		getTemplate: templateN8n,
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa un proyecto con los CLAUDE.md y rules de cada plugin",
	Long: `Configura un proyecto nuevo copiando los CLAUDE.md de ejemplo a su
ubicación correcta y creando las rules necesarias.

Este comando es per-project — ejecutarlo en cada repositorio nuevo.
Para setup global (MCP servers, Engram), usar 'inteliside install'.

Ejemplos:
  cd mi-proyecto
  inteliside init --preset dev         # CLAUDE.md raíz + docs/ + rules
  inteliside init --preset designer    # docs/CLAUDE.md + docs/ux-ui/CLAUDE.md
  inteliside init --preset fullstack   # Todo
  inteliside init --plugin ux-studio   # Solo docs/ux-ui/CLAUDE.md
  inteliside init --dry-run            # Preview sin cambios`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().String("preset", "", "Preset por rol: pm, designer, dev, fullstack, automation, legacy")
	initCmd.Flags().String("plugin", "", "Plugins específicos separados por coma")
	initCmd.Flags().Bool("dry-run", false, "Mostrar plan sin ejecutar")
	initCmd.Flags().BoolP("yes", "y", false, "No preguntar valores, usar defaults")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	presetFlag, _ := cmd.Flags().GetString("preset")
	pluginFlag, _ := cmd.Flags().GetString("plugin")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	autoYes, _ := cmd.Flags().GetBool("yes")

	projectDir, _ := os.Getwd()

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════╗")
	fmt.Println("  ║        Inteliside CLI — Init             ║")
	fmt.Println("  ╚══════════════════════════════════════════╝")
	fmt.Printf("\n  Proyecto: %s\n\n", projectDir)

	// Resolve plugins
	pluginIDs, err := resolveInitPlugins(presetFlag, pluginFlag)
	if err != nil {
		return err
	}

	// Collect variables from user
	vars := make(map[string]string)
	if !autoYes {
		vars = collectProjectVars(pluginIDs)
	}

	// Determine which templates to apply
	var targets []templateTarget
	for _, t := range allTemplates {
		for _, pid := range pluginIDs {
			if t.pluginID == pid {
				targets = append(targets, t)
				break
			}
		}
	}

	// Handle ATL + n8n conflict (both want root CLAUDE.md)
	targets = deduplicateRootCLAUDE(targets)

	// Check for ATL rules
	needsRules := false
	for _, pid := range pluginIDs {
		if pid == "atl-inteliside" {
			needsRules = true
			break
		}
	}

	// Check for GitHub labels
	needsLabels := false
	for _, pid := range pluginIDs {
		if pid == "atl-inteliside" {
			needsLabels = true
			break
		}
	}

	// Show plan
	fmt.Println("  Archivos a crear:")
	fmt.Println("  " + strings.Repeat("─", 50))
	for _, t := range targets {
		dest := filepath.Join(projectDir, t.relPath)
		status := "crear"
		if _, err := os.Stat(dest); err == nil {
			status = "ya existe — SKIP"
		}
		fmt.Printf("    %s (%s)\n", t.relPath, status)
	}
	if needsRules {
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		if entries, err := os.ReadDir(rulesDir); err == nil && len(entries) > 0 {
			fmt.Printf("    .claude/rules/ (ya existe — SKIP)\n")
		} else {
			fmt.Printf("    .claude/rules/ (5 archivos)\n")
		}
	}
	if needsLabels {
		fmt.Printf("    GitHub labels (9 labels ATL)\n")
	}
	fmt.Println()

	if dryRun {
		fmt.Println("  [dry-run] No se realizaron cambios.")
		return nil
	}

	// Backup existing files
	var existingFiles []string
	for _, t := range targets {
		f := filepath.Join(projectDir, t.relPath)
		if _, err := os.Stat(f); err == nil {
			existingFiles = append(existingFiles, f)
		}
	}
	if len(existingFiles) > 0 {
		snap, _ := backup.Create(existingFiles)
		if snap != nil {
			fmt.Printf("  ✓ Backup: %s\n", snap.ID)
		}
	}

	// Write templates
	created := 0
	skipped := 0
	for _, t := range targets {
		dest := filepath.Join(projectDir, t.relPath)
		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("  → %s ya existe, skipping\n", t.relPath)
			skipped++
			continue
		}

		// Create directory
		dir := filepath.Dir(dest)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear %s: %w", dir, err)
		}

		content := t.getTemplate(vars)
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return fmt.Errorf("no se pudo escribir %s: %w", t.relPath, err)
		}
		fmt.Printf("  ✓ %s\n", t.relPath)
		created++
	}

	// Write rules
	if needsRules {
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		if entries, err := os.ReadDir(rulesDir); err != nil || len(entries) == 0 {
			if err := os.MkdirAll(rulesDir, 0755); err != nil {
				return err
			}
			rules := getRuleFilesForInit()
			for name, content := range rules {
				if err := os.WriteFile(filepath.Join(rulesDir, name), []byte(content), 0644); err != nil {
					return err
				}
				fmt.Printf("  ✓ .claude/rules/%s\n", name)
			}
			created += len(rules)
		} else {
			fmt.Println("  → .claude/rules/ ya existe, skipping")
			skipped++
		}
	}

	// Create labels
	if needsLabels {
		fmt.Println("  Creando GitHub labels...")
		dep := catalog.DependencyByID("github-labels")
		if dep != nil {
			ctx := &labelContext{projectDir: projectDir}
			_ = createLabelsFromInit(ctx)
		}
	}

	fmt.Printf("\n  ✓ Init completado: %d creados, %d skipped\n", created, skipped)
	fmt.Println("\n  Siguiente: completa los valores entre corchetes en cada CLAUDE.md")
	fmt.Println()
	return nil
}

type labelContext struct {
	projectDir string
}

func createLabelsFromInit(ctx *labelContext) error {
	labels := []struct{ name, color, desc string }{
		{"atl-task", "0075ca", "Task created by ATL pipeline"},
		{"area:backend", "e4e669", "Backend area"},
		{"area:frontend", "d93f0b", "Frontend area"},
		{"area:db", "0e8a16", "Database area"},
		{"area:test", "5319e7", "Testing area"},
		{"atl:pending", "ededed", "ATL status: pending"},
		{"atl:in-progress", "fbca04", "ATL status: in progress"},
		{"atl:done", "0e8a16", "ATL status: done"},
		{"atl-summary", "cfd3d7", "ATL summary issue"},
	}

	for _, l := range labels {
		out, err := runCmdOutput("gh", "label", "create", l.name,
			"--color", l.color,
			"--description", l.desc,
			"--force",
		)
		if err != nil {
			fmt.Printf("    ⚠ Label '%s': %s\n", l.name, out)
		} else {
			fmt.Printf("    ✓ Label '%s'\n", l.name)
		}
	}
	return nil
}

func runCmdOutput(name string, args ...string) (string, error) {
	out, err := osexec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func resolveInitPlugins(presetFlag, pluginFlag string) ([]string, error) {
	if presetFlag != "" && pluginFlag != "" {
		return nil, fmt.Errorf("usa --preset o --plugin, no ambos")
	}
	if presetFlag != "" {
		preset := catalog.PresetByID(presetFlag)
		if preset == nil {
			return nil, fmt.Errorf("preset desconocido: %s", presetFlag)
		}
		return preset.PluginIDs, nil
	}
	if pluginFlag != "" {
		ids := strings.Split(pluginFlag, ",")
		for _, id := range ids {
			if catalog.PluginByID(strings.TrimSpace(id)) == nil {
				return nil, fmt.Errorf("plugin desconocido: %s", id)
			}
		}
		return ids, nil
	}
	return nil, fmt.Errorf("especifica --preset o --plugin")
}

func collectProjectVars(pluginIDs []string) map[string]string {
	vars := make(map[string]string)
	reader := bufio.NewReader(os.Stdin)

	needsGH := false
	needsEngram := false
	needsFigma := false
	needsN8n := false

	for _, pid := range pluginIDs {
		switch pid {
		case "sdd-wizards", "atl-inteliside", "sdd-intake", "sdd-legacy":
			needsGH = true
		case "ux-studio":
			needsFigma = true
		case "n8n-studio":
			needsN8n = true
		}
		if pid == "atl-inteliside" || pid == "sdd-intake" || pid == "sdd-legacy" || pid == "n8n-studio" {
			needsEngram = true
		}
	}

	fmt.Println("  Configuración del proyecto (Enter para default):")
	fmt.Println("  " + strings.Repeat("─", 50))

	vars["project_name"] = promptVar(reader, "  Nombre del proyecto", "mi-proyecto")

	if needsGH {
		vars["github_owner"] = promptVar(reader, "  GitHub owner (org/usuario)", "")
		vars["github_repo"] = promptVar(reader, "  GitHub repo", "")
	}

	if needsEngram {
		def := vars["project_name"] + "-dev"
		vars["engram_project"] = promptVar(reader, "  Proyecto Engram", def)
	}

	if needsFigma {
		vars["figma_file"] = promptVar(reader, "  URL del archivo Figma", "")
	}

	if needsN8n {
		vars["n8n_dev_url"] = promptVar(reader, "  n8n dev URL", "https://n8n-dev1.codetrain.cloud")
		vars["n8n_prod_url"] = promptVar(reader, "  n8n prod URL", "")
	}

	fmt.Println()
	return vars
}

func promptVar(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func deduplicateRootCLAUDE(targets []templateTarget) []templateTarget {
	// If both ATL and n8n want root CLAUDE.md, prefer ATL (more complete)
	hasATL := false
	hasN8n := false
	for _, t := range targets {
		if t.pluginID == "atl-inteliside" && t.relPath == "CLAUDE.md" {
			hasATL = true
		}
		if t.pluginID == "n8n-studio" && t.relPath == "CLAUDE.md" {
			hasN8n = true
		}
	}

	if hasATL && hasN8n {
		// Remove n8n's root CLAUDE.md, ATL's is more complete
		// n8n config will be appended to ATL's template
		var filtered []templateTarget
		for _, t := range targets {
			if t.pluginID == "n8n-studio" && t.relPath == "CLAUDE.md" {
				continue
			}
			filtered = append(filtered, t)
		}
		return filtered
	}
	return targets
}

func getRuleFilesForInit() map[string]string {
	return map[string]string{
		"atl-workflow.md": `# ATL Workflow

Este proyecto usa ATL Inteliside para ejecución de features.
El flujo es: from-github → init → explore → propose → spec+design → tasks → write-tests → apply → verify → archive.

Reglas:
- Nunca saltarse fases del pipeline
- Cada fase produce artefactos que la siguiente consume
- El orquestador coordina todo el flujo
`,
		"engram-protocol.md": `# Engram Protocol

Memoria persistente compartida entre agentes.

Reglas:
- Guardar decisiones de arquitectura en Engram inmediatamente
- Buscar contexto en Engram antes de proponer cambios
- Dos proyectos: equipo (permanente) y pipeline (efímero)
`,
		"subagent-architecture.md": `# Subagent Architecture

Cada subagente opera en aislamiento (context: fork).

Reglas:
- Analyst y Architect son read-only (no modifican código)
- Test Writer genera tests SIN ver implementación
- Builder es el único que modifica código fuente
- Verifier es read-only y cierra issues/milestones
`,
		"context-monitoring.md": `# Context Monitoring

Gestión del tamaño de contexto durante el pipeline.

Reglas:
- Monitorear uso de contexto entre fases
- Si el contexto supera 70%, resumir y continuar en nuevo fork
- Artefactos grandes van a archivos, no inline
`,
		"team-rules.md": `# Team Rules

Reglas de colaboración entre agentes del equipo ATL.

Reglas:
- El orquestador es la única fuente de verdad del estado
- Los subagentes no se comunican entre sí directamente
- La comunicación es via artefactos (archivos) y Engram
`,
	}
}

// --- Templates ---

func templateATL(vars map[string]string) string {
	name := getVar(vars, "project_name", "[Nombre del Proyecto]")
	engram := getVar(vars, "engram_project", "[nombre-proyecto]-dev")
	owner := getVar(vars, "github_owner", "[tu-org-o-usuario]")
	repo := getVar(vars, "github_repo", "[nombre-repo]")

	return fmt.Sprintf(`# CLAUDE.md — %s

> Configuración del proyecto para Claude Code y ATL Inteliside.
> Completa los valores entre corchetes.

---

## ATL Inteliside

Configuración requerida para el plugin ATL Inteliside. Todos los devs del equipo deben
tener este archivo con los mismos valores para compartir la memoria de Engram.

`+"`"+`yaml
engram_project: "%s"
github_owner: "%s"
github_repo: "%s"
`+"`"+`

> **Nota**: ATL Inteliside deriva automáticamente un segundo proyecto Engram para el pipeline:
> engram_pipeline = "{engram_project}/atl"
>
> - **engram_project** → conocimiento permanente del equipo (decisiones, patrones, bugs)
> - **engram_pipeline** → estado efímero del pipeline de implementación

---

## Proyecto

- **Nombre**: %s
- **Descripción**: [Qué hace el proyecto]
- **Stack**: [ej: Next.js 15 + TypeScript + Drizzle + PostgreSQL]
- **Entorno de desarrollo**: [ej: Node 20+, pnpm, Docker para DB]

---

## Comandos frecuentes

`+"`"+`bash
# Desarrollo
pnpm dev

# Tests
pnpm test

# Build
pnpm build
`+"`"+`

---

## Convenciones del proyecto

### Naming
- Archivos: kebab-case.ts
- Componentes: PascalCase.tsx
- Funciones: camelCase

### Testing
- Archivos de test: *.test.ts junto al archivo testeado

---

## Rules de ATL Inteliside

- @.claude/rules/engram-protocol.md
- @.claude/rules/subagent-architecture.md
- @.claude/rules/atl-workflow.md
- @.claude/rules/context-monitoring.md
- @.claude/rules/team-rules.md

---

*Generado con inteliside init — Marketplace Inteliside*
`, name, engram, owner, repo, name)
}

func templatePM(vars map[string]string) string {
	owner := getVar(vars, "github_owner", "{org-o-usuario}")
	repo := getVar(vars, "github_repo", "{nombre-repo}")
	name := getVar(vars, "project_name", "{nombre}")

	return fmt.Sprintf(`# CLAUDE.md — Documentación y Requerimientos

> Este directorio es el espacio de trabajo del PM.
> Ejecutar Claude Code desde aquí para usar SDD-Wizards.

## Rol

Este CLAUDE.md aplica al PM o equipo de proyecto. Desde este directorio se:
- Crean PRDs con /sdd-wizards:prd-wizard
- Detallan features con /sdd-wizards:feature-spec-wizard

## Config

`+"`"+`yaml
github_owner: "%s"
github_repo: "%s"
`+"`"+`

## Reglas de sesión

- Una conversación por skill — no mezclar PRD y Feature Spec
- Una conversación por feature — cada feature se detalla por separado
- Descargar el output antes de cerrar — el archivo .md es la memoria entre sesiones

## Contexto del producto

- **Producto**: %s
- **Empresa / Cliente**: [nombre]
- **Problema que resuelve**: [descripción breve]
- **Usuarios objetivo**: [quiénes son]
- **Restricciones**: [presupuesto, plazo, integraciones obligatorias]

## Estructura de este directorio

`+"`"+`
docs/
├── CLAUDE.md              ← este archivo (contexto del PM)
├── PRD-{producto}.md      ← PRD generado por prd-wizard
├── feat-{nombre}.md       ← Feature Specs generados
└── ux-ui/                 ← Espacio del Designer
    └── CLAUDE.md
`+"`"+`

---

*Generado con inteliside init — Marketplace Inteliside*
`, owner, repo, name)
}

func templateDesigner(vars map[string]string) string {
	figma := getVar(vars, "figma_file", "{URL del archivo Figma}")
	engram := getVar(vars, "engram_project", "{nombre-proyecto}-dev")

	return fmt.Sprintf(`# CLAUDE.md — Diseño UI/UX

> Este directorio es el espacio de trabajo del Designer.
> Ejecutar Claude Code desde aquí para usar UX Studio.

## Rol

Este CLAUDE.md aplica al Designer. Desde este directorio se:
- Lanza el pipeline completo con /ux-studio:ux-orchestrator
- Ejecuta solo la entrevista de diseño con /ux-studio:ux-discovery

## Config

`+"`"+`yaml
figma_file: "%s"
engram_project: "%s"
`+"`"+`

## Pipeline (v1.1.0 — Stitch)

`+"`"+`
research → discovery → prompt-gen → stitch-gen → figma-import → refine → review
`+"`"+`

## MCP Servers requeridos

Verificar que están configurados:
- **Google Stitch MCP** — generación de diseños UI desde prompts
- **Figma Console MCP** — importación y refinamiento en Figma
- **Chrome DevTools MCP** — research de competencia

## Reglas de diseño

- WCAG AA mínimo (contraste 4.5:1 texto normal, 3:1 texto grande)
- Escala de spacing basada en 4px
- Priorizar productos reales de competencia sobre sitios de diseño
- Componentes con variantes (estados, tamaños)

## Engram

Prefijo ux-studio/ para todos los artefactos:
- ux-studio/research-report
- ux-studio/ux-brief
- ux-studio/stitch-prompts
- ux-studio/stitch-output
- ux-studio/design-system
- ux-studio/screens
- ux-studio/design-review

## Estructura de este directorio

`+"`"+`
docs/ux-ui/
├── CLAUDE.md                  ← este archivo (contexto del Designer)
├── design-spec-{feature}.md   ← Design Specs generados
├── research-{feature}.md      ← Research Reports
└── screenshots/               ← Capturas de referencia y exports
`+"`"+`

---

*Generado con inteliside init — Marketplace Inteliside*
`, figma, engram)
}

func templateN8n(vars map[string]string) string {
	name := getVar(vars, "project_name", "[Nombre del Proyecto]")
	engram := getVar(vars, "engram_project", "[nombre-proyecto]-n8n")
	devURL := getVar(vars, "n8n_dev_url", "[url-de-tu-n8n-dev]")
	prodURL := getVar(vars, "n8n_prod_url", "[url-de-tu-n8n-prod]")

	return fmt.Sprintf(`# CLAUDE.md — %s

> Configuración del proyecto para Claude Code y n8n Studio.
> Completa los valores entre corchetes.

---

## n8n Studio

Configuración requerida para el plugin n8n Studio.

`+"`"+`yaml
engram_project: "%s"
n8n_dev_url: "%s"
n8n_prod_url: "%s"
`+"`"+`

> **Nota**: n8n Studio usa Engram para persistir estado del pipeline entre agentes.

---

## Flujo de trabajo

`+"`"+`
1. /n8n-studio:automation-wizard [descripción]  → Genera Automation Spec
2. /n8n-studio:n8n-orchestrator [ruta-spec.md]  → Construye, prueba y despliega
3. /n8n-studio:n8n-deploy [workflow-id]         → Promueve a producción
`+"`"+`

---

## Directorio de Automation Specs

Los specs generados se guardan en:

`+"`"+`
automation-specs/
`+"`"+`

## Convenciones

- Los workflows generados se nombran con prefijo [N8N-STUDIO]
- Los specs se nombran: automation-spec-[nombre-descriptivo].md
- Las credenciales se configuran manualmente en n8n UI antes de tests

---

*Generado con inteliside init — Marketplace Inteliside*
`, name, engram, devURL, prodURL)
}

func getVar(vars map[string]string, key, fallback string) string {
	if v, ok := vars[key]; ok && v != "" {
		return v
	}
	return fallback
}
