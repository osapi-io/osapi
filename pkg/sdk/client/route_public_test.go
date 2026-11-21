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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type RoutePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *RoutePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *RoutePublicTestSuite) calls() map[string]func(c *client.Client) (any, error) {
	opts := client.RouteConfigOpts{Routes: []client.RouteItem{{To: "10.0.0.0/8", Via: "10.0.1.1"}}}

	return map[string]func(c *client.Client) (any, error){
		"list": func(c *client.Client) (any, error) { return c.Route.List(suite.ctx, "_any") },
		"get":  func(c *client.Client) (any, error) { return c.Route.Get(suite.ctx, "_any", "eth0") },
		"create": func(c *client.Client) (any, error) {
			return c.Route.Create(suite.ctx, "_any", "eth1", opts)
		},
		"update": func(c *client.Client) (any, error) {
			return c.Route.Update(suite.ctx, "_any", "eth1", opts)
		},
		"delete": func(c *client.Client) (any, error) { return c.Route.Delete(suite.ctx, "_any", "eth1") },
	}
}

func (suite *RoutePublicTestSuite) TestRoute() {
	row := `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01","status":"ok",` +
		`"interface_name":"eth1","changed":true}]}`
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		serverURL string
		wantErr   string
	}{
		{name: "when the API answers", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(row))
		}},
		{
			name:    "when the API refuses",
			wantErr: "forbidden",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
		},
		{
			name:    "when the body is empty",
			wantErr: "nil response body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{name: "when the server is unreachable", serverURL: "http://127.0.0.1:0", wantErr: "route"},
	}
	for _, tc := range tests {
		for method, fn := range suite.calls() {
			suite.Run(tc.name+": "+method, func() {
				url := tc.serverURL
				if url == "" {
					server := httptest.NewServer(tc.handler)
					defer server.Close()
					url = server.URL
				}
				resp, err := fn(client.New(url, "test-token", client.WithLogger(slog.Default())))
				if tc.wantErr != "" {
					suite.ErrorContains(err, tc.wantErr)

					return
				}
				suite.Require().NoError(err)
				suite.NotNil(resp)
			})
		}
	}
}

// TestRowsGroupByHost checks that the rows of a list or a get, one per
// route, become one result per host, and a host that failed has no routes.
func (suite *RoutePublicTestSuite) TestRowsGroupByHost() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[` +
			`{"hostname":"web-01","status":"ok","destination":"default","gateway":"10.0.0.1",` +
			`"interface":"eth0","family":"inet","metric":100,"proto":"dhcp"},` +
			`{"hostname":"web-01","status":"ok","destination":"10.0.0.0/24","interface":"eth0",` +
			`"family":"inet","proto":"kernel","scope":"link"},` +
			`{"hostname":"web-02","status":"failed","error":"boom"}]}`))
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))
	want := []client.RouteInfo{
		{
			Destination: "default", Gateway: "10.0.0.1", Interface: "eth0", Family: "inet",
			Metric: 100, Proto: "dhcp",
		},
		{
			Destination: "10.0.0.0/24",
			Interface:   "eth0",
			Family:      "inet",
			Proto:       "kernel",
			Scope:       "link",
		},
	}

	listed, err := c.Route.List(suite.ctx, "_all")
	suite.Require().NoError(err)
	suite.Equal([]client.RouteListResult{
		{Hostname: "web-01", Status: "ok", Routes: want},
		{Hostname: "web-02", Status: "failed", Error: "boom"},
	}, listed.Data.Results)

	got, err := c.Route.Get(suite.ctx, "_all", "eth0")
	suite.Require().NoError(err)
	suite.Equal(
		client.RouteGetResult{Hostname: "web-01", Status: "ok", Routes: want},
		got.Data.Results[0],
	)
}

func (suite *RoutePublicTestSuite) TestMutationFields() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(`{"results":[{"hostname":"web-01","status":"ok","interface_name":"eth1",` +
				`"changed":true},{"hostname":"web-02","status":"skipped","error":"unsupported"}]}`),
		)
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))

	deleted, err := c.Route.Delete(suite.ctx, "_all", "eth1")
	suite.Require().NoError(err)
	suite.Equal([]client.RouteMutationResult{
		{Hostname: "web-01", Status: "ok", Interface: "eth1", Changed: true},
		{Hostname: "web-02", Status: "skipped", Error: "unsupported"},
	}, deleted.Data.Results)
}

// TestRequestBody checks that each route becomes the string the API takes.
func (suite *RoutePublicTestSuite) TestRequestBody() {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = map[string]any{}
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))
	metric := 100

	_, err := c.Route.Create(
		suite.ctx,
		"_any",
		"eth1",
		client.RouteConfigOpts{Routes: []client.RouteItem{
			{To: "10.0.0.0/8", Via: "10.0.1.1"},
			{To: "2001:db8:1::/48", Via: "2001:db8::1", Metric: &metric},
		}},
	)
	suite.Require().NoError(err)
	suite.Equal(map[string]any{"routes": []any{
		"10.0.0.0/8 via 10.0.1.1", "2001:db8:1::/48 via 2001:db8::1 metric 100",
	}}, body)
}

func TestRoutePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RoutePublicTestSuite))
}
