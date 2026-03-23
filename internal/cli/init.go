package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/backup"
	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/Intelliaa/inteliside-cli/internal/embedded"
	"github.com/Intelliaa/inteliside-cli/internal/legacy"
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
		relPath:     ".claude/CLAUDE.md",
		description: ".claude/CLAUDE.md (Dev — ATL Inteliside)",
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
		relPath:     ".claude/CLAUDE.md",
		description: ".claude/CLAUDE.md (Automation — n8n Studio)",
		getTemplate: templateN8n,
	},
	{
		pluginID:    "sdd-legacy",
		relPath:     "docs/legacy/CLAUDE.md",
		description: "docs/legacy/CLAUDE.md (Contexto Legacy — SDD-Legacy)",
		getTemplate: templateLegacy,
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa un proyecto con los CLAUDE.md y rules de cada plugin",
	Long: `Configura un proyecto nuevo copiando los CLAUDE.md de ejemplo a su
ubicacion correcta y creando las rules necesarias.

Si detecta artefactos legacy (CLAUDE.md sin marcadores ATL), los archiva
automaticamente en docs/legacy/ y genera la configuracion nueva.

Este comando es per-project — ejecutarlo en cada repositorio nuevo.
Para setup global (MCP servers, Engram), usar 'inteliside install'.

Ejemplos:
  cd mi-proyecto
  inteliside init --preset dev         # CLAUDE.md raiz + docs/ + rules
  inteliside init --preset legacy      # Archiva legacy + genera ATL
  inteliside init --preset designer    # docs/CLAUDE.md + docs/ux-ui/CLAUDE.md
  inteliside init --preset fullstack   # Todo
  inteliside init --plugin ux-studio   # Solo docs/ux-ui/CLAUDE.md
  inteliside init --dry-run            # Preview sin cambios`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().String("preset", "", "Preset por rol: pm, designer, dev, fullstack, automation, legacy")
	initCmd.Flags().String("plugin", "", "Plugins especificos separados por coma")
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

	// --- Legacy detection ---
	legacyArtifacts := legacy.Detect(projectDir)
	if legacyArtifacts.IsLegacy {
		fmt.Println("  ⚠ Proyecto legacy detectado:")
		fmt.Printf("    CLAUDE.md existente (sin marcadores ATL)\n")
		if legacyArtifacts.DotClaudeMD != "" {
			fmt.Printf("    .claude/CLAUDE.md existente (instrucciones de proyecto)\n")
		}
		if len(legacyArtifacts.RuleFiles) > 0 {
			fmt.Printf("    .claude/rules/ existente (%d archivos)\n", len(legacyArtifacts.RuleFiles))
		}
		if len(legacyArtifacts.DocsFiles) > 0 {
			fmt.Printf("    docs/ existente (%d archivos)\n", len(legacyArtifacts.DocsFiles))
		}
		fmt.Println()
		fmt.Println("  Los artefactos legacy se archivaran en docs/legacy/")
		fmt.Println("  y se generara la configuracion nueva para ATL.")
		fmt.Println()

		if dryRun {
			fmt.Println("  [dry-run] Se archivarian los artefactos legacy en docs/legacy/")
			fmt.Println()
		} else {
			// Backup before archiving
			var legacyFiles []string
			if legacyArtifacts.ClaudeMD != "" {
				legacyFiles = append(legacyFiles, legacyArtifacts.ClaudeMD)
			}
			if legacyArtifacts.DotClaudeMD != "" {
				legacyFiles = append(legacyFiles, legacyArtifacts.DotClaudeMD)
			}
			legacyFiles = append(legacyFiles, legacyArtifacts.RuleFiles...)
			if len(legacyFiles) > 0 {
				snap, _ := backup.Create(legacyFiles)
				if snap != nil {
					fmt.Printf("  ✓ Backup de legacy: %s\n", snap.ID)
				}
			}

			// Archive
			archived, err := legacy.Archive(projectDir, legacyArtifacts)
			if err != nil {
				return fmt.Errorf("error archivando legacy: %w", err)
			}
			for _, a := range archived {
				rel, _ := filepath.Rel(projectDir, a)
				fmt.Printf("  ✓ Archivado: %s\n", rel)
			}
			fmt.Println()
		}
	}

	// Collect variables — auto-detect first, then prompt if interactive
	vars := detectProjectVars(projectDir)

	// Merge extracted legacy vars (these take precedence over auto-detected)
	if legacyArtifacts.IsLegacy {
		for k, v := range legacyArtifacts.ExtractedVars {
			if v != "" {
				vars[k] = v
			}
		}
	}

	// Interactive prompt if not --yes
	if !autoYes {
		vars = collectProjectVars(pluginIDs, vars)
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

	// Check for Stitch MCP (per-project)
	needsStitch := false
	for _, pid := range pluginIDs {
		if pid == "ux-studio" {
			needsStitch = true
			break
		}
	}

	// Show plan
	fmt.Println("  Archivos a crear/actualizar:")
	fmt.Println("  " + strings.Repeat("─", 50))
	for _, t := range targets {
		dest := filepath.Join(projectDir, t.relPath)
		if _, err := os.Stat(dest); err == nil {
			if t.relPath == ".claude/CLAUDE.md" {
				fmt.Printf("    %s (existe — MERGE seccion ATL)\n", t.relPath)
			} else {
				fmt.Printf("    %s (ya existe — SKIP)\n", t.relPath)
			}
		} else {
			fmt.Printf("    %s (crear)\n", t.relPath)
		}
	}
	if needsRules {
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		existing := listExistingRules(rulesDir)
		missing := countMissingRules(existing)
		if missing == 0 && len(existing) > 0 {
			fmt.Printf("    .claude/rules/ (5 rules ATL presentes)\n")
		} else if len(existing) > 0 {
			fmt.Printf("    .claude/rules/ (agregar %d rules ATL faltantes)\n", missing)
		} else {
			fmt.Printf("    .claude/rules/ (5 archivos)\n")
		}
	}
	if needsStitch {
		fmt.Printf("    .claude/settings.json (Stitch MCP per-project)\n")
	}
	if needsLabels {
		fmt.Printf("    GitHub labels (9 labels ATL)\n")
	}
	if legacyArtifacts.IsLegacy {
		fmt.Printf("    CLAUDE.md raiz incluira seccion de referencia legacy\n")
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
	merged := 0
	for _, t := range targets {
		dest := filepath.Join(projectDir, t.relPath)

		// Create directory
		dir := filepath.Dir(dest)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear %s: %w", dir, err)
		}

		existingData, existsErr := os.ReadFile(dest)
		fileExists := existsErr == nil

		if fileExists && t.relPath == ".claude/CLAUDE.md" {
			// MERGE: append ATL section to existing .claude/CLAUDE.md
			existingContent := string(existingData)
			if strings.Contains(existingContent, "<!-- inteliside:atl-config -->") {
				fmt.Printf("  → %s ya tiene seccion ATL, skipping\n", t.relPath)
				skipped++
			} else {
				atlSection := generateATLSection(vars)
				newContent := existingContent + "\n" + atlSection
				if legacyArtifacts.IsLegacy {
					newContent += legacyReferenceSection()
				}
				if err := os.WriteFile(dest, []byte(newContent), 0644); err != nil {
					return fmt.Errorf("no se pudo actualizar %s: %w", t.relPath, err)
				}
				fmt.Printf("  ✓ %s (seccion ATL agregada)\n", t.relPath)
				merged++
			}
		} else if fileExists {
			// Other templates: skip if already exists
			fmt.Printf("  → %s ya existe, skipping\n", t.relPath)
			skipped++
		} else {
			// File doesn't exist: create from template
			content := t.getTemplate(vars)
			if t.relPath == ".claude/CLAUDE.md" && legacyArtifacts.IsLegacy {
				content += legacyReferenceSection()
			}
			if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
				return fmt.Errorf("no se pudo escribir %s: %w", t.relPath, err)
			}
			fmt.Printf("  ✓ %s\n", t.relPath)
			created++
		}
	}

	// Write rules — add missing ATL rules even if directory has other files
	if needsRules {
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return err
		}
		rules, err := embedded.RuleFiles()
		if err != nil {
			return fmt.Errorf("error leyendo rules embebidas: %w", err)
		}
		for name, content := range rules {
			dest := filepath.Join(rulesDir, name)
			if _, err := os.Stat(dest); err == nil {
				// Rule already exists — skip
				continue
			}
			if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Printf("  ✓ .claude/rules/%s\n", name)
			created++
		}
	}

	// Configure Stitch MCP at project level
	if needsStitch {
		settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
		gcpProject := vars["gcp_project"]
		stitchKey := vars["stitch_api_key"]

		if err := writeProjectStitchMCP(settingsPath, stitchKey, gcpProject); err != nil {
			fmt.Printf("  ⚠ Stitch MCP: %v\n", err)
		} else {
			fmt.Println("  ✓ .claude/settings.json (Stitch MCP configurado)")
			created++
		}
	}

	// Create labels (with validation)
	if needsLabels {
		if ok, reason := canCreateLabels(projectDir); ok {
			fmt.Println("  Creando GitHub labels...")
			createLabelsFromInit(projectDir)
		} else {
			fmt.Printf("  ⚠ Labels: saltando — %s\n", reason)
			fmt.Println("    Ejecuta 'inteliside setup atl-inteliside' despues de configurar el remote")
		}
	}

	// Summary
	fmt.Printf("\n  ✓ Init completado: %d creados, %d mergeados, %d skipped\n", created, merged, skipped)

	// Check for pending TODOs
	claudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if data, err := os.ReadFile(claudeMD); err == nil {
		todoCount := embedded.CountTODOs(string(data))
		if todoCount > 0 {
			fmt.Printf("\n  ℹ %d valores pendientes de completar en CLAUDE.md\n", todoCount)
			fmt.Println("    Busca '<!-- TODO:' y reemplaza con los valores reales")
		}
	}

	fmt.Println()
	return nil
}

// --- ATL section for merging into existing CLAUDE.md ---

func generateATLSection(vars map[string]string) string {
	engram := getVar(vars, "engram_project", "<!-- TODO: engram_project -->")
	owner := getVar(vars, "github_owner", "<!-- TODO: github_owner -->")
	repo := getVar(vars, "github_repo", "<!-- TODO: github_repo -->")

	return fmt.Sprintf(`
---

<!-- inteliside:atl-config -->
## ATL Inteliside

Configuracion requerida para el plugin ATL Inteliside. Todos los devs del equipo deben
tener este archivo con los mismos valores para compartir la memoria de Engram.

`+"`"+`yaml
engram_project: "%s"
github_owner: "%s"
github_repo: "%s"
`+"`"+`

> **Nota**: ATL Inteliside deriva automaticamente un segundo proyecto Engram para el pipeline:
> `+"`"+`engram_pipeline = "{engram_project}/atl"`+"`"+`
>
> - **engram_project** → conocimiento permanente del equipo (decisiones, patrones, bugs)
> - **engram_pipeline** → estado efimero del pipeline de implementacion
<!-- /inteliside:atl-config -->

---

## Rules de ATL Inteliside

- @.claude/rules/engram-protocol.md
- @.claude/rules/subagent-architecture.md
- @.claude/rules/atl-workflow.md
- @.claude/rules/context-monitoring.md
- @.claude/rules/team-rules.md
`, engram, owner, repo)
}

// --- Rule helpers ---

func listExistingRules(rulesDir string) map[string]bool {
	existing := make(map[string]bool)
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return existing
	}
	for _, e := range entries {
		if !e.IsDir() {
			existing[e.Name()] = true
		}
	}
	return existing
}

func countMissingRules(existing map[string]bool) int {
	atlRules := embedded.RuleFileNames()
	missing := 0
	for _, name := range atlRules {
		if !existing[name] {
			missing++
		}
	}
	return missing
}

// --- Legacy reference section ---

func legacyReferenceSection() string {
	return `

---

## Documentacion Legacy

Este proyecto fue migrado desde una configuracion legacy. La documentacion original
se encuentra en ` + "`docs/legacy/`" + ` para referencia:

- ` + "`docs/legacy/CLAUDE.md`" + ` — Configuracion original del proyecto
- ` + "`docs/legacy/.claude/rules/`" + ` — Rules originales (si existian)
- ` + "`docs/legacy/docs/`" + ` — Documentos originales del directorio docs/

> **Antes de implementar una feature nueva**: consultar docs/legacy/ para
> reglas de negocio existentes y convenciones del codebase original.
`
}

// --- Label validation (Solution 7) ---

func canCreateLabels(projectDir string) (bool, string) {
	// Check 1: Is this a git repo?
	if _, err := os.Stat(filepath.Join(projectDir, ".git")); os.IsNotExist(err) {
		return false, "no es un repositorio git"
	}

	// Check 2: Has remote configured?
	out, err := osexec.Command("git", "-C", projectDir, "remote", "get-url", "origin").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return false, "no hay remote 'origin' configurado"
	}

	// Check 3: gh authenticated?
	_, err = osexec.Command("gh", "auth", "status").CombinedOutput()
	if err != nil {
		return false, "gh CLI no esta autenticado"
	}

	return true, ""
}

func createLabelsFromInit(projectDir string) {
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
		out, err := osexec.Command("gh", "label", "create", l.name,
			"--color", l.color,
			"--description", l.desc,
			"--force",
		).CombinedOutput()
		if err != nil {
			fmt.Printf("    ⚠ Label '%s': %s\n", l.name, strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("    ✓ Label '%s'\n", l.name)
		}
	}
}

// --- Variable detection and collection (Solution 4) ---

func detectProjectVars(projectDir string) map[string]string {
	vars := make(map[string]string)

	// project_name from directory name
	vars["project_name"] = filepath.Base(projectDir)

	// github_owner and github_repo from git remote
	if out, err := osexec.Command("git", "-C", projectDir, "remote", "get-url", "origin").CombinedOutput(); err == nil {
		owner, repo := parseGitRemote(strings.TrimSpace(string(out)))
		if owner != "" {
			vars["github_owner"] = owner
		}
		if repo != "" {
			vars["github_repo"] = repo
		}
	}

	// engram_project derived from project_name
	vars["engram_project"] = vars["project_name"] + "-dev"

	return vars
}

func parseGitRemote(remote string) (owner, repo string) {
	// Handle SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(remote, "git@") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) == 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			segments := strings.SplitN(path, "/", 2)
			if len(segments) == 2 {
				return segments[0], segments[1]
			}
		}
		return "", ""
	}

	// Handle HTTPS: https://github.com/owner/repo.git
	remote = strings.TrimSuffix(remote, ".git")
	parts := strings.Split(remote, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return "", ""
}

func collectProjectVars(pluginIDs []string, vars map[string]string) map[string]string {
	reader := bufio.NewReader(os.Stdin)

	needsGH := false
	needsEngram := false
	needsFigma := false
	needsStitch := false
	needsN8n := false

	for _, pid := range pluginIDs {
		switch pid {
		case "sdd-wizards", "atl-inteliside", "sdd-intake", "sdd-legacy":
			needsGH = true
		case "ux-studio":
			needsFigma = true
			needsStitch = true
		case "n8n-studio":
			needsN8n = true
		}
		if pid == "atl-inteliside" || pid == "sdd-intake" || pid == "sdd-legacy" || pid == "n8n-studio" {
			needsEngram = true
		}
	}

	fmt.Println("  Configuracion del proyecto (Enter para usar el valor detectado):")
	fmt.Println("  " + strings.Repeat("─", 50))

	vars["project_name"] = promptVar(reader, "  Nombre del proyecto", vars["project_name"])

	if needsGH {
		vars["github_owner"] = promptVar(reader, "  GitHub owner (org/usuario)", vars["github_owner"])
		vars["github_repo"] = promptVar(reader, "  GitHub repo", vars["github_repo"])
	}

	if needsEngram {
		def := vars["engram_project"]
		if def == "" {
			def = vars["project_name"] + "-dev"
		}
		vars["engram_project"] = promptVar(reader, "  Proyecto Engram", def)
	}

	if needsFigma {
		vars["figma_file"] = promptVar(reader, "  URL del archivo Figma", vars["figma_file"])
	}

	if needsStitch {
		vars["gcp_project"] = promptVar(reader, "  Google Cloud Project ID (para Stitch)", vars["gcp_project"])
		vars["stitch_api_key"] = promptVar(reader, "  Stitch API Key (Enter si ya esta en global)", vars["stitch_api_key"])
	}

	if needsN8n {
		def := vars["n8n_dev_url"]
		if def == "" {
			def = "https://n8n-dev1.codetrain.cloud"
		}
		vars["n8n_dev_url"] = promptVar(reader, "  n8n dev URL", def)
		vars["n8n_prod_url"] = promptVar(reader, "  n8n prod URL", vars["n8n_prod_url"])
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

// --- Plugin resolution ---

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

func deduplicateRootCLAUDE(targets []templateTarget) []templateTarget {
	// If both ATL and n8n want root CLAUDE.md, prefer ATL (more complete)
	hasATL := false
	hasN8n := false
	for _, t := range targets {
		if t.pluginID == "atl-inteliside" && t.relPath == ".claude/CLAUDE.md" {
			hasATL = true
		}
		if t.pluginID == "n8n-studio" && t.relPath == ".claude/CLAUDE.md" {
			hasN8n = true
		}
	}

	if hasATL && hasN8n {
		var filtered []templateTarget
		for _, t := range targets {
			if t.pluginID == "n8n-studio" && t.relPath == ".claude/CLAUDE.md" {
				continue
			}
			filtered = append(filtered, t)
		}
		return filtered
	}
	return targets
}

// --- Templates (Solution 5: ATL uses embedded source of truth) ---

func templateATL(vars map[string]string) string {
	tmpl := embedded.ATLTemplate()
	if tmpl == "" {
		// Fallback if embed fails
		return templateATLFallback(vars)
	}
	return embedded.RenderTemplate(tmpl, vars)
}

func templateATLFallback(vars map[string]string) string {
	name := getVar(vars, "project_name", "Mi Proyecto")
	engram := getVar(vars, "engram_project", "mi-proyecto-dev")
	owner := getVar(vars, "github_owner", "mi-org")
	repo := getVar(vars, "github_repo", "mi-repo")

	return fmt.Sprintf(`# CLAUDE.md — %s

<!-- inteliside:atl-config -->
## ATL Inteliside

`+"`"+`yaml
engram_project: "%s"
github_owner: "%s"
github_repo: "%s"
`+"`"+`
<!-- /inteliside:atl-config -->

---

## Proyecto

- **Nombre**: %s

---

*Generado con inteliside init — Marketplace Inteliside*
`, name, engram, owner, repo, name)
}

func templateLegacy(vars map[string]string) string {
	tmpl := embedded.LegacyTemplate()
	if tmpl == "" {
		name := getVar(vars, "project_name", "Mi Proyecto")
		return fmt.Sprintf("# Contexto Legacy — %s\n\nArtefactos archivados del onboarding legacy.\n", name)
	}
	return embedded.RenderTemplate(tmpl, vars)
}

func templatePM(vars map[string]string) string {
	owner := getVar(vars, "github_owner", "")
	repo := getVar(vars, "github_repo", "")
	name := getVar(vars, "project_name", "")

	ownerDisplay := owner
	if ownerDisplay == "" {
		ownerDisplay = "<!-- TODO: github_owner -->"
	}
	repoDisplay := repo
	if repoDisplay == "" {
		repoDisplay = "<!-- TODO: github_repo -->"
	}
	nameDisplay := name
	if nameDisplay == "" {
		nameDisplay = "<!-- TODO: project_name -->"
	}

	return fmt.Sprintf(`# CLAUDE.md — Documentacion y Requerimientos

> Este directorio es el espacio de trabajo del PM.
> Ejecutar Claude Code desde aqui para usar SDD-Wizards.

## Rol

Este CLAUDE.md aplica al PM o equipo de proyecto. Desde este directorio se:
- Crean PRDs con /sdd-wizards:prd-wizard
- Detallan features con /sdd-wizards:feature-spec-wizard

## Config

`+"`"+`yaml
github_owner: "%s"
github_repo: "%s"
`+"`"+`

## Reglas de sesion

- Una conversacion por skill — no mezclar PRD y Feature Spec
- Una conversacion por feature — cada feature se detalla por separado
- Descargar el output antes de cerrar — el archivo .md es la memoria entre sesiones

## Contexto del producto

- **Producto**: %s
- **Empresa / Cliente**: <!-- TODO: nombre_empresa -->
- **Problema que resuelve**: <!-- TODO: problema -->
- **Usuarios objetivo**: <!-- TODO: usuarios -->
- **Restricciones**: <!-- TODO: restricciones -->

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
`, ownerDisplay, repoDisplay, nameDisplay)
}

func templateDesigner(vars map[string]string) string {
	figma := getVar(vars, "figma_file", "")
	engram := getVar(vars, "engram_project", "")

	figmaDisplay := figma
	if figmaDisplay == "" {
		figmaDisplay = "<!-- TODO: figma_file -->"
	}
	engramDisplay := engram
	if engramDisplay == "" {
		engramDisplay = "<!-- TODO: engram_project -->"
	}

	return fmt.Sprintf(`# CLAUDE.md — Diseno UI/UX

> Este directorio es el espacio de trabajo del Designer.
> Ejecutar Claude Code desde aqui para usar UX Studio.

## Rol

Este CLAUDE.md aplica al Designer. Desde este directorio se:
- Lanza el pipeline completo con /ux-studio:ux-orchestrator
- Ejecuta solo la entrevista de diseno con /ux-studio:ux-discovery

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

Verificar que estan configurados:
- **Google Stitch MCP** — generacion de disenos UI desde prompts
- **Figma Console MCP** — importacion y refinamiento en Figma
- **Chrome DevTools MCP** — research de competencia

## Reglas de diseno

- WCAG AA minimo (contraste 4.5:1 texto normal, 3:1 texto grande)
- Escala de spacing basada en 4px
- Priorizar productos reales de competencia sobre sitios de diseno
- Componentes con variantes (estados, tamanos)

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
`, figmaDisplay, engramDisplay)
}

func templateN8n(vars map[string]string) string {
	name := getVar(vars, "project_name", "")
	engram := getVar(vars, "engram_project", "")
	devURL := getVar(vars, "n8n_dev_url", "")
	prodURL := getVar(vars, "n8n_prod_url", "")

	nameDisplay := name
	if nameDisplay == "" {
		nameDisplay = "<!-- TODO: project_name -->"
	}
	engramDisplay := engram
	if engramDisplay == "" {
		engramDisplay = "<!-- TODO: engram_project -->"
	}
	devDisplay := devURL
	if devDisplay == "" {
		devDisplay = "<!-- TODO: n8n_dev_url -->"
	}
	prodDisplay := prodURL
	if prodDisplay == "" {
		prodDisplay = "<!-- TODO: n8n_prod_url -->"
	}

	return fmt.Sprintf(`# CLAUDE.md — %s

> Configuracion del proyecto para Claude Code y n8n Studio.

---

## n8n Studio

Configuracion requerida para el plugin n8n Studio.

`+"`"+`yaml
engram_project: "%s"
n8n_dev_url: "%s"
n8n_prod_url: "%s"
`+"`"+`

> **Nota**: n8n Studio usa Engram para persistir estado del pipeline entre agentes.

---

## Flujo de trabajo

`+"`"+`
1. /n8n-studio:automation-wizard [descripcion]  → Genera Automation Spec
2. /n8n-studio:n8n-orchestrator [ruta-spec.md]  → Construye, prueba y despliega
3. /n8n-studio:n8n-deploy [workflow-id]         → Promueve a produccion
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
`, nameDisplay, engramDisplay, devDisplay, prodDisplay)
}

func writeProjectStitchMCP(settingsPath, stitchKey, gcpProject string) error {
	if gcpProject == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT es requerido para Stitch MCP")
	}

	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Read existing project settings
	existing := make(map[string]any)
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	// Build stitch MCP config
	env := map[string]any{
		"GOOGLE_CLOUD_PROJECT": gcpProject,
	}
	if stitchKey != "" {
		env["STITCH_API_KEY"] = stitchKey
	}

	mcpServers, ok := existing["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
	}
	mcpServers["stitch"] = map[string]any{
		"command": "npx",
		"args":    []any{"-y", "stitch-mcp"},
		"env":     env,
	}
	existing["mcpServers"] = mcpServers

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

func getVar(vars map[string]string, key, fallback string) string {
	if v, ok := vars[key]; ok && v != "" {
		return v
	}
	return fallback
}
