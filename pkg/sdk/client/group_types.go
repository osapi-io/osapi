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

// GroupInfoResult represents a group listing result from a query operation.
type GroupInfoResult struct {
	Hostname string      `json:"hostname"`
	Status   string      `json:"status"`
	Groups   []GroupInfo `json:"groups,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// GroupInfo represents a group on the target node.
type GroupInfo struct {
	Name    string   `json:"name,omitempty"`
	GID     int      `json:"gid,omitempty"`
	Members []string `json:"members,omitempty"`
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

// groupInfoCollectionFromList converts a gen.GroupCollectionResponse
// to a Collection[GroupInfoResult].
func groupInfoCollectionFromList(
	g *gen.GroupCollectionResponse,
) Collection[GroupInfoResult] {
	results := make([]GroupInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		result := GroupInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Groups != nil {
			groups := make([]GroupInfo, 0, len(*r.Groups))
			for _, g := range *r.Groups {
				groups = append(groups, groupInfoFromGen(g))
			}

			result.Groups = groups
		}

		results = append(results, result)
	}

	return Collection[GroupInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// groupInfoCollectionFromGet converts a gen.GroupCollectionResponse (from
// get-by-name) to a Collection[GroupInfoResult].
func groupInfoCollectionFromGet(
	g *gen.GroupCollectionResponse,
) Collection[GroupInfoResult] {
	return groupInfoCollectionFromList(g)
}

// groupMutationCollectionFromCreate converts a gen.GroupMutationResponse
// to a Collection[GroupMutationResult].
func groupMutationCollectionFromCreate(
	g *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromGen(g)
}

// groupMutationCollectionFromUpdate converts a gen.GroupMutationResponse
// to a Collection[GroupMutationResult].
func groupMutationCollectionFromUpdate(
	g *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromGen(g)
}

// groupMutationCollectionFromDelete converts a gen.GroupMutationResponse
// to a Collection[GroupMutationResult].
func groupMutationCollectionFromDelete(
	g *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	return groupMutationCollectionFromGen(g)
}

// groupMutationCollectionFromGen is the shared converter for all group
// mutation response types.
func groupMutationCollectionFromGen(
	g *gen.GroupMutationResponse,
) Collection[GroupMutationResult] {
	results := make([]GroupMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
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
		JobID:   jobIDFromGen(g.JobId),
	}
}

// groupInfoFromGen converts a gen.GroupInfo to a GroupInfo.
func groupInfoFromGen(
	g gen.GroupInfo,
) GroupInfo {
	return GroupInfo{
		Name:    derefString(g.Name),
		GID:     derefInt(g.Gid),
		Members: derefStringSlice(g.Members),
	}
}
