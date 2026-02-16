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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobDeletePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobDeletePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobDeletePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobDeletePublicTestSuite) TestDeleteJobByID() {
	tests := []struct {
		name         string
		request      gen.DeleteJobByIDRequestObject
		mockError    error
		validateFunc func(resp gen.DeleteJobByIDResponseObject)
	}{
		{
			name:    "success",
			request: gen.DeleteJobByIDRequestObject{Id: "job-1"},
			validateFunc: func(resp gen.DeleteJobByIDResponseObject) {
				_, ok := resp.(gen.DeleteJobByID204Response)
				s.True(ok)
			},
		},
		{
			name:      "not found",
			request:   gen.DeleteJobByIDRequestObject{Id: "nonexistent"},
			mockError: fmt.Errorf("job not found: nonexistent"),
			validateFunc: func(resp gen.DeleteJobByIDResponseObject) {
				_, ok := resp.(gen.DeleteJobByID404JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "job client error",
			request:   gen.DeleteJobByIDRequestObject{Id: "job-1"},
			mockError: assert.AnError,
			validateFunc: func(resp gen.DeleteJobByIDResponseObject) {
				_, ok := resp.(gen.DeleteJobByID500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				DeleteJob(gomock.Any(), tt.request.Id).
				Return(tt.mockError)

			resp, err := s.handler.DeleteJobByID(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobDeletePublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobDeletePublicTestSuite))
}
