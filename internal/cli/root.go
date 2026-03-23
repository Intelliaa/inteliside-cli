package cli

import (
	"github.com/spf13/cobra"
)

// version is set at build time by goreleaser via ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "inteliside",
	Short: "CLI del Marketplace de Plugins Inteliside",
	Long: `Inteliside CLI automatiza la instalación y configuración de los plugins
del Marketplace de Inteliside para Claude Code.

Un comando. Cualquier rol. Cualquier plataforma.

Ejecuta sin argumentos para abrir la TUI interactiva, o usa --preset para
instalación directa.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// When run without subcommands, launch TUI
		return runTUI()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
