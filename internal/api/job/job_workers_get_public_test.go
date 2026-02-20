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

package job_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobWorkersGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobWorkersGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobWorkersGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobWorkersGetPublicTestSuite) TestGetJobWorkers() {
	tests := []struct {
		name         string
		mockWorkers  []jobtypes.WorkerInfo
		mockError    error
		validateFunc func(resp gen.GetJobWorkersResponseObject)
	}{
		{
			name: "success with workers",
			mockWorkers: []jobtypes.WorkerInfo{
				{Hostname: "server1"},
				{Hostname: "server2"},
			},
			validateFunc: func(resp gen.GetJobWorkersResponseObject) {
				r, ok := resp.(gen.GetJobWorkers200JSONResponse)
				s.True(ok)
				s.Equal(2, r.Total)
				s.Len(r.Workers, 2)
				s.Equal("server1", r.Workers[0].Hostname)
				s.Equal("server2", r.Workers[1].Hostname)
			},
		},
		{
			name:        "success with no workers",
			mockWorkers: []jobtypes.WorkerInfo{},
			validateFunc: func(resp gen.GetJobWorkersResponseObject) {
				r, ok := resp.(gen.GetJobWorkers200JSONResponse)
				s.True(ok)
				s.Equal(0, r.Total)
				s.Empty(r.Workers)
			},
		},
		{
			name:      "job client error",
			mockError: assert.AnError,
			validateFunc: func(resp gen.GetJobWorkersResponseObject) {
				_, ok := resp.(gen.GetJobWorkers500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				ListWorkers(gomock.Any()).
				Return("", tt.mockWorkers, tt.mockError)

			resp, err := s.handler.GetJobWorkers(s.ctx, gen.GetJobWorkersRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobWorkersGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobWorkersGetPublicTestSuite))
}
