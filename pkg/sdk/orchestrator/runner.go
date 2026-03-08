package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client"
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

	for attempt := range maxAttempts {
		if t.fnr != nil {
			r.mu.Lock()
			results := r.results
			r.mu.Unlock()

			result, err = t.fnr(ctx, client, results)
		} else if t.fn != nil {
			result, err = t.fn(ctx, client)
		} else {
			result, err = r.executeOp(ctx, t.op)
		}

		if err == nil {
			break
		}

		if attempt < maxAttempts-1 {
			r.callOnRetry(t, attempt+1, err)
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

// hostResultsFromResponses builds HostResults from the per-agent Responses
// map returned by the API for broadcast jobs.
func hostResultsFromResponses(
	responses map[string]client.AgentJobResponse,
	agentDurations map[string]time.Duration,
) []HostResult {
	results := make([]HostResult, 0, len(responses))

	for hostname, resp := range responses {
		hr := HostResult{
			Hostname: hostname,
		}

		if resp.Changed != nil {
			hr.Changed = *resp.Changed
		}

		if resp.Error != "" {
			hr.Error = resp.Error
		}

		if resp.Data != nil {
			if m, ok := resp.Data.(map[string]any); ok {
				hr.Data = m
			}
		}

		if d, ok := agentDurations[hostname]; ok {
			hr.JobDuration = d
		}

		results = append(results, hr)
	}

	return results
}

// parseAgentDurations returns the longest agent-side execution duration and
// a per-hostname duration map from AgentStates.
func parseAgentDurations(
	states map[string]client.AgentState,
) (time.Duration, map[string]time.Duration) {
	var longest time.Duration

	perHost := make(map[string]time.Duration, len(states))

	for hostname, s := range states {
		if s.Duration == "" {
			continue
		}

		d, err := time.ParseDuration(s.Duration)
		if err != nil {
			continue
		}

		perHost[hostname] = d

		if d > longest {
			longest = d
		}
	}

	return longest, perHost
}

// DefaultPollInterval is the interval between job status polls.
var DefaultPollInterval = 500 * time.Millisecond

// isCommandOp returns true for command execution operations.
func isCommandOp(
	operation string,
) bool {
	return operation == "command.exec.execute" ||
		operation == "command.shell.execute"
}

// extractHostResults parses per-agent results from a broadcast
// collection response, enriching each result with the agent-side
// execution duration when available.
func extractHostResults(
	data map[string]any,
	agentDurations map[string]time.Duration,
) []HostResult {
	resultsRaw, ok := data["results"]
	if !ok {
		return nil
	}

	items, ok := resultsRaw.([]any)
	if !ok {
		return nil
	}

	hostResults := make([]HostResult, 0, len(items))

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		hr := HostResult{
			Data: m,
		}

		if h, ok := m["hostname"].(string); ok {
			hr.Hostname = h

			if d, ok := agentDurations[h]; ok {
				hr.JobDuration = d
			}
		}

		if c, ok := m["changed"].(bool); ok {
			hr.Changed = c
		}

		if e, ok := m["error"].(string); ok {
			hr.Error = e
		}

		hostResults = append(hostResults, hr)
	}

	return hostResults
}

// countExpectedAgents returns the number of agents expected to respond
// for broadcast targets. Returns 0 for non-broadcast targets.
// Mirrors the server-side CountExpectedAgents logic: excludes agents
// in Cordoned or Draining state, and uses exact-or-dotted-prefix
// matching for label selectors.
func (r *runner) countExpectedAgents(
	ctx context.Context,
	target string,
) int {
	if !IsBroadcastTarget(target) {
		return 0
	}

	resp, err := r.plan.client.Agent.List(ctx)
	if err != nil {
		return 0
	}

	if target == "_all" {
		count := 0
		for _, agent := range resp.Data.Agents {
			if agent.State != "Cordoned" && agent.State != "Draining" {
				count++
			}
		}

		return count
	}

	// Label selector: "key:value" — count agents whose label value
	// equals or is a dotted-prefix of the target value.
	// IsBroadcastTarget already confirmed ":" exists, so SplitN
	// always produces exactly 2 parts.
	parts := strings.SplitN(target, ":", 2)
	key, value := parts[0], parts[1]
	count := 0

	for _, agent := range resp.Data.Agents {
		if agent.State == "Cordoned" || agent.State == "Draining" {
			continue
		}

		if agentVal, ok := agent.Labels[key]; ok {
			if agentVal == value || strings.HasPrefix(agentVal, value+".") {
				count++
			}
		}
	}

	return count
}

// executeOp submits a declarative Op as a job via the SDK and polls
// for completion.
func (r *runner) executeOp(
	ctx context.Context,
	op *Op,
) (*Result, error) {
	client := r.plan.client
	if client == nil {
		return nil, fmt.Errorf(
			"op task %q requires an OSAPI client",
			op.Operation,
		)
	}

	// For broadcast targets, determine how many agents should respond
	// before we consider the job complete.
	expectedAgents := r.countExpectedAgents(ctx, op.Target)

	operation := map[string]interface{}{
		"type": op.Operation,
	}

	if len(op.Params) > 0 {
		operation["data"] = op.Params
	}

	createResp, err := client.Job.Create(ctx, operation, op.Target)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	jobID := createResp.Data.JobID

	result, err := r.pollJob(ctx, jobID, expectedAgents)
	if err != nil {
		return nil, err
	}

	// Only populate per-host results for broadcast targets.
	if IsBroadcastTarget(op.Target) {
		// Prefer HostResults from Responses (populated by pollJob);
		// fall back to parsing an embedded results array in Data.
		if len(result.HostResults) == 0 {
			result.HostResults = extractHostResults(result.Data, result.agentDurations)
		}
	} else {
		// Non-broadcast targets should not show per-host breakdown.
		result.HostResults = nil
	}

	// Non-zero exit for command operations = failure.
	if isCommandOp(op.Operation) {
		if IsBroadcastTarget(op.Target) {
			// Check per-host exit codes for broadcast commands.
			anyFailed := false

			for i, hr := range result.HostResults {
				if ec, ok := hr.Data["exit_code"].(float64); ok && ec != 0 {
					result.HostResults[i].Error = fmt.Sprintf(
						"command exited with code %d",
						int(ec),
					)

					anyFailed = true
				}
			}

			if anyFailed {
				result.Status = StatusFailed

				return result, fmt.Errorf("command failed on one or more hosts")
			}
		} else if exitCode, ok := result.Data["exit_code"].(float64); ok && exitCode != 0 {
			result.Status = StatusFailed

			return result, fmt.Errorf(
				"command exited with code %d",
				int(exitCode),
			)
		}
	}

	return result, nil
}

// pollJob polls a job until it reaches a terminal state.
// When expectedAgents > 0 (broadcast), the poller waits until the
// expected number of per-agent responses have arrived before returning,
// even if the server already reports "completed".
func (r *runner) pollJob(
	ctx context.Context,
	jobID string,
	expectedAgents int,
) (*Result, error) {
	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := r.plan.client.Job.Get(ctx, jobID)
			if err != nil {
				return nil, fmt.Errorf("poll job %s: %w", jobID, err)
			}

			job := resp.Data

			switch job.Status {
			case "completed":
				// For broadcast jobs, keep polling until all expected
				// agents have written their responses.
				if expectedAgents > 0 && len(job.Responses) < expectedAgents {
					continue
				}

				data := make(map[string]any)
				if job.Result != nil {
					if m, ok := job.Result.(map[string]any); ok {
						data = m
					}
				}

				changed := false
				if job.Changed != nil {
					changed = *job.Changed
				}
				delete(data, "changed")

				jobDur, perHost := parseAgentDurations(job.AgentStates)

				result := &Result{
					JobID:          jobID,
					Changed:        changed,
					Data:           data,
					JobDuration:    jobDur,
					agentDurations: perHost,
				}

				// Build HostResults from per-agent Responses.
				if len(job.Responses) > 0 {
					result.HostResults = hostResultsFromResponses(
						job.Responses,
						perHost,
					)
				}

				return result, nil
			case "failed":
				errMsg := "job failed"
				if job.Error != "" {
					errMsg = job.Error
				}

				return nil, fmt.Errorf("job %s: %s", jobID, errMsg)
			}
		}
	}
}

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
