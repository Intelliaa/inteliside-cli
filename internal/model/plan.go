package model

// Step represents a single step in the installation plan.
type Step struct {
	ID          string
	Name        string
	Description string
	Status      StepStatus
	DependsOn   []string
	Execute     func(ctx *InstallContext) error
	Rollback    func() error
}

// StepStatus tracks the state of a step.
type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepDone      StepStatus = "done"
	StepSkipped   StepStatus = "skipped"
	StepFailed    StepStatus = "failed"
)

// Plan is an ordered list of steps to execute.
type Plan struct {
	Steps []Step
}
