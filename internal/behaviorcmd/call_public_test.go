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

package behaviorcmd_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	"github.com/osapi-io/osapi/internal/behaviorcmd/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type kvIn struct {
	Key    string `json:"key"`
	Prefix string `json:"prefix,omitempty"`
	Value  string `json:"value,omitempty"`
}

type kvOut struct {
	Key string `json:"key"`
}

func kvRead(_ context.Context, _ provider, _ kvIn) (kvOut, error) { return kvOut{}, nil }

var kv = behavior.Define("kv", "Key values", behavior.Node, behavior.API,
	behavior.Read("list", kvRead),
	behavior.Read("get", kvRead).Key("key"),
	behavior.Write("update", kvRead).Key("key"),
)

type CallPublicTestSuite struct {
	suite.Suite
}

func (s *CallPublicTestSuite) TestBuild() {
	tests := []struct {
		name       string
		method     string
		input      any
		wantMethod string
		wantPath   string
		wantBody   string
		wantErr    string
	}{
		{
			name:       "when a keyed read is built, the key goes in the path and the rest in the query",
			method:     "get",
			input:      kvIn{Key: "a/b", Prefix: "x"},
			wantMethod: "GET",
			wantPath:   "/api/next/node/web-01/kv/a%2Fb?prefix=x",
		},
		{
			name:       "when a read has no input, there is no query",
			method:     "list",
			wantMethod: "GET",
			wantPath:   "/api/next/node/web-01/kv",
		},
		{
			name:       "when a write is built, its input is the body",
			method:     "update",
			input:      kvIn{Key: "k", Value: "v"},
			wantMethod: "PUT",
			wantPath:   "/api/next/node/web-01/kv/k",
			wantBody:   `{"key":"k","value":"v"}`,
		},
		{
			name:    "when the input cannot be encoded",
			method:  "list",
			input:   map[string]any{"c": make(chan int)},
			wantErr: "encode input",
		},
		{
			name:    "when the input is not an object",
			method:  "list",
			input:   "text",
			wantErr: "must be an object",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			method, path, body, err := behaviorcmd.Build(
				kv,
				kv.Method(tc.method),
				"web-01",
				tc.input,
			)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantMethod, method)
			s.Equal(tc.wantPath, path)
			if tc.wantBody == "" {
				s.Nil(body)
			} else {
				s.JSONEq(tc.wantBody, string(body))
			}
		})
	}
}

func (s *CallPublicTestSuite) TestCall() {
	ok := []byte(`{"job_id":"j","results":[{"hostname":"web-01","key":"k"}]}`)
	tests := []struct {
		name    string
		input   any
		setup   func(c *mocks.MockCaller)
		want    behaviorcmd.Result
		wantErr string
	}{
		{
			name:  "when the API answers, the result is decoded",
			input: kvIn{Key: "k"},
			setup: func(c *mocks.MockCaller) {
				c.EXPECT().Request(gomock.Any(), "GET", "/api/next/node/web-01/kv/k", gomock.Nil()).
					Return(client.NewResponse(json.RawMessage(ok), ok), nil)
			},
			want: behaviorcmd.Result{
				JobID:   "j",
				Results: []map[string]any{{"hostname": "web-01", "key": "k"}},
			},
		},
		{
			name:    "when the request cannot be built, nothing is sent",
			input:   "text",
			setup:   func(*mocks.MockCaller) {},
			wantErr: "must be an object",
		},
		{
			name:  "when the request fails",
			input: kvIn{Key: "k"},
			setup: func(c *mocks.MockCaller) {
				c.EXPECT().Request(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unreachable"))
			},
			wantErr: "unreachable",
		},
		{
			name:  "when the answer is not a result",
			input: kvIn{Key: "k"},
			setup: func(c *mocks.MockCaller) {
				c.EXPECT().Request(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(client.NewResponse(json.RawMessage(`[]`), []byte(`[]`)), nil)
			},
			wantErr: "decode response",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := mocks.NewMockCaller(gomock.NewController(s.T()))
			tc.setup(c)

			got, raw, err := behaviorcmd.Call(
				context.Background(), c, kv, kv.Method("get"), "web-01", tc.input,
			)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
			s.Equal(ok, raw)
		})
	}
}

func TestCallPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CallPublicTestSuite))
}
