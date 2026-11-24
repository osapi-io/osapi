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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// SendPublicTestSuite tests how a typed method sends its request and reads
// the answer, through generated methods of each shape: a path parameter, a
// body, a query, and an answer with no body.
type SendPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *SendPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// serve is a client of a server that records the request and answers with
// handler.
func (s *SendPublicTestSuite) serve(
	handler http.HandlerFunc,
	seen *http.Request,
	body *string,
) *client.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		*seen = *r
		*body = string(data)
		handler(w, r)
	}))
	s.T().Cleanup(server.Close)

	return client.New(server.URL, "token")
}

func (s *SendPublicTestSuite) TestSend() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		call         func(c *client.Client) (any, error)
		validateFunc func(got any, err error, seen *http.Request, body string)
	}{
		{
			name: "when a read succeeds, the answer is the SDK's types",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set(
					"traceresponse",
					"00-26e448c9dbe1c4a8d216a98c675079c9-1a2b3c4d5e6f7a8b-01",
				)
				_, _ = w.Write([]byte(`{"job_id":"7c9e6679-7425-40de-944b-e07fc1f90ae7",` +
					`"results":[{"hostname":"web-01","status":"ok","servers":["1.1.1.1"]}]}`))
			},
			call: func(c *client.Client) (any, error) {
				return c.DNS.Get(s.ctx, "web-01", "eth 0")
			},
			validateFunc: func(got any, err error, seen *http.Request, body string) {
				s.Require().NoError(err)
				resp := got.(*client.Response[client.Collection[client.DnsEntry]])
				s.Equal("7c9e6679-7425-40de-944b-e07fc1f90ae7", resp.Data.JobID)
				r, ok := resp.Data.First()
				s.Require().True(ok)
				s.Equal(client.DnsEntry{
					Hostname: "web-01",
					Status:   "ok",
					Servers:  []string{"1.1.1.1"},
				}, r)
				s.Equal("26e448c9dbe1c4a8d216a98c675079c9", resp.TraceID())
				s.Contains(string(resp.RawJSON()), `"web-01"`)
				s.Equal(http.MethodGet, seen.Method)
				s.Equal("/api/node/web-01/dns/eth%200", seen.URL.EscapedPath())
				s.Empty(body)
			},
		},
		{
			name: "when a write has an input, its key goes in the path and the rest as JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"job_id":"j","results":[]}`))
			},
			call: func(c *client.Client) (any, error) {
				return c.DNS.Update(s.ctx, "_all", client.DNSUpdateInput{
					InterfaceName: "eth0",
					Servers:       []string{"9.9.9.9"},
				})
			},
			validateFunc: func(_ any, err error, seen *http.Request, body string) {
				s.Require().NoError(err)
				s.Equal(http.MethodPut, seen.Method)
				s.Equal("/api/node/_all/dns/eth0", seen.URL.Path)
				s.Equal("application/json", seen.Header.Get("Content-Type"))
				s.JSONEq(`{"servers":["9.9.9.9"]}`, body)
			},
		},
		{
			name: "when a method has a query, each parameter that is set goes in it",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"items":[],"total_items":0}`))
			},
			call: func(c *client.Client) (any, error) {
				return c.Job.List(s.ctx, client.JobListParams{Status: "failed", Limit: 5})
			},
			validateFunc: func(_ any, err error, seen *http.Request, _ string) {
				s.Require().NoError(err)
				s.Equal(url.Values{"status": {"failed"}, "limit": {"5"}}, seen.URL.Query())
			},
		},
		{
			name: "when the answer is not the type of the spec, it is an error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"job_id":7}`))
			},
			call: func(c *client.Client) (any, error) {
				return c.DNS.Get(s.ctx, "web-01", "eth0")
			},
			validateFunc: func(_ any, err error, _ *http.Request, _ string) {
				s.ErrorContains(err, "GET /api/node/web-01/dns/eth0: decode response")
			},
		},
		{
			name: "when the API answers 404, it is a NotFoundError with its message",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"no interface eth9","reason":"not_found"}`))
			},
			call: func(c *client.Client) (any, error) {
				return c.DNS.Get(s.ctx, "web-01", "eth9")
			},
			validateFunc: func(_ any, err error, _ *http.Request, _ string) {
				var notFound *client.NotFoundError
				s.Require().ErrorAs(err, &notFound)
				s.Equal("no interface eth9", notFound.Message)
				s.Equal("not_found", notFound.Reason)
			},
		},
		{
			name: "when an answer has no body, the method returns only the error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			call: func(c *client.Client) (any, error) {
				return nil, c.Agent.Drain(s.ctx, "web-01")
			},
			validateFunc: func(_ any, err error, seen *http.Request, _ string) {
				s.Require().NoError(err)
				s.Equal(http.MethodPost, seen.Method)
				s.Equal("/api/agent/web-01/drain", seen.URL.Path)
			},
		},
		{
			name: "when a method with no answer body fails, it is the SDK's error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":"already draining"}`))
			},
			call: func(c *client.Client) (any, error) {
				return nil, c.Agent.Drain(s.ctx, "web-01")
			},
			validateFunc: func(_ any, err error, _ *http.Request, _ string) {
				var conflict *client.ConflictError
				s.Require().ErrorAs(err, &conflict)
				s.Equal("already draining", conflict.Message)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var seen http.Request
			var body string
			c := s.serve(tc.handler, &seen, &body)
			got, err := tc.call(c)
			tc.validateFunc(got, err, &seen, body)
		})
	}
}

func (s *SendPublicTestSuite) TestSendJSONEncodeError() {
	c := client.New("http://127.0.0.1:0", "token")

	_, err := client.ExportSendJSON(s.ctx, c, http.MethodPut, "/api/x", nil, make(chan int))

	s.ErrorContains(err, "PUT /api/x: encode request")
}

func TestSendPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SendPublicTestSuite))
}
