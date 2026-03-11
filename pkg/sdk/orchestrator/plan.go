package orchestrator

import (
	"context"
	"fmt"
	"strings"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

// Plan is a DAG of tasks with dependency edges.
type Plan struct {
	client       *osapiclient.Client
	tasks        []*Task
	config       PlanConfig
	dockerExecFn ExecFn
}

// NewPlan creates a new plan bound to an OSAPI client.
func NewPlan(
	client *osapiclient.Client,
	opts ...PlanOption,
) *Plan {
	cfg := PlanConfig{
		OnErrorStrategy: StopAll,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Plan{
		client:       client,
		config:       cfg,
		dockerExecFn: cfg.DockerExecFn,
	}
}

// Client returns the OSAPI client bound to this plan.
func (p *Plan) Client() *osapiclient.Client {
	return p.client
}

// Config returns the plan configuration.
func (p *Plan) Config() PlanConfig {
	return p.config
}

// TaskFunc creates a functional task, adds it to the plan, and
// returns it.
func (p *Plan) TaskFunc(
	name string,
	fn TaskFn,
) *Task {
	t := NewTaskFunc(name, fn)
	p.tasks = append(p.tasks, t)

	return t
}

// TaskFuncWithResults creates a functional task that receives
// completed results from prior tasks, adds it to the plan, and
// returns it.
func (p *Plan) TaskFuncWithResults(
	name string,
	fn TaskFnWithResults,
) *Task {
	t := NewTaskFuncWithResults(name, fn)
	p.tasks = append(p.tasks, t)

	return t
}

// Tasks returns all tasks in the plan.
func (p *Plan) Tasks() []*Task {
	return p.tasks
}

// Explain returns a human-readable representation of the execution
// plan showing levels, parallelism, dependencies, and guards.
func (p *Plan) Explain() string {
	levels, err := p.Levels()
	if err != nil {
		return fmt.Sprintf("invalid plan: %s", err)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Plan: %d tasks, %d levels\n", len(p.tasks), len(levels))

	for i, level := range levels {
		if len(level) > 1 {
			fmt.Fprintf(&b, "\nLevel %d (parallel):\n", i)
		} else {
			fmt.Fprintf(&b, "\nLevel %d:\n", i)
		}

		for _, t := range level {
			fmt.Fprintf(&b, "  %s [fn]", t.name)

			if len(t.deps) > 0 {
				names := make([]string, len(t.deps))
				for j, dep := range t.deps {
					names[j] = dep.name
				}

				fmt.Fprintf(&b, " <- %s", strings.Join(names, ", "))
			}

			var flags []string
			if t.requiresChange {
				flags = append(flags, "only-if-changed")
			}

			if t.guard != nil {
				flags = append(flags, "when")
			}

			if len(flags) > 0 {
				fmt.Fprintf(&b, " (%s)", strings.Join(flags, ", "))
			}

			fmt.Fprintln(&b)
		}
	}

	return b.String()
}

// Levels returns the levelized DAG -- tasks grouped into execution
// levels where all tasks in a level can run concurrently.
// Returns an error if the plan fails validation.
func (p *Plan) Levels() ([][]*Task, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	return levelize(p.tasks), nil
}

// Validate checks the plan for errors: duplicate names and cycles.
func (p *Plan) Validate() error {
	names := make(map[string]bool, len(p.tasks))

	for _, t := range p.tasks {
		if names[t.name] {
			return fmt.Errorf("duplicate task name: %q", t.name)
		}

		names[t.name] = true
	}

	return p.detectCycle()
}

// Run validates the plan, resolves the DAG, and executes tasks.
func (p *Plan) Run(
	ctx context.Context,
) (*Report, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("plan validation: %w", err)
	}

	runner := newRunner(p)

	return runner.run(ctx)
}

// detectCycle uses DFS to find cycles in the dependency graph.
func (p *Plan) detectCycle() error {
	const (
		white = 0 // unvisited
		gray  = 1 // in progress
		black = 2 // done
	)

	color := make(map[string]int, len(p.tasks))

	var visit func(t *Task) error
	visit = func(t *Task) error {
		color[t.name] = gray

		for _, dep := range t.deps {
			switch color[dep.name] {
			case gray:
				return fmt.Errorf(
					"cycle detected: %q depends on %q",
					t.name,
					dep.name,
				)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		color[t.name] = black

		return nil
	}

	for _, t := range p.tasks {
		if color[t.name] == white {
			if err := visit(t); err != nil {
				return err
			}
		}
	}

	return nil
}
