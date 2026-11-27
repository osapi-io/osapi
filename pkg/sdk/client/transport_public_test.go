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
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type failingRoundTripper struct{}

func (f *failingRoundTripper) RoundTrip(
	_ *http.Request,
) (*http.Response, error) {
	return nil, fmt.Errorf("transport error")
}

// recordingRoundTripper keeps the request that it was sent.
type recordingRoundTripper struct {
	got *http.Request
}

func (r *recordingRoundTripper) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	r.got = req

	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

type TransportPublicTestSuite struct {
	suite.Suite
}

func (s *TransportPublicTestSuite) TestRoundTripError() {
	tests := []struct {
		name         string
		validateFunc func(*http.Response, error)
	}{
		{
			name: "when base transport fails returns error",
			validateFunc: func(resp *http.Response, err error) {
				s.Error(err)
				s.Contains(err.Error(), "transport error")
				s.Nil(resp)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			transport := client.ExportNewAuthTransport(
				&failingRoundTripper{},
				"Bearer test-token",
				slog.Default(),
			)

			req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
			s.Require().NoError(err)

			resp, err := transport.RoundTrip(req)
			tt.validateFunc(resp, err)
		})
	}
}

// A RoundTripper must not change the request that it is given, so a caller
// that sends the same request again, or a retry, sees it as it was.
func (s *TransportPublicTestSuite) TestRoundTripLeavesTheRequest() {
	tests := []struct {
		name         string
		validateFunc func(sent *http.Request, got *http.Request)
	}{
		{
			name: "when a request is sent the base gets the header and the caller's request has none",
			validateFunc: func(sent *http.Request, got *http.Request) {
				s.Require().NotNil(got)
				s.Equal("Bearer test-token", got.Header.Get("Authorization"))
				s.Empty(sent.Header.Get("Authorization"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			base := &recordingRoundTripper{}
			transport := client.ExportNewAuthTransport(base, "Bearer test-token", slog.Default())

			req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
			s.Require().NoError(err)

			resp, err := transport.RoundTrip(req)
			s.Require().NoError(err)
			_ = resp.Body.Close()

			tt.validateFunc(req, base.got)
		})
	}
}

func TestTransportPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TransportPublicTestSuite))
}
