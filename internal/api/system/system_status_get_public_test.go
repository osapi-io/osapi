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

package system_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apisystem "github.com/retr0h/osapi/internal/api/system"
	"github.com/retr0h/osapi/internal/api/system/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type SystemStatusGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apisystem.System
	ctx           context.Context
}

func (s *SystemStatusGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apisystem.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *SystemStatusGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *SystemStatusGetPublicTestSuite) TestGetSystemStatus() {
	tests := []struct {
		name         string
		request      gen.GetSystemStatusRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetSystemStatusResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetSystemStatusRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemStatus(gomock.Any(), gomock.Any()).
					Return(&jobtypes.SystemStatusResponse{
						Hostname: "test-host",
						Uptime:   time.Hour,
					}, nil)
			},
			validateFunc: func(resp gen.GetSystemStatusResponseObject) {
				_, ok := resp.(gen.GetSystemStatus200JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.GetSystemStatusRequestObject{
				Params: gen.GetSystemStatusParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.GetSystemStatusResponseObject) {
				r, ok := resp.(gen.GetSystemStatus400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name:    "job client error",
			request: gen.GetSystemStatusRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemStatus(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetSystemStatusResponseObject) {
				_, ok := resp.(gen.GetSystemStatus500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.GetSystemStatusRequestObject{
				Params: gen.GetSystemStatusParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemStatusBroadcast(gomock.Any(), gomock.Any()).
					Return([]*jobtypes.SystemStatusResponse{
						{Hostname: "server1", Uptime: time.Hour},
						{Hostname: "server2", Uptime: 2 * time.Hour},
					}, nil)
			},
			validateFunc: func(resp gen.GetSystemStatusResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all error",
			request: gen.GetSystemStatusRequestObject{
				Params: gen.GetSystemStatusParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemStatusBroadcast(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetSystemStatusResponseObject) {
				_, ok := resp.(gen.GetSystemStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetSystemStatus(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestSystemStatusGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SystemStatusGetPublicTestSuite))
}
