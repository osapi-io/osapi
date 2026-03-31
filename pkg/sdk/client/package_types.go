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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// PackageInfoResult represents a package list result from a query operation.
type PackageInfoResult struct {
	Hostname string        `json:"hostname"`
	Status   string        `json:"status"`
	Packages []PackageInfo `json:"packages,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// PackageInfo represents information about an installed package.
type PackageInfo struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

// PackageMutationResult represents the result of a package install, remove,
// or update operation.
type PackageMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// PackageUpdateResult represents an update list result for one host.
type PackageUpdateResult struct {
	Hostname string       `json:"hostname"`
	Status   string       `json:"status"`
	Updates  []UpdateInfo `json:"updates,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// UpdateInfo represents information about an available package update.
type UpdateInfo struct {
	Name           string `json:"name,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	NewVersion     string `json:"new_version,omitempty"`
}

// packageInfoCollectionFromList converts a gen.PackageCollectionResponse
// to a Collection[PackageInfoResult].
func packageInfoCollectionFromList(
	g *gen.PackageCollectionResponse,
) Collection[PackageInfoResult] {
	results := make([]PackageInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PackageInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Packages: packageInfosFromGen(r.Packages),
			Error:    derefString(r.Error),
		})
	}

	return Collection[PackageInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// packageInfoCollectionFromGet converts a gen.PackageCollectionResponse
// from a get-by-name response to a Collection[PackageInfoResult].
func packageInfoCollectionFromGet(
	g *gen.PackageCollectionResponse,
) Collection[PackageInfoResult] {
	return packageInfoCollectionFromList(g)
}

// packageInfosFromGen converts a gen package info slice to SDK PackageInfo
// slice.
func packageInfosFromGen(
	pkgs *[]gen.PackageInfo,
) []PackageInfo {
	if pkgs == nil {
		return nil
	}

	result := make([]PackageInfo, 0, len(*pkgs))
	for _, p := range *pkgs {
		result = append(result, PackageInfo{
			Name:        derefString(p.Name),
			Version:     derefString(p.Version),
			Description: derefString(p.Description),
			Status:      derefString(p.Status),
			Size:        derefInt64(p.Size),
		})
	}

	return result
}

// packageMutationCollectionFromInstall converts a gen.PackageMutationResponse
// to a Collection[PackageMutationResult].
func packageMutationCollectionFromInstall(
	g *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromGen(g)
}

// packageMutationCollectionFromRemove converts a gen.PackageMutationResponse
// to a Collection[PackageMutationResult].
func packageMutationCollectionFromRemove(
	g *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromGen(g)
}

// packageMutationCollectionFromUpdate converts a gen.PackageMutationResponse
// to a Collection[PackageMutationResult].
func packageMutationCollectionFromUpdate(
	g *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	return packageMutationCollectionFromGen(g)
}

// packageMutationCollectionFromGen converts a gen.PackageMutationResponse
// to a Collection[PackageMutationResult].
func packageMutationCollectionFromGen(
	g *gen.PackageMutationResponse,
) Collection[PackageMutationResult] {
	results := make([]PackageMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PackageMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[PackageMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// packageUpdateCollectionFromGen converts a gen.UpdateCollectionResponse
// to a Collection[PackageUpdateResult].
func packageUpdateCollectionFromGen(
	g *gen.UpdateCollectionResponse,
) Collection[PackageUpdateResult] {
	results := make([]PackageUpdateResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PackageUpdateResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Updates:  updateInfosFromGen(r.Updates),
			Error:    derefString(r.Error),
		})
	}

	return Collection[PackageUpdateResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// updateInfosFromGen converts a gen update info slice to SDK UpdateInfo slice.
func updateInfosFromGen(
	updates *[]gen.UpdateInfo,
) []UpdateInfo {
	if updates == nil {
		return nil
	}

	result := make([]UpdateInfo, 0, len(*updates))
	for _, u := range *updates {
		result = append(result, UpdateInfo{
			Name:           derefString(u.Name),
			CurrentVersion: derefString(u.CurrentVersion),
			NewVersion:     derefString(u.NewVersion),
		})
	}

	return result
}
