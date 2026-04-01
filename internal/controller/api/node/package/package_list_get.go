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

// GetNodePackage lists all installed packages on a target node.
func (p *Package) GetNodePackage(
	ctx context.Context,
	request gen.GetNodePackageRequestObject,
) (gen.GetNodePackageResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodePackage500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	p.logger.Debug("package list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return p.getNodePackageBroadcast(ctx, hostname)
	}

	jobID, resp, err := p.JobClient.Query(ctx, hostname, "node", job.OperationPackageList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodePackage500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodePackage200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.PackageEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.PackageEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToPackageEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodePackage200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodePackageBroadcast handles broadcast targets for package list.
func (p *Package) getNodePackageBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodePackageResponseObject, error) {
	jobID, responses, err := p.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationPackageList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodePackage500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.PackageEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.PackageEntry{
				Hostname: h,
				Status:   gen.PackageEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.PackageEntry{
				Hostname: h,
				Status:   gen.PackageEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToPackageEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodePackage200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToPackageEntries converts a job response to gen PackageEntry slice.
func responseToPackageEntries(
	resp *job.Response,
) []gen.PackageEntry {
	var packages []aptProv.Package
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &packages)
	}

	hostname := resp.Hostname
	pkgInfos := make([]gen.PackageInfo, 0, len(packages))
	for _, pkg := range packages {
		name := pkg.Name
		version := pkg.Version
		status := pkg.Status

		info := gen.PackageInfo{
			Name:    &name,
			Version: &version,
			Status:  &status,
		}

		if pkg.Description != "" {
			desc := pkg.Description
			info.Description = &desc
		}
		if pkg.Size > 0 {
			size := pkg.Size
			info.Size = &size
		}

		pkgInfos = append(pkgInfos, info)
	}

	return []gen.PackageEntry{
		{
			Hostname: hostname,
			Status:   gen.PackageEntryStatusOk,
			Packages: &pkgInfos,
		},
	}
}
