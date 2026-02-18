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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apisystem "github.com/retr0h/osapi/internal/api/system"
	"github.com/retr0h/osapi/internal/api/system/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type SystemHostnameGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apisystem.System
	ctx           context.Context
}

func (s *SystemHostnameGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apisystem.New(s.mockJobClient)
	s.ctx = context.Background()
}

func (s *SystemHostnameGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *SystemHostnameGetPublicTestSuite) TestGetSystemHostname() {
	tests := []struct {
		name         string
		request      gen.GetSystemHostnameRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetSystemHostnameResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetSystemHostnameRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemHostname(gomock.Any(), gomock.Any()).
					Return("my-hostname", nil)
			},
			validateFunc: func(resp gen.GetSystemHostnameResponseObject) {
				r, ok := resp.(gen.GetSystemHostname200JSONResponse)
				s.True(ok)
				s.Equal("my-hostname", r.Hostname)
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.GetSystemHostnameRequestObject{
				Params: gen.GetSystemHostnameParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.GetSystemHostnameResponseObject) {
				r, ok := resp.(gen.GetSystemHostname400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name:    "job client error",
			request: gen.GetSystemHostnameRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemHostname(gomock.Any(), gomock.Any()).
					Return("", assert.AnError)
			},
			validateFunc: func(resp gen.GetSystemHostnameResponseObject) {
				_, ok := resp.(gen.GetSystemHostname500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.GetSystemHostnameRequestObject{
				Params: gen.GetSystemHostnameParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemHostnameAll(gomock.Any()).
					Return(map[string]string{
						"server1": "host1",
						"server2": "host2",
					}, nil)
			},
			validateFunc: func(resp gen.GetSystemHostnameResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all error",
			request: gen.GetSystemHostnameRequestObject{
				Params: gen.GetSystemHostnameParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QuerySystemHostnameAll(gomock.Any()).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetSystemHostnameResponseObject) {
				_, ok := resp.(gen.GetSystemHostname500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetSystemHostname(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func strPtr(
	s string,
) *string {
	return &s
}

func TestSystemHostnameGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SystemHostnameGetPublicTestSuite))
}
