package orchestrator

import (
	"context"
	"encoding/json"
	"time"
)

// SetJSONUnmarshalFn overrides the jsonUnmarshalFn injectable for testing.
func SetJSONUnmarshalFn(
	fn func([]byte, any) error,
) {
	jsonUnmarshalFn = fn
}

// ResetJSONUnmarshalFn restores the default jsonUnmarshalFn.
func ResetJSONUnmarshalFn() {
	jsonUnmarshalFn = json.Unmarshal
}

// ExportLevelize exposes the private levelize function for testing.
func ExportLevelize(
	tasks []*Task,
) [][]*Task {
	return levelize(tasks)
}

// ExportRunner wraps a private runner for testing.
type ExportRunner struct {
	r *runner
}

// ExportNewRunner creates a new ExportRunner backed by a private runner.
func ExportNewRunner(
	plan *Plan,
) *ExportRunner {
	return &ExportRunner{r: newRunner(plan)}
}

// Run executes the runner.
func (er *ExportRunner) Run(
	ctx context.Context,
) (*Report, error) {
	return er.r.run(ctx)
}

// GetResult returns the result for the named task.
func (er *ExportRunner) GetResult(
	name string,
) *Result {
	return er.r.results.Get(name)
}

// ExportBackoffDelay exposes the private backoffDelay method on ErrorStrategy
// for testing.
func ExportBackoffDelay(
	initial time.Duration,
	maxInterval time.Duration,
	attempt int,
) time.Duration {
	s := ErrorStrategy{
		kind:            "retry",
		retryCount:      3,
		initialInterval: initial,
		maxInterval:     maxInterval,
	}

	return s.backoffDelay(attempt)
}
