package orchestrator

import (
	"context"
	"strings"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

// Op represents a declarative SDK operation.
type Op struct {
	Operation string
	Target    string
	Params    map[string]any
}

// TaskFn is the signature for functional tasks. The client
// parameter provides access to the OSAPI SDK for making API calls.
type TaskFn func(
	ctx context.Context,
	client *osapiclient.Client,
) (*Result, error)

// TaskFnWithResults is like TaskFn but receives completed task results
// for inter-task data access.
type TaskFnWithResults func(
	ctx context.Context,
	client *osapiclient.Client,
	results Results,
) (*Result, error)

// GuardFn is a predicate that determines if a task should run.
type GuardFn func(results Results) bool

// Task is a unit of work in an orchestration plan.
type Task struct {
	name           string
	op             *Op
	fn             TaskFn
	fnr            TaskFnWithResults
	deps           []*Task
	guard          GuardFn
	guardReason    string
	requiresChange bool
	errorStrategy  *ErrorStrategy
}

// NewTask creates a declarative task wrapping an SDK operation.
func NewTask(
	name string,
	op *Op,
) *Task {
	return &Task{
		name: name,
		op:   op,
	}
}

// NewTaskFunc creates a functional task with custom logic.
func NewTaskFunc(
	name string,
	fn TaskFn,
) *Task {
	return &Task{
		name: name,
		fn:   fn,
	}
}

// NewTaskFuncWithResults creates a functional task that receives
// completed results from prior tasks.
func NewTaskFuncWithResults(
	name string,
	fn TaskFnWithResults,
) *Task {
	return &Task{
		name: name,
		fnr:  fn,
	}
}

// Name returns the task name.
func (t *Task) Name() string {
	return t.name
}

// SetName changes the task name.
func (t *Task) SetName(
	name string,
) {
	t.name = name
}

// IsFunc returns true if this is a functional task.
func (t *Task) IsFunc() bool {
	return t.fn != nil || t.fnr != nil
}

// Operation returns the declarative operation, or nil for functional
// tasks.
func (t *Task) Operation() *Op {
	return t.op
}

// Fn returns the task function, or nil for declarative tasks.
func (t *Task) Fn() TaskFn {
	return t.fn
}

// DependsOn sets this task's dependencies. Returns the task for
// chaining.
func (t *Task) DependsOn(
	deps ...*Task,
) *Task {
	t.deps = append(t.deps, deps...)

	return t
}

// Dependencies returns the task's dependencies.
func (t *Task) Dependencies() []*Task {
	return t.deps
}

// OnlyIfChanged marks this task to only run if at least one
// dependency reported Changed=true.
func (t *Task) OnlyIfChanged() {
	t.requiresChange = true
}

// RequiresChange returns true if OnlyIfChanged was set.
func (t *Task) RequiresChange() bool {
	return t.requiresChange
}

// When sets a custom guard function that determines whether
// this task should execute.
func (t *Task) When(
	fn GuardFn,
) {
	t.guard = fn
}

// WhenWithReason sets a guard with a custom skip reason shown when
// the guard returns false.
func (t *Task) WhenWithReason(
	fn GuardFn,
	reason string,
) {
	t.guard = fn
	t.guardReason = reason
}

// Guard returns the guard function, or nil if none is set.
func (t *Task) Guard() GuardFn {
	return t.guard
}

// OnError sets a per-task error strategy override.
func (t *Task) OnError(
	strategy ErrorStrategy,
) {
	t.errorStrategy = &strategy
}

// ErrorStrategy returns the per-task error strategy, or nil to
// use the plan default.
func (t *Task) ErrorStrategy() *ErrorStrategy {
	return t.errorStrategy
}

// IsBroadcastTarget returns true if the target addresses multiple
// agents (broadcast or label selector).
func IsBroadcastTarget(
	target string,
) bool {
	return target == "_all" || strings.Contains(target, ":")
}
