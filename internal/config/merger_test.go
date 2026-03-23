package config

import (
	"testing"
)

func TestDeepMerge_NewKeysAdded(t *testing.T) {
	dst := map[string]any{
		"existing": "value",
	}
	src := map[string]any{
		"new_key": "new_value",
	}
	result := DeepMerge(dst, src)

	if result["existing"] != "value" {
		t.Error("existing key was modified")
	}
	if result["new_key"] != "new_value" {
		t.Error("new key was not added")
	}
}

func TestDeepMerge_ScalarsNeverOverwritten(t *testing.T) {
	dst := map[string]any{
		"name": "original",
	}
	src := map[string]any{
		"name": "overwritten",
	}
	result := DeepMerge(dst, src)

	if result["name"] != "original" {
		t.Errorf("scalar was overwritten: got %v, want 'original'", result["name"])
	}
}

func TestDeepMerge_NestedObjectsMerged(t *testing.T) {
	dst := map[string]any{
		"mcpServers": map[string]any{
			"existing-server": map[string]any{
				"url": "https://existing.com",
			},
		},
	}
	src := map[string]any{
		"mcpServers": map[string]any{
			"new-server": map[string]any{
				"url": "https://new.com",
			},
		},
	}
	result := DeepMerge(dst, src)

	servers := result["mcpServers"].(map[string]any)
	if _, ok := servers["existing-server"]; !ok {
		t.Error("existing server was removed")
	}
	if _, ok := servers["new-server"]; !ok {
		t.Error("new server was not added")
	}
}

func TestDeepMerge_ArraysUnioned(t *testing.T) {
	dst := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Read", "Write"},
		},
	}
	src := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Write", "Bash(git:*)"},
		},
	}
	result := DeepMerge(dst, src)

	perms := result["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	if len(allow) != 3 {
		t.Errorf("expected 3 permissions, got %d: %v", len(allow), allow)
	}
}

func TestDeepMerge_NilDstCreatesNew(t *testing.T) {
	src := map[string]any{
		"key": "value",
	}
	result := DeepMerge(nil, src)

	if result["key"] != "value" {
		t.Error("nil dst should create new map with src values")
	}
}

func TestDeepMerge_NestedScalarsPreserved(t *testing.T) {
	dst := map[string]any{
		"mcpServers": map[string]any{
			"figma-console": map[string]any{
				"env": map[string]any{
					"FIGMA_ACCESS_TOKEN": "user-token-123",
				},
			},
		},
	}
	src := map[string]any{
		"mcpServers": map[string]any{
			"figma-console": map[string]any{
				"env": map[string]any{
					"FIGMA_ACCESS_TOKEN": "overwrite-attempt",
				},
			},
		},
	}
	result := DeepMerge(dst, src)

	servers := result["mcpServers"].(map[string]any)
	figma := servers["figma-console"].(map[string]any)
	env := figma["env"].(map[string]any)
	if env["FIGMA_ACCESS_TOKEN"] != "user-token-123" {
		t.Error("nested scalar was overwritten")
	}
}
