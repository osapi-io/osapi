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
package dispatch_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type provider struct{}

type kvIn struct {
	Key     string            `json:"key"     validate:"omitempty"`
	Content string            `json:"content" validate:"omitempty" template:"vars"`
	Vars    map[string]string `json:"vars"    validate:"omitempty"`
}

type kvOut struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
}

func list(context.Context, provider, kvIn) ([]kvOut, error) { return nil, nil }
func get(context.Context, provider, kvIn) (kvOut, error)    { return kvOut{}, nil }
func set(context.Context, provider, kvIn) (kvOut, error)    { return kvOut{}, nil }

var kv = behavior.Define("kv", "Keys",
	behavior.Read("list", list),
	behavior.Read("get", get).Key("key"),
	behavior.Write("set", set),
)

// answer is a method's response as the generated type has it.
type answer struct {
	JobID   string           `json:"job_id"`
	Results []map[string]any `json:"results"`
}

type DispatchPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	jobs *jobmocks.MockJobClient
}

func (s *DispatchPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = jobmocks.NewMockJobClient(s.ctrl)
}

func (s *DispatchPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func completed(
	host string,
	data string,
	changed *bool,
) *job.Response {
	return &job.Response{
		Hostname: host,
		Status:   job.StatusCompleted,
		Data:     json.RawMessage(data),
		Changed:  changed,
	}
}

func yes() *bool {
	b := true

	return &b
}

func (s *DispatchPublicTestSuite) TestRun() {
	errBoom := errors.New("boom")
	op := job.OperationType("behavior.kv.")
	tests := []struct {
		name     string
		method   string
		target   string
		input    kvIn
		setup    func()
		wantErr  func(error)
		wantBody string
	}{
		{
			name:   "when a list on one host returns a row per element",
			method: "list",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"list", kvIn{}).
					Return("job-1", completed("server1", `[{"key":"a"},{"key":"b"}]`, nil), nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","key":"a","status":"ok"},` +
				`{"hostname":"server1","key":"b","status":"ok"}]}`,
		},
		{
			name:   "when a read's host answers nothing, it is one row",
			method: "get",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"get", kvIn{}).
					Return("job-1", &job.Response{Hostname: "server1", Status: job.StatusCompleted}, nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a write's output has no changed, the job's is used",
			method: "set",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, op+"set", kvIn{}).
					Return("job-1", completed("server1", `{"key":"a"}`, yes()), nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"changed":true,"hostname":"server1","key":"a","status":"ok"}]}`,
		},
		{
			name:   "when a write's output and job have no changed, the row has none",
			method: "set",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, op+"set", kvIn{}).
					Return("job-1", completed("server1", `{"key":"a"}`, nil), nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","key":"a","status":"ok"}]}`,
		},
		{
			name:   "when a list is empty, the host's row has only its name and status",
			method: "list",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"list", kvIn{}).
					Return("job-1", completed("server1", `[]`, nil), nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a read is broadcast, each host's answer is its own row",
			method: "get",
			target: "_all",
			setup: func() {
				s.jobs.EXPECT().
					QueryBroadcast(gomock.Any(), "_all", job.CategoryBehavior, op+"get", kvIn{}).
					Return("job-1", map[string]*job.Response{
						"server1": completed("server1", `{"key":"a"}`, nil),
					}, nil)
			},
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","key":"a","status":"ok"}]}`,
		},
		{
			name:   "when a broadcast reaches no host, there are no rows",
			method: "set",
			target: "_all",
			setup: func() {
				s.jobs.EXPECT().
					ModifyBroadcast(gomock.Any(), "_all", job.CategoryBehavior, op+"set", kvIn{}).
					Return("job-1", map[string]*job.Response{}, nil)
			},
			wantBody: `{"job_id":"job-1","results":[]}`,
		},
		{
			name:   "when a host fails, skips or times out, its row says so",
			method: "set",
			target: "_all",
			setup: func() {
				s.jobs.EXPECT().
					ModifyBroadcast(gomock.Any(), "_all", job.CategoryBehavior, op+"set", kvIn{}).
					Return("job-1", map[string]*job.Response{
						"server1": {Hostname: "server1", Status: job.StatusFailed, Error: "boom"},
						"server2": {
							Hostname: "server2",
							Status:   job.StatusSkipped,
							Error:    "unsupported",
						},
						"server3": {Hostname: "server3", Status: job.StatusTimeout},
					}, nil)
			},
			wantBody: `{"job_id":"job-1","results":[` +
				`{"error":"boom","hostname":"server1","status":"failed"},` +
				`{"error":"unsupported","hostname":"server2","status":"skipped"},` +
				`{"error":"","hostname":"server3","status":"timeout"}]}`,
		},
		{
			name: "when a broadcast fails, it is an internal error", method: "get", target: "_all",
			setup: func() {
				s.jobs.EXPECT().
					QueryBroadcast(gomock.Any(), "_all", job.CategoryBehavior, op+"get", kvIn{}).
					Return("", nil, errBoom)
			},
			wantErr: func(err error) {
				var internal *endpoint.InternalError
				s.Require().ErrorAs(err, &internal)
				s.ErrorIs(internal.Cause, errBoom)
			},
		},
		{
			name:   "when the host does not have it, it is ErrNotFound",
			method: "get",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"get", kvIn{}).
					Return("", nil, fmt.Errorf("on server1: %w", job.ErrNotFound))
			},
			wantErr: func(err error) { s.ErrorIs(err, behavior.ErrNotFound) },
		},
		{
			name:   "when osapi did not deploy it, it is ErrNotManaged",
			method: "get",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"get", kvIn{}).
					Return("", nil, fmt.Errorf("on server1: %w", job.ErrNotManaged))
			},
			wantErr: func(err error) { s.ErrorIs(err, behavior.ErrNotManaged) },
		},
		{
			name:   "when the host lacks a program, it is ErrNotInstalled",
			method: "get",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, op+"get", kvIn{}).
					Return("", nil, fmt.Errorf("on server1: %w", job.ErrNotInstalled))
			},
			wantErr: func(err error) { s.ErrorIs(err, behavior.ErrNotInstalled) },
		},
		{
			name:   "when the job runs and fails, the host's row says why",
			method: "set",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, op+"set", kvIn{}).
					Return("job-1", nil, job.NewResponseError(&job.Response{
						Hostname: "server1", Status: job.StatusFailed, Error: "refused",
					}))
			},
			wantBody: `{"job_id":"job-1","results":[{"error":"refused","hostname":"server1","status":"failed"}]}`,
		},
		{
			name:   "when the job cannot be sent, it is an internal error",
			method: "set",
			target: "server1",
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, op+"set", kvIn{}).
					Return("", nil, errBoom)
			},
			wantErr: func(err error) {
				var internal *endpoint.InternalError
				s.ErrorAs(err, &internal)
			},
		},
		{
			name:    "when a template reads a var it does not have, it is ErrInvalid and no job is sent",
			method:  "set",
			target:  "server1",
			input:   kvIn{Content: "{{.Vars.missing}}"},
			setup:   func() {},
			wantErr: func(err error) { s.ErrorIs(err, dispatch.ErrInvalid) },
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			got, err := dispatch.Run[answer](
				context.Background(),
				dispatch.Deps{JobClient: s.jobs},
				kv,
				tc.method,
				tc.target,
				tc.input,
			)
			if tc.wantErr != nil {
				tc.wantErr(err)

				return
			}
			s.Require().NoError(err)
			data, _ := json.Marshal(got)
			s.JSONEq(tc.wantBody, sortedBody(data))
		})
	}
}

// sortedBody is a response with its rows in hostname order, as a broadcast
// answers in the order of a map.
func sortedBody(
	data []byte,
) string {
	var a answer
	_ = json.Unmarshal(data, &a)
	sort.SliceStable(a.Results, func(i, j int) bool {
		return fmt.Sprint(a.Results[i]["hostname"]) < fmt.Sprint(a.Results[j]["hostname"])
	})
	out, _ := json.Marshal(a)

	return string(out)
}

// pwIn is the input of a write whose controller step hashes its password.
type pwIn struct {
	Name         string `json:"name"`
	Password     string `json:"password,omitempty"      secret:"true"`
	PasswordHash string `json:"password_hash,omitempty"               before:"set"`
}

func setPassword(context.Context, provider, pwIn) (kvOut, error) { return kvOut{}, nil }

// TestRunBefore: the controller step runs before the job is queued, so the
// job data has the hash and not the plain password; a step that fails
// queues nothing.
func (s *DispatchPublicTestSuite) TestRunBefore() {
	account := behavior.Define("account", "Accounts",
		behavior.Write("update", setPassword).Before(
			func(_ context.Context, in pwIn) (pwIn, error) {
				if in.Password == "boom" {
					return pwIn{}, errors.New("no salt")
				}
				in.PasswordHash, in.Password = "$6$hash", ""

				return in, nil
			}),
	)
	op := job.OperationType("behavior.account.update")

	s.Run("when the step runs, the job has the hash and not the password", func() {
		s.jobs.EXPECT().
			Modify(gomock.Any(), "server1", job.CategoryBehavior, op,
				pwIn{Name: "web", PasswordHash: "$6$hash"}).
			Return("job-1", completed("server1", `{"changed":true}`, yes()), nil)

		_, err := dispatch.Run[answer](context.Background(), dispatch.Deps{JobClient: s.jobs},
			account, "update", "server1", pwIn{Name: "web", Password: "hunter2"})
		s.NoError(err)
	})

	s.Run("when the step fails, no job is queued", func() {
		_, err := dispatch.Run[answer](context.Background(), dispatch.Deps{JobClient: s.jobs},
			account, "update", "server1", pwIn{Name: "web", Password: "boom"})
		s.ErrorContains(err, "the controller step failed")
	})
}

func TestDispatchPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DispatchPublicTestSuite))
}
