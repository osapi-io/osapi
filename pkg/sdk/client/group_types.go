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

// GroupInfoResult is one group on one host, from a list or a get. A host
// that failed or skipped has a row with its status and no group.
type GroupInfoResult struct {
	Hostname string   `json:"hostname"`
	Status   string   `json:"status"`
	Name     string   `json:"name,omitempty"`
	GID      int      `json:"gid,omitempty"`
	Members  []string `json:"members,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// GroupMutationResult represents the result of a group create, update, or
// delete operation.
type GroupMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// GroupCreateOpts contains options for creating a group.
type GroupCreateOpts struct {
	// Name is the group name. Required.
	Name string
	// GID is the numeric group ID. If zero, the system assigns one.
	GID int
	// System creates a system group.
	System bool
}

// GroupUpdateOpts contains options for updating a group.
type GroupUpdateOpts struct {
	// Members is the list of group member usernames (replaces existing).
	Members []string
}

// groupInfoCollectionFromGen converts the rows of a list or a get.
func groupInfoCollectionFromGen(
	rows []gen.GroupAccount,
	jobID *openapi_types.UUID,
) Collection[GroupInfoResult] {
	results := make([]GroupInfoResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, GroupInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			GID:      derefInt(r.Gid),
			Members:  derefStringSlice(r.Members),
			Error:    derefString(r.Error),
		})
	}

	return Collection[GroupInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}

// groupMutationCollectionFromGen converts the rows of a create, an update or
// a delete.
func groupMutationCollectionFromGen(
	rows []gen.GroupResult,
	jobID *openapi_types.UUID,
) Collection[GroupMutationResult] {
	results := make([]GroupMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, GroupMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[GroupMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}
