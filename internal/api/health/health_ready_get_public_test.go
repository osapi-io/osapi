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
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api/health"
	"github.com/retr0h/osapi/internal/api/health/gen"
)

type HealthReadyGetPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *HealthReadyGetPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *HealthReadyGetPublicTestSuite) TestGetHealthReady() {
	tests := []struct {
		name         string
		checker      health.Checker
		validateFunc func(resp gen.GetHealthReadyResponseObject)
	}{
		{
			name: "ready when all checks pass",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthReadyResponseObject) {
				r, ok := resp.(gen.GetHealthReady200JSONResponse)
				s.True(ok)
				s.Equal("ready", r.Status)
			},
		},
		{
			name: "not ready when NATS check fails",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("nats not connected") },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthReadyResponseObject) {
				r, ok := resp.(gen.GetHealthReady503JSONResponse)
				s.True(ok)
				s.Equal("not_ready", r.Status)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "nats not connected")
			},
		},
		{
			name: "not ready when KV check fails",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return fmt.Errorf("kv bucket not accessible") },
			},
			validateFunc: func(resp gen.GetHealthReadyResponseObject) {
				r, ok := resp.(gen.GetHealthReady503JSONResponse)
				s.True(ok)
				s.Equal("not_ready", r.Status)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "kv bucket not accessible")
			},
		},
		{
			name: "not ready when both checks fail",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("nats down") },
				KVCheck:   func() error { return fmt.Errorf("kv down") },
			},
			validateFunc: func(resp gen.GetHealthReadyResponseObject) {
				r, ok := resp.(gen.GetHealthReady503JSONResponse)
				s.True(ok)
				s.Equal("not_ready", r.Status)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "nats down")
				s.Contains(*r.Error, "kv down")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := health.New(slog.Default(), tt.checker, time.Now(), "0.1.0", nil)

			resp, err := handler.GetHealthReady(s.ctx, gen.GetHealthReadyRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestHealthReadyGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HealthReadyGetPublicTestSuite))
}
