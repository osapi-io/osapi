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
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
}

func (s *JobGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *JobGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobGetPublicTestSuite) TestGetJobByID() {
	tests := []struct {
		name         string
		request      gen.GetJobByIDRequestObject
		mockJob      *jobtypes.QueuedJob
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.GetJobByIDResponseObject)
	}{
		{
			name: "success with basic fields",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:      "550e8400-e29b-41d4-a716-446655440000",
				Status:  "completed",
				Created: "2025-06-14T10:00:00Z",
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", r.Id.String())
				s.Equal("completed", *r.Status)
				s.Nil(r.Operation)
				s.Nil(r.Error)
				s.Nil(r.Hostname)
				s.Nil(r.UpdatedAt)
				s.Nil(r.Result)
			},
		},
		{
			name: "success with all optional fields",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("660e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:        "660e8400-e29b-41d4-a716-446655440000",
				Status:    "failed",
				Created:   "2025-06-14T10:00:00Z",
				Operation: map[string]interface{}{"type": "system.hostname.get"},
				Error:     "disk full",
				Hostname:  "worker-1",
				UpdatedAt: "2025-06-14T10:05:00Z",
				Result:    json.RawMessage(`{"hostname":"server-01"}`),
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Equal("660e8400-e29b-41d4-a716-446655440000", r.Id.String())
				s.Equal("failed", *r.Status)
				s.NotNil(r.Operation)
				s.Equal("system.hostname.get", (*r.Operation)["type"])
				s.NotNil(r.Error)
				s.Equal("disk full", *r.Error)
				s.NotNil(r.Hostname)
				s.Equal("worker-1", *r.Hostname)
				s.NotNil(r.UpdatedAt)
				s.Equal("2025-06-14T10:05:00Z", *r.UpdatedAt)
				s.NotNil(r.Result)
			},
		},
		{
			name: "not found",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("770e8400-e29b-41d4-a716-446655440000"),
			},
			mockError:  fmt.Errorf("job not found: 770e8400-e29b-41d4-a716-446655440000"),
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				_, ok := resp.(gen.GetJobByID404JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "job client error",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("880e8400-e29b-41d4-a716-446655440000"),
			},
			mockError:  assert.AnError,
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				_, ok := resp.(gen.GetJobByID500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast job with multiple responses",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("990e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:      "990e8400-e29b-41d4-a716-446655440000",
				Status:  "completed",
				Created: "2025-06-14T10:00:00Z",
				Responses: map[string]jobtypes.Response{
					"server1": {
						Status:   "completed",
						Hostname: "server1",
						Data:     json.RawMessage(`{"hostname":"server1"}`),
					},
					"server2": {
						Status:   "completed",
						Hostname: "server2",
						Data:     json.RawMessage(`{"hostname":"server2"}`),
					},
				},
				WorkerStates: map[string]jobtypes.WorkerState{
					"server1": {
						Status:   "completed",
						Duration: "1.5s",
					},
					"server2": {
						Status:   "completed",
						Duration: "2.1s",
					},
				},
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Equal("990e8400-e29b-41d4-a716-446655440000", r.Id.String())
				s.Equal("completed", *r.Status)
				s.NotNil(r.Responses)
				s.Len(*r.Responses, 2)
				s.NotNil(r.WorkerStates)
				s.Len(*r.WorkerStates, 2)
			},
		},
		{
			name: "single response omits responses map",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:      "aa0e8400-e29b-41d4-a716-446655440000",
				Status:  "completed",
				Created: "2025-06-14T10:00:00Z",
				Responses: map[string]jobtypes.Response{
					"server1": {
						Status:   "completed",
						Hostname: "server1",
						Data:     json.RawMessage(`{"hostname":"server1"}`),
					},
				},
				Result: json.RawMessage(`{"hostname":"server1"}`),
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Nil(r.Responses)
				s.NotNil(r.Result)
			},
		},
		{
			name: "worker states with errors",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:      "bb0e8400-e29b-41d4-a716-446655440000",
				Status:  "partial_failure",
				Created: "2025-06-14T10:00:00Z",
				Responses: map[string]jobtypes.Response{
					"server1": {
						Status:   "completed",
						Hostname: "server1",
						Data:     json.RawMessage(`{"hostname":"server1"}`),
					},
					"server2": {
						Status:   "failed",
						Hostname: "server2",
						Error:    "disk full",
					},
				},
				WorkerStates: map[string]jobtypes.WorkerState{
					"server1": {
						Status:   "completed",
						Duration: "1.5s",
					},
					"server2": {
						Status: "failed",
						Error:  "disk full",
					},
				},
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.NotNil(r.Responses)
				s.Len(*r.Responses, 2)
				s.NotNil(r.WorkerStates)
				ws := *r.WorkerStates
				s.NotNil(ws["server2"].Error)
				s.Equal("disk full", *ws["server2"].Error)
			},
		},
		{
			name: "success with timeline events",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:      "cc0e8400-e29b-41d4-a716-446655440000",
				Status:  "failed",
				Created: "2026-02-19T10:00:00Z",
				Timeline: []jobtypes.TimelineEvent{
					{
						Timestamp: time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC),
						Event:     "submitted",
						Hostname:  "_api",
						Message:   "Job submitted to queue",
					},
					{
						Timestamp: time.Date(2026, 2, 19, 10, 0, 1, 0, time.UTC),
						Event:     "acknowledged",
						Hostname:  "worker-1",
						Message:   "Job acknowledged by worker worker-1",
					},
					{
						Timestamp: time.Date(2026, 2, 19, 10, 0, 3, 0, time.UTC),
						Event:     "failed",
						Hostname:  "worker-1",
						Message:   "Job failed on worker-1",
						Error:     "timeout",
					},
					{
						Timestamp: time.Date(2026, 2, 19, 10, 5, 0, 0, time.UTC),
						Event:     "retried",
						Hostname:  "_api",
						Message:   "Job retried as dd0e8400-e29b-41d4-a716-446655440000",
					},
				},
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Equal("cc0e8400-e29b-41d4-a716-446655440000", r.Id.String())
				s.Equal("failed", *r.Status)
				s.NotNil(r.Timeline)
				tl := *r.Timeline
				s.Len(tl, 4)

				s.Equal("submitted", *tl[0].Event)
				s.Equal("_api", *tl[0].Hostname)
				s.Equal("Job submitted to queue", *tl[0].Message)
				s.Nil(tl[0].Error)

				s.Equal("acknowledged", *tl[1].Event)
				s.Equal("worker-1", *tl[1].Hostname)

				s.Equal("failed", *tl[2].Event)
				s.NotNil(tl[2].Error)
				s.Equal("timeout", *tl[2].Error)

				s.Equal("retried", *tl[3].Event)
				s.Contains(*tl[3].Message, "dd0e8400")
			},
		},
		{
			name: "empty timeline omits field",
			request: gen.GetJobByIDRequestObject{
				Id: uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440000"),
			},
			mockJob: &jobtypes.QueuedJob{
				ID:       "dd0e8400-e29b-41d4-a716-446655440000",
				Status:   "completed",
				Created:  "2026-02-19T10:00:00Z",
				Timeline: nil,
			},
			expectMock: true,
			validateFunc: func(resp gen.GetJobByIDResponseObject) {
				r, ok := resp.(gen.GetJobByID200JSONResponse)
				s.True(ok)
				s.Nil(r.Timeline)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				s.mockJobClient.EXPECT().
					GetJobStatus(gomock.Any(), tt.request.Id.String()).
					Return(tt.mockJob, tt.mockError)
			}

			resp, err := s.handler.GetJobByID(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestJobGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobGetPublicTestSuite))
}
