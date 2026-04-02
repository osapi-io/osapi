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

// GetNodePackageUpdate lists available package updates on a target node.
func (p *Package) GetNodePackageUpdate(
	ctx context.Context,
	request gen.GetNodePackageUpdateRequestObject,
) (gen.GetNodePackageUpdateResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodePackageUpdate400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	p.logger.Debug("package list updates",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return p.getNodePackageUpdateBroadcast(ctx, hostname)
	}

	jobID, resp, err := p.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationPackageListUpdates,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodePackageUpdate500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodePackageUpdate200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.UpdateEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.Skipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToUpdateEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodePackageUpdate200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodePackageUpdateBroadcast handles broadcast targets for list updates.
func (p *Package) getNodePackageUpdateBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodePackageUpdateResponseObject, error) {
	jobID, responses, err := p.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationPackageListUpdates,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodePackageUpdate500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.UpdateEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.UpdateEntry{
				Hostname: h,
				Status:   gen.Failed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.UpdateEntry{
				Hostname: h,
				Status:   gen.Skipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToUpdateEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodePackageUpdate200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToUpdateEntries converts a job response to gen UpdateEntry slice.
func responseToUpdateEntries(
	resp *job.Response,
) []gen.UpdateEntry {
	var updates []aptProv.Update
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &updates)
	}

	hostname := resp.Hostname
	updateInfos := make([]gen.UpdateInfo, 0, len(updates))
	for _, u := range updates {
		name := u.Name
		currentVersion := u.CurrentVersion
		newVersion := u.NewVersion

		updateInfos = append(updateInfos, gen.UpdateInfo{
			Name:           &name,
			CurrentVersion: &currentVersion,
			NewVersion:     &newVersion,
		})
	}

	return []gen.UpdateEntry{
		{
			Hostname: hostname,
			Status:   gen.Ok,
			Updates:  &updateInfos,
		},
	}
}
