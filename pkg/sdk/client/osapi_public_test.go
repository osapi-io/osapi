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

	"github.com/osapi-io/osapi/pkg/sdk/client"
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
	customTransport := &http.Transport{}

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
				suite.NotNil(c.Status)
				suite.NotNil(c.Hostname)
				suite.NotNil(c.NetworkDNS)
				suite.NotNil(c.NodeFile)
				suite.NotNil(c.File)
				suite.NotNil(c.Job)
				suite.NotNil(c.Health)
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
		{
			name: "when no transport provided uses its own transport",
			opts: func() []client.Option {
				return nil
			},
			validateFunc: func(c *client.Client) {
				base := client.ExportBaseTransport(c)
				suite.IsType(&http.Transport{}, base)
				suite.NotSame(http.DefaultTransport, base)
			},
		},
		{
			name: "when custom transport provided uses the caller transport",
			opts: func() []client.Option {
				return []client.Option{
					client.WithHTTPTransport(customTransport),
				}
			},
			validateFunc: func(c *client.Client) {
				suite.Same(customTransport, client.ExportBaseTransport(c))
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

func (suite *ClientPublicTestSuite) TestCloneTransport() {
	tests := []struct {
		name         string
		base         func() http.RoundTripper
		validateFunc func(base, got http.RoundTripper)
	}{
		{
			name: "when base is an http.Transport returns a clone",
			base: func() http.RoundTripper {
				return &http.Transport{MaxIdleConns: 7}
			},
			validateFunc: func(base, got http.RoundTripper) {
				suite.NotSame(base, got)
				suite.Require().IsType(&http.Transport{}, got)
				suite.Equal(7, got.(*http.Transport).MaxIdleConns)
			},
		},
		{
			name: "when base is another RoundTripper returns it as is",
			base: func() http.RoundTripper {
				return roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return nil, nil
				})
			},
			validateFunc: func(base, got http.RoundTripper) {
				suite.NotNil(got)
				suite.IsType(base, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			base := tc.base()
			tc.validateFunc(base, client.ExportCloneTransport(base))
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(
	r *http.Request,
) (*http.Response, error) {
	return f(r)
}

func TestClientPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ClientPublicTestSuite))
}
