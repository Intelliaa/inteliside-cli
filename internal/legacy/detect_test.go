package legacy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_NoClaudeMD(t *testing.T) {
	dir := t.TempDir()
	a := Detect(dir)
	if a.IsLegacy {
		t.Error("expected IsLegacy=false when no CLAUDE.md exists")
	}
}

func TestDetect_WithATLMarkers(t *testing.T) {
	dir := t.TempDir()
	content := `# CLAUDE.md
<!-- inteliside:atl-config -->
## ATL Inteliside
engram_project: "test-dev"
<!-- /inteliside:atl-config -->
`
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0644)

	a := Detect(dir)
	if a.IsLegacy {
		t.Error("expected IsLegacy=false when ATL markers present")
	}
}

func TestDetect_Legacy(t *testing.T) {
	dir := t.TempDir()
	content := `# CLAUDE.md — Mi Proyecto Legacy

## Proyecto
- **Stack**: Next.js 14 + TypeScript + Prisma

## Comandos
` + "```bash\npnpm dev\npnpm test\npnpm build\n```" + `
`
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0644)

	// Create .claude/rules/
	rulesDir := filepath.Join(dir, ".claude", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "naming.md"), []byte("# Naming conventions"), 0644)
	os.WriteFile(filepath.Join(rulesDir, "testing.md"), []byte("# Testing rules"), 0644)

	// Create docs/
	docsDir := filepath.Join(dir, "docs")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "audit-report.md"), []byte("# Audit"), 0644)

	a := Detect(dir)
	if !a.IsLegacy {
		t.Fatal("expected IsLegacy=true")
	}
	if len(a.RuleFiles) != 2 {
		t.Errorf("expected 2 rule files, got %d", len(a.RuleFiles))
	}
	if len(a.DocsFiles) != 1 {
		t.Errorf("expected 1 docs file, got %d", len(a.DocsFiles))
	}
}

func TestExtractVars_Stack(t *testing.T) {
	content := `# CLAUDE.md
- **Stack**: Next.js 15 + TypeScript + Drizzle + PostgreSQL
- **Descripcion**: Una app de gestion
`
	vars := ExtractVars(content)
	if vars["project_stack"] != "Next.js 15 + TypeScript + Drizzle + PostgreSQL" {
		t.Errorf("unexpected stack: %q", vars["project_stack"])
	}
	if vars["project_description"] != "Una app de gestion" {
		t.Errorf("unexpected description: %q", vars["project_description"])
	}
}

func TestExtractVars_EngramProject(t *testing.T) {
	content := `## Config
engram_project: "mi-app-dev"
github_owner: "Intelliaa"
github_repo: "mi-app"
`
	vars := ExtractVars(content)
	if vars["engram_project"] != "mi-app-dev" {
		t.Errorf("unexpected engram_project: %q", vars["engram_project"])
	}
	if vars["github_owner"] != "Intelliaa" {
		t.Errorf("unexpected github_owner: %q", vars["github_owner"])
	}
	if vars["github_repo"] != "mi-app" {
		t.Errorf("unexpected github_repo: %q", vars["github_repo"])
	}
}

func TestExtractVars_Commands(t *testing.T) {
	content := "# Commands\n```bash\npnpm dev\npnpm test:unit\npnpm build\npnpm db:migrate\n```\n"
	vars := ExtractVars(content)
	if vars["cmd_dev"] != "pnpm dev" {
		t.Errorf("unexpected cmd_dev: %q", vars["cmd_dev"])
	}
	if vars["cmd_test"] != "pnpm test:unit" {
		t.Errorf("unexpected cmd_test: %q", vars["cmd_test"])
	}
	if vars["cmd_build"] != "pnpm build" {
		t.Errorf("unexpected cmd_build: %q", vars["cmd_build"])
	}
	if vars["cmd_db"] != "pnpm db:migrate" {
		t.Errorf("unexpected cmd_db: %q", vars["cmd_db"])
	}
}

func TestArchive_MovesFiles(t *testing.T) {
	dir := t.TempDir()

	// Setup legacy files
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Legacy CLAUDE"), 0644)

	rulesDir := filepath.Join(dir, ".claude", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "naming.md"), []byte("# Naming"), 0644)

	docsDir := filepath.Join(dir, "docs")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "audit.md"), []byte("# Audit"), 0644)

	artifacts := &Artifacts{
		IsLegacy:  true,
		ClaudeMD:  filepath.Join(dir, "CLAUDE.md"),
		RulesDir:  rulesDir,
		RuleFiles: []string{filepath.Join(rulesDir, "naming.md")},
		DocsDir:   docsDir,
		DocsFiles: []string{filepath.Join(docsDir, "audit.md")},
	}

	archived, err := Archive(dir, artifacts)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}

	if len(archived) == 0 {
		t.Fatal("expected at least one archived file")
	}

	// Original files should be gone
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should have been moved")
	}

	// Legacy dir should have the files
	legacyClaude := filepath.Join(dir, "docs", "legacy", "CLAUDE.md")
	if _, err := os.Stat(legacyClaude); err != nil {
		t.Errorf("docs/legacy/CLAUDE.md should exist: %v", err)
	}

	legacyRules := filepath.Join(dir, "docs", "legacy", ".claude", "rules", "naming.md")
	if _, err := os.Stat(legacyRules); err != nil {
		t.Errorf("docs/legacy/.claude/rules/naming.md should exist: %v", err)
	}

	legacyDocs := filepath.Join(dir, "docs", "legacy", "docs", "audit.md")
	if _, err := os.Stat(legacyDocs); err != nil {
		t.Errorf("docs/legacy/docs/audit.md should exist: %v", err)
	}
}

func TestArchive_HandlesNoRules(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Legacy"), 0644)

	artifacts := &Artifacts{
		IsLegacy: true,
		ClaudeMD: filepath.Join(dir, "CLAUDE.md"),
	}

	archived, err := Archive(dir, artifacts)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}
	if len(archived) != 1 {
		t.Errorf("expected 1 archived file, got %d", len(archived))
	}
}

func TestDetect_WithDotClaudeMD(t *testing.T) {
	dir := t.TempDir()

	// Root CLAUDE.md (legacy)
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Legacy Project"), 0644)

	// .claude/CLAUDE.md (project instructions)
	dotClaudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)
	os.WriteFile(filepath.Join(dotClaudeDir, "CLAUDE.md"), []byte(`# Project Instructions
- **Stack**: Remix + Drizzle + PostgreSQL
github_owner: "MyOrg"
github_repo: "my-legacy-app"
`), 0644)

	a := Detect(dir)
	if !a.IsLegacy {
		t.Fatal("expected IsLegacy=true")
	}
	if a.DotClaudeMD == "" {
		t.Error("expected DotClaudeMD to be detected")
	}
	// Vars from .claude/CLAUDE.md should be extracted
	if a.ExtractedVars["github_owner"] != "MyOrg" {
		t.Errorf("expected github_owner 'MyOrg', got %q", a.ExtractedVars["github_owner"])
	}
	if a.ExtractedVars["github_repo"] != "my-legacy-app" {
		t.Errorf("expected github_repo 'my-legacy-app', got %q", a.ExtractedVars["github_repo"])
	}
}

func TestArchive_MovesDotClaudeMD(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Legacy"), 0644)
	dotClaudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)
	dotClaudeMD := filepath.Join(dotClaudeDir, "CLAUDE.md")
	os.WriteFile(dotClaudeMD, []byte("# Project Instructions"), 0644)

	artifacts := &Artifacts{
		IsLegacy:    true,
		ClaudeMD:    filepath.Join(dir, "CLAUDE.md"),
		DotClaudeMD: dotClaudeMD,
	}

	archived, err := Archive(dir, artifacts)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}
	if len(archived) != 2 {
		t.Errorf("expected 2 archived files, got %d", len(archived))
	}

	// Original should be gone
	if _, err := os.Stat(dotClaudeMD); err == nil {
		t.Error(".claude/CLAUDE.md should have been moved")
	}

	// Should be in docs/legacy/
	dest := filepath.Join(dir, "docs", "legacy", ".claude", "CLAUDE.md")
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("docs/legacy/.claude/CLAUDE.md should exist: %v", err)
	}
}

// parseGitRemote tests are in internal/cli/ package since the function lives there
