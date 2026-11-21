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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type DNSPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *DNSPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *DNSPublicTestSuite) calls() map[string]func(c *client.Client) (any, error) {
	return map[string]func(c *client.Client) (any, error){
		"get": func(c *client.Client) (any, error) { return c.DNS.Get(suite.ctx, "_any", "eth0") },
		"update": func(c *client.Client) (any, error) {
			return c.DNS.Update(
				suite.ctx,
				"_any",
				"eth0",
				[]string{"9.9.9.9"},
				[]string{"a.example"},
				true,
			)
		},
		"update with nothing optional": func(c *client.Client) (any, error) {
			return c.DNS.Update(suite.ctx, "_any", "eth0", nil, nil, false)
		},
		"delete": func(c *client.Client) (any, error) { return c.DNS.Delete(suite.ctx, "_any", "eth0") },
	}
}

func (suite *DNSPublicTestSuite) TestDNS() {
	row := `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01","status":"ok",` +
		`"servers":["1.1.1.1"],"search_domains":["a.example"],"interface_name":"eth0","changed":true}]}`
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
		{name: "when the server is unreachable", serverURL: "http://127.0.0.1:0", wantErr: "dns"},
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

func (suite *DNSPublicTestSuite) TestFields() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(
				`{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01",` +
					`"status":"ok","servers":["1.1.1.1"],"search_domains":["a.example"],"changed":true}]}`,
			),
		)
	}))
	defer server.Close()
	c := client.New(server.URL, "test-token", client.WithLogger(slog.Default()))

	got, err := c.DNS.Get(suite.ctx, "_any", "eth0")
	suite.Require().NoError(err)
	suite.Equal(client.DNSConfig{
		Hostname:      "web-01",
		Status:        "ok",
		Servers:       []string{"1.1.1.1"},
		SearchDomains: []string{"a.example"},
	}, got.Data.Results[0])

	deleted, err := c.DNS.Delete(suite.ctx, "_any", "eth0")
	suite.Require().NoError(err)
	suite.Equal(
		client.DNSDeleteResult{Hostname: "web-01", Status: "ok", Changed: true},
		deleted.Data.Results[0],
	)
}

func (suite *DNSPublicTestSuite) TestEmptyLists() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"hostname":"web-01","status":"ok"}]}`))
	}))
	defer server.Close()
	got, err := client.New(
		server.URL,
		"t",
		client.WithLogger(slog.Default()),
	).DNS.Get(
		suite.ctx,
		"_any",
		"eth0",
	)
	suite.Require().NoError(err)
	suite.Nil(got.Data.Results[0].Servers)
}

func TestDNSPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DNSPublicTestSuite))
}
