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

package ntp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/ntp/gen"
	"github.com/retr0h/osapi/internal/job"
	ntpProv "github.com/retr0h/osapi/internal/provider/node/ntp"
)

// GetNodeNtp gets NTP sync status and configured servers on a target node.
func (s *Ntp) GetNodeNtp(
	ctx context.Context,
	request gen.GetNodeNtpRequestObject,
) (gen.GetNodeNtpResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("ntp get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNtpBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationNtpGet,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNtp200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.NtpStatusEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.NtpStatusEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var status ntpProv.Status
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &status)
	}

	jobUUID := uuid.MustParse(jobID)
	agentHostname := resp.Hostname

	return gen.GetNodeNtp200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.NtpStatusEntry{
			statusToEntry(agentHostname, &status),
		},
	}, nil
}

// getNodeNtpBroadcast handles broadcast targets for NTP get.
func (s *Ntp) getNodeNtpBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeNtpResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationNtpGet,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.NtpStatusEntry, 0)
	for host, resp := range responses {
		item := gen.NtpStatusEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.NtpStatusEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.NtpStatusEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			var status ntpProv.Status
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &status)
			}
			item = statusToEntry(host, &status)
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeNtp200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// statusToEntry converts an NTP Status to a gen NtpStatusEntry.
func statusToEntry(
	hostname string,
	status *ntpProv.Status,
) gen.NtpStatusEntry {
	entry := gen.NtpStatusEntry{
		Hostname:     hostname,
		Status:       gen.NtpStatusEntryStatusOk,
		Synchronized: &status.Synchronized,
	}

	if status.Stratum != 0 {
		stratum := status.Stratum
		entry.Stratum = &stratum
	}

	if status.Offset != "" {
		offset := status.Offset
		entry.Offset = &offset
	}

	if status.CurrentSource != "" {
		src := status.CurrentSource
		entry.CurrentSource = &src
	}

	if len(status.Servers) > 0 {
		servers := status.Servers
		entry.Servers = &servers
	}

	return entry
}
