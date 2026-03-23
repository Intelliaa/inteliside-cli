package cli

import (
	"fmt"
	"os"

	"github.com/Intelliaa/inteliside-cli/internal/verify"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifica el estado de las dependencias y configuración",
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, _ := cmd.Flags().GetString("project-dir")
		if projectDir == "" {
			projectDir, _ = os.Getwd()
		}

		fmt.Println()
		fmt.Println("  ╔══════════════════════════════════════════╗")
		fmt.Println("  ║        Inteliside CLI — Verify           ║")
		fmt.Println("  ╚══════════════════════════════════════════╝")

		checks := verify.RunAll(projectDir)
		verify.PrintResults(checks)
		fmt.Println()
	},
}

func init() {
	verifyCmd.Flags().String("project-dir", "", "Directorio del proyecto a verificar")
}
