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

package client

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type JobsTestSuite struct {
	suite.Suite
}

func (s *JobsTestSuite) TestComputeStatusFromKeyNames() {
	tests := []struct {
		name              string
		keys              []string
		expectedOrderIDs  []string
		expectedStatuses  map[string]string
		validateOrderFunc func(ids []string)
	}{
		{
			name:             "empty keys",
			keys:             []string{},
			expectedOrderIDs: nil,
			expectedStatuses: map[string]string{},
		},
		{
			name:             "only jobs keys no status events",
			keys:             []string{"jobs.job-1", "jobs.job-2"},
			expectedOrderIDs: []string{"job-2", "job-1"},
			expectedStatuses: map[string]string{},
		},
		{
			name: "single agent completed",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.acknowledged.agent1.101",
				"status.job-1.started.agent1.102",
				"status.job-1.completed.agent1.103",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
			},
		},
		{
			name: "single agent failed",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.acknowledged.agent1.101",
				"status.job-1.started.agent1.102",
				"status.job-1.failed.agent1.103",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "failed",
			},
		},
		{
			name: "single agent processing",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.acknowledged.agent1.101",
				"status.job-1.started.agent1.102",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "processing",
			},
		},
		{
			name: "submitted only via api",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "submitted",
			},
		},
		{
			name: "acknowledged only agent shows processing",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.acknowledged.agent1.101",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "processing",
			},
		},
		{
			name: "multi-agent partial failure",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.completed.agent1.101",
				"status.job-1.failed.agent2.102",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "partial_failure",
			},
		},
		{
			name: "multi-agent all completed",
			keys: []string{
				"jobs.job-1",
				"status.job-1.completed.agent1.101",
				"status.job-1.completed.agent2.102",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
			},
		},
		{
			name: "multi-agent one still processing",
			keys: []string{
				"jobs.job-1",
				"status.job-1.completed.agent1.101",
				"status.job-1.started.agent2.102",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "processing",
			},
		},
		{
			name: "retried counts as completed",
			keys: []string{
				"jobs.job-1",
				"status.job-1.submitted._api.100",
				"status.job-1.retried.agent1.101",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
			},
		},
		{
			name: "multiple jobs mixed statuses",
			keys: []string{
				"jobs.job-1",
				"jobs.job-2",
				"jobs.job-3",
				"status.job-1.completed.agent1.101",
				"status.job-2.started.agent1.201",
				"status.job-3.failed.agent1.301",
			},
			expectedOrderIDs: []string{"job-3", "job-2", "job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
				"job-2": "processing",
				"job-3": "failed",
			},
		},
		{
			name: "malformed status key skipped",
			keys: []string{
				"jobs.job-1",
				"status.incomplete",
				"status.job-1.completed.agent1.101",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
			},
		},
		{
			name: "non-job non-status keys ignored",
			keys: []string{
				"jobs.job-1",
				"responses.job-1.agent1.100",
				"status.job-1.completed.agent1.101",
			},
			expectedOrderIDs: []string{"job-1"},
			expectedStatuses: map[string]string{
				"job-1": "completed",
			},
		},
		{
			name:             "empty job ID after trim skipped",
			keys:             []string{"jobs."},
			expectedOrderIDs: nil,
			expectedStatuses: map[string]string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			orderedIDs, jobStatuses := computeStatusFromKeyNames(tt.keys)

			s.Equal(tt.expectedOrderIDs, orderedIDs)

			actualStatuses := make(map[string]string, len(jobStatuses))
			for id, info := range jobStatuses {
				actualStatuses[id] = info.Status
			}
			s.Equal(tt.expectedStatuses, actualStatuses)
		})
	}
}

func TestJobsTestSuite(t *testing.T) {
	suite.Run(t, new(JobsTestSuite))
}
