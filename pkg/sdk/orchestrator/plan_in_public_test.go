// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type PlanInPublicTestSuite struct {
	suite.Suite
}

func TestPlanInPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PlanInPublicTestSuite))
}

func (s *PlanInPublicTestSuite) TestDocker() {
	tests := []struct {
		name         string
		execFn       orchestrator.ExecFn
		targetName   string
		image        string
		validateFunc func(target *orchestrator.DockerTarget)
	}{
		{
			name: "returns target with correct name and image",
			execFn: func(
				_ context.Context,
				_ string,
				_ []string,
			) (string, string, int, error) {
				return "", "", 0, nil
			},
			targetName: "web-01",
			image:      "ubuntu:24.04",
			validateFunc: func(target *orchestrator.DockerTarget) {
				s.Equal("web-01", target.Name())
				s.Equal("ubuntu:24.04", target.Image())
				s.Equal("docker", target.Runtime())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(
				nil,
				orchestrator.WithDockerExecFn(tt.execFn),
			)

			target := plan.Docker(tt.targetName, tt.image)
			tt.validateFunc(target)
		})
	}
}

func (s *PlanInPublicTestSuite) TestDockerPanicsWithoutExecFn() {
	plan := orchestrator.NewPlan(nil)

	s.Panics(func() {
		plan.Docker("web", "ubuntu:24.04")
	})
}

func (s *PlanInPublicTestSuite) TestIn() {
	tests := []struct {
		name         string
		validateFunc func(sp *orchestrator.ScopedPlan)
	}{
		{
			name: "returns scoped plan with target",
			validateFunc: func(sp *orchestrator.ScopedPlan) {
				s.NotNil(sp)
				s.Equal("web", sp.Target().Name())
				s.Equal("docker", sp.Target().Runtime())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			execFn := func(
				_ context.Context,
				_ string,
				_ []string,
			) (string, string, int, error) {
				return "", "", 0, nil
			}

			plan := orchestrator.NewPlan(
				nil,
				orchestrator.WithDockerExecFn(execFn),
			)
			target := plan.Docker("web", "ubuntu:24.04")
			sp := plan.In(target)

			tt.validateFunc(sp)
		})
	}
}

func (s *PlanInPublicTestSuite) TestScopedPlanTaskFunc() {
	execFn := func(
		_ context.Context,
		_ string,
		_ []string,
	) (string, string, int, error) {
		return "", "", 0, nil
	}

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.WithDockerExecFn(execFn),
	)
	target := plan.Docker("web", "ubuntu:24.04")
	sp := plan.In(target)

	task := sp.TaskFunc("install-pkg", func(
		_ context.Context,
		_ *osapiclient.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	s.Equal("install-pkg", task.Name())
	s.Len(plan.Tasks(), 1)
	s.Equal(task, plan.Tasks()[0])
}

func (s *PlanInPublicTestSuite) TestScopedPlanTaskFuncWithResults() {
	execFn := func(
		_ context.Context,
		_ string,
		_ []string,
	) (string, string, int, error) {
		return "", "", 0, nil
	}

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.WithDockerExecFn(execFn),
	)
	target := plan.Docker("web", "ubuntu:24.04")
	sp := plan.In(target)

	task := sp.TaskFuncWithResults("check-status", func(
		_ context.Context,
		_ *osapiclient.Client,
		_ orchestrator.Results,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	s.Equal("check-status", task.Name())
	s.Len(plan.Tasks(), 1)
	s.Equal(task, plan.Tasks()[0])
}
