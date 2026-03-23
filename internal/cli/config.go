package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gestionar configuración",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Muestra la configuración actual relevante para los plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()

		fmt.Println()
		fmt.Println("  ╔══════════════════════════════════════════╗")
		fmt.Println("  ║        Inteliside CLI — Config           ║")
		fmt.Println("  ╚══════════════════════════════════════════╝")
		fmt.Println()

		// ~/.claude/settings.json
		settingsPath := filepath.Join(home, ".claude", "settings.json")
		printJSONSection("~/.claude/settings.json", settingsPath, []string{"mcpServers", "permissions", "enabledPlugins"})

		// ~/.claude.json
		claudeJSONPath := filepath.Join(home, ".claude.json")
		printJSONSection("~/.claude.json", claudeJSONPath, []string{"mcpServers"})

		// Project CLAUDE.md (if in a project)
		projectDir, _ := cmd.Flags().GetString("project-dir")
		if projectDir == "" {
			projectDir, _ = os.Getwd()
		}
		claudeMD := filepath.Join(projectDir, "CLAUDE.md")
		if data, err := os.ReadFile(claudeMD); err == nil {
			if strings.Contains(string(data), "inteliside:atl-config") {
				fmt.Println("  CLAUDE.md (ATL section):")
				fmt.Println("  " + strings.Repeat("─", 50))
				// Extract just the ATL section
				content := string(data)
				start := strings.Index(content, "<!-- inteliside:atl-config -->")
				end := strings.Index(content, "<!-- /inteliside:atl-config -->")
				if start >= 0 && end >= 0 {
					section := content[start : end+len("<!-- /inteliside:atl-config -->")]
					for _, line := range strings.Split(section, "\n") {
						fmt.Printf("    %s\n", line)
					}
				}
				fmt.Println()
			}
		}

		// .claude/rules/
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		if entries, err := os.ReadDir(rulesDir); err == nil {
			fmt.Println("  .claude/rules/:")
			fmt.Println("  " + strings.Repeat("─", 50))
			for _, e := range entries {
				fmt.Printf("    • %s\n", e.Name())
			}
			fmt.Println()
		}

		// Backups
		backupDir := filepath.Join(home, ".inteliside", "backups")
		if entries, err := os.ReadDir(backupDir); err == nil && len(entries) > 0 {
			fmt.Println("  Backups disponibles:")
			fmt.Println("  " + strings.Repeat("─", 50))
			for _, e := range entries {
				if e.IsDir() {
					fmt.Printf("    • %s\n", e.Name())
				}
			}
			fmt.Println()
		}

		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset <plugin>",
	Short: "Elimina la configuración de un plugin específico",
	Long: `Elimina las entradas de configuración que un plugin agregó.
No elimina archivos de usuario — solo las secciones marcadas por inteliside.

Ejemplos:
  inteliside config reset ux-studio       # Elimina figma-console de settings.json
  inteliside config reset n8n-studio      # Elimina n8n-mcp de claude.json
  inteliside config reset atl-inteliside  # Elimina sección ATL de CLAUDE.md`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginID := args[0]
		plugin := catalog.PluginByID(pluginID)
		if plugin == nil {
			return fmt.Errorf("plugin desconocido: %s", pluginID)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectDir, _ := cmd.Flags().GetString("project-dir")
		if projectDir == "" {
			projectDir, _ = os.Getwd()
		}
		home, _ := os.UserHomeDir()

		fmt.Printf("\n  Reseteando configuración de %s...\n\n", plugin.Name)

		switch pluginID {
		case "ux-studio":
			path := filepath.Join(home, ".claude", "settings.json")
			if dryRun {
				fmt.Println("  [dry-run] Eliminaría 'figma-console' y 'stitch' de", path)
			} else {
				if err := removeJSONKey(path, "mcpServers", "figma-console"); err != nil {
					return err
				}
				fmt.Println("  ✓ Eliminado figma-console de settings.json")
				if err := removeJSONKey(path, "mcpServers", "stitch"); err != nil {
					return err
				}
				fmt.Println("  ✓ Eliminado stitch de settings.json")
			}

		case "n8n-studio":
			path := filepath.Join(home, ".claude.json")
			if dryRun {
				fmt.Println("  [dry-run] Eliminaría 'n8n-mcp' de", path)
			} else {
				if err := removeJSONKey(path, "mcpServers", "n8n-mcp"); err != nil {
					return err
				}
				fmt.Println("  ✓ Eliminado n8n-mcp de claude.json")
			}

		case "atl-inteliside":
			claudeMD := filepath.Join(projectDir, "CLAUDE.md")
			if dryRun {
				fmt.Println("  [dry-run] Eliminaría sección ATL de CLAUDE.md")
				fmt.Println("  [dry-run] No eliminaría .claude/rules/ (contiene archivos del usuario)")
			} else {
				if err := removeATLSection(claudeMD); err != nil {
					fmt.Printf("  ⚠ CLAUDE.md: %v\n", err)
				} else {
					fmt.Println("  ✓ Eliminada sección ATL de CLAUDE.md")
				}
				fmt.Println("  ℹ .claude/rules/ no se elimina — puede contener archivos personalizados")
			}

		default:
			fmt.Printf("  ℹ %s no tiene configuración específica para resetear\n", plugin.Name)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configResetCmd)
	configShowCmd.Flags().String("project-dir", "", "Directorio del proyecto")
	configResetCmd.Flags().Bool("dry-run", false, "Mostrar cambios sin ejecutar")
	configResetCmd.Flags().String("project-dir", "", "Directorio del proyecto")
	rootCmd.AddCommand(configCmd)
}

// helpers

func printJSONSection(label, path string, keys []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return
	}

	fmt.Printf("  %s:\n", label)
	fmt.Println("  " + strings.Repeat("─", 50))

	for _, key := range keys {
		val, ok := root[key]
		if !ok {
			continue
		}
		pretty, _ := json.MarshalIndent(val, "    ", "  ")
		fmt.Printf("    %s: %s\n", key, string(pretty))
	}
	fmt.Println()
}

func removeJSONKey(path, topKey, subKey string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // file doesn't exist, nothing to remove
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	if top, ok := root[topKey].(map[string]any); ok {
		delete(top, subKey)
		root[topKey] = top
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func removeATLSection(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("archivo no encontrado: %s", path)
	}

	content := string(data)
	marker := "<!-- inteliside:atl-config -->"
	endMarker := "<!-- /inteliside:atl-config -->"

	start := strings.Index(content, marker)
	end := strings.Index(content, endMarker)
	if start < 0 || end < 0 {
		return fmt.Errorf("sección ATL no encontrada")
	}

	// Remove the section plus surrounding whitespace
	before := strings.TrimRight(content[:start], "\n")
	after := strings.TrimLeft(content[end+len(endMarker):], "\n")
	result := before + "\n" + after

	return os.WriteFile(path, []byte(result), 0644)
}
