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
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobListPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobListPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobListPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobListPublicTestSuite) TestGetJob() {
	completedStatus := gen.Completed
	invalidStatus := gen.GetJobParamsStatus("bogus")

	tests := []struct {
		name         string
		request      gen.GetJobRequestObject
		mockResult   *jobclient.ListJobsResult
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.GetJobResponseObject)
	}{
		{
			name: "validation error invalid status",
			request: gen.GetJobRequestObject{
				Params: gen.GetJobParams{Status: &invalidStatus},
			},
			expectMock: false,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "invalid status filter")
			},
		},
		{
			name: "success with filter",
			request: gen.GetJobRequestObject{
				Params: gen.GetJobParams{Status: &completedStatus},
			},
			mockResult: &jobclient.ListJobsResult{
				Jobs: []*jobtypes.QueuedJob{
					{
						ID:     "550e8400-e29b-41d4-a716-446655440000",
						Status: "completed",
					},
				},
				TotalCount: 1,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob200JSONResponse)
				s.True(ok)
				s.Equal(1, *r.TotalItems)
				s.Len(*r.Items, 1)
			},
		},
		{
			name:    "success without filter",
			request: gen.GetJobRequestObject{},
			mockResult: &jobclient.ListJobsResult{
				Jobs: []*jobtypes.QueuedJob{
					{ID: "550e8400-e29b-41d4-a716-446655440000", Status: "completed"},
					{ID: "660e8400-e29b-41d4-a716-446655440000", Status: "processing"},
				},
				TotalCount: 2,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob200JSONResponse)
				s.True(ok)
				s.Equal(2, *r.TotalItems)
			},
		},
		{
			name:    "success with all optional fields",
			request: gen.GetJobRequestObject{},
			mockResult: &jobclient.ListJobsResult{
				Jobs: []*jobtypes.QueuedJob{
					{
						ID:        "550e8400-e29b-41d4-a716-446655440000",
						Status:    "failed",
						Created:   "2025-06-14T10:00:00Z",
						Operation: map[string]interface{}{"type": "network.dns.get"},
						Error:     "timeout",
						Hostname:  "worker-2",
						UpdatedAt: "2025-06-14T10:05:00Z",
						Result:    json.RawMessage(`{"servers":["8.8.8.8"]}`),
					},
				},
				TotalCount: 1,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob200JSONResponse)
				s.True(ok)
				s.Equal(1, *r.TotalItems)
				item := (*r.Items)[0]
				s.Equal("550e8400-e29b-41d4-a716-446655440000", item.Id.String())
				s.NotNil(item.Operation)
				s.NotNil(item.Error)
				s.Equal("timeout", *item.Error)
				s.NotNil(item.Hostname)
				s.Equal("worker-2", *item.Hostname)
				s.NotNil(item.UpdatedAt)
				s.NotNil(item.Result)
			},
		},
		{
			name: "explicit limit and offset params",
			request: func() gen.GetJobRequestObject {
				limit := 5
				offset := 20
				return gen.GetJobRequestObject{
					Params: gen.GetJobParams{Limit: &limit, Offset: &offset},
				}
			}(),
			mockResult: &jobclient.ListJobsResult{
				Jobs:       []*jobtypes.QueuedJob{},
				TotalCount: 50,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob200JSONResponse)
				s.True(ok)
				s.Equal(50, *r.TotalItems)
			},
		},
		{
			name:       "job client error",
			request:    gen.GetJobRequestObject{},
			mockError:  assert.AnError,
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				_, ok := resp.(gen.GetJob500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "total items reflects total count not page size",
			request: gen.GetJobRequestObject{},
			mockResult: &jobclient.ListJobsResult{
				Jobs: []*jobtypes.QueuedJob{
					{ID: "550e8400-e29b-41d4-a716-446655440000", Status: "completed"},
				},
				TotalCount: 50,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobResponseObject) {
				r, ok := resp.(gen.GetJob200JSONResponse)
				s.True(ok)
				s.Equal(50, *r.TotalItems)
				s.Len(*r.Items, 1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				var statusFilter string
				if tt.request.Params.Status != nil {
					statusFilter = string(*tt.request.Params.Status)
				}
				expectedLimit := 10
				if tt.request.Params.Limit != nil {
					expectedLimit = *tt.request.Params.Limit
				}
				expectedOffset := 0
				if tt.request.Params.Offset != nil {
					expectedOffset = *tt.request.Params.Offset
				}
				s.mockJobClient.EXPECT().
					ListJobs(gomock.Any(), statusFilter, expectedLimit, expectedOffset).
					Return(tt.mockResult, tt.mockError)
			}

			resp, err := s.handler.GetJob(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobListPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobListPublicTestSuite))
}
