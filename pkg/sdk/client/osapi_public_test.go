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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

type ClientPublicTestSuite struct {
	suite.Suite

	server *httptest.Server
}

func (suite *ClientPublicTestSuite) SetupTest() {
	suite.server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}),
	)
}

func (suite *ClientPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *ClientPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		opts         func() []client.Option
		validateFunc func(*client.Client)
	}{
		{
			name: "when creating client returns all services",
			opts: func() []client.Option {
				return nil
			},
			validateFunc: func(c *client.Client) {
				suite.NotNil(c)
				suite.NotNil(c.Node)
				suite.NotNil(c.Job)
				suite.NotNil(c.Health)
				suite.NotNil(c.Audit)
				suite.NotNil(c.File)
			},
		},
		{
			name: "when custom transport provided creates client",
			opts: func() []client.Option {
				return []client.Option{
					client.WithHTTPTransport(&http.Transport{}),
					client.WithLogger(slog.Default()),
				}
			},
			validateFunc: func(c *client.Client) {
				suite.NotNil(c)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			c := client.New(suite.server.URL, "test-token", tc.opts()...)
			tc.validateFunc(c)
		})
	}
}

func TestClientPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ClientPublicTestSuite))
}
