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

// ServiceInfoResult represents a service list result from a single agent.
type ServiceInfoResult struct {
	Hostname string        `json:"hostname"`
	Status   string        `json:"status"`
	Services []ServiceInfo `json:"services,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ServiceInfo represents a single systemd service entry.
type ServiceInfo struct {
	Name        string `json:"name,omitempty"`
	Status      string `json:"status,omitempty"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

// ServiceGetResult represents a service get result from a single agent.
type ServiceGetResult struct {
	Hostname string       `json:"hostname"`
	Status   string       `json:"status"`
	Service  *ServiceInfo `json:"service,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// ServiceMutationResult represents the result of a service mutation or
// action operation.
type ServiceMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// ServiceCreateOpts contains options for creating a service unit file.
type ServiceCreateOpts struct {
	// Name is the service unit name (required).
	Name string
	// Object is the object store reference for the unit file (required).
	Object string
}

// ServiceUpdateOpts contains options for updating a service unit file.
type ServiceUpdateOpts struct {
	// Object is the new object store reference for the unit file (required).
	Object string
}

// serviceListCollectionFromGen converts a gen.ServiceListResponse
// to a Collection[ServiceInfoResult].
func serviceListCollectionFromGen(
	g *gen.ServiceListResponse,
) Collection[ServiceInfoResult] {
	results := make([]ServiceInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		result := ServiceInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Services != nil {
			svcs := make([]ServiceInfo, 0, len(*r.Services))
			for _, s := range *r.Services {
				svcs = append(svcs, serviceInfoFromGen(s))
			}
			result.Services = svcs
		}

		results = append(results, result)
	}

	return Collection[ServiceInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// serviceInfoFromGen converts a gen.ServiceInfo to a ServiceInfo.
func serviceInfoFromGen(
	s gen.ServiceInfo,
) ServiceInfo {
	return ServiceInfo{
		Name:        derefString(s.Name),
		Status:      derefString(s.Status),
		Enabled:     derefBool(s.Enabled),
		Description: derefString(s.Description),
		PID:         derefInt(s.Pid),
	}
}

// serviceGetCollectionFromGen converts a gen.ServiceGetResponse
// to a Collection[ServiceGetResult].
func serviceGetCollectionFromGen(
	g *gen.ServiceGetResponse,
) Collection[ServiceGetResult] {
	results := make([]ServiceGetResult, 0, len(g.Results))
	for _, r := range g.Results {
		result := ServiceGetResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Service != nil {
			info := serviceInfoFromGen(*r.Service)
			result.Service = &info
		}

		results = append(results, result)
	}

	return Collection[ServiceGetResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// serviceMutationCollectionFromGen converts a gen.ServiceMutationResponse
// to a Collection[ServiceMutationResult].
func serviceMutationCollectionFromGen(
	g *gen.ServiceMutationResponse,
) Collection[ServiceMutationResult] {
	results := make([]ServiceMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, serviceMutationResultFromGen(r))
	}

	return Collection[ServiceMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// serviceMutationResultFromGen converts a single gen.ServiceMutationEntry
// to a ServiceMutationResult.
func serviceMutationResultFromGen(
	r gen.ServiceMutationEntry,
) ServiceMutationResult {
	return ServiceMutationResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Name:     derefString(r.Name),
		Changed:  derefBool(r.Changed),
		Error:    derefString(r.Error),
	}
}
