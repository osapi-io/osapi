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
package job_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	jobgroup "github.com/osapi-io/osapi/internal/controller/endpoints/job"
	"github.com/osapi-io/osapi/internal/controller/endpoints/job/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/job/mocks"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/validation"
)

const (
	jobID    = "550e8400-e29b-41d4-a716-446655440000"
	newJobID = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
)

// signingKey signs every token a case sends.
var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-job-group"}

var when = time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }

// queued is a job with each field the list and the get read.
func queued() *job.QueuedJob {
	return &job.QueuedJob{
		ID: jobID, Status: "completed", Created: "2026-11-17T08:00:00Z",
		Operation: map[string]any{
			"type": "user.create",
			"data": map[string]any{"name": "ops", "password": "hunter2"},
		},
		Result:    json.RawMessage(`{"ok":true}`),
		Changed:   ptr(true),
		Hostname:  "web-01",
		UpdatedAt: "2026-11-17T08:00:01Z",
	}
}

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	jobs *mocks.MockJobs
}

func (s *MethodsPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "web-01", Labels: map[string]string{"group": "web"}},
		}, nil
	})
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = mocks.NewMockJobs(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps() jobgroup.Deps {
	return jobgroup.Deps{Jobs: s.jobs}
}

// TestSpecStatuses: the spec has an enum for the status of a job and for the
// status of a job on a host, so a client checks a value when it compiles.
func (s *MethodsPublicTestSuite) TestSpecStatuses() {
	doc, err := endpoint.OpenAPI(jobgroup.Group)
	s.Require().NoError(err)

	job := doc.Components.Schemas["JobStatus"].Value.Enum
	s.Contains(job, string(types.JobStatusSkipped))
	s.NotContains(job, string(types.AgentJobStatusTimeout))
	host := doc.Components.Schemas["AgentJobStatus"].Value.Enum
	s.Contains(host, string(types.AgentJobStatusTimeout))
	s.NotContains(host, string(types.JobStatusPartialFailure))
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		in           jobgroup.ListInput
		scope        []string
		setup        func()
		validateFunc func(jobgroup.ListJobsResponse, error)
	}{
		{
			name: "when the caller gives no limit, it reads the default page and masks secrets",
			setup: func() {
				s.jobs.EXPECT().
					ListJobs(gomock.Any(), types.JobStatus(""), client.DefaultPageSize, 0, gomock.Nil()).
					Return(&client.ListJobsResult{
						Jobs: []*job.QueuedJob{
							queued(),
							{ID: newJobID, Status: "submitted"},
						},
						TotalCount:   42,
						StatusCounts: map[types.JobStatus]int{"completed": 41, "submitted": 1},
					}, nil)
			},
			validateFunc: func(got jobgroup.ListJobsResponse, err error) {
				s.Require().NoError(err)
				s.Equal(42, got.TotalItems)
				s.Equal(41, got.StatusCounts["completed"])
				s.Require().Len(got.Items, 2)
				first := got.Items[0]
				s.Equal(jobID, first.ID)
				s.Equal("web-01", first.Hostname)
				s.Equal(
					"web-01",
					first.Target,
					"a job with no stored target has the host that answered",
				)
				s.Equal(map[string]any{"ok": true}, first.Result)
				s.Equal(ptr(true), first.Changed)
				data, _ := first.Operation.Data.(map[string]any)
				s.NotEqual("hunter2", data["password"])
				s.Nil(got.Items[1].Result)
				s.Nil(got.Items[1].Operation)
			},
		},
		{
			name: "when the caller gives a status, a limit and an offset, it reads that page",
			in:   jobgroup.ListInput{Status: "failed", Limit: 3, Offset: 6},
			setup: func() {
				s.jobs.EXPECT().ListJobs(gomock.Any(), types.JobStatusFailed, 3, 6, gomock.Nil()).
					Return(&client.ListJobsResult{}, nil)
			},
			validateFunc: func(got jobgroup.ListJobsResponse, err error) {
				s.Require().NoError(err)
				s.Empty(got.Items)
			},
		},
		{
			name: "when the job client fails, the caller gets a message without its error",
			setup: func() {
				s.jobs.EXPECT().
					ListJobs(gomock.Any(), types.JobStatus(""), client.DefaultPageSize, 0, gomock.Nil()).
					Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ jobgroup.ListJobsResponse, err error) {
				s.EqualError(err, "failed to list jobs")
			},
		},
		{
			name:  "when the token is scoped, the job client keeps only the jobs in its scope",
			scope: []string{"group:web"},
			setup: func() {
				s.jobs.EXPECT().
					ListJobs(gomock.Any(), types.JobStatus(""), client.DefaultPageSize, 0, gomock.Not(gomock.Nil())).
					DoAndReturn(func(
						_ context.Context, _ types.JobStatus, _ int, _ int, keep func(string) bool,
					) (*client.ListJobsResult, error) {
						s.True(keep("group:web.dev"))
						s.True(keep("web-01"), "a host whose agent has a label in the scope")
						s.False(keep("group:db"))
						s.False(keep("_all"))
						s.False(keep("_any"))
						s.False(keep(""), "a job with no stored target")

						return &client.ListJobsResult{TotalCount: 1}, nil
					})
			},
			validateFunc: func(got jobgroup.ListJobsResponse, err error) {
				s.Require().NoError(err)
				s.Equal(1, got.TotalItems)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(jobgroup.List(ctx, s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		scope        []string
		setup        func()
		validateFunc func(jobgroup.DetailResponse, error)
	}{
		{
			name: "when the job is a broadcast, it has each agent's answer, state and the timeline",
			setup: func() {
				qj := queued()
				qj.Responses = map[string]job.Response{
					"web-01": {
						Status: types.AgentJobStatusCompleted, Hostname: "web-01",
						Data: json.RawMessage(`{"uptime":5}`), Changed: ptr(false),
					},
					"web-02": {
						Status: types.AgentJobStatusFailed, Hostname: "web-02", Error: "boom",
						Steps: []types.Step{
							{
								Behavior: "netplan",
								Method:   "update",
								Changed:  ptr(true),
								Steps: []types.Step{
									{Behavior: "managedfile", Method: "update", Error: "boom"},
								},
							},
						},
					},
				}
				qj.AgentStates = map[string]job.AgentState{
					"web-01": {Status: "completed", Duration: "1.2s"},
					"web-02": {Status: "failed", Error: "boom"},
				}
				qj.Timeline = []job.TimelineEvent{
					{Timestamp: when, Event: "submitted", Hostname: "_all", Message: "sent"},
					{Timestamp: when, Event: "failed", Hostname: "web-02", Error: "boom"},
				}
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(got jobgroup.DetailResponse, err error) {
				s.Require().NoError(err)
				s.Require().Len(got.Responses, 2, "a list, sorted by hostname")
				s.Equal("web-01", got.Responses[0].Hostname)
				s.Equal(map[string]any{"uptime": float64(5)}, got.Responses[0].Data)
				s.Equal(ptr(false), got.Responses[0].Changed)
				s.Equal("web-02", got.Responses[1].Hostname)
				s.Equal("boom", got.Responses[1].Error)
				s.Nil(got.Responses[1].Data)
				s.Nil(got.Responses[0].Steps, "a method that called no other behavior")
				s.Len(got.Responses[1].Steps, 1)
				s.Equal(
					[]string{"web-02: netplan update (changed) → managedfile update (failed)"},
					got.Chain,
					"one line for each host whose method called other behaviors",
				)
				s.Require().Len(got.AgentStates, 2)
				s.Equal(
					jobgroup.AgentState{Hostname: "web-01", Status: "completed", Duration: "1.2s"},
					got.AgentStates[0],
				)
				s.Equal("boom", got.AgentStates[1].Error)
				s.Require().Len(got.Timeline, 2)
				s.Equal("2026-11-17T08:00:00Z", got.Timeline[0].Timestamp)
				s.Equal("boom", got.Timeline[1].Error)
			},
		},
		{
			name: "when the job has no agents and no timeline, those fields are absent",
			setup: func() {
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(queued(), nil)
			},
			validateFunc: func(got jobgroup.DetailResponse, err error) {
				s.Require().NoError(err)
				s.Nil(got.Responses)
				s.Nil(got.AgentStates)
				s.Nil(got.Timeline)
			},
		},
		{
			name: "when the job is not there, it is ErrNotFound and names the ID",
			setup: func() {
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(nil, job.ErrNotFound)
			},
			validateFunc: func(_ jobgroup.DetailResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNotFound)
				s.ErrorContains(err, jobID)
			},
		},
		{
			name: "when the job client fails, the caller gets a message without its error",
			setup: func() {
				s.jobs.EXPECT().
					GetJobStatus(gomock.Any(), jobID).
					Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ jobgroup.DetailResponse, err error) {
				s.EqualError(err, "failed to get job")
			},
		},
		{
			name:  "when a scoped token gets a job in its scope, it is the job",
			scope: []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "group:web.dev"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(got jobgroup.DetailResponse, err error) {
				s.Require().NoError(err)
				s.Equal("group:web.dev", got.Target)
			},
		},
		{
			name:  "when a scoped token gets a job outside its scope, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "group:db"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(_ jobgroup.DetailResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNotFound)
			},
		},
		{
			name:  "when a scoped token gets a job sent to _all, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "_all"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(_ jobgroup.DetailResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNotFound)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(
				jobgroup.Get(ctx, s.deps(), jobgroup.GetInput{ID: jobID}),
			)
		})
	}
}

func (s *MethodsPublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		scope        []string
		setup        func()
		validateFunc func(error)
	}{
		{
			name:         "when the job is there, it goes",
			setup:        func() { s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(nil) },
			validateFunc: func(err error) { s.NoError(err) },
		},
		{
			name: "when the job is not there, it is ErrNotFound",
			setup: func() {
				s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(job.ErrNotFound)
			},
			validateFunc: func(err error) { s.ErrorIs(err, jobgroup.ErrNotFound) },
		},
		{
			name: "when the job client fails, the caller gets a message without its error",
			setup: func() {
				s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(errors.New("nats: timeout"))
			},
			validateFunc: func(err error) { s.EqualError(err, "failed to delete job") },
		},
		{
			name:  "when a scoped token deletes a job in its scope, it goes",
			scope: []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "group:web"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
				s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(nil)
			},
			validateFunc: func(err error) { s.NoError(err) },
		},
		{
			name:  "when a scoped token deletes a job outside its scope, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "_all"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(err error) { s.ErrorIs(err, jobgroup.ErrNotFound) },
		},
		{
			name:  "when a scoped token deletes a job that is not there, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(nil, job.ErrNotFound)
			},
			validateFunc: func(err error) { s.ErrorIs(err, jobgroup.ErrNotFound) },
		},
		{
			name:  "when a scoped token deletes and the job client fails, it is a message without its error",
			scope: []string{"group:web"},
			setup: func() {
				s.jobs.EXPECT().
					GetJobStatus(gomock.Any(), jobID).
					Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(err error) { s.EqualError(err, "failed to get job") },
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			_, err := jobgroup.Delete(
				authtoken.WithTargets(context.Background(), tc.scope),
				s.deps(),
				jobgroup.DeleteInput{ID: jobID},
			)
			tc.validateFunc(err)
		})
	}
}

func (s *MethodsPublicTestSuite) TestRetry() {
	tests := []struct {
		name         string
		target       string
		scope        []string
		setup        func()
		validateFunc func(jobgroup.CreateJobResponse, error)
	}{
		{
			name:  "when a scoped token retries without a target, it is ErrOutOfScope",
			scope: []string{"group:web"},
			setup: func() {},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrOutOfScope)
				s.ErrorContains(err, "_any")
			},
		},
		{
			name:   "when a scoped token retries to a target outside its scope, it is ErrOutOfScope",
			target: "group:db",
			scope:  []string{"group:web"},
			setup:  func() {},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrOutOfScope)
			},
		},
		{
			name:   "when a scoped token retries to a target in its scope, the job is sent",
			target: "group:web.dev",
			scope:  []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "group:web"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
				s.jobs.EXPECT().
					RetryJob(gomock.Any(), jobID, "group:web.dev").
					Return(&client.CreateJobResult{JobID: newJobID}, nil)
			},
			validateFunc: func(got jobgroup.CreateJobResponse, err error) {
				s.Require().NoError(err)
				s.Equal(newJobID, got.JobID)
			},
		},
		{
			name:   "when a scoped token retries a job outside its scope, it is ErrNotFound",
			target: "group:web.dev",
			scope:  []string{"group:web"},
			setup: func() {
				qj := queued()
				qj.Target = "group:db"
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(qj, nil)
			},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNotFound)
			},
		},
		{
			name:   "when the job is there, it answers the new job",
			target: "web-01",
			setup: func() {
				s.jobs.EXPECT().
					RetryJob(gomock.Any(), jobID, "web-01").
					Return(&client.CreateJobResult{
						JobID:     newJobID,
						Status:    "submitted",
						Revision:  7,
						Timestamp: "2026-11-17T08:00:00Z",
					}, nil)
			},
			validateFunc: func(got jobgroup.CreateJobResponse, err error) {
				s.Require().NoError(err)
				s.Equal(jobgroup.CreateJobResponse{
					JobID:     newJobID,
					Status:    "submitted",
					Revision:  7,
					Timestamp: "2026-11-17T08:00:00Z",
				}, got)
			},
		},
		{
			name: "when the job is not there, it is ErrNotFound",
			setup: func() {
				s.jobs.EXPECT().RetryJob(gomock.Any(), jobID, "").Return(nil, job.ErrNotFound)
			},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNotFound)
			},
		},
		{
			name: "when the job has no operation data, it is ErrNoOperation",
			setup: func() {
				s.jobs.EXPECT().
					RetryJob(gomock.Any(), jobID, "").
					Return(nil, job.ErrNoOperationData)
			},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.ErrorIs(err, jobgroup.ErrNoOperation)
				s.ErrorContains(err, jobID)
			},
		},
		{
			name: "when the job client fails, the caller gets a message without its error",
			setup: func() {
				s.jobs.EXPECT().
					RetryJob(gomock.Any(), jobID, "").
					Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ jobgroup.CreateJobResponse, err error) {
				s.EqualError(err, "failed to retry job")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(jobgroup.Retry(ctx, s.deps(), jobgroup.RetryInput{
				ID: jobID, Body: jobgroup.RetryJobRequest{TargetHostname: tc.target},
			}))
		})
	}
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

	return t
}

// scopedToken is a token with perms that may act only on targets.
func scopedToken(
	targets []string,
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, targets)

	return t
}

// TestRoutes runs requests through the generated routes, with the auth, the
// decoding, the validation and the status mapping of the endpoint runtime.
func (s *MethodsPublicTestSuite) TestDeadLetters() {
	dls := []job.DeadLetter{
		{Seq: 3, JobID: "job-web", Deliveries: 5, Consumer: "jobs-agent"},
		{Seq: 2, JobID: "job-db", Deliveries: 5},
		{Seq: 1, Deliveries: 5},
	}
	jobs := func() {
		s.jobs.EXPECT().GetJobStatus(gomock.Any(), "job-web").Return(&job.QueuedJob{
			ID: "job-web", Target: "web-01", Error: "agent did not answer",
			Operation: map[string]interface{}{"type": "node.hostname.get"},
		}, nil).AnyTimes()
		s.jobs.EXPECT().GetJobStatus(gomock.Any(), "job-db").
			Return(nil, errors.New("not found")).AnyTimes()
	}
	tests := []struct {
		name         string
		in           jobgroup.DeadLetterListInput
		scope        []string
		setup        func()
		validateFunc func(jobgroup.DeadLetterList, error)
	}{
		{
			name: "each dead letter, with its job when the job is there",
			setup: func() {
				s.jobs.EXPECT().ListDeadLetters(gomock.Any(), 20).Return(dls, 654, nil)
				jobs()
			},
			validateFunc: func(got jobgroup.DeadLetterList, err error) {
				s.Require().NoError(err)
				s.Equal(654, got.TotalItems)
				s.Require().Len(got.Items, 3)
				s.Equal("web-01", got.Items[0].Target)
				s.Equal("node.hostname.get", got.Items[0].Operation.Type)
				s.Equal("agent did not answer", got.Items[0].Error)
				s.Empty(got.Items[1].Target, "a job that is gone leaves its fields empty")
				s.Empty(got.Items[2].JobID)
			},
		},
		{
			name:  "a scoped token sees only the dead letters of its targets",
			in:    jobgroup.DeadLetterListInput{Limit: 5},
			scope: []string{"web-01"},
			setup: func() {
				s.jobs.EXPECT().ListDeadLetters(gomock.Any(), 5).Return(dls, 654, nil)
				jobs()
			},
			validateFunc: func(got jobgroup.DeadLetterList, err error) {
				s.Require().NoError(err)
				s.Equal(1, got.TotalItems, "the total counts only what the token sees")
				s.Require().Len(got.Items, 1)
				s.Equal("job-web", got.Items[0].JobID)
			},
		},
		{
			name: "a stream that cannot be read is a 500",
			setup: func() {
				s.jobs.EXPECT().
					ListDeadLetters(gomock.Any(), 20).
					Return(nil, 0, errors.New("nats down"))
			},
			validateFunc: func(_ jobgroup.DeadLetterList, err error) {
				s.ErrorContains(err, "failed to list the dead letters")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(jobgroup.DeadLetters(ctx, s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		token    string
		setup    func()
		wantCode int
		wantBody string
		wantLoc  string
	}{
		{
			name:   "when the token has job:read, list is a 200",
			method: http.MethodGet,
			path:   "/api/job?status=completed&limit=5&offset=10",
			token:  token("job:read"),
			setup: func() {
				s.jobs.EXPECT().
					ListJobs(gomock.Any(), types.JobStatusCompleted, 5, 10, gomock.Nil()).
					Return(&client.ListJobsResult{TotalCount: 0}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{}`,
		},
		{
			name:     "when the status is not one of the five, it is a 400",
			method:   http.MethodGet,
			path:     "/api/job?status=lost",
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when the limit is over 100, it is a 400",
			method:   http.MethodGet,
			path:     "/api/job?limit=101",
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when the offset is negative, it is a 400",
			method:   http.MethodGet,
			path:     "/api/job?offset=-1",
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when there is no token, it is a 401",
			method:   http.MethodGet,
			path:     "/api/job",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks job:read, it is a 403",
			method:   http.MethodGet,
			path:     "/api/job/" + jobID,
			token:    token("agent:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:   "when the job is there, get is a 200",
			method: http.MethodGet,
			path:   "/api/job/" + jobID,
			token:  token("job:read"),
			setup: func() {
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).
					Return(&job.QueuedJob{ID: jobID, Status: "completed"}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"id":"` + jobID + `","status":"completed"}`,
		},
		{
			name:     "when the ID is not a UUID, it is a 400",
			method:   http.MethodGet,
			path:     "/api/job/not-a-uuid",
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when the job is not there, it is a 404",
			method: http.MethodGet,
			path:   "/api/job/" + jobID,
			token:  token("job:read"),
			setup: func() {
				s.jobs.EXPECT().GetJobStatus(gomock.Any(), jobID).Return(nil, job.ErrNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "when the token has only job:read, delete is a 403",
			method:   http.MethodDelete,
			path:     "/api/job/" + jobID,
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:   "when the token has job:write, delete is a 204",
			method: http.MethodDelete,
			path:   "/api/job/" + jobID,
			token:  token("job:write"),
			setup: func() {
				s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(nil)
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:   "when a retry has no body, it goes to the default target and is a 201",
			method: http.MethodPost,
			path:   "/api/job/" + jobID + "/retry",
			token:  token("job:write"),
			setup: func() {
				s.jobs.EXPECT().RetryJob(gomock.Any(), jobID, "").
					Return(&client.CreateJobResult{JobID: newJobID, Status: "submitted"}, nil)
			},
			wantCode: http.StatusCreated,
			wantBody: `{"job_id":"` + newJobID + `","status":"submitted"}`,
			wantLoc:  "/api/job/" + newJobID,
		},
		{
			name:   "when a retry names a target, the new job goes there",
			method: http.MethodPost,
			path:   "/api/job/" + jobID + "/retry",
			body:   `{"target_hostname":"_all"}`,
			token:  token("job:write"),
			setup: func() {
				s.jobs.EXPECT().RetryJob(gomock.Any(), jobID, "_all").
					Return(&client.CreateJobResult{JobID: newJobID, Status: "submitted"}, nil)
			},
			wantCode: http.StatusCreated,
			wantLoc:  "/api/job/" + newJobID,
		},
		{
			name:     "when a scoped token retries to _any, it is a 403",
			method:   http.MethodPost,
			path:     "/api/job/" + jobID + "/retry",
			token:    scopedToken([]string{"group:web"}, "job:write"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "when the retry's target is not a valid target, it is a 400",
			method:   http.MethodPost,
			path:     "/api/job/" + jobID + "/retry",
			body:     `{"target_hostname":"not a target!"}`,
			token:    token("job:write"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when the old job has no operation data, the retry is a 400",
			method: http.MethodPost,
			path:   "/api/job/" + jobID + "/retry",
			token:  token("job:write"),
			setup: func() {
				s.jobs.EXPECT().
					RetryJob(gomock.Any(), jobID, "").
					Return(nil, job.ErrNoOperationData)
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when the job client fails, it is a 500 that does not name its error",
			method: http.MethodDelete,
			path:   "/api/job/" + jobID,
			token:  token("job:write"),
			setup: func() {
				s.jobs.EXPECT().DeleteJob(gomock.Any(), jobID).Return(errors.New("nats: timeout"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"failed to delete job","reason":"server_error"}`,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, s.deps())
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
			s.Equal(tc.wantLoc, rec.Header().Get("Location"))
		})
	}
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
