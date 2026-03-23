package catalog

import (
	"testing"
)

func TestAllPlugins_Returns7(t *testing.T) {
	plugins := AllPlugins()
	if len(plugins) != 7 {
		t.Errorf("expected 7 plugins, got %d", len(plugins))
	}
}

func TestPluginByID_Found(t *testing.T) {
	p := PluginByID("atl-inteliside")
	if p == nil {
		t.Fatal("expected to find atl-inteliside")
	}
	if p.Name != "ATL Inteliside" {
		t.Errorf("unexpected name: %s", p.Name)
	}
}

func TestPluginByID_NotFound(t *testing.T) {
	p := PluginByID("nonexistent")
	if p != nil {
		t.Error("expected nil for nonexistent plugin")
	}
}

func TestAllPresets_Returns8(t *testing.T) {
	presets := AllPresets()
	if len(presets) != 8 {
		t.Errorf("expected 8 presets, got %d", len(presets))
	}
}

func TestPresetByID_DevHasCorrectPlugins(t *testing.T) {
	p := PresetByID("dev")
	if p == nil {
		t.Fatal("expected to find dev preset")
	}
	if len(p.PluginIDs) != 2 {
		t.Errorf("dev preset should have 2 plugins, got %d", len(p.PluginIDs))
	}
	hasATL := false
	hasSDD := false
	for _, id := range p.PluginIDs {
		if id == "atl-inteliside" {
			hasATL = true
		}
		if id == "sdd-wizards" {
			hasSDD = true
		}
	}
	if !hasATL || !hasSDD {
		t.Errorf("dev preset should include atl-inteliside and sdd-wizards, got %v", p.PluginIDs)
	}
}

func TestPresetByID_FullstackHasAll(t *testing.T) {
	p := PresetByID("fullstack")
	if p == nil {
		t.Fatal("expected to find fullstack preset")
	}
	if len(p.PluginIDs) != 7 {
		t.Errorf("fullstack preset should have 7 plugins, got %d", len(p.PluginIDs))
	}
}

func TestAllDependencies_NotEmpty(t *testing.T) {
	deps := AllDependencies()
	if len(deps) == 0 {
		t.Error("expected dependencies")
	}
}

func TestDependencyByID_Found(t *testing.T) {
	d := DependencyByID("gh-cli")
	if d == nil {
		t.Fatal("expected to find gh-cli dependency")
	}
	if d.CheckFn == nil {
		t.Error("gh-cli should have a CheckFn")
	}
}

func TestPluginDepsExist(t *testing.T) {
	// Every dep referenced by a plugin should exist in the catalog
	for _, plugin := range AllPlugins() {
		for _, depID := range plugin.Deps {
			d := DependencyByID(depID)
			if d == nil {
				t.Errorf("plugin %s references unknown dep: %s", plugin.ID, depID)
			}
		}
	}
}

func TestPresetPluginsExist(t *testing.T) {
	// Every plugin referenced by a preset should exist in the catalog
	for _, preset := range AllPresets() {
		for _, pluginID := range preset.PluginIDs {
			p := PluginByID(pluginID)
			if p == nil {
				t.Errorf("preset %s references unknown plugin: %s", preset.ID, pluginID)
			}
		}
	}
}
