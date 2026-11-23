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

package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type DockerPublicTestSuite struct {
	suite.Suite
}

// dockerCall is one method of the DockerService, with what it sends, the row
// the server answers and the result that the row converts to.
type dockerCall struct {
	name   string
	method string
	path   string
	query  string
	body   string
	row    string
	run    func(*client.DockerService) (any, error)
	want   any
}

func (s *DockerPublicTestSuite) calls() []dockerCall {
	ctx := context.Background()

	return []dockerCall{
		{
			name: "docker list", method: http.MethodGet, path: "/api/node/web1/container/docker",
			query: "state=running",
			row: `{"hostname":"web1","status":"ok","id":"abc","name":"web","image":"nginx",` +
				`"state":"running","created":"2026-11-22T10:00:00Z","health":"healthy",` +
				`"ports":["80/tcp"],"mounts":["/a:/b"],"networks":["bridge"],"env_names":["PATH"]}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.List(ctx, "web1", "running")
				return unwrap(r, err)
			},
			want: []client.DockerResult{{
				Hostname: "web1", Status: "ok", ID: "abc", Name: "web", Image: "nginx",
				State: "running", Created: "2026-11-22T10:00:00Z", Health: "healthy",
				Ports: []string{"80/tcp"}, Mounts: []string{"/a:/b"}, Networks: []string{"bridge"},
				EnvNames: []string{"PATH"},
			}},
		},
		{
			name:   "docker list of all",
			method: http.MethodGet,
			path:   "/api/node/web1/container/docker",
			row:    `{"hostname":"web1","status":"skipped","error":"docker does not answer"}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.List(ctx, "web1", "")
				return unwrap(r, err)
			},
			want: []client.DockerResult{
				{Hostname: "web1", Status: "skipped", Error: "docker does not answer"},
			},
		},
		{
			name: "docker get", method: http.MethodGet, path: "/api/node/web1/container/docker/web",
			row: `{"hostname":"web1","status":"ok","name":"web"}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Get(ctx, "web1", "web")
				return unwrap(r, err)
			},
			want: []client.DockerResult{{Hostname: "web1", Status: "ok", Name: "web"}},
		},
		{
			name:   "docker create",
			method: http.MethodPost,
			path:   "/api/node/web1/container/docker",
			body: `{"auto_start":true,"command":["sleep","1"],"container_hostname":"box","dns":["1.1.1.1"],` +
				`"env":{"A":"1"},"image":"alpine","name":"web","ports":["80:80"],"volumes":["/a:/b"]}`,
			row: `{"hostname":"web1","status":"ok","name":"web","id":"abc","changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Create(ctx, "web1", client.DockerCreateOpts{
					Name: "web", Image: "alpine", Hostname: "box", DNS: []string{"1.1.1.1"},
					Command: []string{"sleep", "1"}, Env: map[string]string{"A": "1"},
					Ports: []string{"80:80"}, Volumes: []string{"/a:/b"}, AutoStart: true,
				})
				return unwrap(r, err)
			},
			want: []client.DockerActionResult{
				{Hostname: "web1", Status: "ok", Name: "web", ID: "abc", Changed: true},
			},
		},
		{
			name:   "docker create with no options",
			method: http.MethodPost,
			path:   "/api/node/web1/container/docker",
			body:   `{"image":"alpine","name":"web"}`,
			row:    `{"hostname":"web1","status":"ok","name":"web"}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Create(
					ctx,
					"web1",
					client.DockerCreateOpts{Name: "web", Image: "alpine"},
				)
				return unwrap(r, err)
			},
			want: []client.DockerActionResult{{Hostname: "web1", Status: "ok", Name: "web"}},
		},
		{
			name:   "docker remove",
			method: http.MethodDelete,
			path:   "/api/node/web1/container/docker/web",
			row:    `{"hostname":"web1","status":"ok","name":"web","changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Remove(ctx, "web1", "web")
				return unwrap(r, err)
			},
			want: []client.DockerActionResult{
				{Hostname: "web1", Status: "ok", Name: "web", Changed: true},
			},
		},
		{
			name:   "docker start",
			method: http.MethodPost,
			path:   "/api/node/web1/container/docker/web/start",
			row:    `{"hostname":"web1","status":"ok","name":"web"}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Start(ctx, "web1", "web")
				return unwrap(r, err)
			},
			want: []client.DockerActionResult{{Hostname: "web1", Status: "ok", Name: "web"}},
		},
		{
			name:   "docker stop",
			method: http.MethodPost,
			path:   "/api/node/web1/container/docker/web/stop",
			body:   `{"timeout":5}`,
			row:    `{"hostname":"web1","status":"ok","name":"web","changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Stop(ctx, "web1", "web", 5)
				return unwrap(r, err)
			},
			want: []client.DockerActionResult{
				{Hostname: "web1", Status: "ok", Name: "web", Changed: true},
			},
		},
		{
			name:   "docker exec",
			method: http.MethodPost,
			path:   "/api/node/web1/container/docker/web/exec",
			body:   `{"command":["ls"],"env":{"A":"1"},"working_dir":"/tmp"}`,
			row:    `{"hostname":"web1","status":"ok","stdout":"a","stderr":"b","exit_code":2,"changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Exec(ctx, "web1", "web", client.DockerExecOpts{
					Command: []string{"ls"}, Env: map[string]string{"A": "1"}, WorkingDir: "/tmp",
				})
				return unwrap(r, err)
			},
			want: []client.DockerExecResult{
				{
					Hostname: "web1",
					Status:   "ok",
					Stdout:   "a",
					Stderr:   "b",
					ExitCode: 2,
					Changed:  true,
				},
			},
		},
		{
			name:   "docker image list",
			method: http.MethodGet,
			path:   "/api/node/web1/container/image",
			row:    `{"hostname":"web1","status":"ok","id":"91ef","repository":"alpine","tag":"3.20","size":"7.8MB"}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.ListImages(ctx, "web1")
				return unwrap(r, err)
			},
			want: []client.DockerImageResult{
				{
					Hostname:   "web1",
					Status:     "ok",
					ID:         "91ef",
					Repository: "alpine",
					Tag:        "3.20",
					Size:       "7.8MB",
				},
			},
		},
		{
			name:   "docker pull",
			method: http.MethodPost,
			path:   "/api/node/web1/container/image/pull",
			body:   `{"image":"alpine"}`,
			row:    `{"hostname":"web1","status":"ok","image":"alpine","image_id":"sha256:x","size":9,"changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.Pull(ctx, "web1", "alpine")
				return unwrap(r, err)
			},
			want: []client.DockerPullResult{
				{
					Hostname: "web1",
					Status:   "ok",
					Image:    "alpine",
					ImageID:  "sha256:x",
					Size:     9,
					Changed:  true,
				},
			},
		},
		{
			name:   "docker image remove",
			method: http.MethodDelete,
			path:   "/api/node/web1/container/image/alpine",
			body:   `{}`,
			row:    `{"hostname":"web1","status":"ok","image":"alpine","changed":true}`,
			run: func(d *client.DockerService) (any, error) {
				r, err := d.RemoveImage(ctx, "web1", "alpine", false)
				return unwrap(r, err)
			},
			want: []client.DockerImageRemoveResult{
				{Hostname: "web1", Status: "ok", Image: "alpine", Changed: true},
			},
		},
	}
}

// unwrap is the rows of a response, for a comparison of any method's.
func unwrap[T any](
	r *client.Response[client.Collection[T]],
	err error,
) (any, error) {
	if err != nil {
		return nil, err
	}

	return r.Data.Results, nil
}

func (s *DockerPublicTestSuite) serve(
	h http.HandlerFunc,
) *client.DockerService {
	srv := httptest.NewServer(h)
	s.T().Cleanup(srv.Close)

	return client.New(srv.URL, "test-token", client.WithLogger(slog.Default())).Docker
}

// TestCalls: each method sends its request, and converts each row of the
// answer.
func (s *DockerPublicTestSuite) TestCalls() {
	for _, c := range s.calls() {
		s.Run(c.name, func() {
			d := s.serve(func(w http.ResponseWriter, r *http.Request) {
				s.Equal(c.method, r.Method)
				s.Equal(c.path, r.URL.Path)
				s.Equal(c.query, r.URL.RawQuery)
				if c.body != "" {
					got, _ := io.ReadAll(r.Body)
					s.JSONEq(c.body, string(got))
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(
					[]byte(
						`{"job_id":"00000000-0000-0000-0000-000000000001","results":[` + c.row + `]}`,
					),
				)
			})
			got, err := c.run(d)
			s.Require().NoError(err)
			s.Equal(c.want, got)
		})
	}
}

// TestFailures: each method answers a status that is an error, a body that
// is nil and a request that fails as an error.
func (s *DockerPublicTestSuite) TestFailures() {
	for _, c := range s.calls() {
		s.Run(c.name+" with a 403", func() {
			d := s.serve(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			})
			_, err := c.run(d)
			var target *client.AuthError
			s.True(errors.As(err, &target), "err is %T: %v", err, err)
		})
		s.Run(c.name+" with no body", func() {
			d := s.serve(
				func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
			)
			_, err := c.run(d)
			var target *client.UnexpectedStatusError
			s.True(errors.As(err, &target), "err is %T: %v", err, err)
		})
		s.Run(c.name+" that cannot connect", func() {
			d := client.New("http://127.0.0.1:0", "t", client.WithLogger(slog.Default())).Docker
			_, err := c.run(d)
			s.ErrorContains(err, c.name[:len("docker ")])
		})
	}
}

// TestRowsAreJSON: a row converts to the same JSON the server sent.
func (s *DockerPublicTestSuite) TestRowsAreJSON() {
	b, err := json.Marshal(client.DockerExecResult{Hostname: "h", Status: "ok"})
	s.Require().NoError(err)
	s.JSONEq(`{"hostname":"h","status":"ok","exit_code":0,"changed":false}`, string(b))
}

func TestDockerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DockerPublicTestSuite))
}
