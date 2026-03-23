package model

// Preset represents a role-based collection of plugins.
type Preset struct {
	ID          string
	Name        string
	Description string
	PluginIDs   []string
}
