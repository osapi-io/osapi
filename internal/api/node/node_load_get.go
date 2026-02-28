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
	"github.com/retr0h/osapi/internal/provider/node/load"
)

// GetNodeLoad get the node load averages API endpoint.
func (s *Node) GetNodeLoad(
	ctx context.Context,
	request gen.GetNodeLoadRequestObject,
) (gen.GetNodeLoadResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeLoad400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("load get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeLoadBroadcast(ctx, hostname)
	}

	jobID, loadStats, agentHostname, err := s.JobClient.QueryNodeLoad(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLoad500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := buildLoadResultItem(agentHostname, loadStats)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLoad200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.LoadResultItem{*resp},
	}, nil
}

// getNodeLoadBroadcast handles broadcast targets (_all or label) for node load.
func (s *Node) getNodeLoadBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeLoadResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryNodeLoadBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLoad500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.LoadResultItem
	for host, loadStats := range results {
		responses = append(responses, *buildLoadResultItem(host, loadStats))
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.LoadResultItem{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeLoad200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// buildLoadResultItem converts load.AverageStats to a LoadResultItem.
func buildLoadResultItem(
	hostname string,
	loadStats *load.AverageStats,
) *gen.LoadResultItem {
	item := &gen.LoadResultItem{
		Hostname: hostname,
	}

	if loadStats != nil {
		item.LoadAverage = &gen.LoadAverageResponse{
			N1min:  loadStats.Load1,
			N5min:  loadStats.Load5,
			N15min: loadStats.Load15,
		}
	}

	return item
}
