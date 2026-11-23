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
	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

type StatusPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *StatusPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *StatusPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.NodeStatus]], error)
	}{
		{
			name:   "when requesting status returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"web-01"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("web-01", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when a read failed the SDK carries which one",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"results":[{"hostname":"web-01","status":"ok","name":"web-01",` +
						`"uptime":{"uptime":"1 day","seconds":86400},"disks":[{"name":"/","total":1}],` +
						`"load_average":{"1min":1},"memory":{"total":1},"os_info":{"distribution":"ubuntu"},` +
						`"errors":["memory: permission denied"]}]}`),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Len(resp.Data.Results, 1)
				r := resp.Data.Results[0]
				suite.Equal([]string{"memory: permission denied"}, r.Errors)
				suite.Equal("web-01", r.Name)
				suite.Equal("1 day", r.Uptime)
				suite.Len(r.Disks, 1)
				suite.Equal("ubuntu", r.OSInfo.Distribution)
			},
		},
		{
			name:   "when every read succeeded the SDK reports no field errors",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"web-01"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Len(resp.Data.Results, 1)
				suite.Nil(resp.Data.Results[0].Errors)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get status")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Status.Get(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *StatusPublicTestSuite) TestConditionTypesInAPIEnum() {
	tests := []struct {
		name      string
		condition client.ConditionType
	}{
		{name: "MemoryPressure", condition: client.ConditionMemoryPressure},
		{name: "HighLoad", condition: client.ConditionHighLoad},
		{name: "DiskPressure", condition: client.ConditionDiskPressure},
		{name: "DrainStuck", condition: client.ConditionDrainStuck},
		{name: "ConsumersUnavailable", condition: client.ConditionConsumersUnavailable},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.True(gen.NodeConditionType(tc.condition).Valid())
		})
	}
}

func TestStatusPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StatusPublicTestSuite))
}
