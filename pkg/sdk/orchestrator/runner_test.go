package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

type RunnerTestSuite struct {
	suite.Suite
}

func TestRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}

func (s *RunnerTestSuite) TestLevelize() {
	tests := []struct {
		name       string
		setup      func() []*Task
		wantLevels int
	}{
		{
			name: "linear chain has 3 levels",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(b)

				return []*Task{a, b, c}
			},
			wantLevels: 3,
		},
		{
			name: "diamond has 3 levels",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				d := NewTask("d", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(a)
				d.DependsOn(b, c)

				return []*Task{a, b, c, d}
			},
			wantLevels: 3,
		},
		{
			name: "independent tasks in 1 level",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})

				return []*Task{a, b}
			},
			wantLevels: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tasks := tt.setup()
			levels := levelize(tasks)
			s.Len(levels, tt.wantLevels)
		})
	}
}

func (s *RunnerTestSuite) TestRunTaskStoresResultForAllPaths() {
	tests := []struct {
		name       string
		setup      func() *Plan
		taskName   string
		wantStatus Status
	}{
		{
			name: "OnlyIfChanged skip stores StatusSkipped",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				// dep returns Changed=false, so child with
				// OnlyIfChanged should be skipped.
				dep := plan.TaskFunc("dep", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: false}, nil
				})

				child := plan.TaskFunc("child", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: true}, nil
				})
				child.DependsOn(dep)
				child.OnlyIfChanged()

				return plan
			},
			taskName:   "child",
			wantStatus: StatusSkipped,
		},
		{
			name: "failed task stores StatusFailed",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				plan.TaskFunc("failing", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return nil, fmt.Errorf("deliberate error")
				})

				return plan
			},
			taskName:   "failing",
			wantStatus: StatusFailed,
		},
		{
			name: "guard-false skip stores StatusSkipped",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				plan.TaskFunc("guarded", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: true}, nil
				}).When(func(_ Results) bool {
					return false
				})

				return plan
			},
			taskName:   "guarded",
			wantStatus: StatusSkipped,
		},
		{
			name: "dependency-failed skip stores StatusSkipped",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				dep := plan.TaskFunc("dep", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return nil, fmt.Errorf("deliberate error")
				})

				child := plan.TaskFunc("child", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: true}, nil
				})
				child.DependsOn(dep)

				return plan
			},
			taskName:   "child",
			wantStatus: StatusSkipped,
		},
		{
			name: "successful changed task stores StatusChanged",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				plan.TaskFunc("ok", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: true}, nil
				})

				return plan
			},
			taskName:   "ok",
			wantStatus: StatusChanged,
		},
		{
			name: "successful unchanged task stores StatusUnchanged",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(Continue))

				plan.TaskFunc("ok", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: false}, nil
				})

				return plan
			},
			taskName:   "ok",
			wantStatus: StatusUnchanged,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := tt.setup()
			runner := newRunner(plan)

			_, err := runner.run(context.Background())
			// Some plans produce errors (e.g. StopAll with a
			// failing task); we don't assert on err here because
			// we only care about the results map.
			_ = err

			result := runner.results.Get(tt.taskName)
			s.NotNil(
				result,
				"results map should contain entry for %q",
				tt.taskName,
			)
			s.Equal(
				tt.wantStatus,
				result.Status,
				"result status for %q",
				tt.taskName,
			)
		})
	}
}

func (s *RunnerTestSuite) TestDownstreamGuardInspectsSkippedStatus() {
	tests := []struct {
		name            string
		setup           func() (*Plan, *bool)
		observerName    string
		wantGuardCalled bool
		wantTaskStatus  Status
	}{
		{
			name: "guard can see guard-skipped task status",
			setup: func() (*Plan, *bool) {
				plan := NewPlan(nil, OnError(Continue))
				guardCalled := false

				// This task is skipped because its guard
				// returns false.
				guarded := plan.TaskFunc("guarded", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: true}, nil
				})
				guarded.When(func(_ Results) bool {
					return false
				})

				// Observer depends on guarded so it runs in a
				// later level. Its guard inspects the skipped
				// task's status.
				observer := plan.TaskFunc("observer", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{Changed: false}, nil
				})
				observer.DependsOn(guarded)
				observer.When(func(r Results) bool {
					guardCalled = true
					res := r.Get("guarded")

					return res != nil && res.Status == StatusSkipped
				})

				return plan, &guardCalled
			},
			observerName:    "observer",
			wantGuardCalled: true,
			// Observer runs because the guard sees the skipped
			// status and returns true.
			wantTaskStatus: StatusUnchanged,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan, guardCalled := tt.setup()
			runner := newRunner(plan)

			_, err := runner.run(context.Background())
			_ = err

			s.Equal(
				tt.wantGuardCalled,
				*guardCalled,
				"guard should have been called",
			)

			result := runner.results.Get(tt.observerName)
			s.NotNil(
				result,
				"observer should have a result entry",
			)
			s.Equal(
				tt.wantTaskStatus,
				result.Status,
				"observer task status",
			)
		})
	}
}

func (s *RunnerTestSuite) TestTaskFuncWithResultsReceivesResults() {
	tests := []struct {
		name        string
		setup       func() (*Plan, *string)
		wantCapture string
	}{
		{
			name: "receives upstream result data",
			setup: func() (*Plan, *string) {
				plan := NewPlan(nil, OnError(StopAll))
				var captured string

				a := plan.TaskFunc("a", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{
						Changed: true,
						Data:    map[string]any{"hostname": "web-01"},
					}, nil
				})

				b := plan.TaskFuncWithResults("b", func(
					_ context.Context,
					_ *osapiclient.Client,
					results Results,
				) (*Result, error) {
					r := results.Get("a")
					if r != nil {
						if h, ok := r.Data["hostname"].(string); ok {
							captured = h
						}
					}

					return &Result{Changed: false}, nil
				})
				b.DependsOn(a)

				return plan, &captured
			},
			wantCapture: "web-01",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan, captured := tt.setup()

			_, err := plan.Run(context.Background())

			s.Require().NoError(err)
			s.Equal(tt.wantCapture, *captured)
		})
	}
}

func (s *RunnerTestSuite) TestTaskResultCarriesData() {
	tests := []struct {
		name     string
		setup    func() *Plan
		taskName string
		wantKey  string
		wantVal  any
	}{
		{
			name: "success result includes data",
			setup: func() *Plan {
				plan := NewPlan(nil, OnError(StopAll))

				plan.TaskFunc("a", func(
					_ context.Context,
					_ *osapiclient.Client,
				) (*Result, error) {
					return &Result{
						Changed: true,
						Data:    map[string]any{"stdout": "hello"},
					}, nil
				})

				return plan
			},
			taskName: "a",
			wantKey:  "stdout",
			wantVal:  "hello",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := tt.setup()

			report, err := plan.Run(context.Background())

			s.Require().NoError(err)

			var found bool
			for _, tr := range report.Tasks {
				if tr.Name == tt.taskName {
					found = true
					s.Equal(tt.wantVal, tr.Data[tt.wantKey])
				}
			}

			s.True(found, "task %q should be in report", tt.taskName)
		})
	}
}
