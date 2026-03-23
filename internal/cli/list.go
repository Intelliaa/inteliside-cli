package cli

import (
	"fmt"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista plugins y presets disponibles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  Plugins disponibles:")
		fmt.Println("  " + strings.Repeat("─", 60))
		for _, p := range catalog.AllPlugins() {
			fmt.Printf("  %-18s [%s] %s\n", p.ID, p.Role, p.Description)
		}

		fmt.Println()
		fmt.Println("  Presets por rol:")
		fmt.Println("  " + strings.Repeat("─", 60))
		for _, p := range catalog.AllPresets() {
			if p.ID == "custom" {
				continue
			}
			plugins := strings.Join(p.PluginIDs, ", ")
			fmt.Printf("  %-12s %s\n             → %s\n", p.ID, p.Description, plugins)
		}
		fmt.Println()
	},
}
