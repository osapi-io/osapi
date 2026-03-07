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

package osapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type errorReader struct{}

func (e *errorReader) Read(
	_ []byte,
) (int, error) {
	return 0, fmt.Errorf("read error")
}

func (e *errorReader) Close() error {
	return nil
}

type MetricsTestSuite struct {
	suite.Suite
}

func (s *MetricsTestSuite) TestGetReadBodyError() {
	tests := []struct {
		name         string
		validateFunc func(string, error)
	}{
		{
			name: "when body read fails returns error",
			validateFunc: func(body string, err error) {
				s.Error(err)
				s.Contains(err.Error(), "reading metrics response")
				s.Empty(body)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusOK)
				}),
			)
			defer server.Close()

			sut := &MetricsService{
				baseURL: server.URL,
			}

			origTransport := http.DefaultTransport
			http.DefaultTransport = &readErrorTransport{}
			defer func() { http.DefaultTransport = origTransport }()

			body, err := sut.Get(context.Background())
			tt.validateFunc(body, err)
		})
	}
}

type readErrorTransport struct{}

func (t *readErrorTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&errorReader{}),
		Request:    req,
	}, nil
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}
