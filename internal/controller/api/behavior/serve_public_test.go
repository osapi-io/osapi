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

package behavior_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	apibehavior "github.com/osapi-io/osapi/internal/controller/api/behavior"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/validation"
)

type ServePublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	jobs *jobmocks.MockJobClient
}

func (s *ServePublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{{Hostname: "server1"}, {Hostname: "server2"}}, nil
	})
}

func (s *ServePublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = jobmocks.NewMockJobClient(s.ctrl)
}

func (s *ServePublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

// serve sends one request to method's handler, decoding its input with
// decode, the way a generated handler does.
func (s *ServePublicTestSuite) serve(
	method string,
	target string,
	decode apibehavior.Decoder[kvIn],
) *httptest.ResponseRecorder {
	h := apibehavior.Serve(
		apibehavior.NewServer(slog.Default(), s.jobs, signingKey, nil), kv, method, decode,
	)

	e := echo.New()
	e.GET("/:hostname", h)

	return send(e, http.MethodGet, "/"+target, "")
}

func changed(
	b bool,
) *bool {
	return &b
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

func (s *ServePublicTestSuite) TestServe() {
	in := kvIn{Key: "k"}
	decoded := func(_ *echo.Context) (kvIn, string, bool) { return in, "", true }

	tests := []struct {
		name     string
		method   string
		target   string
		decode   apibehavior.Decoder[kvIn]
		setup    func()
		wantCode int
		wantBody string
	}{
		{
			name:   "when a list on one host returns a row per element",
			method: "list",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, job.OperationType("behavior.kv.list"), in).
					Return("job-1", completed("server1", `[{"key":"a"},{"key":"b"}]`, nil), nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"key":"a","hostname":"server1","status":"ok"},{"key":"b","hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a read's host answers nothing, it is one row",
			method: "get",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", &job.Response{Hostname: "server1", Status: job.StatusCompleted}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a write's output has changed, the row has its value",
			method: "create",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, job.OperationType("behavior.kv.create"), in).
					Return("job-1", completed("server1", `{"key":"k","changed":false}`, changed(true)), nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"key":"k","changed":false,"hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a write's output has no changed, the job's is used",
			method: "update",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", completed("server1", `{"key":"k"}`, changed(true)), nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"key":"k","changed":true,"hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a write's output and job have no changed, the row has none",
			method: "update",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Modify(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", completed("server1", `{"key":"k"}`, nil), nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"key":"k","hostname":"server1","status":"ok"}]}`,
		},
		{
			name:   "when a read is broadcast, each host's answer is its own row",
			method: "get",
			target: "_all",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					QueryBroadcast(gomock.Any(), "_all", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", map[string]*job.Response{
						"server1": {
							Hostname: "server1",
							Status:   job.StatusSkipped,
							Error:    "unsupported",
						},
					}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"skipped","error":"unsupported"}]}`,
		},
		{
			name:   "when a broadcast reaches no host, there are no rows",
			method: "list",
			target: "_all",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					QueryBroadcast(gomock.Any(), "_all", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", map[string]*job.Response{}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[]}`,
		},
		{
			name:   "when a write is broadcast, a host that fails or times out is a row",
			method: "update",
			target: "_all",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					ModifyBroadcast(gomock.Any(), "_all", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", map[string]*job.Response{
						"server1": {Hostname: "server1", Status: job.StatusFailed, Error: "boom"},
					}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"failed","error":"boom"}]}`,
		},
		{
			name:   "when a host times out, its row says timeout",
			method: "get",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("job-1", &job.Response{Hostname: "server1", Status: job.StatusTimeout}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"job_id":"job-1","results":[{"hostname":"server1","status":"timeout","error":""}]}`,
		},
		{
			name:   "when a broadcast fails, it is a 500",
			method: "update",
			target: "_all",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					ModifyBroadcast(gomock.Any(), "_all", job.CategoryBehavior, gomock.Any(), in).
					Return("", nil, errors.New("nats unreachable"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"nats unreachable"}`,
		},
		{
			name:   "when the host does not have it, it is a 404",
			method: "get",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("", nil, fmt.Errorf("on server1: %w", job.ErrNotFound))
			},
			wantCode: http.StatusNotFound,
			wantBody: `{"error":"on server1: not found"}`,
		},
		{
			name:   "when the job fails, it is a 500",
			method: "get",
			target: "server1",
			decode: decoded,
			setup: func() {
				s.jobs.EXPECT().
					Query(gomock.Any(), "server1", job.CategoryBehavior, gomock.Any(), in).
					Return("", nil, errors.New("boom"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"boom"}`,
		},
		{
			name:     "when the target is not an agent, it is a 400 and no job is queued",
			method:   "get",
			target:   "nowhere",
			decode:   decoded,
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when the decoder refuses the request, it is a 400 with its reason",
			method: "get",
			target: "server1",
			decode: func(_ *echo.Context) (kvIn, string, bool) {
				return kvIn{}, "key is required", false
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"key is required"}`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			rec := s.serve(tc.method, tc.target, tc.decode)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
		})
	}
}

func TestServePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServePublicTestSuite))
}
