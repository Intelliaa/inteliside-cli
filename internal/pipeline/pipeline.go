package pipeline

import (
	"fmt"
	"time"

	"github.com/Intelliaa/inteliside-cli/internal/model"
)

// Result holds the outcome of running a pipeline.
type Result struct {
	Completed []string
	Skipped   []string
	Failed    string
	Error     error
	Duration  time.Duration
}

// Run executes all steps in order. If a step fails, execution stops.
func Run(steps []model.Step, ctx *model.InstallContext) Result {
	start := time.Now()
	var completed, skipped []string

	for i := range steps {
		step := &steps[i]
		step.Status = model.StepRunning

		if err := step.Execute(ctx); err != nil {
			step.Status = model.StepFailed
			return Result{
				Completed: completed,
				Skipped:   skipped,
				Failed:    step.ID,
				Error:     fmt.Errorf("paso '%s' falló: %w", step.Name, err),
				Duration:  time.Since(start),
			}
		}

		step.Status = model.StepDone
		completed = append(completed, step.ID)
	}

	return Result{
		Completed: completed,
		Skipped:   skipped,
		Duration:  time.Since(start),
	}
}
