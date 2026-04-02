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

package packageapi

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/package/gen"
	"github.com/retr0h/osapi/internal/job"
	aptProv "github.com/retr0h/osapi/internal/provider/node/apt"
)

// PostNodePackageUpdate refreshes package sources on a target node.
func (p *Package) PostNodePackageUpdate(
	ctx context.Context,
	request gen.PostNodePackageUpdateRequestObject,
) (gen.PostNodePackageUpdateResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodePackageUpdate400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	p.logger.Debug("package update",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return p.postNodePackageUpdateBroadcast(ctx, hostname)
	}

	jobID, resp, err := p.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationPackageUpdate,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodePackageUpdate500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodePackageUpdate200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.PackageMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.PackageMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result aptProv.Result
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	return gen.PostNodePackageUpdate200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.PackageMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.PackageMutationResultStatusOk,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodePackageUpdateBroadcast handles broadcast targets for package update.
func (p *Package) postNodePackageUpdateBroadcast(
	ctx context.Context,
	target string,
) (gen.PostNodePackageUpdateResponseObject, error) {
	jobID, responses, err := p.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationPackageUpdate,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodePackageUpdate500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.PackageMutationResult
	for host, resp := range responses {
		item := gen.PackageMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.PackageMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.PackageMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.PackageMutationResultStatusOk
			var result aptProv.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodePackageUpdate200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
