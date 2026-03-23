package deps

import (
	"testing"
)

func TestResolve_SinglePlugin(t *testing.T) {
	steps, err := Resolve([]string{"n8n-studio"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) < 2 {
		t.Errorf("expected at least 2 steps, got %d", len(steps))
	}
	// Last step should be marketplace-register
	last := steps[len(steps)-1]
	if last.ID != "marketplace-register" {
		t.Errorf("last step should be marketplace-register, got %s", last.ID)
	}
}

func TestResolve_DevPreset(t *testing.T) {
	steps, err := Resolve([]string{"atl-inteliside", "sdd-wizards"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have deps + 1 marketplace step
	hasMarketplace := false
	for _, s := range steps {
		if s.ID == "marketplace-register" {
			hasMarketplace = true
		}
	}
	if !hasMarketplace {
		t.Error("expected marketplace-register step")
	}
}

func TestResolve_UnknownPlugin(t *testing.T) {
	_, err := Resolve([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown plugin")
	}
}

func TestResolve_MarketplaceIsLast(t *testing.T) {
	steps, err := Resolve([]string{"atl-inteliside"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last := steps[len(steps)-1]
	if last.ID != "marketplace-register" {
		t.Errorf("marketplace-register should be last, got %s", last.ID)
	}

	// All other steps should come before marketplace
	for i, s := range steps[:len(steps)-1] {
		if s.ID == "marketplace-register" {
			t.Errorf("marketplace-register at index %d should only be at the end", i)
		}
	}
}

func TestResolve_NoDuplicateDeps(t *testing.T) {
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
