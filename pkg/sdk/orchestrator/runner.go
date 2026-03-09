package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// runner executes a validated plan.
type runner struct {
	plan    *Plan
	results Results
	failed  map[string]bool
	mu      sync.Mutex
}

// newRunner creates a runner for the plan.
func newRunner(
	plan *Plan,
) *runner {
	return &runner{
		plan:    plan,
		results: make(Results),
		failed:  make(map[string]bool),
	}
}

// run executes the plan by levelizing the DAG and running each
// level in parallel.
func (r *runner) run(
	ctx context.Context,
) (*Report, error) {
	start := time.Now()
	levels := levelize(r.plan.tasks)

	r.callBeforePlan(buildPlanSummary(r.plan.tasks, levels))

	var taskResults []TaskResult

	for i, level := range levels {
		r.callBeforeLevel(i, level, len(level) > 1)

		results, err := r.runLevel(ctx, level)
		taskResults = append(taskResults, results...)

		r.callAfterLevel(i, results)

		if err != nil {
			report := &Report{
				Tasks:    taskResults,
				Duration: time.Since(start),
			}

			r.callAfterPlan(report)

			return report, err
		}
	}

	report := &Report{
		Tasks:    taskResults,
		Duration: time.Since(start),
	}

	r.callAfterPlan(report)

	return report, nil
}

// hook returns the plan's hooks or nil.
func (r *runner) hook() *Hooks {
	return r.plan.config.Hooks
}

// callBeforePlan invokes the BeforePlan hook if set.
func (r *runner) callBeforePlan(
	summary PlanSummary,
) {
	if h := r.hook(); h != nil && h.BeforePlan != nil {
		h.BeforePlan(summary)
	}
}

// buildPlanSummary creates a PlanSummary from tasks and levels.
func buildPlanSummary(
	tasks []*Task,
	levels [][]*Task,
) PlanSummary {
	steps := make([]StepSummary, len(levels))
	for i, level := range levels {
		names := make([]string, len(level))
		for j, t := range level {
			names[j] = t.name
		}

		steps[i] = StepSummary{
			Tasks:    names,
			Parallel: len(level) > 1,
		}
	}

	return PlanSummary{
		TotalTasks: len(tasks),
		Steps:      steps,
	}
}

// callAfterPlan invokes the AfterPlan hook if set.
func (r *runner) callAfterPlan(
	report *Report,
) {
	if h := r.hook(); h != nil && h.AfterPlan != nil {
		h.AfterPlan(report)
	}
}

// callBeforeLevel invokes the BeforeLevel hook if set.
func (r *runner) callBeforeLevel(
	level int,
	tasks []*Task,
	parallel bool,
) {
	if h := r.hook(); h != nil && h.BeforeLevel != nil {
		h.BeforeLevel(level, tasks, parallel)
	}
}

// callAfterLevel invokes the AfterLevel hook if set.
func (r *runner) callAfterLevel(
	level int,
	results []TaskResult,
) {
	if h := r.hook(); h != nil && h.AfterLevel != nil {
		h.AfterLevel(level, results)
	}
}

// callBeforeTask invokes the BeforeTask hook if set.
func (r *runner) callBeforeTask(
	task *Task,
) {
	if h := r.hook(); h != nil && h.BeforeTask != nil {
		h.BeforeTask(task)
	}
}

// callAfterTask invokes the AfterTask hook if set.
func (r *runner) callAfterTask(
	task *Task,
	result TaskResult,
) {
	if h := r.hook(); h != nil && h.AfterTask != nil {
		h.AfterTask(task, result)
	}
}

// callOnRetry invokes the OnRetry hook if set.
func (r *runner) callOnRetry(
	task *Task,
	attempt int,
	err error,
) {
	if h := r.hook(); h != nil && h.OnRetry != nil {
		h.OnRetry(task, attempt, err)
	}
}

// callOnSkip invokes the OnSkip hook if set.
func (r *runner) callOnSkip(
	task *Task,
	reason string,
) {
	if h := r.hook(); h != nil && h.OnSkip != nil {
		h.OnSkip(task, reason)
	}
}

// effectiveStrategy returns the error strategy for a task,
// checking the per-task override before falling back to the
// plan-level default.
func (r *runner) effectiveStrategy(
	t *Task,
) ErrorStrategy {
	if t.errorStrategy != nil {
		return *t.errorStrategy
	}

	return r.plan.config.OnErrorStrategy
}

// runLevel executes all tasks in a level concurrently.
func (r *runner) runLevel(
	ctx context.Context,
	tasks []*Task,
) ([]TaskResult, error) {
	results := make([]TaskResult, len(tasks))

	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)

		go func() {
			defer wg.Done()

			results[i] = r.runTask(ctx, t)
		}()
	}

	wg.Wait()

	for i, tr := range results {
		if tr.Status == StatusFailed {
			strategy := r.effectiveStrategy(tasks[i])
			if strategy.kind != "continue" {
				return results, tr.Error
			}
		}
	}

	return results, nil
}

// runTask executes a single task with guard checks.
func (r *runner) runTask(
	ctx context.Context,
	t *Task,
) TaskResult {
	start := time.Now()

	// Skip if any dependency failed — unless the task has a When guard,
	// which may intentionally inspect failure status (e.g. alert-on-failure).
	if t.guard == nil {
		r.mu.Lock()

		for _, dep := range t.deps {
			if r.failed[dep.name] {
				r.failed[t.name] = true
				r.results[t.name] = &Result{Status: StatusSkipped}
				r.mu.Unlock()

				tr := TaskResult{
					Name:     t.name,
					Status:   StatusSkipped,
					Duration: time.Since(start),
				}
				r.callOnSkip(t, "dependency failed")
				r.callAfterTask(t, tr)

				return tr
			}
		}

		r.mu.Unlock()
	}

	if t.requiresChange {
		anyChanged := false

		r.mu.Lock()

		for _, dep := range t.deps {
			if res := r.results.Get(dep.name); res != nil && res.Changed {
				anyChanged = true

				break
			}
		}

		r.mu.Unlock()

		if !anyChanged {
			r.mu.Lock()
			r.results[t.name] = &Result{Status: StatusSkipped}
			r.mu.Unlock()

			tr := TaskResult{
				Name:     t.name,
				Status:   StatusSkipped,
				Duration: time.Since(start),
			}

			r.callOnSkip(t, "no dependencies changed")
			r.callAfterTask(t, tr)

			return tr
		}
	}

	if t.guard != nil {
		r.mu.Lock()
		shouldRun := t.guard(r.results)
		r.mu.Unlock()

		if !shouldRun {
			r.mu.Lock()
			r.results[t.name] = &Result{Status: StatusSkipped}
			r.mu.Unlock()

			tr := TaskResult{
				Name:     t.name,
				Status:   StatusSkipped,
				Duration: time.Since(start),
			}

			reason := "guard returned false"
			if t.guardReason != "" {
				reason = t.guardReason
			}
			r.callOnSkip(t, reason)
			r.callAfterTask(t, tr)

			return tr
		}
	}

	r.callBeforeTask(t)

	strategy := r.effectiveStrategy(t)
	maxAttempts := 1

	if strategy.kind == "retry" {
		maxAttempts = strategy.retryCount + 1
	}

	var result *Result
	var err error

	client := r.plan.client

retryLoop:
	for attempt := range maxAttempts {
		if t.fnr != nil {
			r.mu.Lock()
			results := r.results
			r.mu.Unlock()

			result, err = t.fnr(ctx, client, results)
		} else if t.fn != nil {
			result, err = t.fn(ctx, client)
		} else {
			result = nil
			err = fmt.Errorf("task %q: Op-based tasks are no longer supported; use TaskFunc instead", t.name)
		}

		if err == nil {
			break
		}

		if attempt < maxAttempts-1 {
			r.callOnRetry(t, attempt+1, err)

			// Backoff between retries if configured.
			if strategy.initialInterval > 0 {
				delay := strategy.backoffDelay(attempt)

				select {
				case <-ctx.Done():
					err = ctx.Err()

					break retryLoop
				case <-time.After(delay):
				}
			}
		}
	}

	elapsed := time.Since(start)

	if err != nil {
		// Preserve the full result for partial failures (e.g. broadcast
		// commands where some hosts succeeded and others failed).
		// Guards like OnlyIfChanged and OnlyIfFailed inspect r.results,
		// so they need Changed, HostResults, etc.
		failedResult := &Result{Status: StatusFailed}
		if result != nil {
			result.Status = StatusFailed
			failedResult = result
		}

		r.mu.Lock()
		r.failed[t.name] = true
		r.results[t.name] = failedResult
		r.mu.Unlock()

		tr := TaskResult{
			Name:     t.name,
			Status:   StatusFailed,
			Duration: elapsed,
			Error:    err,
		}

		if result != nil {
			tr.JobID = result.JobID
			tr.JobDuration = result.JobDuration
			tr.Data = result.Data
			tr.HostResults = result.HostResults
			tr.Changed = result.Changed
		}

		r.callAfterTask(t, tr)

		return tr
	}

	status := StatusUnchanged
	if result.Changed {
		status = StatusChanged
	}

	result.Status = status

	r.mu.Lock()
	r.results[t.name] = result
	r.mu.Unlock()

	tr := TaskResult{
		JobID:       result.JobID,
		Name:        t.name,
		Status:      status,
		Changed:     result.Changed,
		Duration:    elapsed,
		JobDuration: result.JobDuration,
		Data:        result.Data,
		HostResults: result.HostResults,
	}

	r.callAfterTask(t, tr)

	return tr
}

// DefaultPollInterval is the default interval between job status polls.
// Deprecated: kept for backward compatibility; will be removed in a future release.
var DefaultPollInterval = 500 * time.Millisecond

// levelize groups tasks into levels where all tasks in a level can
// run concurrently (all dependencies are in earlier levels).
func levelize(
	tasks []*Task,
) [][]*Task {
	level := make(map[string]int, len(tasks))

	var computeLevel func(t *Task) int
	computeLevel = func(t *Task) int {
		if l, ok := level[t.name]; ok {
			return l
		}

		maxDep := -1

		for _, dep := range t.deps {
			depLevel := computeLevel(dep)
			if depLevel > maxDep {
				maxDep = depLevel
			}
		}

		level[t.name] = maxDep + 1

		return maxDep + 1
	}

	maxLevel := 0

	for _, t := range tasks {
		l := computeLevel(t)
		if l > maxLevel {
			maxLevel = l
		}
	}

	levels := make([][]*Task, maxLevel+1)

	for _, t := range tasks {
		l := level[t.name]
		levels[l] = append(levels[l], t)
	}

	return levels
}
