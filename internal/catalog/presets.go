package catalog

import "github.com/Intelliaa/inteliside-cli/internal/model"

// AllPresets returns all role-based presets.
func AllPresets() []model.Preset {
	return []model.Preset{
		{
			ID:          "pm",
			Name:        "PM",
			Description: "Product Manager — levantamiento de requerimientos",
			PluginIDs:   []string{"sdd-wizards"},
		},
		{
			ID:          "designer",
			Name:        "Designer",
			Description: "Diseñador UI/UX — investigación y diseño en Figma",
			PluginIDs:   []string{"ux-studio", "sdd-wizards"},
		},
		{
			ID:          "dev",
			Name:        "Dev",
			Description: "Desarrollador — ejecución de features con TDD",
			PluginIDs:   []string{"atl-inteliside", "sdd-wizards"},
		},
		{
			ID:          "fullstack",
			Name:        "Full Stack",
			Description: "Todo el equipo — los 7 plugins del marketplace",
			PluginIDs:   []string{"sdd-wizards", "ux-studio", "atl-inteliside", "sdd-intake", "sdd-legacy", "n8n-studio", "sales-engine"},
		},
		{
			ID:          "automation",
			Name:        "Automation",
			Description: "Automatización — workflows n8n guiados",
			PluginIDs:   []string{"n8n-studio"},
		},
		{
			ID:          "legacy",
			Name:        "Legacy",
			Description: "Legacy onboarding + desarrollo",
			PluginIDs:   []string{"sdd-legacy", "atl-inteliside"},
		},
		{
			ID:          "sales",
			Name:        "Sales",
			Description: "Ventas B2B — prospección, outreach y cierre",
			PluginIDs:   []string{"sales-engine"},
		},
		{
			ID:          "custom",
			Name:        "Custom",
			Description: "Selección manual de plugins",
			PluginIDs:   nil, // populated at runtime via TUI
		},
	}
}

// PresetByID returns a preset by its ID, or nil if not found.
func PresetByID(id string) *model.Preset {
	for _, p := range AllPresets() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}
