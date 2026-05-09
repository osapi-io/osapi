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
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
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

	s.logger.Debug(
		"disk get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeDiskBroadcast(ctx, hostname)
	}

	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationNodeDiskGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeDisk500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeDisk200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DiskResultItem{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.DiskResultItemStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var diskResp job.NodeDiskResponse
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &diskResp)
	}

	resp := buildDiskResultItem(rawResp.Hostname, &diskResp)
	resp.Status = gen.DiskResultItemStatusOk
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
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationNodeDiskGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeDisk500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.DiskResultItem
	for host, resp := range responses {
		item := gen.DiskResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DiskResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DiskResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DiskResultItemStatusOk
			var diskResp job.NodeDiskResponse
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &diskResp)
			}
			built := buildDiskResultItem(host, &diskResp)
			item.Disks = built.Disks
			item.Changed = built.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeDisk200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
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

	changed := false

	return &gen.DiskResultItem{
		Hostname: hostname,
		Disks:    &disksSlice,
		Changed:  &changed,
	}
}
