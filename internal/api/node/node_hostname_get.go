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
	"github.com/retr0h/osapi/internal/validation"
)

// GetNodeHostname get the node hostname API endpoint.
func (s *Node) GetNodeHostname(
	ctx context.Context,
	request gen.GetNodeHostnameRequestObject,
) (gen.GetNodeHostnameResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetNodeHostname400JSONResponse{Error: &errMsg}, nil
	}

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	s.logger.Debug("routing",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeHostnameBroadcast(ctx, hostname)
	}

	jobID, result, worker, err := s.JobClient.QueryNodeHostname(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	displayHostname := result
	if displayHostname == "" && worker != nil {
		displayHostname = worker.Hostname
	}

	resp := gen.HostnameResponse{Hostname: displayHostname}
	if worker != nil && len(worker.Labels) > 0 {
		resp.Labels = &worker.Labels
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeHostname200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.HostnameResponse{resp},
	}, nil
}

// getNodeHostnameBroadcast handles broadcast targets (_all or label) for node hostname.
func (s *Node) getNodeHostnameBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeHostnameResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryNodeHostnameBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.HostnameResponse
	for _, w := range results {
		r := gen.HostnameResponse{Hostname: w.Hostname}
		if len(w.Labels) > 0 {
			r.Labels = &w.Labels
		}
		responses = append(responses, r)
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.HostnameResponse{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeHostname200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}
