package deps

import (
	"strings"
	"testing"
)

func TestResolve_SinglePlugin(t *testing.T) {
	steps, err := Resolve([]string{"n8n-studio"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// n8n-studio only needs n8n-mcp + the plugin step
	if len(steps) < 2 {
		t.Errorf("expected at least 2 steps, got %d", len(steps))
	}
	// Last step should be the plugin
	last := steps[len(steps)-1]
	if !strings.HasPrefix(last.ID, "plugin:") {
		t.Errorf("last step should be a plugin step, got %s", last.ID)
	}
}

func TestResolve_DevPreset(t *testing.T) {
	steps, err := Resolve([]string{"atl-inteliside", "sdd-wizards"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have deps + 2 plugin steps
	pluginSteps := 0
	for _, s := range steps {
		if strings.HasPrefix(s.ID, "plugin:") {
			pluginSteps++
		}
	}
	if pluginSteps != 2 {
		t.Errorf("expected 2 plugin steps, got %d", pluginSteps)
	}
}

func TestResolve_UnknownPlugin(t *testing.T) {
	_, err := Resolve([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown plugin")
	}
}

func TestResolve_DepsBeforePlugins(t *testing.T) {
	steps, err := Resolve([]string{"atl-inteliside"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pluginIdx := -1
	for i, s := range steps {
		if s.ID == "plugin:atl-inteliside" {
			pluginIdx = i
			break
		}
	}
	if pluginIdx < 0 {
		t.Fatal("plugin step not found")
	}

	// All dep steps should come before the plugin step
	for i, s := range steps[:pluginIdx] {
		if strings.HasPrefix(s.ID, "plugin:") {
			t.Errorf("dep step at index %d is a plugin step: %s", i, s.ID)
		}
	}
}

func TestResolve_NoDuplicateDeps(t *testing.T) {
	// fullstack installs all 6 plugins — deps should be deduplicated
	steps, err := Resolve([]string{"sdd-wizards", "ux-studio", "atl-inteliside", "sdd-intake", "sdd-legacy", "n8n-studio"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]bool)
	for _, s := range steps {
		if seen[s.ID] {
			t.Errorf("duplicate step: %s", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestResolve_TransitiveDeps(t *testing.T) {
	// gh-repo-scope requires gh-auth requires gh-cli
	steps, err := Resolve([]string{"sdd-wizards"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ghCLIIdx := -1
	ghAuthIdx := -1
	ghScopeIdx := -1

	for i, s := range steps {
		switch s.ID {
		case "gh-cli":
			ghCLIIdx = i
		case "gh-auth":
			ghAuthIdx = i
		case "gh-repo-scope":
			ghScopeIdx = i
		}
	}

	if ghCLIIdx < 0 || ghAuthIdx < 0 || ghScopeIdx < 0 {
		t.Fatal("expected gh-cli, gh-auth, gh-repo-scope in steps")
	}

	if ghCLIIdx >= ghAuthIdx {
		t.Error("gh-cli should come before gh-auth")
	}
	if ghAuthIdx >= ghScopeIdx {
		t.Error("gh-auth should come before gh-repo-scope")
	}
}
