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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api/health"
	"github.com/retr0h/osapi/internal/api/health/gen"
)

type HealthGetPublicTestSuite struct {
	suite.Suite

	handler *health.Health
	ctx     context.Context
}

func (s *HealthGetPublicTestSuite) SetupTest() {
	s.handler = health.New(
		&health.NATSChecker{},
		time.Now(),
		"0.1.0",
	)
	s.ctx = context.Background()
}

func (s *HealthGetPublicTestSuite) TestGetHealth() {
	tests := []struct {
		name         string
		validateFunc func(resp gen.GetHealthResponseObject)
	}{
		{
			name: "returns ok status",
			validateFunc: func(resp gen.GetHealthResponseObject) {
				r, ok := resp.(gen.GetHealth200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := s.handler.GetHealth(s.ctx, gen.GetHealthRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestHealthGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HealthGetPublicTestSuite))
}
