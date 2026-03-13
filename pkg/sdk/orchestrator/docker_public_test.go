package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type DockerPublicTestSuite struct {
	suite.Suite
}

func (s *DockerPublicTestSuite) TestDockerPull() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		image        string
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "pull-image",
			target:   "_any",
			image:    "ubuntu:24.04",
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("pull-image", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerPull(tt.taskName, tt.target, tt.image)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerCreate() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		body         gen.DockerCreateRequest
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "create-container",
			target:   "_any",
			body:     gen.DockerCreateRequest{Image: "nginx:latest"},
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("create-container", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerCreate(tt.taskName, tt.target, tt.body)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerStart() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		id           string
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "start-container",
			target:   "_any",
			id:       "abc123",
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("start-container", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerStart(tt.taskName, tt.target, tt.id)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerStop() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		id           string
		body         gen.DockerStopRequest
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "stop-container",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerStopRequest{},
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("stop-container", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerStop(tt.taskName, tt.target, tt.id, tt.body)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerRemove() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		id           string
		params       *gen.DeleteNodeContainerDockerByIDParams
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "remove-container",
			target:   "_any",
			id:       "abc123",
			params:   nil,
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("remove-container", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerRemove(tt.taskName, tt.target, tt.id, tt.params)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerExec() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		id           string
		body         gen.DockerExecRequest
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "exec-cmd",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerExecRequest{Command: []string{"hostname"}},
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("exec-cmd", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerExec(tt.taskName, tt.target, tt.id, tt.body)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerInspect() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		id           string
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "inspect-container",
			target:   "_any",
			id:       "abc123",
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("inspect-container", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerInspect(tt.taskName, tt.target, tt.id)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func (s *DockerPublicTestSuite) TestDockerList() {
	tests := []struct {
		name         string
		taskName     string
		target       string
		params       *gen.GetNodeContainerDockerParams
		validateFunc func(*orchestrator.Task)
	}{
		{
			name:     "creates task with correct name",
			taskName: "list-containers",
			target:   "_any",
			params:   nil,
			validateFunc: func(task *orchestrator.Task) {
				s.NotNil(task)
				s.Equal("list-containers", task.Name())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			task := plan.DockerList(tt.taskName, tt.target, tt.params)
			tt.validateFunc(task)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func TestDockerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DockerPublicTestSuite))
}
