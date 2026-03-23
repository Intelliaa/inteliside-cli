package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/backup"
	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/Intelliaa/inteliside-cli/internal/deps"
	"github.com/Intelliaa/inteliside-cli/internal/model"
	"github.com/Intelliaa/inteliside-cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <plugin>",
	Short: "Re-ejecuta la post-configuración de un plugin específico",
	Long: `Re-configura un plugin individual sin reinstalar todo.
Útil para actualizar tokens, re-crear labels, o regenerar archivos.

Ejemplos:
  inteliside setup ux-studio        # Re-configura Figma MCP
  inteliside setup atl-inteliside   # Re-crea labels, rules, CLAUDE.md
  inteliside setup n8n-studio       # Re-registra n8n MCP`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginID := args[0]
		plugin := catalog.PluginByID(pluginID)
		if plugin == nil {
			return fmt.Errorf("plugin desconocido: %s\nDisponibles: %s", pluginID, strings.Join(catalog.PluginIDs(), ", "))
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectDir, _ := cmd.Flags().GetString("project-dir")
		if projectDir == "" {
			projectDir, _ = os.Getwd()
		}

		fmt.Println()
		fmt.Printf("  Re-configurando %s...\n\n", plugin.Name)

		// Resolve only this plugin's deps
		steps, err := deps.Resolve([]string{pluginID})
		if err != nil {
			return fmt.Errorf("error resolviendo dependencias: %w", err)
		}

		if dryRun {
			fmt.Println("  Plan:")
			for i, s := range steps {
				fmt.Printf("    %d. %s\n", i+1, s.Name)
			}
			fmt.Println("\n  [dry-run] No se realizaron cambios.")
			return nil
		}

		// Backup
		targets := backup.TargetFiles(projectDir)
		snap, _ := backup.Create(targets)
		if snap != nil && len(snap.Files) > 0 {
			fmt.Printf("  ✓ Backup: %s\n", snap.ID)
		}

		ctx := model.NewInstallContext(projectDir)
		fmt.Println("  " + strings.Repeat("─", 40))
		result := pipeline.Run(steps, ctx)
		fmt.Println("  " + strings.Repeat("─", 40))

		if result.Error != nil {
			return fmt.Errorf("setup falló: %w", result.Error)
		}

		fmt.Printf("\n  ✓ %s re-configurado exitosamente (%s)\n\n", plugin.Name, result.Duration.Round(1e6))
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Mostrar plan sin ejecutar")
	setupCmd.Flags().String("project-dir", "", "Directorio del proyecto destino")
	rootCmd.AddCommand(setupCmd)
}
