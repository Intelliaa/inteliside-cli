package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/spf13/cobra"
)

const marketplaceRepo = "Intelliaa/marketplace-plugins-inteliside"

type marketplaceJSON struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Plugins     []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"plugins"`
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Verifica actualizaciones del marketplace y sincroniza plugins",
	Long: `Compara las versiones locales del catálogo con el marketplace remoto
en GitHub y muestra si hay actualizaciones disponibles.

Ejemplos:
  inteliside sync          # Verificar actualizaciones
  inteliside sync --apply  # Aplicar actualizaciones`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		fmt.Println()
		fmt.Println("  ╔══════════════════════════════════════════╗")
		fmt.Println("  ║        Inteliside CLI — Sync             ║")
		fmt.Println("  ╚══════════════════════════════════════════╝")
		fmt.Println()

		// Fetch remote marketplace.json
		fmt.Println("  Consultando marketplace remoto...")
		remote, err := fetchMarketplace()
		if err != nil {
			return fmt.Errorf("no se pudo obtener marketplace: %w", err)
		}

		fmt.Printf("  Marketplace: %s v%s\n\n", remote.Name, remote.Version)

		// Compare versions
		localPlugins := catalog.AllPlugins()
		localVersions := make(map[string]string)
		for _, p := range localPlugins {
			localVersions[p.ID] = "1.0.0" // hardcoded in catalog for now
		}

		hasUpdates := false
		for _, rp := range remote.Plugins {
			localVer, exists := localVersions[rp.Name]
			if !exists {
				fmt.Printf("  + %s v%s — %s\n", rp.Name, rp.Version, warnStyleCLI("NUEVO"))
				hasUpdates = true
				continue
			}
			if localVer != rp.Version {
				fmt.Printf("  ↑ %s %s → %s\n", rp.Name, localVer, successStyleCLI(rp.Version))
				hasUpdates = true
			} else {
				fmt.Printf("  ✓ %s v%s — al día\n", rp.Name, localVer)
			}
		}

		fmt.Println()
		if !hasUpdates {
			fmt.Println("  Todo al día. No hay actualizaciones.")
		} else if apply {
			fmt.Println("  Aplicando actualizaciones...")
			if err := applySync(); err != nil {
				return err
			}
			fmt.Println("  ✓ Plugins actualizados. Reinicia Claude Code.")
		} else {
			fmt.Println("  Hay actualizaciones disponibles.")
			fmt.Println("  Ejecuta: inteliside sync --apply")
		}
		fmt.Println()

		return nil
	},
}

func init() {
	syncCmd.Flags().Bool("apply", false, "Aplicar actualizaciones encontradas")
	rootCmd.AddCommand(syncCmd)
}

func fetchMarketplace() (*marketplaceJSON, error) {
	// Try gh api first (works with auth)
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/contents/.claude-plugin/marketplace.json", marketplaceRepo),
		"--jq", ".content",
	).Output()
	if err != nil {
		// Fallback: try local clone if exists
		home, _ := os.UserHomeDir()
		localPath := filepath.Join(home, ".inteliside", "marketplace-cache", "marketplace.json")
		data, ferr := os.ReadFile(localPath)
		if ferr != nil {
			return nil, fmt.Errorf("gh api falló y no hay cache local: %w", err)
		}
		var m marketplaceJSON
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}

	// gh api returns base64 content, decode it
	decoded, err := decodeBase64Content(strings.TrimSpace(string(out)))
	if err != nil {
		return nil, fmt.Errorf("error decodificando contenido: %w", err)
	}

	// Cache for offline use
	cacheDir := filepath.Join(homeDir(), ".inteliside", "marketplace-cache")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "marketplace.json"), decoded, 0644)

	var m marketplaceJSON
	if err := json.Unmarshal(decoded, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func decodeBase64Content(encoded string) ([]byte, error) {
	// Remove newlines that GitHub API adds
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, "\\n", "")
	encoded = strings.TrimSpace(encoded)

	// Use standard library base64
	import_encoding := exec.Command("bash", "-c",
		fmt.Sprintf("echo '%s' | base64 --decode", encoded))
	return import_encoding.Output()
}

func applySync() error {
	// Re-add marketplace to get latest versions
	cmd := exec.Command("claude", "plugin", "marketplace", "add", marketplaceRepo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

func successStyleCLI(s string) string {
	return fmt.Sprintf("\033[32;1m%s\033[0m", s)
}

func warnStyleCLI(s string) string {
	return fmt.Sprintf("\033[33;1m%s\033[0m", s)
}
