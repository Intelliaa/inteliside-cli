package deps

import (
	"fmt"

	"github.com/Intelliaa/inteliside-cli/internal/catalog"
	"github.com/Intelliaa/inteliside-cli/internal/model"
)

// Resolve takes a list of plugin IDs and returns an ordered list of steps
// (dependencies + plugin installs) using topological sort.
func Resolve(pluginIDs []string) ([]model.Step, error) {
	// Collect all unique dependency IDs from selected plugins
	seen := make(map[string]bool)
	var depIDs []string

	for _, pid := range pluginIDs {
		plugin := catalog.PluginByID(pid)
		if plugin == nil {
			return nil, fmt.Errorf("plugin desconocido: %s", pid)
		}
		for _, did := range plugin.Deps {
			if !seen[did] {
				seen[did] = true
				depIDs = append(depIDs, did)
			}
		}
	}

	// Expand transitive dependencies
	allDeps, err := expandDeps(depIDs)
	if err != nil {
		return nil, err
	}

	// Topological sort
	sorted, err := topoSort(allDeps)
	if err != nil {
		return nil, err
	}

	// Convert to steps
	var steps []model.Step
	for _, dep := range sorted {
		d := dep // capture
		steps = append(steps, model.Step{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			Status:      model.StepPending,
			DependsOn:   d.Requires,
			Execute: func(ctx *model.InstallContext) error {
				ok, msg, err := d.CheckFn()
				if err != nil {
					return fmt.Errorf("error verificando %s: %w", d.Name, err)
				}
				if ok {
					fmt.Printf("  ✓ %s — %s\n", d.Name, msg)
					return nil
				}
				fmt.Printf("  → Instalando %s (%s)...\n", d.Name, msg)
				return d.InstallFn(ctx)
			},
		})
	}

	// Register marketplace and activate plugins in Claude Code
	mktDep := catalog.DependencyByID("marketplace-register")
	if mktDep != nil {
		d := *mktDep
		steps = append(steps, model.Step{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			Status:      model.StepPending,
			Execute: func(ctx *model.InstallContext) error {
				// Pass plugin IDs to context so the installer knows which to activate
				ctx.PluginIDs = pluginIDs
				fmt.Printf("  → Registrando marketplace y activando %d plugins...\n", len(pluginIDs))
				return d.InstallFn(ctx)
			},
		})
	}

	return steps, nil
}

// expandDeps recursively collects all transitive dependencies.
func expandDeps(depIDs []string) ([]model.Dependency, error) {
	result := make(map[string]model.Dependency)
	var expand func(ids []string) error

	expand = func(ids []string) error {
		for _, id := range ids {
			if _, exists := result[id]; exists {
				continue
			}
			dep := catalog.DependencyByID(id)
			if dep == nil {
				return fmt.Errorf("dependencia desconocida: %s", id)
			}
			result[id] = *dep
			if len(dep.Requires) > 0 {
				if err := expand(dep.Requires); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := expand(depIDs); err != nil {
		return nil, err
	}

	deps := make([]model.Dependency, 0, len(result))
	for _, d := range result {
		deps = append(deps, d)
	}
	return deps, nil
}

// topoSort performs a topological sort on dependencies.
func topoSort(deps []model.Dependency) ([]model.Dependency, error) {
	byID := make(map[string]model.Dependency)
	for _, d := range deps {
		byID[d.ID] = d
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var order []model.Dependency

	var visit func(id string) error
	visit = func(id string) error {
		if inStack[id] {
			return fmt.Errorf("dependencia circular detectada: %s", id)
		}
		if visited[id] {
			return nil
		}
		inStack[id] = true

		d, ok := byID[id]
		if !ok {
			// Not in our set, skip (external dep already satisfied)
			inStack[id] = false
			return nil
		}

		for _, req := range d.Requires {
			if err := visit(req); err != nil {
				return err
			}
		}

		visited[id] = true
		inStack[id] = false
		order = append(order, d)
		return nil
	}

	for _, d := range deps {
		if err := visit(d.ID); err != nil {
			return nil, err
		}
	}

	return order, nil
}
