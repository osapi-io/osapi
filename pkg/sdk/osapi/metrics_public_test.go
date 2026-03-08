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

package osapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/osapi"
)

type MetricsPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *MetricsPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *MetricsPublicTestSuite) TestGet() {
	closedServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	closedServerURL := closedServer.URL
	closedServer.Close()

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		ctx          context.Context
		validateFunc func(string, error)
	}{
		{
			name: "when server returns metrics returns text body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("# HELP go_goroutines\n"))
			},
			ctx: suite.ctx,
			validateFunc: func(body string, err error) {
				suite.NoError(err)
				suite.Equal("# HELP go_goroutines\n", body)
			},
		},
		{
			name: "when server returns non-200 returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			ctx: suite.ctx,
			validateFunc: func(body string, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "metrics endpoint returned status")
				suite.Empty(body)
			},
		},
		{
			name:      "when server is unreachable returns error",
			serverURL: closedServerURL,
			ctx:       suite.ctx,
			validateFunc: func(body string, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "fetching metrics")
				suite.Empty(body)
			},
		},
		{
			name: "when request creation fails returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			ctx: nil,
			validateFunc: func(body string, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "creating metrics request")
				suite.Empty(body)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var targetURL string

			if tc.serverURL != "" {
				targetURL = tc.serverURL
			} else {
				server := httptest.NewServer(tc.handler)
				defer server.Close()
				targetURL = server.URL
			}

			sut := osapi.New(
				targetURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			//nolint:staticcheck // nil context intentionally triggers NewRequestWithContext error
			body, err := sut.Metrics.Get(tc.ctx)
			tc.validateFunc(body, err)
		})
	}
}

func TestMetricsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsPublicTestSuite))
}
