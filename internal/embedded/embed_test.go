package embedded

import (
	"strings"
	"testing"
)

func TestRuleFiles_ReturnsAllFive(t *testing.T) {
	rules, err := RuleFiles()
	if err != nil {
		t.Fatalf("RuleFiles() error: %v", err)
	}
	if len(rules) != 5 {
		t.Errorf("expected 5 rule files, got %d", len(rules))
	}

	expected := []string{
		"atl-workflow.md",
		"context-monitoring.md",
		"engram-protocol.md",
		"subagent-architecture.md",
		"team-rules.md",
	}
	for _, name := range expected {
		content, ok := rules[name]
		if !ok {
			t.Errorf("missing rule file: %s", name)
			continue
		}
		if len(content) < 50 {
			t.Errorf("rule file %s seems too short (%d bytes) — might be a dummy", name, len(content))
		}
	}
}

func TestRuleFileNames(t *testing.T) {
	names := RuleFileNames()
	if len(names) != 5 {
		t.Errorf("expected 5 rule file names, got %d: %v", len(names), names)
	}
}

func TestATLTemplate_HasPlaceholders(t *testing.T) {
	tmpl := ATLTemplate()
	if tmpl == "" {
		t.Fatal("ATLTemplate() returned empty string")
	}

	placeholders := []string{
		"{{project_name}}",
		"{{engram_project}}",
		"{{github_owner}}",
		"{{github_repo}}",
	}
	for _, p := range placeholders {
		if !strings.Contains(tmpl, p) {
			t.Errorf("ATLTemplate missing placeholder: %s", p)
		}
	}
}

func TestATLTemplate_HasMarkers(t *testing.T) {
	tmpl := ATLTemplate()
	if !strings.Contains(tmpl, "<!-- inteliside:atl-config -->") {
		t.Error("ATLTemplate missing opening ATL marker")
	}
	if !strings.Contains(tmpl, "<!-- /inteliside:atl-config -->") {
		t.Error("ATLTemplate missing closing ATL marker")
	}
}

func TestLegacyTemplate_NotEmpty(t *testing.T) {
	tmpl := LegacyTemplate()
	if tmpl == "" {
		t.Fatal("LegacyTemplate() returned empty string")
	}
	if !strings.Contains(tmpl, "{{project_name}}") {
		t.Error("LegacyTemplate missing project_name placeholder")
	}
}

func TestRenderTemplate_Substitutes(t *testing.T) {
	tmpl := "Hello {{name}}, welcome to {{place}}"
	vars := map[string]string{
		"name":  "Raul",
		"place": "Inteliside",
	}
	result := RenderTemplate(tmpl, vars)
	expected := "Hello Raul, welcome to Inteliside"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestRenderTemplate_UnresolvedBecomeTODO(t *testing.T) {
	tmpl := "Stack: {{project_stack}}"
	vars := map[string]string{} // no vars provided
	result := RenderTemplate(tmpl, vars)
	if !strings.Contains(result, "<!-- TODO: project_stack -->") {
		t.Errorf("unresolved placeholder not converted to TODO: %q", result)
	}
}

func TestRenderTemplate_EmptyVarBecomesTodo(t *testing.T) {
	tmpl := "Owner: {{github_owner}}"
	vars := map[string]string{"github_owner": ""}
	result := RenderTemplate(tmpl, vars)
	if !strings.Contains(result, "<!-- TODO: github_owner -->") {
		t.Errorf("empty var should become TODO: %q", result)
	}
}

func TestCountTODOs(t *testing.T) {
	content := "Hello <!-- TODO: name --> and <!-- TODO: place --> end"
	count := CountTODOs(content)
	if count != 2 {
		t.Errorf("expected 2 TODOs, got %d", count)
	}
}

func TestCountTODOs_None(t *testing.T) {
	content := "Hello world, no placeholders here"
	count := CountTODOs(content)
	if count != 0 {
		t.Errorf("expected 0 TODOs, got %d", count)
	}
}
