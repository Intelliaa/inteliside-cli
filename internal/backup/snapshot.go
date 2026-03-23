package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manifest describes a backup snapshot.
type Manifest struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files"`
}

// BackupDir returns the backup directory path.
func BackupDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".inteliside", "backups")
}

// Create backs up the given files into a timestamped snapshot.
func Create(files []string) (*Manifest, error) {
	id := time.Now().Format("20060102-150405")
	snapDir := filepath.Join(BackupDir(), id)
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		return nil, fmt.Errorf("no se pudo crear directorio de backup: %w", err)
	}

	var backedUp []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue // file doesn't exist yet, skip
		}
		dest := filepath.Join(snapDir, filepath.Base(f))
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return nil, fmt.Errorf("no se pudo copiar %s: %w", f, err)
		}
		backedUp = append(backedUp, f)
	}

	manifest := &Manifest{
		ID:        id,
		CreatedAt: time.Now(),
		Files:     backedUp,
	}

	mdata, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(snapDir, "manifest.json"), mdata, 0644); err != nil {
		return nil, err
	}

	return manifest, nil
}

// List returns all available backup manifests, newest first.
func List() ([]Manifest, error) {
	dir := BackupDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil // no backups dir = no backups
	}

	var manifests []Manifest
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if !e.IsDir() {
			continue
		}
		mpath := filepath.Join(dir, e.Name(), "manifest.json")
		data, err := os.ReadFile(mpath)
		if err != nil {
			continue
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

// Restore copies backed-up files back to their original locations.
func Restore(id string) error {
	snapDir := filepath.Join(BackupDir(), id)
	mpath := filepath.Join(snapDir, "manifest.json")
	data, err := os.ReadFile(mpath)
	if err != nil {
		return fmt.Errorf("backup '%s' no encontrado", id)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("manifest corrupto: %w", err)
	}

	restored := 0
	for _, origPath := range m.Files {
		backupFile := filepath.Join(snapDir, filepath.Base(origPath))
		content, err := os.ReadFile(backupFile)
		if err != nil {
			fmt.Printf("  ⚠ No se pudo leer backup de %s: %v\n", filepath.Base(origPath), err)
			continue
		}

		dir := filepath.Dir(origPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("  ⚠ No se pudo crear directorio %s: %v\n", dir, err)
			continue
		}

		if err := os.WriteFile(origPath, content, 0644); err != nil {
			fmt.Printf("  ⚠ No se pudo restaurar %s: %v\n", origPath, err)
			continue
		}
		fmt.Printf("  ✓ Restaurado: %s\n", origPath)
		restored++
	}

	fmt.Printf("\n  %d/%d archivos restaurados desde backup %s\n", restored, len(m.Files), id)
	return nil
}

// TargetFiles returns all files that the CLI may modify.
func TargetFiles(projectDir string) []string {
	home, _ := os.UserHomeDir()
	files := []string{
		filepath.Join(home, ".claude", "settings.json"),
		filepath.Join(home, ".claude.json"),
	}
	if projectDir != "" {
		files = append(files, filepath.Join(projectDir, "CLAUDE.md"))
		// Include docs CLAUDE.md files
		files = append(files, filepath.Join(projectDir, "docs", "CLAUDE.md"))
		files = append(files, filepath.Join(projectDir, "docs", "ux-ui", "CLAUDE.md"))
		files = append(files, filepath.Join(projectDir, "docs", "legacy", "CLAUDE.md"))
		// Include rule files from .claude/rules/
		rulesDir := filepath.Join(projectDir, ".claude", "rules")
		if entries, err := os.ReadDir(rulesDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					files = append(files, filepath.Join(rulesDir, e.Name()))
				}
			}
		}
	}
	return files
}
