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

package osapi

import (
	"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
)

// JobCreated represents a newly created job response.
type JobCreated struct {
	JobID     string
	Status    string
	Revision  int64
	Timestamp string
}

// JobDetail represents a job's full details.
type JobDetail struct {
	ID          string
	Status      string
	Hostname    string
	Created     string
	UpdatedAt   string
	Error       string
	Operation   map[string]any
	Result      any
	AgentStates map[string]AgentState
	Responses   map[string]AgentJobResponse
	Timeline    []TimelineEvent
}

// AgentState represents an agent's processing state for a broadcast job.
type AgentState struct {
	Status   string
	Duration string
	Error    string
}

// AgentJobResponse represents an agent's response data for a broadcast job.
type AgentJobResponse struct {
	Hostname string
	Status   string
	Error    string
	Data     any
}

// JobList is a paginated list of jobs.
type JobList struct {
	Items        []JobDetail
	TotalItems   int
	StatusCounts map[string]int
}

// QueueStats represents job queue statistics.
type QueueStats struct {
	TotalJobs    int
	DlqCount     int
	StatusCounts map[string]int
}

// jobCreatedFromGen converts a gen.CreateJobResponse to a JobCreated.
func jobCreatedFromGen(
	g *gen.CreateJobResponse,
) JobCreated {
	j := JobCreated{
		JobID:  g.JobId.String(),
		Status: g.Status,
	}

	if g.Revision != nil {
		j.Revision = *g.Revision
	}

	if g.Timestamp != nil {
		j.Timestamp = *g.Timestamp
	}

	return j
}

// jobDetailFromGen converts a gen.JobDetailResponse to a JobDetail.
func jobDetailFromGen(
	g *gen.JobDetailResponse,
) JobDetail {
	j := JobDetail{}

	if g.Id != nil {
		j.ID = g.Id.String()
	}

	if g.Status != nil {
		j.Status = *g.Status
	}

	if g.Hostname != nil {
		j.Hostname = *g.Hostname
	}

	if g.Created != nil {
		j.Created = *g.Created
	}

	if g.UpdatedAt != nil {
		j.UpdatedAt = *g.UpdatedAt
	}

	if g.Error != nil {
		j.Error = *g.Error
	}

	if g.Operation != nil {
		j.Operation = *g.Operation
	}

	j.Result = g.Result

	if g.AgentStates != nil {
		states := make(map[string]AgentState, len(*g.AgentStates))
		for k, v := range *g.AgentStates {
			as := AgentState{}

			if v.Status != nil {
				as.Status = *v.Status
			}

			if v.Duration != nil {
				as.Duration = *v.Duration
			}

			if v.Error != nil {
				as.Error = *v.Error
			}

			states[k] = as
		}

		j.AgentStates = states
	}

	if g.Responses != nil {
		responses := make(map[string]AgentJobResponse, len(*g.Responses))
		for k, v := range *g.Responses {
			r := AgentJobResponse{
				Data: v.Data,
			}

			if v.Hostname != nil {
				r.Hostname = *v.Hostname
			}

			if v.Status != nil {
				r.Status = *v.Status
			}

			if v.Error != nil {
				r.Error = *v.Error
			}

			responses[k] = r
		}

		j.Responses = responses
	}

	if g.Timeline != nil {
		timeline := make([]TimelineEvent, 0, len(*g.Timeline))
		for _, v := range *g.Timeline {
			te := TimelineEvent{}

			if v.Timestamp != nil {
				te.Timestamp = *v.Timestamp
			}

			if v.Event != nil {
				te.Event = *v.Event
			}

			if v.Hostname != nil {
				te.Hostname = *v.Hostname
			}

			if v.Message != nil {
				te.Message = *v.Message
			}

			if v.Error != nil {
				te.Error = *v.Error
			}

			timeline = append(timeline, te)
		}

		j.Timeline = timeline
	}

	return j
}

// jobListFromGen converts a gen.ListJobsResponse to a JobList.
func jobListFromGen(
	g *gen.ListJobsResponse,
) JobList {
	jl := JobList{}

	if g.TotalItems != nil {
		jl.TotalItems = *g.TotalItems
	}

	if g.StatusCounts != nil {
		jl.StatusCounts = *g.StatusCounts
	}

	if g.Items != nil {
		items := make([]JobDetail, 0, len(*g.Items))
		for i := range *g.Items {
			items = append(items, jobDetailFromGen(&(*g.Items)[i]))
		}

		jl.Items = items
	}

	return jl
}

// queueStatsFromGen converts a gen.QueueStatsResponse to QueueStats.
func queueStatsFromGen(
	g *gen.QueueStatsResponse,
) QueueStats {
	qs := QueueStats{}

	if g.TotalJobs != nil {
		qs.TotalJobs = *g.TotalJobs
	}

	if g.DlqCount != nil {
		qs.DlqCount = *g.DlqCount
	}

	if g.StatusCounts != nil {
		qs.StatusCounts = *g.StatusCounts
	}

	return qs
}
