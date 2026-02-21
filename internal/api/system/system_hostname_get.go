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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/system/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// GetSystemHostname get the system hostname API endpoint.
func (s *System) GetSystemHostname(
	ctx context.Context,
	request gen.GetSystemHostnameRequestObject,
) (gen.GetSystemHostnameResponseObject, error) {
	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.GetSystemHostname400JSONResponse{Error: &errMsg}, nil
		}
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
		return s.getSystemHostnameBroadcast(ctx, hostname)
	}

	jobID, result, workerHostname, err := s.JobClient.QuerySystemHostname(ctx, hostname)
	if err != nil {
		errMsg := err.Error()
		return gen.GetSystemHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	displayHostname := result
	if displayHostname == "" {
		displayHostname = workerHostname
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetSystemHostname200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.HostnameResponse{
			{Hostname: displayHostname},
		},
	}, nil
}

// getSystemHostnameBroadcast handles broadcast targets (_all or label) for system hostname.
func (s *System) getSystemHostnameBroadcast(
	ctx context.Context,
	target string,
) (gen.GetSystemHostnameResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QuerySystemHostnameBroadcast(ctx, target)
	if err != nil {
		errMsg := err.Error()
		return gen.GetSystemHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.HostnameResponse
	for _, h := range results {
		responses = append(responses, gen.HostnameResponse{Hostname: h})
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.HostnameResponse{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetSystemHostname200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}
