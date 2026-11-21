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

func (suite *RoutePublicTestSuite) calls() map[string]call {
	opts := client.RouteOpts{Destination: "198.51.100.0/24", Via: "192.0.2.1", Metric: 10}

	return map[string]call{
		"list": func(c *client.Client) (any, error) {
			return c.Route.List(suite.ctx, "_any", "eth0")
		},
		"create": func(c *client.Client) (any, error) {
			return c.Route.Create(suite.ctx, "_any", "eth0", opts)
		},
		"update": func(c *client.Client) (any, error) {
			return c.Route.Update(
				suite.ctx,
				"_any",
				"eth0",
				client.RouteOpts{Destination: "198.51.100.0/24", Via: "192.0.2.2"},
			)
		},
		"delete": func(c *client.Client) (any, error) {
			return c.Route.Delete(suite.ctx, "_any", "eth0", "198.51.100.0/24")
		},
	}
}

const routeRow = `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01",` +
	`"status":"ok","destination":"198.51.100.0/24","gateway":"192.0.2.1","interface":"eth0",` +
	`"family":"inet","metric":10,"proto":"static","changed":true}]}`

func (suite *RoutePublicTestSuite) TestRoute() {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		serverURL string
		wantErr   string
	}{
		{
			name: "when the API answers",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(routeRow))
			},
		},
		{
			name: "when the API refuses",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			wantErr: "forbidden",
		},
		{
			name: "when the body is empty",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: "nil response body",
		},
		{
			name:      "when the server is unreachable",
			serverURL: "http://127.0.0.1:0",
			wantErr:   "route",
		},
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
				c := client.New(url, "test-token", client.WithLogger(slog.Default()))

				resp, err := fn(c)
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

func (suite *RoutePublicTestSuite) TestFields() {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = nil
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(routeRow))
	}))
	defer server.Close()
	c := client.New(server.URL, "test-token", client.WithLogger(slog.Default()))

	list, err := c.Route.List(suite.ctx, "_any", "eth0")
	suite.Require().NoError(err)
	suite.Equal(client.RouteInfo{
		Hostname: "web-01", Status: "ok", Destination: "198.51.100.0/24", Gateway: "192.0.2.1",
		Interface: "eth0", Family: "inet", Metric: 10, Proto: "static",
	}, list.Data.Results[0])

	created, err := c.Route.Create(suite.ctx, "_any", "eth0",
		client.RouteOpts{Destination: "198.51.100.0/24", Via: "192.0.2.1"})
	suite.Require().NoError(err)
	suite.Equal(client.RouteMutationResult{
		Hostname: "web-01", Status: "ok", Destination: "198.51.100.0/24", Changed: true,
	}, created.Data.Results[0])
	suite.NotContains(body, "metric", "a metric of 0 is none")

	_, err = c.Route.Create(suite.ctx, "_any", "eth0",
		client.RouteOpts{Destination: "198.51.100.0/24", Via: "192.0.2.1", Metric: 10})
	suite.Require().NoError(err)
	suite.Equal(float64(10), body["metric"])
}

func TestRoutePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RoutePublicTestSuite))
}
