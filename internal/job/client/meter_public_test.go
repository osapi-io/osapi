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
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type MeterPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *MeterPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	var err error
	s.jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)
}

func (s *MeterPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *MeterPublicTestSuite) TestSetMeterProvider() {
	tests := []struct {
		name string
	}{
		{
			name: "creates OTEL instruments without panic",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			promExp, err := promexporter.New()
			s.Require().NoError(err)

			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithReader(promExp),
			)
			defer func() { _ = mp.Shutdown(context.Background()) }()

			s.NotPanics(func() {
				s.jobsClient.SetMeterProvider(mp)
			})
		})
	}
}

func TestMeterPublicTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MeterPublicTestSuite))
}
