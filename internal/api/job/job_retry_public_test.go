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
	"fmt"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobRetryPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobRetryPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobRetryPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobRetryPublicTestSuite) TestRetryJobByID() {
	targetHostname := "_any"

	tests := []struct {
		name         string
		request      gen.RetryJobByIDRequestObject
		mockResult   *client.CreateJobResult
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.RetryJobByIDResponseObject)
	}{
		{
			name: "success",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				Body: &gen.RetryJobByIDJSONRequestBody{
					TargetHostname: &targetHostname,
				},
			},
			mockResult: &client.CreateJobResult{
				JobID:     "660e8400-e29b-41d4-a716-446655440000",
				Status:    "created",
				Revision:  1,
				Timestamp: "2026-02-19T00:00:00Z",
			},
			expectMock: true,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				r, ok := resp.(gen.RetryJobByID201JSONResponse)
				s.True(ok)
				s.Equal("660e8400-e29b-41d4-a716-446655440000", r.JobId)
				s.Equal("created", r.Status)
			},
		},
		{
			name: "success with nil body",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			mockResult: &client.CreateJobResult{
				JobID:     "770e8400-e29b-41d4-a716-446655440000",
				Status:    "created",
				Revision:  1,
				Timestamp: "2026-02-19T00:00:00Z",
			},
			expectMock: true,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				r, ok := resp.(gen.RetryJobByID201JSONResponse)
				s.True(ok)
				s.Equal("770e8400-e29b-41d4-a716-446655440000", r.JobId)
			},
		},
		{
			name: "validation error empty target hostname",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				Body: func() *gen.RetryJobByIDJSONRequestBody {
					s := ""
					return &gen.RetryJobByIDJSONRequestBody{
						TargetHostname: &s,
					}
				}(),
			},
			expectMock: false,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				r, ok := resp.(gen.RetryJobByID400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
			},
		},
		{
			name: "not found",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("660e8400-e29b-41d4-a716-446655440000"),
			},
			mockError:  fmt.Errorf("job not found: 660e8400-e29b-41d4-a716-446655440000"),
			expectMock: true,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				_, ok := resp.(gen.RetryJobByID404JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "no operation data",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("770e8400-e29b-41d4-a716-446655440000"),
			},
			mockError: fmt.Errorf(
				"job has no operation data: 770e8400-e29b-41d4-a716-446655440000",
			),
			expectMock: true,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				r, ok := resp.(gen.RetryJobByID400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "no operation data")
			},
		},
		{
			name: "job client error",
			request: gen.RetryJobByIDRequestObject{
				Id: uuid.MustParse("880e8400-e29b-41d4-a716-446655440000"),
			},
			mockError:  fmt.Errorf("failed to create retry job: connection refused"),
			expectMock: true,
			validateFunc: func(resp gen.RetryJobByIDResponseObject) {
				_, ok := resp.(gen.RetryJobByID500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				var expectedTarget string
				if tt.request.Body != nil && tt.request.Body.TargetHostname != nil {
					expectedTarget = *tt.request.Body.TargetHostname
				}
				s.mockJobClient.EXPECT().
					RetryJob(gomock.Any(), tt.request.Id.String(), expectedTarget).
					Return(tt.mockResult, tt.mockError)
			}

			resp, err := s.handler.RetryJobByID(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobRetryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobRetryPublicTestSuite))
}
