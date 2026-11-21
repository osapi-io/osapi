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

type PackagePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *PackagePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

// call runs one PackageService method against c and returns its error and
// its first result as JSON-shaped fields.
type call func(c *client.Client) (any, error)

func (suite *PackagePublicTestSuite) calls() map[string]call {
	return map[string]call{
		"list": func(c *client.Client) (any, error) {
			return c.Package.List(suite.ctx, "_any")
		},
		"get": func(c *client.Client) (any, error) {
			return c.Package.Get(suite.ctx, "_any", "curl")
		},
		"install": func(c *client.Client) (any, error) {
			return c.Package.Install(suite.ctx, "_any", "curl", "8.5.0")
		},
		"remove": func(c *client.Client) (any, error) {
			return c.Package.Remove(suite.ctx, "_any", "curl")
		},
		"list updates": func(c *client.Client) (any, error) {
			return c.Package.ListUpdates(suite.ctx, "_any")
		},
		"update": func(c *client.Client) (any, error) {
			return c.Package.Update(suite.ctx, "_any")
		},
	}
}

func (suite *PackagePublicTestSuite) TestPackage() {
	row := `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01","status":"ok",` +
		`"name":"curl","version":"8.5.0","architecture":"amd64","manager":"dpkg","description":"curl","size":512000,` +
		`"changed":true,"current_version":"8.5.0","new_version":"8.5.1"}]}`
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
				_, _ = w.Write([]byte(row))
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
			wantErr:   "package",
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

func (suite *PackagePublicTestSuite) TestFields() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(
				`{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01",` +
					`"status":"ok","name":"curl","version":"8.5.0","architecture":"amd64","manager":"dpkg",` +
					`"description":"curl","size":512000,"changed":true,"current_version":"8.5.0","new_version":"8.5.1"}]}`,
			),
		)
	}))
	defer server.Close()
	c := client.New(server.URL, "test-token", client.WithLogger(slog.Default()))

	list, err := c.Package.List(suite.ctx, "_any")
	suite.Require().NoError(err)
	suite.Equal(client.PackageInfo{
		Hostname: "web-01", Status: "ok", Name: "curl", Version: "8.5.0",
		Architecture: "amd64", Manager: "dpkg", Description: "curl", Size: 512000,
	}, list.Data.Results[0])
	suite.Equal("00000000-0000-0000-0000-000000000001", list.Data.JobID)

	installed, err := c.Package.Install(suite.ctx, "_any", "curl", "")
	suite.Require().NoError(err)
	suite.Equal(client.PackageMutationResult{
		Hostname: "web-01", Status: "ok", Name: "curl", Changed: true,
	}, installed.Data.Results[0])

	updates, err := c.Package.ListUpdates(suite.ctx, "_any")
	suite.Require().NoError(err)
	suite.Equal(client.PackageUpdate{
		Hostname:       "web-01",
		Status:         "ok",
		Name:           "curl",
		CurrentVersion: "8.5.0",
		NewVersion:     "8.5.1",
	}, updates.Data.Results[0])

	index, err := c.Package.Update(suite.ctx, "_any")
	suite.Require().NoError(err)
	suite.Equal(client.PackageIndexResult{Hostname: "web-01", Status: "ok", Changed: true},
		index.Data.Results[0])
}

func TestPackagePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PackagePublicTestSuite))
}
