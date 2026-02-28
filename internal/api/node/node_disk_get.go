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

package node

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
)

// GetNodeDisk get the node disk usage API endpoint.
func (s *Node) GetNodeDisk(
	ctx context.Context,
	request gen.GetNodeDiskRequestObject,
) (gen.GetNodeDiskResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeDisk400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("disk get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeDiskBroadcast(ctx, hostname)
	}

	jobID, diskResp, agentHostname, err := s.JobClient.QueryNodeDisk(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeDisk500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := buildDiskResultItem(agentHostname, diskResp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeDisk200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.DiskResultItem{*resp},
	}, nil
}

// getNodeDiskBroadcast handles broadcast targets (_all or label) for node disk.
func (s *Node) getNodeDiskBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeDiskResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryNodeDiskBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeDisk500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.DiskResultItem
	for host, diskResp := range results {
		responses = append(responses, *buildDiskResultItem(host, diskResp))
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.DiskResultItem{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeDisk200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// buildDiskResultItem converts a NodeDiskResponse to a DiskResultItem.
func buildDiskResultItem(
	hostname string,
	diskResp *job.NodeDiskResponse,
) *gen.DiskResultItem {
	disksSlice := make(gen.DisksResponse, 0, len(diskResp.Disks))
	for _, d := range diskResp.Disks {
		disk := gen.DiskResponse{
			Name:  d.Name,
			Total: uint64ToInt(d.Total),
			Used:  uint64ToInt(d.Used),
			Free:  uint64ToInt(d.Free),
		}
		disksSlice = append(disksSlice, disk)
	}

	return &gen.DiskResultItem{
		Hostname: hostname,
		Disks:    &disksSlice,
	}
}
