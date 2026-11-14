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

// Package apitest runs the HTTP cases each behavior's generated
// gen/api_<method>_public_test.go lists. Each case is one request to the
// real handler and middleware with the job client mocked, the way the
// hand-written domains' handler tests run; the response is checked for its
// status, its body when the case gives one, and against the generated spec.
package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	behaviorapi "github.com/osapi-io/osapi/internal/controller/api/behavior"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/validation"
)

// signingKey signs every token a case sends.
const signingKey = "behavior-api-test-signing-key"

// jobID is the job every mocked job client answers with.
const jobID = "job-1"

// ErrUnreachable is the job client failing to reach any agent.
var ErrUnreachable = errors.New("agent unreachable")

// Token is the token a case sends.
type Token int

const (
	// Read is a token with the built-in read role.
	Read Token = iota
	// Write is a token with the built-in write role.
	Write
	// NoToken sends no Authorization header.
	NoToken
	// Lacking is a token whose permissions grant nothing a behavior needs.
	Lacking
	// Admin is a token with the built-in admin role.
	Admin
)

// Request is the HTTP request a case sends.
type Request struct {
	Method string
	Path   string
	Body   string
}

// Case is one request and what the job client is asked and answers.
type Case struct {
	Name    string
	Request Request
	Token   Token
	// Job is what the job client expects; nil expects no job, because the
	// request must be refused before one is queued.
	Job      *Job
	WantCode int
	// WantBody, when set, is the JSON body, with the results in hostname
	// order.
	WantBody string
}

// Job is one call the job client expects and its answer.
type Job struct {
	write   bool
	target  string
	op      job.OperationType
	input   any
	replies []*job.Response
	err     error
}

// Query expects a read of op on target with input, answered by replies: one
// for a single host, one per host for a broadcast.
func Query(
	target string,
	op string,
	input any,
	replies ...*job.Response,
) *Job {
	return &Job{target: target, op: operation(op), input: input, replies: replies}
}

// Modify expects a write of op on target with input, answered by replies.
func Modify(
	target string,
	op string,
	input any,
	replies ...*job.Response,
) *Job {
	j := Query(target, op, input, replies...)
	j.write = true

	return j
}

// Fails makes the job client answer err in place of replies.
func (j *Job) Fails(
	err error,
) *Job {
	j.err = err

	return j
}

// NotFound is the error a host answers when it does not have the key.
func NotFound(
	host string,
) error {
	return fmt.Errorf("on %s: %w", host, job.ErrNotFound)
}

// Completed is host's answer carrying data, a method's output as JSON. A
// write's changed is read from it.
func Completed(
	host string,
	data string,
) *job.Response {
	r := &job.Response{Hostname: host, Status: job.StatusCompleted, Data: json.RawMessage(data)}
	var out struct {
		Changed *bool `json:"changed"`
	}
	if json.Unmarshal(r.Data, &out) == nil {
		r.Changed = out.Changed
	}

	return r
}

// Skipped is host's answer when the behavior does not run on its platform.
func Skipped(
	host string,
) *job.Response {
	return &job.Response{Hostname: host, Status: job.StatusSkipped, Error: "unsupported"}
}

// Failed is host's answer when the method fails there with msg.
func Failed(
	host string,
	msg string,
) *job.Response {
	return &job.Response{Hostname: host, Status: job.StatusFailed, Error: msg}
}

// TimedOut is host's answer when it never replied.
func TimedOut(
	host string,
) *job.Response {
	return &job.Response{Hostname: host, Status: job.StatusTimeout}
}

// Harness serves one behavior's generated routes.
type Harness struct {
	routes func(s *behaviorapi.Server) []behaviorapi.Route
	router routers.Router
}

// New returns a harness for routes, with server1 and server2 as the agents
// a target can name, and the generated spec loaded to check responses
// against.
func New(
	t *testing.T,
	routes func(s *behaviorapi.Server) []behaviorapi.Route,
) *Harness {
	validation.RegisterTargetValidator(func(context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{{Hostname: "server1"}, {Hostname: "server2"}}, nil
	})

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromFile(filepath.Join(root(t), behavior.SpecPath))
	require.NoError(t, err)
	// The spec's paths are served under any host.
	spec.Servers = nil
	router, err := legacy.NewRouter(spec)
	require.NoError(t, err)

	return &Harness{routes: routes, router: router}
}

// Run sends each case to a fresh server and checks its response.
func (h *Harness) Run(
	t *testing.T,
	cases []Case,
) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			jobs := jobmocks.NewMockJobClient(gomock.NewController(t))
			if tc.Job != nil {
				tc.Job.expect(jobs)
			}

			rec := h.serve(jobs, request(tc))

			require.Equal(t, tc.WantCode, rec.Code, rec.Body.String())
			if tc.WantBody != "" {
				assert.JSONEq(t, tc.WantBody, sorted(t, rec.Body.Bytes()))
			}
			h.conforms(t, request(tc), rec)
		})
	}
}

// expect sets the call on jobs.
func (j *Job) expect(
	jobs *jobmocks.MockJobClient,
) {
	cat := job.CategoryBehavior
	switch {
	case job.IsBroadcastTarget(j.target) && j.write:
		jobs.EXPECT().ModifyBroadcast(gomock.Any(), j.target, cat, j.op, j.input).
			Return(jobID, byHost(j.replies), j.err)
	case job.IsBroadcastTarget(j.target):
		jobs.EXPECT().QueryBroadcast(gomock.Any(), j.target, cat, j.op, j.input).
			Return(jobID, byHost(j.replies), j.err)
	case j.write:
		jobs.EXPECT().Modify(gomock.Any(), j.target, cat, j.op, j.input).
			Return(jobID, first(j.replies), j.err)
	default:
		jobs.EXPECT().Query(gomock.Any(), j.target, cat, j.op, j.input).
			Return(jobID, first(j.replies), j.err)
	}
}

func (h *Harness) serve(
	jobs *jobmocks.MockJobClient,
	req *http.Request,
) *httptest.ResponseRecorder {
	a := api.New(config.Config{}, slog.Default())
	s := behaviorapi.NewServer(slog.Default(), jobs, signingKey, nil)
	s.Mount(a.Echo, h.routes(s))

	rec := httptest.NewRecorder()
	a.Echo.ServeHTTP(rec, req)

	return rec
}

// conforms checks the response against the spec.
func (h *Harness) conforms(
	t *testing.T,
	req *http.Request,
	rec *httptest.ResponseRecorder,
) {
	route, params, err := h.router.FindRoute(req)
	require.NoError(t, err, "the spec has no route for %s %s", req.Method, req.URL)

	assert.NoError(
		t,
		openapi3filter.ValidateResponse(
			context.Background(),
			&openapi3filter.ResponseValidationInput{
				RequestValidationInput: &openapi3filter.RequestValidationInput{
					Request:    req,
					PathParams: params,
					Route:      route,
					Options: &openapi3filter.Options{
						AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
					},
				},
				Status: rec.Code,
				Header: rec.Header(),
				Body:   io.NopCloser(bytes.NewReader(rec.Body.Bytes())),
			},
		),
		"the response does not match the spec: %s",
		rec.Body.String(),
	)
}

// request builds tc's HTTP request.
func request(
	tc Case,
) *http.Request {
	var body io.Reader = http.NoBody
	if tc.Request.Body != "" {
		body = strings.NewReader(tc.Request.Body)
	}

	req := httptest.NewRequest(tc.Request.Method, tc.Request.Path, body)
	req.Header.Set("Content-Type", "application/json")
	if tok := token(tc.Token); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	return req
}

// token signs the token tok names.
func token(
	tok Token,
) string {
	var roles, perms []string
	switch tok {
	case NoToken:
		return ""
	case Read:
		roles = []string{"read"}
	case Write:
		roles = []string{"write"}
	case Lacking:
		roles, perms = []string{"admin"}, []string{"node:read"}
	case Admin:
		roles = []string{"admin"}
	}
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, roles, "apitest", perms)

	return t
}

// sorted is a response body with its results in hostname order, so a
// broadcast compares the same however the hosts answered.
func sorted(
	t *testing.T,
	data []byte,
) string {
	var body behaviorapi.Response
	require.NoError(t, json.Unmarshal(data, &body))
	sort.SliceStable(body.Results, func(i, j int) bool {
		return fmt.Sprint(body.Results[i]["hostname"]) < fmt.Sprint(body.Results[j]["hostname"])
	})
	out, _ := json.Marshal(body)

	return string(out)
}

// operation is the job operation for a behavior's op, such as sysctl.update.
func operation(
	op string,
) job.OperationType {
	return job.OperationType(job.CategoryBehavior + "." + op)
}

func first(
	replies []*job.Response,
) *job.Response {
	if len(replies) == 0 {
		return nil
	}

	return replies[0]
}

func byHost(
	replies []*job.Response,
) map[string]*job.Response {
	if len(replies) == 0 {
		return nil
	}
	out := make(map[string]*job.Response, len(replies))
	for _, r := range replies {
		out[r.Hostname] = r
	}

	return out
}

// root is the repository root: the nearest directory above the test's
// working directory with a go.mod.
func root(
	t *testing.T,
) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "no go.mod above the test")
		dir = parent
	}
}
