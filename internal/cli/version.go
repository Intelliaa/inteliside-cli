package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Muestra la versión de inteliside CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("inteliside v%s\n", version)
	},
}
