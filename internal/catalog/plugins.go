package catalog

import "github.com/Intelliaa/inteliside-cli/internal/model"

// AllPlugins returns the full catalog of marketplace plugins.
func AllPlugins() []model.Plugin {
	return []model.Plugin{
		{
			ID:          "sdd-wizards",
			Name:        "SDD-Wizards",
			Description: "Flujo guiado de levantamiento de requerimientos (PM). PRD → Feature Spec → GitHub Projects.",
			Role:        model.RolePM,
			Deps:        []string{"gh-cli", "gh-auth", "gh-repo-scope"},
		},
		{
			ID:          "ux-studio",
			Name:        "UX Studio",
			Description: "Pipeline de diseño UI/UX. Genera diseños via Stitch, importa a Figma y refina.",
			Role:        model.RoleDesigner,
			Deps:        []string{"gh-cli", "gh-auth", "gh-repo-scope", "node-runtime", "figma-mcp"},
		},
		{
			ID:          "atl-inteliside",
			Name:        "ATL Inteliside",
			Description: "Motor de ejecución de features con 5 subagentes, TDD y verificación automática.",
			Role:        model.RoleDev,
			Deps:        []string{"gh-cli", "gh-auth", "gh-repo-scope", "engram-binary", "engram-plugin"},
		},
		{
			ID:          "sdd-intake",
			Name:        "SDD-Intake",
			Description: "Procesamiento de requerimientos de clientes existentes → Feature Specs enriquecidos.",
			Role:        model.RolePM,
			Deps:        []string{"gh-cli", "gh-auth", "gh-repo-scope", "engram-binary", "engram-plugin"},
		},
		{
			ID:          "sdd-legacy",
			Name:        "SDD-Legacy",
			Description: "Onboarding de proyectos legacy al flujo SDD. Audita, extrae features y reglas de negocio.",
			Role:        model.RoleDev,
			Deps:        []string{"gh-cli", "gh-auth", "gh-repo-scope", "engram-binary", "engram-plugin"},
		},
		{
			ID:          "n8n-studio",
			Name:        "n8n Studio",
			Description: "Automatización guiada de workflows n8n. Wizard → Build → Test → Deploy.",
			Role:        model.RoleAutomation,
			Deps:        []string{"n8n-mcp"},
		},
		{
			ID:          "sales-engine",
			Name:        "Sales Engine",
			Description: "Motor de ventas B2B automatizado. Coach diario, research, outreach, propuestas y métricas.",
			Role:        model.RoleSales,
			Deps:        []string{"engram-binary", "engram-plugin"},
		},
	}
}

// PluginByID returns a plugin by its ID, or nil if not found.
func PluginByID(id string) *model.Plugin {
	for _, p := range AllPlugins() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// PluginIDs returns all plugin IDs.
func PluginIDs() []string {
	plugins := AllPlugins()
	ids := make([]string, len(plugins))
	for i, p := range plugins {
		ids[i] = p.ID
	}
	return ids
}
