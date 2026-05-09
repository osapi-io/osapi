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
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/package/gen"
	"github.com/retr0h/osapi/internal/job"
	aptProv "github.com/retr0h/osapi/internal/provider/node/apt"
)

// GetNodePackageByName gets a single package by name on a target node.
func (p *Package) GetNodePackageByName(
	ctx context.Context,
	request gen.GetNodePackageByNameRequestObject,
) (gen.GetNodePackageByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodePackageByName400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	p.logger.Debug(
		"package get",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return p.getNodePackageByNameBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := p.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationPackageGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "not installed") {
			return gen.GetNodePackageByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodePackageByName500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodePackageByName200JSONResponse{
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

	var pkg aptProv.Package
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &pkg)
	}

	jobUUID := uuid.MustParse(jobID)
	agentHostname := resp.Hostname
	pkgName := pkg.Name
	version := pkg.Version
	status := pkg.Status

	info := gen.PackageInfo{
		Name:    &pkgName,
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

	pkgInfos := []gen.PackageInfo{info}

	return gen.GetNodePackageByName200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.PackageEntry{
			{
				Hostname: agentHostname,
				Status:   gen.PackageEntryStatusOk,
				Packages: &pkgInfos,
			},
		},
	}, nil
}

// getNodePackageByNameBroadcast handles broadcast targets for package get.
func (p *Package) getNodePackageByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodePackageByNameResponseObject, error) {
	jobID, responses, err := p.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationPackageGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodePackageByName500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.PackageEntry, 0)
	for host, resp := range responses {
		item := gen.PackageEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.PackageEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.PackageEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.PackageEntryStatusOk
			var pkg aptProv.Package
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &pkg)
			}
			pkgName := pkg.Name
			version := pkg.Version
			status := pkg.Status
			info := gen.PackageInfo{
				Name:    &pkgName,
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
			pkgInfos := []gen.PackageInfo{info}
			item.Packages = &pkgInfos
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodePackageByName200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}
