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

// ServiceInfo is one service on one host.
type ServiceInfo struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Name        string `json:"name,omitempty"`
	State       string `json:"state,omitempty"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
}

// ServiceMutationResult is a service after a write or an action.
type ServiceMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// ServiceCreateOpts names the service and the object that holds its unit
// file.
type ServiceCreateOpts struct {
	// Name is the service, such as app for osapi-app.service (required).
	Name string
	// Object is the object that holds the unit file (required).
	Object string
}

// ServiceUpdateOpts names the object that holds the new unit file.
type ServiceUpdateOpts struct {
	// Object is the object that holds the unit file (required).
	Object string
}

func serviceInfosFromGen(
	jobID *openapi_types.UUID,
	rows []gen.ServiceEntry,
) Collection[ServiceInfo] {
	results := make([]ServiceInfo, 0, len(rows))
	for _, r := range rows {
		results = append(results, ServiceInfo{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Error:       derefString(r.Error),
			Name:        derefString(r.Name),
			State:       derefString(r.State),
			Enabled:     derefBool(r.Enabled),
			Description: derefString(r.Description),
		})
	}

	return Collection[ServiceInfo]{Results: results, JobID: jobIDFromGen(jobID)}
}
