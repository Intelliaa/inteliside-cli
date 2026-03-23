package model

// Dependency represents an external dependency that a plugin requires.
type Dependency struct {
	ID          string
	Name        string
	Description string
	Requires    []string // IDs of other dependencies this one depends on
	CheckFn     CheckFunc
	InstallFn   InstallFunc
}

// CheckFunc verifies if a dependency is already satisfied.
// Returns true if the dependency is met.
type CheckFunc func() (bool, string, error)

// InstallFunc installs or configures a dependency.
// Returns an error if installation fails.
type InstallFunc func(ctx *InstallContext) error

// InstallContext holds state shared across installation steps.
type InstallContext struct {
	ProjectDir string
	DryRun     bool
	Verbose    bool
	AutoYes    bool
	Secrets    map[string]string // collected tokens/keys
}

// NewInstallContext creates a new install context with defaults.
func NewInstallContext(projectDir string) *InstallContext {
	return &InstallContext{
		ProjectDir: projectDir,
		Secrets:    make(map[string]string),
	}
}
