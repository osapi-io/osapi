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

package nats_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health/mocks"
	"github.com/osapi-io/osapi/internal/controller/endpoints/nats"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestStatus() {
	ctrl := gomock.NewController(s.T())
	reporter := mocks.NewMockReporter(ctrl)
	reporter.EXPECT().
		Status(gomock.Any()).
		Return(health.StatusResponse{Registry: []health.ComponentEntry{
			{
				Type:       "controller",
				Hostname:   "hub",
				Components: []health.ComponentEntry{{Type: "api", Status: "ok"}},
			},
			{Type: "nats", Hostname: "hub"},
			{Type: "agent", Hostname: "web-01"},
			{Type: "agent", Hostname: "web-02"},
		}})

	got, err := nats.Status(context.Background(), nats.Deps{Health: reporter}, nats.StatusInput{})

	s.Require().NoError(err)
	s.Require().Len(got.Items, 1, "only the nats components")
	s.Equal("nats", got.Items[0].Type)
	s.Equal("hub", got.Items[0].Hostname)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
