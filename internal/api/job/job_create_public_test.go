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
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobCreatePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobCreatePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobCreatePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobCreatePublicTestSuite) TestPostJob() {
	tests := []struct {
		name         string
		request      gen.PostJobRequestObject
		mockResult   *client.CreateJobResult
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.PostJobResponseObject)
	}{
		{
			name: "success",
			request: gen.PostJobRequestObject{
				Body: &gen.CreateJobRequest{
					Operation:      map[string]interface{}{"type": "system.hostname.get"},
					TargetHostname: "_any",
				},
			},
			mockResult: &client.CreateJobResult{
				JobID:     "550e8400-e29b-41d4-a716-446655440000",
				Status:    "created",
				Revision:  1,
				Timestamp: "2025-06-14T10:00:00Z",
			},
			expectMock: true,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob201JSONResponse)
				s.True(ok)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", r.JobId.String())
				s.Equal("created", r.Status)
			},
		},
		{
			name: "validation error missing operation",
			request: gen.PostJobRequestObject{
				Body: &gen.PostJobJSONRequestBody{
					TargetHostname: "_any",
				},
			},
			expectMock: false,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Operation")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error empty target hostname",
			request: gen.PostJobRequestObject{
				Body: &gen.PostJobJSONRequestBody{
					Operation: map[string]interface{}{"type": "test"},
				},
			},
			expectMock: false,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "job client error",
			request: gen.PostJobRequestObject{
				Body: &gen.CreateJobRequest{
					Operation:      map[string]interface{}{"type": "invalid"},
					TargetHostname: "_any",
				},
			},
			mockError:  assert.AnError,
			expectMock: true,
			validateFunc: func(resp gen.PostJobResponseObject) {
				_, ok := resp.(gen.PostJob500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				s.mockJobClient.EXPECT().
					CreateJob(gomock.Any(), tt.request.Body.Operation, tt.request.Body.TargetHostname).
					Return(tt.mockResult, tt.mockError)
			}

			resp, err := s.handler.PostJob(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobCreatePublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobCreatePublicTestSuite))
}
