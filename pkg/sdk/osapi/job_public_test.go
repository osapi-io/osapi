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
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/osapi"
)

type JobPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *JobPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *JobPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		operation    map[string]interface{}
		target       string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when creating job returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
					),
				)
			},
			operation: map[string]interface{}{"type": "system.hostname.get"},
			target:    "_any",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.JobID)
				suite.Equal("pending", resp.Data.Status)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"validation failed"}`))
			},
			operation: map[string]interface{}{},
			target:    "_any",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:      "when HTTP request fails returns error",
			serverURL: "http://127.0.0.1:0",
			operation: map[string]interface{}{},
			target:    "_any",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "create job")
			},
		},
		{
			name: "when server returns 201 with empty body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			operation: map[string]interface{}{},
			target:    "_any",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				server    *httptest.Server
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
			} else {
				server = httptest.NewServer(tc.handler)
				defer server.Close()
				serverURL = server.URL
			}

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Create(suite.ctx, tc.operation, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		id           string
		validateFunc func(*osapi.Response[osapi.JobDetail], error)
	}{
		{
			name: "when valid UUID returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"id":"550e8400-e29b-41d4-a716-446655440000","status":"completed"}`,
					),
				)
			},
			id: "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.ID)
				suite.Equal("completed", resp.Data.Status)
			},
		},
		{
			name: "when invalid UUID returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"id":"550e8400-e29b-41d4-a716-446655440000","status":"completed"}`,
					),
				)
			},
			id: "not-a-uuid",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
		{
			name:      "when HTTP request fails returns error",
			serverURL: "http://127.0.0.1:0",
			id:        "00000000-0000-0000-0000-000000000000",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get job")
			},
		},
		{
			name: "when server returns 200 with empty body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			id: "00000000-0000-0000-0000-000000000000",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
		{
			name: "when server returns 404 returns NotFoundError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"job not found"}`))
			},
			id: "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
				suite.Equal("job not found", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				server    *httptest.Server
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
			} else {
				server = httptest.NewServer(tc.handler)
				defer server.Close()
				serverURL = server.URL
			}

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Get(suite.ctx, tc.id)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		id           string
		validateFunc func(error)
	}{
		{
			name: "when valid UUID returns no error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			id: "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name: "when invalid UUID returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			id: "not-a-uuid",
			validateFunc: func(err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
		{
			name:      "when HTTP request fails returns error",
			serverURL: "http://127.0.0.1:0",
			id:        "00000000-0000-0000-0000-000000000000",
			validateFunc: func(err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "delete job")
			},
		},
		{
			name: "when server returns 404 returns NotFoundError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"job not found"}`))
			},
			id: "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(err error) {
				suite.Error(err)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
				suite.Equal("job not found", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				server    *httptest.Server
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
			} else {
				server = httptest.NewServer(tc.handler)
				defer server.Close()
				serverURL = server.URL
			}

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			err := sut.Job.Delete(suite.ctx, tc.id)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		params       osapi.ListParams
		validateFunc func(*osapi.Response[osapi.JobList], error)
	}{
		{
			name: "when no filters returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"items":[],"total_items":0}`))
			},
			params: osapi.ListParams{},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(0, resp.Data.TotalItems)
				suite.Empty(resp.Data.Items)
			},
		},
		{
			name: "when all filters provided returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"items":[],"total_items":0}`))
			},
			params: osapi.ListParams{
				Status: "completed",
				Limit:  10,
				Offset: 5,
			},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:      "when HTTP request fails returns error",
			serverURL: "http://127.0.0.1:0",
			params:    osapi.ListParams{},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "list jobs")
			},
		},
		{
			name: "when server returns 401 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			},
			params: osapi.ListParams{},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusUnauthorized, target.StatusCode)
			},
		},
		{
			name: "when server returns 200 with empty body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			params: osapi.ListParams{},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				server    *httptest.Server
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
			} else {
				server = httptest.NewServer(tc.handler)
				defer server.Close()
				serverURL = server.URL
			}

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.List(suite.ctx, tc.params)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetry() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		id           string
		target       string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when valid UUID with empty target returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
					),
				)
			},
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.JobID)
				suite.Equal("pending", resp.Data.Status)
			},
		},
		{
			name: "when valid UUID with target returns response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
					),
				)
			},
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "web-01",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name: "when invalid UUID returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
					),
				)
			},
			id:     "not-a-uuid",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
		{
			name:      "when HTTP request fails returns error",
			serverURL: "http://127.0.0.1:0",
			id:        "00000000-0000-0000-0000-000000000000",
			target:    "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "retry job")
			},
		},
		{
			name: "when server returns 404 returns NotFoundError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"job not found"}`))
			},
			id:     "00000000-0000-0000-0000-000000000000",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name: "when server returns 201 with empty body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			id:     "00000000-0000-0000-0000-000000000000",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				server    *httptest.Server
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
			} else {
				server = httptest.NewServer(tc.handler)
				defer server.Close()
				serverURL = server.URL
			}

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Retry(suite.ctx, tc.id, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func TestJobPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobPublicTestSuite))
}
