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
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type LogPublicTestSuite struct {
	suite.Suite
}

// serve runs call against a server that answers with handler, or against
// url when it is set.
func (s *LogPublicTestSuite) serve(
	handler http.HandlerFunc,
	url string,
	call func(*client.Client) error,
) error {
	if url == "" {
		server := httptest.NewServer(handler)
		defer server.Close()
		url = server.URL
	}

	return call(client.New(url, "test-token", client.WithLogger(slog.Default())))
}

func answer(
	code int,
	body string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}
}

func (s *LogPublicTestSuite) TestEntries() {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		url       string
		wantQuery string
		want      []client.LogEntryResult
		wantErr   any
		wantMsg   string
	}{
		{
			name: "when the hosts answer, each entry is a row, with the filters in the query",
			handler: answer(http.StatusOK, `{"job_id":"00000000-0000-0000-0000-000000000001",`+
				`"results":[{"hostname":"web-01","status":"ok","timestamp":"2024-03-31T23:50:45Z",`+
				`"unit":"nginx","priority":"info","message":"Started nginx","pid":1234}]}`),
			wantQuery: "unit=nginx.service&lines=5&since=1h&priority=err",
			want: []client.LogEntryResult{{
				Hostname: "web-01", Status: "ok", Timestamp: "2024-03-31T23:50:45Z",
				Unit: "nginx", Priority: "info", Message: "Started nginx", PID: 1234,
			}},
		},
		{
			name:    "when the token may not read the log",
			handler: answer(http.StatusForbidden, `{"error":"forbidden"}`),
			wantErr: &client.AuthError{},
		},
		{
			name:    "when the controller fails",
			handler: answer(http.StatusInternalServerError, `{"error":"x"}`),
			wantErr: &client.ServerError{},
		},
		{
			name:    "when the call does not reach the controller",
			url:     "http://127.0.0.1:0",
			wantMsg: "log entries",
		},
		{
			name:    "when the answer has no body",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
			wantErr: &client.UnexpectedStatusError{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var query string
			handler := tc.handler
			if handler != nil {
				inner := handler
				handler = func(w http.ResponseWriter, r *http.Request) {
					query = r.URL.RawQuery
					inner(w, r)
				}
			}
			var resp *client.Response[client.Collection[client.LogEntryResult]]
			err := s.serve(handler, tc.url, func(c *client.Client) error {
				var err error
				resp, err = c.Log.Entries(context.Background(), "_all", client.LogEntryOpts{
					Unit: "nginx.service", Lines: 5, Since: "1h", Priority: "err",
				})

				return err
			})
			switch {
			case tc.wantErr != nil:
				s.Require().Error(err)
				s.True(errors.As(err, &tc.wantErr), "err is %T: %v", err, err)
			case tc.wantMsg != "":
				s.ErrorContains(err, tc.wantMsg)
			default:
				s.Require().NoError(err)
				s.Equal(tc.wantQuery, query)
				s.Equal(tc.want, resp.Data.Results)
				s.Equal("00000000-0000-0000-0000-000000000001", resp.Data.JobID)
			}
		})
	}
}

func (s *LogPublicTestSuite) TestEntriesWithNoFilters() {
	var query string
	err := s.serve(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		answer(http.StatusOK, `{"results":[]}`)(w, r)
	}, "", func(c *client.Client) error {
		_, err := c.Log.Entries(context.Background(), "web-01", client.LogEntryOpts{})

		return err
	})
	s.NoError(err)
	s.Empty(query, "no filter is sent")
}

func (s *LogPublicTestSuite) TestSources() {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		url     string
		want    []client.LogSourceResult
		wantErr any
		wantMsg string
	}{
		{
			name: "when the hosts answer, each source is a row",
			handler: answer(
				http.StatusOK,
				`{"results":[{"hostname":"web-01","status":"ok","name":"sshd"},`+
					`{"hostname":"mac-01","status":"skipped","error":"unsupported"}]}`,
			),
			want: []client.LogSourceResult{
				{Hostname: "web-01", Status: "ok", Name: "sshd"},
				{Hostname: "mac-01", Status: "skipped", Error: "unsupported"},
			},
		},
		{
			name:    "when the token is not valid",
			handler: answer(http.StatusUnauthorized, `{"error":"unauthorized"}`),
			wantErr: &client.AuthError{},
		},
		{
			name:    "when the call does not reach the controller",
			url:     "http://127.0.0.1:0",
			wantMsg: "log sources",
		},
		{
			name:    "when the answer has no body",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
			wantErr: &client.UnexpectedStatusError{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var resp *client.Response[client.Collection[client.LogSourceResult]]
			err := s.serve(tc.handler, tc.url, func(c *client.Client) error {
				var err error
				resp, err = c.Log.Sources(context.Background(), "_all")

				return err
			})
			switch {
			case tc.wantErr != nil:
				s.Require().Error(err)
				s.True(errors.As(err, &tc.wantErr), "err is %T: %v", err, err)
			case tc.wantMsg != "":
				s.ErrorContains(err, tc.wantMsg)
			default:
				s.Require().NoError(err)
				s.Equal(tc.want, resp.Data.Results)
			}
		})
	}
}

func TestLogPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LogPublicTestSuite))
}
