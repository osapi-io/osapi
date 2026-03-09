package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type TaskPublicTestSuite struct {
	suite.Suite
}

func TestTaskPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TaskPublicTestSuite))
}

// noop is a no-op TaskFn for tests that only need a valid task.
func noop(
	_ context.Context,
	_ *osapiclient.Client,
) (*orchestrator.Result, error) {
	return &orchestrator.Result{}, nil
}

func (s *TaskPublicTestSuite) TestDependsOn() {
	tests := []struct {
		name       string
		setupDeps  func(a, b, c *orchestrator.Task)
		checkTask  string
		wantDepLen int
	}{
		{
			name: "single dependency",
			setupDeps: func(a, b, _ *orchestrator.Task) {
				b.DependsOn(a)
			},
			checkTask:  "b",
			wantDepLen: 1,
		},
		{
			name: "multiple dependencies",
			setupDeps: func(a, b, c *orchestrator.Task) {
				c.DependsOn(a, b)
			},
			checkTask:  "c",
			wantDepLen: 2,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := orchestrator.NewTaskFunc("a", noop)
			b := orchestrator.NewTaskFunc("b", noop)
			c := orchestrator.NewTaskFunc("c", noop)
			tt.setupDeps(a, b, c)

			tasks := map[string]*orchestrator.Task{"a": a, "b": b, "c": c}
			s.Len(tasks[tt.checkTask].Dependencies(), tt.wantDepLen)
		})
	}
}

func (s *TaskPublicTestSuite) TestOnlyIfChanged() {
	task := orchestrator.NewTaskFunc("t", noop)
	dep := orchestrator.NewTaskFunc("dep", noop)
	task.DependsOn(dep).OnlyIfChanged()

	s.True(task.RequiresChange())
}

func (s *TaskPublicTestSuite) TestWhen() {
	task := orchestrator.NewTaskFunc("t", noop)
	called := false
	task.When(func(_ orchestrator.Results) bool {
		called = true

		return true
	})

	guard := task.Guard()
	s.NotNil(guard)
	s.True(guard(orchestrator.Results{}))
	s.True(called)
}

func (s *TaskPublicTestSuite) TestTaskFunc() {
	fn := func(
		_ context.Context,
		_ *osapiclient.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}

	task := orchestrator.NewTaskFunc("custom", fn)
	s.Equal("custom", task.Name())
	s.True(task.IsFunc())
}

func (s *TaskPublicTestSuite) TestSetName() {
	tests := []struct {
		name     string
		initial  string
		renamed  string
		wantName string
	}{
		{
			name:     "changes task name",
			initial:  "original",
			renamed:  "renamed",
			wantName: "renamed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			task := orchestrator.NewTaskFunc(tt.initial, noop)
			task.SetName(tt.renamed)
			s.Equal(tt.wantName, task.Name())
		})
	}
}

func (s *TaskPublicTestSuite) TestWhenWithReason() {
	tests := []struct {
		name        string
		guardResult bool
		reason      string
	}{
		{
			name:        "sets guard and reason when guard returns false",
			guardResult: false,
			reason:      "host is unreachable",
		},
		{
			name:        "sets guard and reason when guard returns true",
			guardResult: true,
			reason:      "custom reason",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			task := orchestrator.NewTaskFunc("t", noop)
			task.WhenWithReason(func(_ orchestrator.Results) bool {
				return tt.guardResult
			}, tt.reason)

			guard := task.Guard()
			s.NotNil(guard)
			s.Equal(tt.guardResult, guard(orchestrator.Results{}))
		})
	}
}

func (s *TaskPublicTestSuite) TestOnErrorOverride() {
	task := orchestrator.NewTaskFunc("t", noop)
	task.OnError(orchestrator.Continue)

	s.NotNil(task.ErrorStrategy())
	s.Equal("continue", task.ErrorStrategy().String())
}

func (s *TaskPublicTestSuite) TestFn() {
	fnTask := orchestrator.NewTaskFunc("fn", func(
		_ context.Context,
		_ *osapiclient.Client,
	) (*orchestrator.Result, error) {
		return nil, nil
	})
	s.NotNil(fnTask.Fn())
}
