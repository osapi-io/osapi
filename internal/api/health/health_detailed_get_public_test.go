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

package health_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api/health"
	"github.com/retr0h/osapi/internal/api/health/gen"
)

type stubChecker struct{}

func (s *stubChecker) CheckHealth(
	_ context.Context,
) error {
	return nil
}

type HealthDetailedGetPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *HealthDetailedGetPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *HealthDetailedGetPublicTestSuite) TestGetHealthDetailed() {
	tests := []struct {
		name         string
		checker      health.Checker
		validateFunc func(resp gen.GetHealthDetailedResponseObject)
	}{
		{
			name: "all components healthy",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("ok", r.Components["kv"].Status)
				s.Equal("0.1.0", r.Version)
				s.NotEmpty(r.Uptime)
			},
		},
		{
			name: "NATS unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("NATS not connected") },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("error", r.Components["nats"].Status)
				s.Require().NotNil(r.Components["nats"].Error)
				s.Contains(*r.Components["nats"].Error, "NATS not connected")
				s.Equal("ok", r.Components["kv"].Status)
			},
		},
		{
			name: "KV unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return fmt.Errorf("KV bucket not accessible") },
			},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("error", r.Components["kv"].Status)
				s.Require().NotNil(r.Components["kv"].Error)
				s.Contains(*r.Components["kv"].Error, "KV bucket not accessible")
			},
		},
		{
			name: "both unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("NATS down") },
				KVCheck:   func() error { return fmt.Errorf("KV down") },
			},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("error", r.Components["nats"].Status)
				s.Equal("error", r.Components["kv"].Status)
			},
		},
		{
			name: "includes version and uptime",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed200JSONResponse)
				s.True(ok)
				s.Equal("0.1.0", r.Version)
				s.NotEmpty(r.Uptime)
			},
		},
		{
			name:    "non-NATSChecker returns ok with nil components",
			checker: &stubChecker{},
			validateFunc: func(resp gen.GetHealthDetailedResponseObject) {
				r, ok := resp.(gen.GetHealthDetailed200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("ok", r.Components["kv"].Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := health.New(tt.checker, time.Now(), "0.1.0")

			resp, err := handler.GetHealthDetailed(s.ctx, gen.GetHealthDetailedRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestHealthDetailedGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HealthDetailedGetPublicTestSuite))
}
