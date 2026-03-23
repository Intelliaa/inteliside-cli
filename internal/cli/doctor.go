package cli

import (
	"fmt"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/Intelliaa/inteliside-cli/internal/system"
	"github.com/spf13/cobra"
)

type diagnosis struct {
	name    string
	ok      bool
	detail  string
	fix     string // actionable fix command
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnostica dependencias faltantes y sugiere soluciones",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Println("  ╔══════════════════════════════════════════╗")
		fmt.Println("  ║        Inteliside CLI — Doctor           ║")
		fmt.Println("  ╚══════════════════════════════════════════╝")
		fmt.Println()

		diagnoses := runDiagnosis()

		var ok, missing int
		for _, d := range diagnoses {
			if d.ok {
				fmt.Printf("  ✓ %s — %s\n", d.name, d.detail)
				ok++
			} else {
				fmt.Printf("  ✗ %s — %s\n", d.name, d.detail)
				if d.fix != "" {
					fmt.Printf("    %s %s\n", dimCLI("Fix:"), d.fix)
				}
				missing++
			}
		}

		fmt.Println()
		fmt.Println("  " + strings.Repeat("─", 50))
		fmt.Printf("  Resultado: %d ok, %d por resolver\n", ok, missing)

		if missing > 0 {
			fmt.Println()
			fmt.Println("  Opciones rápidas:")
			fmt.Println("    inteliside install --preset dev     # Instala todo para desarrollo")
			fmt.Println("    inteliside install --preset pm      # Solo para PM")
			fmt.Println("    inteliside setup <plugin>           # Re-configura un plugin específico")
		} else {
			fmt.Println("\n  Todo en orden. Ejecuta 'inteliside verify' para health check completo.")
		}
		fmt.Println()

		return nil
	},
}

func runDiagnosis() []diagnosis {
	var results []diagnosis
	platform := system.DetectPlatform()

	// gh CLI
	ghDep := catalog.DependencyByID("gh-cli")
	if ghDep != nil {
		ok, detail, _ := ghDep.CheckFn()
		fix := ""
		if !ok {
			switch platform {
			case system.PlatformMacOS:
				fix = "brew install gh"
			case system.PlatformLinux:
				fix = "https://cli.github.com/"
			default:
				fix = "winget install GitHub.cli"
			}
		}
		results = append(results, diagnosis{"GitHub CLI", ok, detail, fix})
	}

	// gh auth
	if hasDiag(results, "GitHub CLI") {
		authDep := catalog.DependencyByID("gh-auth")
		if authDep != nil {
			ok, detail, _ := authDep.CheckFn()
			fix := ""
			if !ok {
				fix = "gh auth login"
			}
			results = append(results, diagnosis{"GitHub Auth", ok, detail, fix})
		}
	}

	// gh repo scope
	if hasDiag(results, "GitHub Auth") {
		scopeDep := catalog.DependencyByID("gh-repo-scope")
		if scopeDep != nil {
			ok, detail, _ := scopeDep.CheckFn()
			fix := ""
			if !ok {
				fix = "gh auth refresh -s repo,project"
			}
			results = append(results, diagnosis{"GitHub Scopes", ok, detail, fix})
		}
	}

	// Node.js
	nodeDep := catalog.DependencyByID("node-runtime")
	if nodeDep != nil {
		ok, detail, _ := nodeDep.CheckFn()
		fix := ""
		if !ok {
			fix = "https://nodejs.org/ o: curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash"
		}
		results = append(results, diagnosis{"Node.js", ok, detail, fix})
	}

	// Engram
	engramDep := catalog.DependencyByID("engram-binary")
	if engramDep != nil {
		ok, detail, _ := engramDep.CheckFn()
		fix := ""
		if !ok {
			switch platform {
			case system.PlatformMacOS:
				fix = "brew install gentleman-programming/tap/engram"
			default:
				fix = "go install github.com/Gentleman-Programming/engram@latest"
			}
		}
		results = append(results, diagnosis{"Engram", ok, detail, fix})
	}

	// Engram plugin
	engramPlugDep := catalog.DependencyByID("engram-plugin")
	if engramPlugDep != nil {
		ok, detail, _ := engramPlugDep.CheckFn()
		fix := ""
		if !ok {
			fix = "En Claude Code: /plugin marketplace add Gentleman-Programming/engram && /plugin install engram"
		}
		results = append(results, diagnosis{"Engram Plugin", ok, detail, fix})
	}

	// Figma MCP
	figmaDep := catalog.DependencyByID("figma-mcp")
	if figmaDep != nil {
		ok, detail, _ := figmaDep.CheckFn()
		fix := ""
		if !ok {
			fix = "inteliside install --plugin ux-studio"
		}
		results = append(results, diagnosis{"Figma Console MCP", ok, detail, fix})
	}

	// n8n MCP
	n8nDep := catalog.DependencyByID("n8n-mcp")
	if n8nDep != nil {
		ok, detail, _ := n8nDep.CheckFn()
		fix := ""
		if !ok {
			fix = "inteliside install --plugin n8n-studio"
		}
		results = append(results, diagnosis{"n8n MCP", ok, detail, fix})
	}

	return results
}

func hasDiag(results []diagnosis, name string) bool {
	for _, d := range results {
		if d.name == name && d.ok {
			return true
		}
	}
	return false
}

func dimCLI(s string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", s)
}
