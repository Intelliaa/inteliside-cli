package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/model"
	"github.com/Intelliaa/inteliside-cli/internal/system"
)

const (
	defaultN8nMCPURL   = "https://n8n-mcp.codetrain.cloud/mcp"
	defaultN8nMCPToken = "vt+T+8qEHrtahKldrjJH462vFfr6ExD7ssV0LbAtjOE="
)

// AllDependencies returns every dependency the plugins may need.
func AllDependencies() []model.Dependency {
	return []model.Dependency{
		ghCLI(),
		ghAuth(),
		ghRepoScope(),
		nodeRuntime(),
		engramBinary(),
		engramPlugin(),
		figmaMCP(),
		stitchMCP(),
		n8nMCP(),
		githubLabels(),
		claudeRules(),
		claudeMDTemplate(),
	}
}

// DependencyByID returns a dependency by ID, or nil.
func DependencyByID(id string) *model.Dependency {
	for _, d := range AllDependencies() {
		if d.ID == id {
			return &d
		}
	}
	return nil
}

// --- gh CLI ---

func ghCLI() model.Dependency {
	return model.Dependency{
		ID:          "gh-cli",
		Name:        "GitHub CLI",
		Description: "gh CLI para interactuar con GitHub",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			if !system.HasCommand("gh") {
				return false, "gh no encontrado en PATH", nil
			}
			out, err := system.RunCommand("gh", "--version")
			if err != nil {
				return false, "no se pudo verificar version de gh", err
			}
			return true, out, nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Instalaría gh CLI")
				return nil
			}
			switch system.DetectPlatform() {
			case system.PlatformMacOS:
				return runInstall("brew", "install", "gh")
			case system.PlatformLinux:
				return fmt.Errorf("instala gh manualmente: https://cli.github.com/")
			default:
				return fmt.Errorf("instala gh manualmente: https://cli.github.com/")
			}
		},
	}
}

func ghAuth() model.Dependency {
	return model.Dependency{
		ID:          "gh-auth",
		Name:        "GitHub Auth",
		Description: "gh autenticado con una cuenta de GitHub",
		Requires:    []string{"gh-cli"},
		CheckFn: func() (bool, string, error) {
			out, err := system.RunCommand("gh", "auth", "status")
			if err != nil {
				return false, "gh no autenticado", nil
			}
			return true, out, nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Ejecutaría: gh auth login")
				return nil
			}
			fmt.Println("  Abriendo autenticación de GitHub...")
			cmd := exec.Command("gh", "auth", "login")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

func ghRepoScope() model.Dependency {
	return model.Dependency{
		ID:          "gh-repo-scope",
		Name:        "GitHub Repo Scope",
		Description: "gh con scope 'repo' habilitado",
		Requires:    []string{"gh-auth"},
		CheckFn: func() (bool, string, error) {
			out, err := system.RunCommand("gh", "auth", "status")
			if err != nil {
				return false, "gh no autenticado", nil
			}
			if strings.Contains(out, "repo") || strings.Contains(out, "Token scopes: ''") {
				return true, "scope repo disponible", nil
			}
			return false, "scope 'repo' no encontrado", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Ejecutaría: gh auth refresh -s repo,project")
				return nil
			}
			fmt.Println("  Actualizando scopes de GitHub...")
			cmd := exec.Command("gh", "auth", "refresh", "-s", "repo,project")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

// --- Node.js ---

func nodeRuntime() model.Dependency {
	return model.Dependency{
		ID:          "node-runtime",
		Name:        "Node.js",
		Description: "Node.js >= 18 (requerido para Figma Console MCP)",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			if !system.HasCommand("node") {
				return false, "node no encontrado en PATH", nil
			}
			out, err := system.RunCommand("node", "--version")
			if err != nil {
				return false, "no se pudo verificar version de node", err
			}
			return true, out, nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			return fmt.Errorf("instala Node.js >= 18: https://nodejs.org/ o usa nvm")
		},
	}
}

// --- Engram ---

func engramBinary() model.Dependency {
	return model.Dependency{
		ID:          "engram-binary",
		Name:        "Engram",
		Description: "Engram binary para memoria persistente",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			if !system.HasCommand("engram") {
				return false, "engram no encontrado en PATH", nil
			}
			out, _ := system.RunCommand("engram", "version")
			return true, out, nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Instalaría engram via brew")
				return nil
			}
			switch system.DetectPlatform() {
			case system.PlatformMacOS:
				fmt.Println("  Instalando engram via Homebrew...")
				return runInstall("brew", "install", "gentleman-programming/tap/engram")
			default:
				return fmt.Errorf("instala engram: go install github.com/Gentleman-Programming/engram@latest")
			}
		},
	}
}

func engramPlugin() model.Dependency {
	return model.Dependency{
		ID:          "engram-plugin",
		Name:        "Engram Plugin",
		Description: "Engram conectado como plugin de Claude Code",
		Requires:    []string{"engram-binary"},
		CheckFn: func() (bool, string, error) {
			home, _ := os.UserHomeDir()
			settingsPath := filepath.Join(home, ".claude", "settings.json")
			data, err := os.ReadFile(settingsPath)
			if err != nil {
				return false, "no se pudo leer ~/.claude/settings.json", nil
			}
			if strings.Contains(string(data), "engram") {
				return true, "engram encontrado en settings.json", nil
			}
			return false, "engram no configurado en Claude Code", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Ejecutaría: claude plugin install engram")
				return nil
			}
			fmt.Println("  Conectando engram a Claude Code...")
			fmt.Println("  Ejecuta manualmente en Claude Code:")
			fmt.Println("    /plugin marketplace add Gentleman-Programming/engram")
			fmt.Println("    /plugin install engram")
			return nil
		},
	}
}

// --- Figma MCP ---

func figmaMCP() model.Dependency {
	return model.Dependency{
		ID:          "figma-mcp",
		Name:        "Figma Console MCP",
		Description: "MCP server para crear diseños en Figma",
		Requires:    []string{"node-runtime"},
		CheckFn: func() (bool, string, error) {
			home, _ := os.UserHomeDir()
			// Claude Code reads MCP servers from ~/.claude.json, NOT ~/.claude/settings.json
			claudeJSON := filepath.Join(home, ".claude.json")
			data, err := os.ReadFile(claudeJSON)
			if err != nil {
				return false, "no se pudo leer ~/.claude.json", nil
			}
			if strings.Contains(string(data), "figma-console") {
				return true, "figma-console MCP configurado", nil
			}
			return false, "figma-console MCP no configurado", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Configuraría Figma Console MCP en ~/.claude.json")
				return nil
			}
			token := ctx.Secrets["figma_token"]
			if token == "" {
				return fmt.Errorf("se requiere FIGMA_ACCESS_TOKEN. Genera uno en Figma → Settings → Personal access tokens")
			}
			return mergeFigmaMCP(token)
		},
	}
}

func mergeFigmaMCP(token string) error {
	home, _ := os.UserHomeDir()
	claudeJSON := filepath.Join(home, ".claude.json")
	return mergeJSONKey(claudeJSON, "mcpServers", "figma-console", map[string]any{
		"command": "npx",
		"args":    []any{"-y", "@anthropic-ai/figma-console-mcp"},
		"env": map[string]any{
			"FIGMA_ACCESS_TOKEN": token,
		},
	})
}

// --- Google Stitch MCP ---

func stitchMCP() model.Dependency {
	return model.Dependency{
		ID:          "stitch-mcp",
		Name:        "Google Stitch MCP",
		Description: "MCP server para generar diseños UI desde prompts (Google Stitch)",
		Requires:    []string{"node-runtime"},
		CheckFn: func() (bool, string, error) {
			home, _ := os.UserHomeDir()
			claudeJSON := filepath.Join(home, ".claude.json")
			data, err := os.ReadFile(claudeJSON)
			if err != nil {
				return false, "no se pudo leer ~/.claude.json", nil
			}
			if strings.Contains(string(data), "stitch") {
				return true, "stitch MCP configurado", nil
			}
			return false, "stitch MCP no configurado", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.DryRun {
				fmt.Println("  [dry-run] Configuraría Google Stitch MCP en ~/.claude.json")
				return nil
			}
			token := ctx.Secrets["stitch_api_key"]
			if token == "" {
				return fmt.Errorf("se requiere STITCH_API_KEY. Obtén una en https://stitch.google.com/")
			}
			return mergeStitchMCP(token)
		},
	}
}

func mergeStitchMCP(token string) error {
	home, _ := os.UserHomeDir()
	claudeJSON := filepath.Join(home, ".claude.json")
	return mergeJSONKey(claudeJSON, "mcpServers", "stitch", map[string]any{
		"command": "npx",
		"args":    []any{"-y", "@anthropic-ai/stitch-mcp"},
		"env": map[string]any{
			"STITCH_API_KEY": token,
		},
	})
}

// --- n8n MCP ---

func n8nMCP() model.Dependency {
	return model.Dependency{
		ID:          "n8n-mcp",
		Name:        "n8n MCP",
		Description: "MCP server para n8n (remoto, hosteado por Inteliside)",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			home, _ := os.UserHomeDir()
			// Check ~/.claude.json (user scope, where claude mcp add writes)
			claudeJSON := filepath.Join(home, ".claude.json")
			data, err := os.ReadFile(claudeJSON)
			if err == nil && strings.Contains(string(data), "n8n-mcp") {
				return true, "n8n-mcp configurado en ~/.claude.json", nil
			}
			// Also check settings.json
			settingsPath := filepath.Join(home, ".claude", "settings.json")
			data, err = os.ReadFile(settingsPath)
			if err == nil && strings.Contains(string(data), "n8n-mcp") {
				return true, "n8n-mcp configurado en settings.json", nil
			}
			return false, "n8n-mcp no configurado", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			mcpURL := defaultN8nMCPURL
			mcpToken := defaultN8nMCPToken
			if v, ok := ctx.Secrets["n8n_mcp_url"]; ok && v != "" {
				mcpURL = v
			}
			if v, ok := ctx.Secrets["n8n_mcp_token"]; ok && v != "" {
				mcpToken = v
			}

			if ctx.DryRun {
				fmt.Printf("  [dry-run] Registraría n8n MCP: %s\n", mcpURL)
				return nil
			}

			fmt.Println("  Registrando n8n MCP server...")
			return runInstall("claude", "mcp", "add",
				"--transport", "http",
				"n8n-mcp", mcpURL,
				"--header", fmt.Sprintf("Authorization: Bearer %s", mcpToken),
				"--scope", "user",
			)
		},
	}
}

// --- GitHub Labels (ATL) ---

func githubLabels() model.Dependency {
	return model.Dependency{
		ID:          "github-labels",
		Name:        "GitHub Labels",
		Description: "Labels de ATL en el repositorio GitHub del proyecto",
		Requires:    []string{"gh-auth"},
		CheckFn: func() (bool, string, error) {
			// Can only check in project context
			return false, "requiere verificación en contexto de proyecto", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			labels := []struct {
				name, color, desc string
			}{
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

			if ctx.DryRun {
				fmt.Printf("  [dry-run] Crearía %d labels en GitHub\n", len(labels))
				return nil
			}

			fmt.Println("  Creando labels en GitHub...")
			for _, l := range labels {
				_, err := system.RunCommand("gh", "label", "create", l.name,
					"--color", l.color,
					"--description", l.desc,
					"--force",
				)
				if err != nil {
					fmt.Printf("    ⚠ Label '%s': %v\n", l.name, err)
				} else {
					fmt.Printf("    ✓ Label '%s'\n", l.name)
				}
			}
			return nil
		},
	}
}

// --- Claude Rules (ATL) ---

func claudeRules() model.Dependency {
	return model.Dependency{
		ID:          "claude-rules",
		Name:        "Claude Rules",
		Description: "Archivos .claude/rules/ para el flujo ATL",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			// Project-specific, check later
			return false, "requiere verificación en contexto de proyecto", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.ProjectDir == "" {
				return fmt.Errorf("se requiere --project-dir para instalar claude rules")
			}
			rulesDir := filepath.Join(ctx.ProjectDir, ".claude", "rules")

			if ctx.DryRun {
				fmt.Printf("  [dry-run] Copiaría rules a %s\n", rulesDir)
				return nil
			}

			if err := os.MkdirAll(rulesDir, 0755); err != nil {
				return fmt.Errorf("no se pudo crear %s: %w", rulesDir, err)
			}

			rules := getRuleFiles()
			for name, content := range rules {
				dest := filepath.Join(rulesDir, name)
				if _, err := os.Stat(dest); err == nil {
					fmt.Printf("    → Skipping %s (ya existe)\n", name)
					continue
				}
				if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
					return fmt.Errorf("no se pudo escribir %s: %w", name, err)
				}
				fmt.Printf("    ✓ %s\n", name)
			}
			return nil
		},
	}
}

// --- CLAUDE.md Template (ATL) ---

func claudeMDTemplate() model.Dependency {
	return model.Dependency{
		ID:          "claude-md-template",
		Name:        "CLAUDE.md Template",
		Description: "Template de CLAUDE.md con configuración ATL",
		Requires:    nil,
		CheckFn: func() (bool, string, error) {
			return false, "requiere verificación en contexto de proyecto", nil
		},
		InstallFn: func(ctx *model.InstallContext) error {
			if ctx.ProjectDir == "" {
				return fmt.Errorf("se requiere --project-dir para generar CLAUDE.md")
			}

			if ctx.DryRun {
				fmt.Println("  [dry-run] Generaría sección ATL en CLAUDE.md")
				return nil
			}

			claudeMD := filepath.Join(ctx.ProjectDir, "CLAUDE.md")
			return appendATLSection(claudeMD, ctx.Secrets)
		},
	}
}

// --- helpers ---

func runInstall(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func mergeJSONKey(path, topKey, subKey string, value any) error {
	data, err := os.ReadFile(path)
	var root map[string]any
	if err != nil {
		root = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &root); err != nil {
			root = make(map[string]any)
		}
	}

	top, ok := root[topKey].(map[string]any)
	if !ok {
		top = make(map[string]any)
	}
	top[subKey] = value
	root[topKey] = top

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func appendATLSection(path string, secrets map[string]string) error {
	marker := "<!-- inteliside:atl-config -->"
	endMarker := "<!-- /inteliside:atl-config -->"

	engramProject := secrets["engram_project"]
	if engramProject == "" {
		engramProject = "my-project-dev"
	}
	githubOwner := secrets["github_owner"]
	if githubOwner == "" {
		githubOwner = "my-org"
	}
	githubRepo := secrets["github_repo"]
	if githubRepo == "" {
		githubRepo = "my-repo"
	}

	section := fmt.Sprintf(`
%s
## ATL Inteliside — Configuración

Variables del pipeline de desarrollo:

- engram_project: "%s"
- github_owner: "%s"
- github_repo: "%s"

### Rules

Los archivos en .claude/rules/ definen el flujo ATL:
- atl-workflow.md — Fases del pipeline
- engram-protocol.md — Protocolo de memoria
- subagent-architecture.md — Reglas de aislamiento
- context-monitoring.md — Gestión de contexto
- team-rules.md — Colaboración de equipo
%s
`, marker, engramProject, githubOwner, githubRepo, endMarker)

	existing, err := os.ReadFile(path)
	if err != nil {
		// No existing file, create new
		return os.WriteFile(path, []byte(strings.TrimSpace(section)+"\n"), 0644)
	}

	content := string(existing)
	if strings.Contains(content, marker) {
		// Replace existing section
		start := strings.Index(content, marker)
		end := strings.Index(content, endMarker) + len(endMarker)
		content = content[:start] + strings.TrimSpace(section) + content[end:]
	} else {
		// Append section
		content = content + "\n" + section
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func getRuleFiles() map[string]string {
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
