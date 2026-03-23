package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/system"
)

// Check represents a single verification check.
type Check struct {
	Name   string
	Status string // "ok", "warn", "fail"
	Detail string
}

// RunAll performs all health checks and returns results.
func RunAll(projectDir string) []Check {
	var checks []Check

	checks = append(checks, checkGH()...)
	checks = append(checks, checkEngram()...)
	checks = append(checks, checkMCPs()...)

	if projectDir != "" {
		checks = append(checks, checkProjectFiles(projectDir)...)
	}

	return checks
}

// PrintResults displays check results.
func PrintResults(checks []Check) {
	fmt.Println("\n  Health Check Results:")
	fmt.Println("  " + strings.Repeat("─", 50))

	for _, c := range checks {
		icon := "✓"
		if c.Status == "warn" {
			icon = "⚠"
		} else if c.Status == "fail" {
			icon = "✗"
		}
		fmt.Printf("  %s %s — %s\n", icon, c.Name, c.Detail)
	}

	ok, warn, fail := 0, 0, 0
	for _, c := range checks {
		switch c.Status {
		case "ok":
			ok++
		case "warn":
			warn++
		case "fail":
			fail++
		}
	}
	fmt.Printf("\n  Total: %d ok, %d warnings, %d failed\n", ok, warn, fail)
}

func checkGH() []Check {
	var checks []Check

	if !system.HasCommand("gh") {
		checks = append(checks, Check{"GitHub CLI", "fail", "gh no encontrado"})
		return checks
	}
	checks = append(checks, Check{"GitHub CLI", "ok", "instalado"})

	_, err := system.RunCommand("gh", "auth", "status")
	if err != nil {
		checks = append(checks, Check{"GitHub Auth", "fail", "no autenticado"})
	} else {
		checks = append(checks, Check{"GitHub Auth", "ok", "autenticado"})
	}

	return checks
}

func checkEngram() []Check {
	var checks []Check

	if !system.HasCommand("engram") {
		checks = append(checks, Check{"Engram", "warn", "no instalado"})
		return checks
	}
	checks = append(checks, Check{"Engram", "ok", "instalado"})

	return checks
}

func checkMCPs() []Check {
	var checks []Check
	home, _ := os.UserHomeDir()

	// All MCP servers live in ~/.claude.json
	claudeJSON, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err == nil {
		content := string(claudeJSON)
		if strings.Contains(content, "figma-console") {
			checks = append(checks, Check{"Figma Console MCP", "ok", "configurado"})
		}
		if strings.Contains(content, "stitch") {
			checks = append(checks, Check{"Google Stitch MCP", "ok", "configurado"})
		}
		if strings.Contains(content, "n8n-mcp") {
			checks = append(checks, Check{"n8n MCP", "ok", "configurado"})
		}
	}

	return checks
}

func checkProjectFiles(dir string) []Check {
	var checks []Check

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	if _, err := os.Stat(claudeMD); err == nil {
		data, _ := os.ReadFile(claudeMD)
		if strings.Contains(string(data), "inteliside:atl-config") {
			checks = append(checks, Check{"CLAUDE.md (ATL)", "ok", "sección ATL presente"})
		}
	}

	rulesDir := filepath.Join(dir, ".claude", "rules")
	if entries, err := os.ReadDir(rulesDir); err == nil && len(entries) > 0 {
		checks = append(checks, Check{".claude/rules/", "ok", fmt.Sprintf("%d archivos", len(entries))})
	}

	return checks
}
