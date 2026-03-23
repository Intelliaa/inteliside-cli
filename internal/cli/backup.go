package cli

import (
	"fmt"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/backup"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Gestionar backups de configuración",
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todos los backups disponibles",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Println("  Backups disponibles:")
		fmt.Println("  " + strings.Repeat("─", 50))

		manifests, err := backup.List()
		if err != nil {
			return err
		}

		if len(manifests) == 0 {
			fmt.Println("  (ninguno)")
			fmt.Println()
			return nil
		}

		for _, m := range manifests {
			fmt.Printf("  %s  (%d archivos)  %s\n",
				m.ID,
				len(m.Files),
				m.CreatedAt.Format("02 Jan 2006 15:04:05"),
			)
			for _, f := range m.Files {
				fmt.Printf("    • %s\n", f)
			}
		}
		fmt.Println()
		fmt.Println("  Para restaurar: inteliside backup restore <id>")
		fmt.Println()
		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-id>",
	Short: "Restaura archivos desde un backup",
	Long: `Restaura todos los archivos de configuración desde un snapshot previo.

Ejemplos:
  inteliside backup list                  # Ver backups disponibles
  inteliside backup restore 20260322-143022  # Restaurar un backup específico`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		fmt.Println()
		fmt.Printf("  Restaurando backup %s...\n\n", id)

		if err := backup.Restore(id); err != nil {
			return err
		}

		fmt.Println("\n  Reinicia Claude Code para aplicar los cambios.")
		fmt.Println()
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	rootCmd.AddCommand(backupCmd)
}
