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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type RequestPublicTestSuite struct {
	suite.Suite
}

func (s *RequestPublicTestSuite) TestRequest() {
	tests := []struct {
		name         string
		baseURL      string
		method       string
		path         string
		body         []byte
		handler      http.HandlerFunc
		validateFunc func(resp *client.Response[json.RawMessage], err error, seen *http.Request, body string)
	}{
		{
			name:   "when a read succeeds, the body is returned as it came",
			method: http.MethodGet,
			path:   "/api/node/_all/kv?prefix=a",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"job_id":"j","results":[]}`))
			},
			validateFunc: func(resp *client.Response[json.RawMessage], err error, seen *http.Request, _ string) {
				s.Require().NoError(err)
				s.JSONEq(`{"job_id":"j","results":[]}`, string(resp.Data))
				s.Equal(`{"job_id":"j","results":[]}`, string(resp.RawJSON()))
				s.Equal("/api/node/_all/kv?prefix=a", seen.URL.RequestURI())
				s.Equal("Bearer token", seen.Header.Get("Authorization"))
				s.Empty(seen.Header.Get("Content-Type"))
			},
		},
		{
			name:   "when a write sends a body, it goes as JSON",
			method: http.MethodPut,
			body:   []byte(`{"value":"v"}`),
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{}`))
			},
			validateFunc: func(_ *client.Response[json.RawMessage], err error, seen *http.Request, body string) {
				s.Require().NoError(err)
				s.Equal(http.MethodPut, seen.Method)
				s.Equal("application/json", seen.Header.Get("Content-Type"))
				s.JSONEq(`{"value":"v"}`, body)
			},
		},
		{
			name:   "when the API refuses the request, it is the SDK's error",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"bad key"}`))
			},
			validateFunc: func(_ *client.Response[json.RawMessage], err error, _ *http.Request, _ string) {
				var v *client.ValidationError
				s.ErrorAs(err, &v)
				s.ErrorContains(err, "bad key")
			},
		},
		{
			name:   "when the response body is cut short",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "100")
				_, _ = w.Write([]byte(`{"res`))
			},
			validateFunc: func(_ *client.Response[json.RawMessage], err error, _ *http.Request, _ string) {
				s.ErrorContains(err, "read response")
			},
		},
		{
			name:    "when the URL is not valid",
			baseURL: "http://bad\x7f",
			method:  http.MethodGet,
			validateFunc: func(_ *client.Response[json.RawMessage], err error, _ *http.Request, _ string) {
				s.ErrorContains(err, "GET /api/node/_all/kv")
			},
		},
		{
			name:    "when the server cannot be reached",
			baseURL: "http://127.0.0.1:1",
			method:  http.MethodGet,
			validateFunc: func(_ *client.Response[json.RawMessage], err error, _ *http.Request, _ string) {
				s.ErrorContains(err, "GET /api/node/_all/kv")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var (
				seen *http.Request
				body string
			)
			url := tc.baseURL
			if tc.handler != nil {
				srv := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						raw, _ := io.ReadAll(r.Body)
						seen, body = r, string(raw)
						tc.handler(w, r)
					}),
				)
				defer srv.Close()
				url = srv.URL
			}

			path := tc.path
			if path == "" {
				path = "/api/node/_all/kv"
			}
			resp, err := client.New(url, "token").
				Request(context.Background(), tc.method, path, tc.body)
			tc.validateFunc(resp, err, seen, body)
		})
	}
}

func TestRequestPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RequestPublicTestSuite))
}
