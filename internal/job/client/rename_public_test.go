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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type RenamePublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
}

func (s *RenamePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *RenamePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// TestFollowRename records the new name of a host that a hostname update
// renamed, so its later answers verify; any other answer changes nothing.
func (s *RenamePublicTestSuite) TestFollowRename() {
	update := job.OperationType(job.CategoryBehavior + ".hostname.update")
	renamed, err := json.Marshal(map[string]any{"name": "db-01", "changed": true})
	s.Require().NoError(err)
	tests := []struct {
		name      string
		req       *job.Request
		resp      *job.Response
		setupMock func(*jobmocks.MockAgentKeyStore)
	}{
		{
			name: "when a hostname update completes with a new name, the record takes it",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed,
				Hostname: "web-01", MachineID: "machine-001",
			},
			setupMock: func(m *jobmocks.MockAgentKeyStore) {
				m.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "db-01").Return(nil)
			},
		},
		{
			name: "when the store refuses the rename, the answer still stands",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed,
				Hostname: "web-01", MachineID: "machine-001",
			},
			setupMock: func(m *jobmocks.MockAgentKeyStore) {
				m.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "db-01").
					Return(errors.New("kv error"))
			},
		},
		{
			name: "when the job is another operation, nothing changes",
			req: &job.Request{
				Operation: job.OperationType(job.CategoryBehavior + ".hostname.get"),
			},
			resp: &job.Response{Status: types.AgentJobStatusCompleted, Data: renamed},
		},
		{
			name: "when the update failed, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{Status: types.AgentJobStatusFailed, Data: renamed},
		},
		{
			name: "when the answer has no name, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{Status: types.AgentJobStatusCompleted, Data: json.RawMessage(`{}`)},
		},
		{
			name: "when the name is the one it has, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed, Hostname: "db-01",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			if tt.setupMock != nil {
				tt.setupMock(store)
			}
			c, err := newTrusted(jobmocks.NewMockNATSClient(s.mockCtrl), &client.Options{
				Timeout: time.Second, JobsBucket: "jobs", StreamName: "JOBS",
			})
			s.Require().NoError(err)
			c.SetAgentKeyStore(store)

			client.ExportFollowRename(context.Background(), c, tt.req, tt.resp)
		})
	}
}

func TestRenamePublicTestSuite(t *testing.T) {
	suite.Run(t, new(RenamePublicTestSuite))
}
