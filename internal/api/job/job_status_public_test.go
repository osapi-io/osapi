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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobStatusPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobStatusPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobStatusPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobStatusPublicTestSuite) TestGetJobStatus() {
	tests := []struct {
		name         string
		mockStats    *jobtypes.QueueStats
		mockError    error
		validateFunc func(resp gen.GetJobStatusResponseObject)
	}{
		{
			name: "success",
			mockStats: &jobtypes.QueueStats{
				TotalJobs: 42,
				StatusCounts: map[string]int{
					"completed": 30,
					"failed":    5,
				},
				OperationCounts: map[string]int{
					"system.hostname.get": 15,
				},
				DLQCount: 2,
			},
			validateFunc: func(resp gen.GetJobStatusResponseObject) {
				r, ok := resp.(gen.GetJobStatus200JSONResponse)
				s.True(ok)
				s.Equal(42, *r.TotalJobs)
				s.Equal(2, *r.DlqCount)
			},
		},
		{
			name:      "job client error",
			mockError: assert.AnError,
			validateFunc: func(resp gen.GetJobStatusResponseObject) {
				_, ok := resp.(gen.GetJobStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				GetQueueStats(gomock.Any()).
				Return(tt.mockStats, tt.mockError)

			resp, err := s.handler.GetJobStatus(s.ctx, gen.GetJobStatusRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobStatusPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobStatusPublicTestSuite))
}
