package legacy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Artifacts holds paths to detected legacy files in a project.
type Artifacts struct {
	ClaudeMD      string            // path to existing CLAUDE.md
	ClaudeMDBody  string            // content of existing CLAUDE.md
	RulesDir      string            // path to existing .claude/rules/
	RuleFiles     []string          // individual rule file paths
	DocsDir       string            // path to existing docs/
	DocsFiles     []string          // files in docs/
	IsLegacy      bool              // true if CLAUDE.md exists but lacks ATL markers
	ExtractedVars map[string]string // parsed from legacy CLAUDE.md
}

const atlMarker = "<!-- inteliside:atl-config -->"

// Detect checks whether a project directory contains legacy artifacts
// (CLAUDE.md without ATL markers, .claude/rules/, docs/).
func Detect(projectDir string) *Artifacts {
	a := &Artifacts{
		ExtractedVars: make(map[string]string),
	}

	claudeMD := filepath.Join(projectDir, "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		return a // no CLAUDE.md → not legacy
	}

	content := string(data)

	// If it already has ATL markers, this is an ATL project — not legacy
	if strings.Contains(content, atlMarker) {
		return a
	}

	// It has a CLAUDE.md without ATL markers → legacy
	a.IsLegacy = true
	a.ClaudeMD = claudeMD
	a.ClaudeMDBody = content

	// Check for .claude/rules/
	rulesDir := filepath.Join(projectDir, ".claude", "rules")
	if entries, err := os.ReadDir(rulesDir); err == nil && len(entries) > 0 {
		a.RulesDir = rulesDir
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				a.RuleFiles = append(a.RuleFiles, filepath.Join(rulesDir, e.Name()))
			}
		}
	}

	// Check for docs/
	docsDir := filepath.Join(projectDir, "docs")
	if entries, err := os.ReadDir(docsDir); err == nil && len(entries) > 0 {
		a.DocsDir = docsDir
		for _, e := range entries {
			a.DocsFiles = append(a.DocsFiles, filepath.Join(docsDir, e.Name()))
		}
	}

	// Extract vars from legacy CLAUDE.md
	a.ExtractedVars = ExtractVars(content)

	return a
}

// ExtractVars parses a legacy CLAUDE.md and extracts useful values
// (stack, commands, conventions, env vars) to pre-fill the new ATL template.
func ExtractVars(content string) map[string]string {
	vars := make(map[string]string)

	// Extract stack
	if v := extractSection(content, "Stack"); v != "" {
		vars["project_stack"] = v
	}

	// Extract description
	if v := extractSection(content, "Descripcion"); v != "" {
		vars["project_description"] = v
	}
	if v := extractSection(content, "Descripción"); v != "" {
		vars["project_description"] = v
	}

	// Extract dev commands from code blocks
	vars["cmd_dev"] = extractCommand(content, "dev")
	vars["cmd_test"] = extractCommand(content, "test")
	vars["cmd_build"] = extractCommand(content, "build")
	vars["cmd_db"] = extractCommand(content, "db")

	// Extract test framework
	if v := extractSection(content, "Framework"); v != "" {
		vars["test_framework"] = v
	}

	// Extract env vars from code blocks
	vars["env_vars"] = extractEnvVars(content)

	// Extract engram_project if present
	if v := extractYAMLValue(content, "engram_project"); v != "" {
		vars["engram_project"] = v
	}

	// Extract github_owner/repo if present
	if v := extractYAMLValue(content, "github_owner"); v != "" {
		vars["github_owner"] = v
	}
	if v := extractYAMLValue(content, "github_repo"); v != "" {
		vars["github_repo"] = v
	}

	// Extract dev environment
	if v := extractSection(content, "Entorno"); v != "" {
		vars["dev_environment"] = v
	}

	// Extract folder structure (look for indented tree blocks)
	vars["folder_structure"] = extractFolderStructure(content)

	return vars
}

// Archive moves legacy artifacts to docs/legacy/ preserving structure.
// Returns the list of archived file paths.
func Archive(projectDir string, artifacts *Artifacts) ([]string, error) {
	legacyDir := filepath.Join(projectDir, "docs", "legacy")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		return nil, fmt.Errorf("no se pudo crear docs/legacy/: %w", err)
	}

	var archived []string

	// Archive CLAUDE.md
	if artifacts.ClaudeMD != "" {
		dest := filepath.Join(legacyDir, "CLAUDE.md")
		if err := moveFile(artifacts.ClaudeMD, dest); err != nil {
			return archived, fmt.Errorf("no se pudo archivar CLAUDE.md: %w", err)
		}
		archived = append(archived, dest)
	}

	// Archive .claude/rules/
	if artifacts.RulesDir != "" && len(artifacts.RuleFiles) > 0 {
		rulesArchive := filepath.Join(legacyDir, ".claude", "rules")
		if err := os.MkdirAll(rulesArchive, 0755); err != nil {
			return archived, err
		}
		for _, rulePath := range artifacts.RuleFiles {
			dest := filepath.Join(rulesArchive, filepath.Base(rulePath))
			if err := moveFile(rulePath, dest); err != nil {
				fmt.Printf("  ⚠ No se pudo archivar %s: %v\n", filepath.Base(rulePath), err)
				continue
			}
			archived = append(archived, dest)
		}
		// Remove empty rules dir
		os.Remove(artifacts.RulesDir)
	}

	// Archive docs/ contents (excluding legacy/ itself)
	if artifacts.DocsDir != "" {
		docsArchive := filepath.Join(legacyDir, "docs")
		if err := os.MkdirAll(docsArchive, 0755); err != nil {
			return archived, err
		}
		for _, docPath := range artifacts.DocsFiles {
			name := filepath.Base(docPath)
			if name == "legacy" {
				continue // don't archive ourselves
			}
			info, err := os.Stat(docPath)
			if err != nil {
				continue
			}
			if info.IsDir() {
				dest := filepath.Join(docsArchive, name)
				if err := moveDir(docPath, dest); err != nil {
					fmt.Printf("  ⚠ No se pudo archivar docs/%s: %v\n", name, err)
					continue
				}
				archived = append(archived, dest)
			} else {
				dest := filepath.Join(docsArchive, name)
				if err := moveFile(docPath, dest); err != nil {
					fmt.Printf("  ⚠ No se pudo archivar docs/%s: %v\n", name, err)
					continue
				}
				archived = append(archived, dest)
			}
		}
	}

	return archived, nil
}

// --- helpers ---

func extractSection(content, label string) string {
	// Look for "**Label**: value" or "- **Label**: value"
	re := regexp.MustCompile(`(?i)\*\*` + regexp.QuoteMeta(label) + `\*\*\s*:\s*(.+)`)
	if m := re.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractYAMLValue(content, key string) string {
	// Match key: "value" or key: value inside code blocks
	re := regexp.MustCompile(`(?m)` + regexp.QuoteMeta(key) + `\s*:\s*"?([^"\n]+)"?`)
	if m := re.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractCommand(content, keyword string) string {
	// Look for lines with the keyword inside bash code blocks
	lines := strings.Split(content, "\n")
	inBash := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```bash") || strings.HasPrefix(trimmed, "```shell") {
			inBash = true
			continue
		}
		if strings.HasPrefix(trimmed, "```") && inBash {
			inBash = false
			continue
		}
		if inBash && strings.Contains(strings.ToLower(trimmed), keyword) && !strings.HasPrefix(trimmed, "#") {
			return trimmed
		}
	}
	return ""
}

func extractEnvVars(content string) string {
	lines := strings.Split(content, "\n")
	inBash := false
	var envLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```bash") || strings.HasPrefix(trimmed, "```shell") || strings.HasPrefix(trimmed, "```") && inBash {
			if inBash {
				inBash = false
			} else {
				inBash = true
			}
			continue
		}
		if inBash && strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "#") {
			envLines = append(envLines, trimmed)
		}
	}
	return strings.Join(envLines, "\n")
}

func extractFolderStructure(content string) string {
	lines := strings.Split(content, "\n")
	inBlock := false
	var structLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				// Check if we captured something useful
				if len(structLines) > 2 {
					return strings.Join(structLines, "\n")
				}
				structLines = nil
			}
			inBlock = !inBlock
			continue
		}
		if inBlock && (strings.Contains(line, "├") || strings.Contains(line, "└") || strings.Contains(line, "│")) {
			structLines = append(structLines, line)
		}
	}
	return ""
}

// moveFile copies src to dst then removes src. Falls back to copy if rename fails.
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	// Try rename first (atomic, same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback: copy + remove
	return copyAndRemove(src, dst)
}

func copyAndRemove(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	out.Close()
	in.Close()
	return os.Remove(src)
}

// moveDir moves an entire directory tree.
func moveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback: walk and copy
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyAndRemove(path, target)
	})
}
