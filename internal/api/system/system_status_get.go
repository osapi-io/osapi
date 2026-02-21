// Copyright (c) 2024 John Dewey

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

package system

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/system/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// GetSystemStatus get the system status API endpoint.
func (s *System) GetSystemStatus(
	ctx context.Context,
	request gen.GetSystemStatusRequestObject,
) (gen.GetSystemStatusResponseObject, error) {
	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.GetSystemStatus400JSONResponse{Error: &errMsg}, nil
		}
	}

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	if job.IsBroadcastTarget(hostname) {
		return s.getSystemStatusBroadcast(ctx, hostname)
	}

	jobID, status, err := s.JobClient.QuerySystemStatus(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetSystemStatus500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := buildSystemStatusResponse(status)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetSystemStatus200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.SystemStatusResponse{*resp},
	}, nil
}

// getSystemStatusBroadcast handles broadcast targets (_all or label) for system status.
func (s *System) getSystemStatusBroadcast(
	ctx context.Context,
	target string,
) (gen.GetSystemStatusResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QuerySystemStatusBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetSystemStatus500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.SystemStatusResponse
	for _, status := range results {
		responses = append(responses, *buildSystemStatusResponse(status))
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.SystemStatusResponse{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetSystemStatus200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// buildSystemStatusResponse converts a job.SystemStatusResponse to the API response.
func buildSystemStatusResponse(
	status *job.SystemStatusResponse,
) *gen.SystemStatusResponse {
	disksSlice := make(gen.DisksResponse, 0, len(status.DiskUsage))
	for _, d := range status.DiskUsage {
		disk := gen.DiskResponse{
			Name:  d.Name,
			Total: uint64ToInt(d.Total),
			Used:  uint64ToInt(d.Used),
			Free:  uint64ToInt(d.Free),
		}
		disksSlice = append(disksSlice, disk)
	}

	uptime := formatDuration(status.Uptime)
	resp := gen.SystemStatusResponse{
		Hostname: status.Hostname,
		Uptime:   &uptime,
		Disks:    &disksSlice,
	}

	if status.LoadAverages != nil {
		resp.LoadAverage = &gen.LoadAverageResponse{
			N1min:  status.LoadAverages.Load1,
			N5min:  status.LoadAverages.Load5,
			N15min: status.LoadAverages.Load15,
		}
	}

	if status.OSInfo != nil {
		resp.OsInfo = &gen.OSInfoResponse{
			Distribution: status.OSInfo.Distribution,
			Version:      status.OSInfo.Version,
		}
	}

	if status.MemoryStats != nil {
		resp.Memory = &gen.MemoryResponse{
			Total: uint64ToInt(status.MemoryStats.Total),
			Free:  uint64ToInt(status.MemoryStats.Free),
			Used:  uint64ToInt(status.MemoryStats.Cached),
		}
	}

	return &resp
}

func formatDuration(
	d time.Duration,
) string {
	totalMinutes := int(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	// Pluralize days, hours, and minutes
	dayStr := "day"
	if days != 1 {
		dayStr = "days"
	}

	hourStr := "hour"
	if hours != 1 {
		hourStr = "hours"
	}

	minuteStr := "minute"
	if minutes != 1 {
		minuteStr = "minutes"
	}

	// Format the result as a string
	return fmt.Sprintf("%d %s, %d %s, %d %s", days, dayStr, hours, hourStr, minutes, minuteStr)
}

// uint64ToInt convert uint64 to int, with overflow protection.
func uint64ToInt(
	value uint64,
) int {
	maxInt := int(^uint(0) >> 1) // maximum value of int based on the architecture
	if value > uint64(maxInt) {  // check for overflow
		return maxInt // return max int to prevent overflow
	}
	return int(value) // conversion within bounds
}
