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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type ProcessPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *ProcessPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

// reply is a handler that answers with status and body.
func reply(
	status int,
	body string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if body != "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// serve runs call against handler, or against serverURL when it is set.
func (suite *ProcessPublicTestSuite) serve(
	handler http.HandlerFunc,
	serverURL string,
	call func(c *client.Client) error,
) error {
	if serverURL == "" {
		server := httptest.NewServer(handler)
		defer server.Close()
		serverURL = server.URL
	}

	return call(client.New(serverURL, "test-token", client.WithLogger(slog.Default())))
}

const processRow = `{"job_id":"00000000-0000-0000-0000-000000000001","results":[` +
	`{"hostname":"agent1","status":"ok","pid":1234,"name":"sleep","user":"alice",` +
	`"state":"sleep","cpu_percent":5,"mem_percent":1,"mem_rss":20971520,` +
	`"command":"sleep 300","start_time":"2026-01-02T02:55:45Z"}]}`

var sleeping = client.ProcessInfoResult{
	Hostname: "agent1", Status: "ok", PID: 1234, Name: "sleep", User: "alice",
	State: "sleep", CPUPercent: 5, MemPercent: 1, MemRSS: 20971520,
	Command: "sleep 300", StartTime: "2026-01-02T02:55:45Z",
}

// failure is an answer that each method turns into an error, and a test of
// that error.
type failure struct {
	name      string
	handler   http.HandlerFunc
	serverURL string
	is        func(err error) bool
}

// failures are the failures of the method op, whose error names op when
// the call does not reach the server.
func failures(
	op string,
) []failure {
	return []failure{
		{
			"403", reply(http.StatusForbidden, `{"error":"forbidden"}`), "",
			func(err error) bool { var t *client.AuthError; return errors.As(err, &t) },
		},
		{
			"404", reply(http.StatusNotFound, `{"error":"not found"}`), "",
			func(err error) bool { var t *client.NotFoundError; return errors.As(err, &t) },
		},
		{
			"500", reply(http.StatusInternalServerError, `{"error":"boom"}`), "",
			func(err error) bool { var t *client.ServerError; return errors.As(err, &t) },
		},
		{
			"200 with no body", reply(http.StatusOK, ""), "",
			func(err error) bool { var t *client.UnexpectedStatusError; return errors.As(err, &t) },
		},
		{
			"no server", nil, "http://127.0.0.1:0",
			func(err error) bool { return strings.Contains(err.Error(), op) },
		},
	}
}

func (suite *ProcessPublicTestSuite) TestList() {
	err := suite.serve(reply(http.StatusOK, processRow), "", func(c *client.Client) error {
		resp, err := c.Process.List(suite.ctx, "_any")
		suite.Require().NoError(err)
		suite.Equal("00000000-0000-0000-0000-000000000001", resp.Data.JobID)
		suite.Equal([]client.ProcessInfoResult{sleeping}, resp.Data.Results)

		return nil
	})
	suite.NoError(err)

	for _, tc := range failures("process list") {
		suite.Run(tc.name, func() {
			err := suite.serve(tc.handler, tc.serverURL, func(c *client.Client) error {
				resp, err := c.Process.List(suite.ctx, "_any")
				suite.Nil(resp)

				return err
			})
			suite.Require().Error(err)
			suite.True(tc.is(err), "%T: %v", err, err)
		})
	}
}

func (suite *ProcessPublicTestSuite) TestGet() {
	err := suite.serve(reply(http.StatusOK, processRow), "", func(c *client.Client) error {
		resp, err := c.Process.Get(suite.ctx, "_any", 1234)
		suite.Require().NoError(err)
		suite.Equal([]client.ProcessInfoResult{sleeping}, resp.Data.Results)

		return nil
	})
	suite.NoError(err)

	for _, tc := range failures("process get") {
		suite.Run(tc.name, func() {
			err := suite.serve(tc.handler, tc.serverURL, func(c *client.Client) error {
				resp, err := c.Process.Get(suite.ctx, "_any", 1234)
				suite.Nil(resp)

				return err
			})
			suite.Require().Error(err)
			suite.True(tc.is(err), "%T: %v", err, err)
		})
	}
}

func (suite *ProcessPublicTestSuite) TestSignal() {
	var body string
	handler := func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 64)
		n, _ := r.Body.Read(b)
		body = string(b[:n])
		reply(http.StatusOK, `{"results":[{"hostname":"agent1","status":"ok",`+
			`"pid":1234,"signal":"TERM","changed":true}]}`)(w, r)
	}
	err := suite.serve(handler, "", func(c *client.Client) error {
		resp, err := c.Process.Signal(
			suite.ctx,
			"_any",
			1234,
			client.ProcessSignalOpts{Signal: "TERM"},
		)
		suite.Require().NoError(err)
		suite.Equal([]client.ProcessSignalResult{
			{Hostname: "agent1", Status: "ok", PID: 1234, Signal: "TERM", Changed: true},
		}, resp.Data.Results)

		return nil
	})
	suite.NoError(err)
	suite.JSONEq(`{"signal":"TERM"}`, body)

	for _, tc := range failures("process signal") {
		suite.Run(tc.name, func() {
			err := suite.serve(tc.handler, tc.serverURL, func(c *client.Client) error {
				resp, err := c.Process.Signal(suite.ctx, "_any", 1234,
					client.ProcessSignalOpts{Signal: "TERM"})
				suite.Nil(resp)

				return err
			})
			suite.Require().Error(err)
			suite.True(tc.is(err), "%T: %v", err, err)
		})
	}
}

func TestProcessPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessPublicTestSuite))
}
