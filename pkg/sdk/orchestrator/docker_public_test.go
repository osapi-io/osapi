package orchestrator_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "pull-image",
			target:   "_any",
			image:    "ubuntu:24.04",
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("pull-image", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "pull-image",
			target:   "_any",
			image:    "alpine:latest",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"h1","image_id":"sha256:abc","tag":"latest","size":1024,"changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000001", result.JobID)
				s.True(result.Changed)
				s.Equal("sha256:abc", result.Data["image_id"])
				s.Equal("latest", result.Data["tag"])
				s.Equal(int64(1024), result.Data["size"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "pull-image",
			target:   "_any",
			image:    "alpine:latest",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerPull(tt.taskName, tt.target, tt.image)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "create-container",
			target:   "_any",
			body:     gen.DockerCreateRequest{Image: "nginx:latest"},
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("create-container", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "create-container",
			target:   "_any",
			body:     gen.DockerCreateRequest{Image: "nginx:latest"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000002","results":[{"hostname":"h1","id":"c1","name":"web","image":"nginx:latest","state":"created","changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000002", result.JobID)
				s.True(result.Changed)
				s.Equal("c1", result.Data["id"])
				s.Equal("web", result.Data["name"])
				s.Equal("nginx:latest", result.Data["image"])
				s.Equal("created", result.Data["state"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "create-container",
			target:   "_any",
			body:     gen.DockerCreateRequest{Image: "nginx:latest"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerCreate(tt.taskName, tt.target, tt.body)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "start-container",
			target:   "_any",
			id:       "abc123",
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("start-container", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "start-container",
			target:   "_any",
			id:       "abc123",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000003","results":[{"hostname":"h1","id":"abc123","message":"started","changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000003", result.JobID)
				s.True(result.Changed)
				s.Equal("abc123", result.Data["id"])
				s.Equal("started", result.Data["message"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "start-container",
			target:   "_any",
			id:       "abc123",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerStart(tt.taskName, tt.target, tt.id)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "stop-container",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerStopRequest{},
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("stop-container", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "stop-container",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerStopRequest{},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000004","results":[{"hostname":"h1","id":"abc123","message":"stopped","changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000004", result.JobID)
				s.True(result.Changed)
				s.Equal("abc123", result.Data["id"])
				s.Equal("stopped", result.Data["message"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "stop-container",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerStopRequest{},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerStop(tt.taskName, tt.target, tt.id, tt.body)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "remove-container",
			target:   "_any",
			id:       "abc123",
			params:   nil,
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("remove-container", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "remove-container",
			target:   "_any",
			id:       "abc123",
			params:   nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000005","results":[{"hostname":"h1","id":"abc123","message":"removed","changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000005", result.JobID)
				s.True(result.Changed)
				s.Equal("abc123", result.Data["id"])
				s.Equal("removed", result.Data["message"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "remove-container",
			target:   "_any",
			id:       "abc123",
			params:   nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerRemove(tt.taskName, tt.target, tt.id, tt.params)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "exec-cmd",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerExecRequest{Command: []string{"hostname"}},
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("exec-cmd", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "exec-cmd",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerExecRequest{Command: []string{"hostname"}},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000006","results":[{"hostname":"h1","stdout":"web-01\n","stderr":"","exit_code":0,"changed":true}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000006", result.JobID)
				s.True(result.Changed)
				s.Equal("web-01\n", result.Data["stdout"])
				s.Equal("", result.Data["stderr"])
				s.Equal(0, result.Data["exit_code"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "exec-cmd",
			target:   "_any",
			id:       "abc123",
			body:     gen.DockerExecRequest{Command: []string{"hostname"}},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerExec(tt.taskName, tt.target, tt.id, tt.body)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "inspect-container",
			target:   "_any",
			id:       "abc123",
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("inspect-container", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "inspect-container",
			target:   "_any",
			id:       "abc123",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000007","results":[{"hostname":"h1","id":"abc123","name":"web","image":"nginx:latest","state":"running"}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000007", result.JobID)
				s.False(result.Changed)
				s.Equal("abc123", result.Data["id"])
				s.Equal("web", result.Data["name"])
				s.Equal("nginx:latest", result.Data["image"])
				s.Equal("running", result.Data["state"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "inspect-container",
			target:   "_any",
			id:       "abc123",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerInspect(tt.taskName, tt.target, tt.id)
			tt.validateFunc(task, c)
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
		handler      http.HandlerFunc
		validateFunc func(*orchestrator.Task, *osapiclient.Client)
	}{
		{
			name:     "creates task with correct name",
			taskName: "list-containers",
			target:   "_any",
			params:   nil,
			validateFunc: func(
				task *orchestrator.Task,
				_ *osapiclient.Client,
			) {
				s.NotNil(task)
				s.Equal("list-containers", task.Name())
			},
		},
		{
			name:     "executes closure and returns result",
			taskName: "list-containers",
			target:   "_any",
			params:   nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(
					`{"job_id":"00000000-0000-0000-0000-000000000008","results":[{"hostname":"h1","containers":[{"id":"c1","name":"web","image":"nginx","state":"running"}]}]}`,
				))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.NoError(err)
				s.Equal("00000000-0000-0000-0000-000000000008", result.JobID)
				s.False(result.Changed)
				s.NotNil(result.Data["containers"])
			},
		},
		{
			name:     "returns error when SDK call fails",
			taskName: "list-containers",
			target:   "_any",
			params:   nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(
				task *orchestrator.Task,
				c *osapiclient.Client,
			) {
				result, err := task.Fn()(context.Background(), c)
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c *osapiclient.Client
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				c = osapiclient.New(srv.URL, "token")
			}

			plan := orchestrator.NewPlan(c)
			task := plan.DockerList(tt.taskName, tt.target, tt.params)
			tt.validateFunc(task, c)
			s.Len(plan.Tasks(), 1)
		})
	}
}

func TestDockerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DockerPublicTestSuite))
}
