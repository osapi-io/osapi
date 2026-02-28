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
	"github.com/retr0h/osapi/internal/provider/node/host"
)

// GetNodeOS get the node OS info API endpoint.
func (s *Node) GetNodeOS(
	ctx context.Context,
	request gen.GetNodeOSRequestObject,
) (gen.GetNodeOSResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeOS400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("os get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeOSBroadcast(ctx, hostname)
	}

	jobID, osInfo, agentHostname, err := s.JobClient.QueryNodeOS(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeOS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := buildOSInfoResultItem(agentHostname, osInfo)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeOS200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.OSInfoResultItem{*resp},
	}, nil
}

// getNodeOSBroadcast handles broadcast targets (_all or label) for node OS info.
func (s *Node) getNodeOSBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeOSResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryNodeOSBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeOS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.OSInfoResultItem
	for host, osInfo := range results {
		responses = append(responses, *buildOSInfoResultItem(host, osInfo))
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.OSInfoResultItem{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeOS200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// buildOSInfoResultItem converts host.OSInfo to an OSInfoResultItem.
func buildOSInfoResultItem(
	hostname string,
	osInfo *host.OSInfo,
) *gen.OSInfoResultItem {
	item := &gen.OSInfoResultItem{
		Hostname: hostname,
	}

	if osInfo != nil {
		item.OsInfo = &gen.OSInfoResponse{
			Distribution: osInfo.Distribution,
			Version:      osInfo.Version,
		}
	}

	return item
}
