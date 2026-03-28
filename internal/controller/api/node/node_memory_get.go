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
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// GetNodeMemory get the node memory stats API endpoint.
func (s *Node) GetNodeMemory(
	ctx context.Context,
	request gen.GetNodeMemoryRequestObject,
) (gen.GetNodeMemoryResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeMemory400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("memory get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeMemoryBroadcast(ctx, hostname)
	}

	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationNodeMemoryGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeMemory500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var memStats mem.Result
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &memStats)
	}

	resp := buildMemoryResultItem(rawResp.Hostname, &memStats)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeMemory200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.MemoryResultItem{*resp},
	}, nil
}

// getNodeMemoryBroadcast handles broadcast targets (_all or label) for node memory.
func (s *Node) getNodeMemoryBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeMemoryResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationNodeMemoryGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeMemory500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.MemoryResultItem
	for host, resp := range responses {
		item := gen.MemoryResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.MemoryResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.MemoryResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.MemoryResultItemStatusOk
			var memStats mem.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &memStats)
			}
			built := buildMemoryResultItem(host, &memStats)
			item.Memory = built.Memory
			item.Changed = built.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeMemory200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// buildMemoryResultItem converts mem.Result to a MemoryResultItem.
func buildMemoryResultItem(
	hostname string,
	memStats *mem.Result,
) *gen.MemoryResultItem {
	changed := false
	item := &gen.MemoryResultItem{
		Hostname: hostname,
		Changed:  &changed,
	}

	if memStats != nil {
		item.Memory = &gen.MemoryResponse{
			Total: uint64ToInt(memStats.Total),
			Free:  uint64ToInt(memStats.Free),
			Used:  uint64ToInt(memStats.Cached),
		}
	}

	return item
}
