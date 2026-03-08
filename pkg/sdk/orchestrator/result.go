// Package orchestrator provides DAG-based task orchestration primitives.
package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

// Status represents the outcome of a task execution.
type Status string

// Task execution statuses.
const (
	// StatusPending and StatusRunning are reserved for future
	// streaming status support. The runner does not currently
	// assign these — tasks go directly to a terminal status.
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusChanged   Status = "changed"
	StatusUnchanged Status = "unchanged"
	StatusSkipped   Status = "skipped"
	StatusFailed    Status = "failed"
)

// HostResult represents a single host's response within a broadcast
// operation.
type HostResult struct {
	Hostname    string
	Changed     bool
	Error       string
	Data        map[string]any
	JobDuration time.Duration
}

// Result is the outcome of a single task execution.
type Result struct {
	JobID          string
	Changed        bool
	Data           map[string]any
	Status         Status
	JobDuration    time.Duration
	HostResults    []HostResult
	agentDurations map[string]time.Duration // per-host durations from AgentStates
}

// TaskResult records the full execution details of a task.
type TaskResult struct {
	JobID       string
	Name        string
	Status      Status
	Changed     bool
	Duration    time.Duration
	JobDuration time.Duration
	Error       error
	Data        map[string]any
	HostResults []HostResult
}

// Results is a map of task name to Result, used for conditional logic.
type Results map[string]*Result

// Get returns the Result for the named task, or nil if not found.
func (r Results) Get(
	name string,
) *Result {
	return r[name]
}

// StepSummary describes a single execution step (DAG level).
type StepSummary struct {
	Tasks    []string
	Parallel bool
}

// PlanSummary describes the execution plan before it runs.
type PlanSummary struct {
	TotalTasks int
	Steps      []StepSummary
}

// Report is the aggregate output of a plan execution.
type Report struct {
	Tasks    []TaskResult
	Duration time.Duration
}

// Summary returns a human-readable summary of the report.
func (r *Report) Summary() string {
	var changed, unchanged, skipped, failed int

	for _, t := range r.Tasks {
		switch t.Status {
		case StatusChanged:
			changed++
		case StatusUnchanged:
			unchanged++
		case StatusSkipped:
			skipped++
		case StatusFailed:
			failed++
		}
	}

	parts := []string{
		fmt.Sprintf("%d tasks", len(r.Tasks)),
	}

	if changed > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", changed))
	}

	if unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", unchanged))
	}

	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}

	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}

	return strings.Join(parts, ", ")
}
