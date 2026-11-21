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

package client

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// PackageInfo is one installed package on one host.
type PackageInfo struct {
	Hostname     string `json:"hostname"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
	Name         string `json:"name,omitempty"`
	Version      string `json:"version,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Manager      string `json:"manager,omitempty"`
	Description  string `json:"description,omitempty"`
	Size         int64  `json:"size,omitempty"`
}

// PackageMutationResult is a package after an install or a removal.
type PackageMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
}

// PackageUpdate is one package the index of a host has a newer version of.
type PackageUpdate struct {
	Hostname       string `json:"hostname"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Name           string `json:"name,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	NewVersion     string `json:"new_version,omitempty"`
}

// PackageIndexResult is the index of a host after a refresh.
type PackageIndexResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

func packageInfosFromGen(
	jobID *openapi_types.UUID,
	rows []gen.PackageInstalled,
) Collection[PackageInfo] {
	results := make([]PackageInfo, 0, len(rows))
	for _, r := range rows {
		results = append(results, PackageInfo{
			Hostname:     r.Hostname,
			Status:       string(r.Status),
			Error:        derefString(r.Error),
			Name:         derefString(r.Name),
			Version:      derefString(r.Version),
			Architecture: derefString(r.Architecture),
			Manager:      derefString(r.Manager),
			Description:  derefString(r.Description),
			Size:         derefInt64(r.Size),
		})
	}

	return Collection[PackageInfo]{Results: results, JobID: jobIDFromGen(jobID)}
}

func packageMutationsFromGen(
	jobID *openapi_types.UUID,
	rows []gen.PackageResult,
) Collection[PackageMutationResult] {
	results := make([]PackageMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, PackageMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[PackageMutationResult]{Results: results, JobID: jobIDFromGen(jobID)}
}

func packageUpdatesFromGen(
	g *gen.UpdatepackagesCollectionResponse,
) Collection[PackageUpdate] {
	results := make([]PackageUpdate, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PackageUpdate{
			Hostname:       r.Hostname,
			Status:         string(r.Status),
			Error:          derefString(r.Error),
			Name:           derefString(r.Name),
			CurrentVersion: derefString(r.CurrentVersion),
			NewVersion:     derefString(r.NewVersion),
		})
	}

	return Collection[PackageUpdate]{Results: results, JobID: jobIDFromGen(g.JobId)}
}

func packageIndexFromGen(
	g *gen.UpdatepackagesUpdateResponse,
) Collection[PackageIndexResult] {
	results := make([]PackageIndexResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PackageIndexResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[PackageIndexResult]{Results: results, JobID: jobIDFromGen(g.JobId)}
}
