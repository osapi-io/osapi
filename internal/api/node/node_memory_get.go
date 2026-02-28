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

	jobID, memStats, agentHostname, err := s.JobClient.QueryNodeMemory(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeMemory500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := buildMemoryResultItem(agentHostname, memStats)
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
	jobID, results, errs, err := s.JobClient.QueryNodeMemoryBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeMemory500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.MemoryResultItem
	for host, memStats := range results {
		responses = append(responses, *buildMemoryResultItem(host, memStats))
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.MemoryResultItem{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeMemory200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// buildMemoryResultItem converts mem.Stats to a MemoryResultItem.
func buildMemoryResultItem(
	hostname string,
	memStats *mem.Stats,
) *gen.MemoryResultItem {
	item := &gen.MemoryResultItem{
		Hostname: hostname,
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
