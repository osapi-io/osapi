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

package job

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them.
var (
	sampleID      = "550e8400-e29b-41d4-a716-446655440000"
	sampleChanged = true

	sampleDetail = JobDetailResponse{
		ID:      sampleID,
		Status:  "completed",
		Created: "2026-01-15T09:58:12Z",
		Target:  "web-01",
		Operation: &Operation{
			Type: "node.sysctl.update",
			Data: map[string]any{"key": "net.ipv4.ip_forward", "value": "1"},
		},
		Hostname:  "web-01",
		UpdatedAt: "2026-01-15T09:58:13Z",
		Changed:   &sampleChanged,
		Responses: []Response{{
			Hostname: "web-01", Status: "completed", Changed: &sampleChanged,
			Data: map[string]any{"key": "net.ipv4.ip_forward", "changed": true},
		}},
		AgentStates: []AgentState{{Hostname: "web-01", Status: "completed", Duration: "184ms"}},
		Timeline: []TimelineEvent{
			{Timestamp: "2026-01-15T09:58:12Z", Event: "submitted", Message: "job created"},
			{Timestamp: "2026-01-15T09:58:12Z", Event: "acknowledged", Hostname: "web-01"},
			{Timestamp: "2026-01-15T09:58:13Z", Event: "completed", Hostname: "web-01"},
		},
	}

	sampleList = ListJobsResponse{
		TotalItems:   2,
		StatusCounts: map[string]int{"completed": 1, "failed": 1},
		Items: []JobDetailResponse{
			sampleDetail,
			{
				ID:        "7c9e6679-7425-40de-944b-e07fc1f90ae7",
				Status:    "failed",
				Created:   "2026-01-15T09:41:03Z",
				Target:    "_all",
				Operation: &Operation{Type: "node.package.update"},
				Error:     "apt-get: unable to lock the administration directory",
			},
		},
	}

	sampleRetry = CreateJobResponse{
		JobID:     "9b2e4c6a-8d0f-4a1b-b3c5-d7e9f1a2b4c6",
		Status:    "created",
		Revision:  42,
		Timestamp: "2026-01-15T10:00:00Z",
	}
)
