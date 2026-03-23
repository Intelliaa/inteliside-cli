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

var (
	flagPreset     string
	flagPlugins    string
	flagDryRun     bool
	flagYes        bool
	flagVerbose    bool
	flagProjectDir string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala plugins con sus dependencias y post-configuración",
	Long: `Instala plugins del marketplace con resolución automática de dependencias.

Ejemplos:
  inteliside install --preset dev
  inteliside install --preset fullstack --dry-run
  inteliside install --plugin atl-inteliside,n8n-studio
  inteliside install --preset dev --project-dir ./my-project`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().StringVar(&flagPreset, "preset", "", "Preset por rol: pm, designer, dev, fullstack, automation, legacy")
	installCmd.Flags().StringVar(&flagPlugins, "plugin", "", "Plugins específicos separados por coma")
	installCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Mostrar plan sin ejecutar")
	installCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Confirmar todo automáticamente")
	installCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Output detallado")
	installCmd.Flags().StringVar(&flagProjectDir, "project-dir", "", "Directorio del proyecto destino")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Determine which plugins to install
	pluginIDs, err := resolvePluginIDs()
	if err != nil {
		return err
	}

	if len(pluginIDs) == 0 {
		return fmt.Errorf("no se seleccionaron plugins. Usa --preset o --plugin")
	}

	// Print header
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════╗")
	fmt.Println("  ║        Inteliside CLI — Install          ║")
	fmt.Println("  ╚══════════════════════════════════════════╝")
	fmt.Println()

	// Show what will be installed
	fmt.Println("  Plugins a instalar:")
	for _, id := range pluginIDs {
		p := catalog.PluginByID(id)
		if p != nil {
			fmt.Printf("    • %s — %s\n", p.Name, p.Description)
		}
	}
	fmt.Println()

	// Resolve dependencies
	steps, err := deps.Resolve(pluginIDs)
	if err != nil {
		return fmt.Errorf("error resolviendo dependencias: %w", err)
	}

	fmt.Printf("  Plan de instalación: %d pasos\n", len(steps))
	if flagDryRun || flagVerbose {
		for i, s := range steps {
			fmt.Printf("    %d. %s — %s\n", i+1, s.Name, s.Description)
		}
	}
	fmt.Println()

	if flagDryRun {
		fmt.Println("  [dry-run] No se realizaron cambios.")
		return nil
	}

	// Create backup
	projectDir := flagProjectDir
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	targets := backup.TargetFiles(projectDir)
	snap, err := backup.Create(targets)
	if err != nil {
		fmt.Printf("  ⚠ No se pudo crear backup: %v\n", err)
	} else if snap != nil && len(snap.Files) > 0 {
		fmt.Printf("  ✓ Backup creado: %s (%d archivos)\n", snap.ID, len(snap.Files))
	}

	// Build install context
	ctx := model.NewInstallContext(projectDir)
	ctx.DryRun = flagDryRun
	ctx.Verbose = flagVerbose
	ctx.AutoYes = flagYes

	// Execute pipeline
	fmt.Println("\n  Ejecutando instalación...")
	fmt.Println("  " + strings.Repeat("─", 50))

	result := pipeline.Run(steps, ctx)

	fmt.Println("  " + strings.Repeat("─", 50))

	if result.Error != nil {
		fmt.Printf("\n  ✗ Instalación falló en '%s': %v\n", result.Failed, result.Error)
		fmt.Printf("  Completados: %d, Duración: %s\n", len(result.Completed), result.Duration.Round(100*1e6))
		if snap != nil {
			fmt.Printf("  Restaura con: inteliside backup restore %s\n", snap.ID)
		}
		return result.Error
	}

	fmt.Printf("\n  ✓ Instalación completada exitosamente\n")
	fmt.Printf("  %d pasos, %s\n", len(result.Completed), result.Duration.Round(100*1e6))
	fmt.Println("\n  Siguiente paso: reinicia Claude Code y verifica con 'inteliside verify'")

	return nil
}

func resolvePluginIDs() ([]string, error) {
	if flagPreset != "" && flagPlugins != "" {
		return nil, fmt.Errorf("usa --preset o --plugin, no ambos")
	}

	if flagPreset != "" {
		preset := catalog.PresetByID(flagPreset)
		if preset == nil {
			return nil, fmt.Errorf("preset desconocido: %s (opciones: pm, designer, dev, fullstack, automation, legacy)", flagPreset)
		}
		return preset.PluginIDs, nil
	}

	if flagPlugins != "" {
		ids := strings.Split(flagPlugins, ",")
		for _, id := range ids {
			if catalog.PluginByID(strings.TrimSpace(id)) == nil {
				return nil, fmt.Errorf("plugin desconocido: %s", id)
			}
		}
		return ids, nil
	}

	return nil, nil
}
