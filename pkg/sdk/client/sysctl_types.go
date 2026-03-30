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

// SysctlEntryResult represents a sysctl entry from a query operation.
type SysctlEntryResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Error    string `json:"error,omitempty"`
}

// SysctlMutationResult represents the result of a sysctl create, update, or delete.
type SysctlMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Key      string `json:"key,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// SysctlCreateOpts contains options for creating a sysctl parameter.
type SysctlCreateOpts struct {
	// Key is the sysctl parameter name (e.g., "net.ipv4.ip_forward"). Required.
	Key string
	// Value is the parameter value. Required.
	Value string
}

// SysctlUpdateOpts contains options for updating a sysctl parameter.
type SysctlUpdateOpts struct {
	// Value is the new parameter value. Required.
	Value string
}

// sysctlEntryCollectionFromGen converts a gen.SysctlCollectionResponse
// to a Collection[SysctlEntryResult].
func sysctlEntryCollectionFromGen(
	g *gen.SysctlCollectionResponse,
) Collection[SysctlEntryResult] {
	results := make([]SysctlEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, SysctlEntryResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Key:      derefString(r.Key),
			Value:    derefString(r.Value),
			Error:    derefString(r.Error),
		})
	}

	return Collection[SysctlEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sysctlEntryCollectionFromGet converts a gen.SysctlGetResponse
// to a Collection[SysctlEntryResult].
func sysctlEntryCollectionFromGet(
	g *gen.SysctlGetResponse,
) Collection[SysctlEntryResult] {
	results := make([]SysctlEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, SysctlEntryResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Key:      derefString(r.Key),
			Value:    derefString(r.Value),
			Error:    derefString(r.Error),
		})
	}

	return Collection[SysctlEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sysctlMutationCollectionFromCreate converts a gen.SysctlCreateResponse
// to a Collection[SysctlMutationResult].
func sysctlMutationCollectionFromCreate(
	g *gen.SysctlCreateResponse,
) Collection[SysctlMutationResult] {
	results := make([]SysctlMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, SysctlMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Key:      derefString(r.Key),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[SysctlMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sysctlMutationCollectionFromUpdate converts a gen.SysctlUpdateResponse
// to a Collection[SysctlMutationResult].
func sysctlMutationCollectionFromUpdate(
	g *gen.SysctlUpdateResponse,
) Collection[SysctlMutationResult] {
	results := make([]SysctlMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, SysctlMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Key:      derefString(r.Key),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[SysctlMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sysctlMutationCollectionFromDelete converts a gen.SysctlDeleteResponse
// to a Collection[SysctlMutationResult].
func sysctlMutationCollectionFromDelete(
	g *gen.SysctlDeleteResponse,
) Collection[SysctlMutationResult] {
	results := make([]SysctlMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, SysctlMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Key:      derefString(r.Key),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[SysctlMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
