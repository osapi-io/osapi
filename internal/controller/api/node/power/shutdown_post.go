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

package power

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/power/gen"
	"github.com/retr0h/osapi/internal/job"
	powerProv "github.com/retr0h/osapi/internal/provider/node/power"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodePowerShutdown initiates a shutdown on the target node.
func (s *Power) PostNodePowerShutdown(
	ctx context.Context,
	request gen.PostNodePowerShutdownRequestObject,
) (gen.PostNodePowerShutdownResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodePowerShutdown400JSONResponse{Error: &errMsg}, nil
	}

	opts := powerProv.Opts{}
	if request.Body != nil {
		// Defense in depth: current fields use omitempty so validation
		// always passes, but guards against future field additions.
		if errMsg, ok := validation.Struct(request.Body); !ok {
			return gen.PostNodePowerShutdown400JSONResponse{Error: &errMsg}, nil
		}

		if request.Body.Delay != nil {
			opts.Delay = *request.Body.Delay
		}
		if request.Body.Message != nil {
			opts.Message = *request.Body.Message
		}
	}

	hostname := request.Hostname

	s.logger.Debug(
		"power shutdown",
		slog.String("target", hostname),
		slog.Int("delay", opts.Delay),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodePowerShutdownBroadcast(ctx, hostname, opts)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationPowerShutdown,
		opts,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodePowerShutdown500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodePowerShutdown200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.PowerResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.Skipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result powerProv.Result
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname
	action := result.Action
	delay := result.Delay

	return gen.PostNodePowerShutdown200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.PowerResult{
			{
				Hostname: agentHostname,
				Status:   gen.Ok,
				Changed:  changed,
				Action:   &action,
				Delay:    &delay,
			},
		},
	}, nil
}

// postNodePowerShutdownBroadcast handles broadcast targets for power shutdown.
func (s *Power) postNodePowerShutdownBroadcast(
	ctx context.Context,
	target string,
	opts powerProv.Opts,
) (gen.PostNodePowerShutdownResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationPowerShutdown,
		opts,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodePowerShutdown500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.PowerResult
	for host, resp := range responses {
		item := gen.PowerResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.Failed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.Skipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.Ok
			item.Changed = resp.Changed

			var result powerProv.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			action := result.Action
			delay := result.Delay
			item.Action = &action
			item.Delay = &delay
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodePowerShutdown200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
