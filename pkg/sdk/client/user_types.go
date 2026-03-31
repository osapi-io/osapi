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

// UserInfoResult represents a user listing result from a query operation.
type UserInfoResult struct {
	Hostname string     `json:"hostname"`
	Status   string     `json:"status"`
	Users    []UserInfo `json:"users,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// UserInfo represents a user account on the target node.
type UserInfo struct {
	Name   string   `json:"name,omitempty"`
	UID    int      `json:"uid,omitempty"`
	GID    int      `json:"gid,omitempty"`
	Home   string   `json:"home,omitempty"`
	Shell  string   `json:"shell,omitempty"`
	Groups []string `json:"groups,omitempty"`
	Locked bool     `json:"locked,omitempty"`
}

// UserMutationResult represents the result of a user create, update, delete,
// or password operation.
type UserMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// UserCreateOpts contains options for creating a user account.
type UserCreateOpts struct {
	// Name is the username. Required.
	Name string
	// UID is the numeric user ID. If zero, the system assigns one.
	UID int
	// GID is the primary group ID. If zero, a group matching the username is created.
	GID int
	// Home is the home directory path.
	Home string
	// Shell is the login shell path.
	Shell string
	// Groups is the list of supplementary group names.
	Groups []string
	// Password is the initial password (plaintext, hashed by the agent).
	Password string
	// System creates a system account.
	System bool
}

// UserUpdateOpts contains options for updating a user account.
type UserUpdateOpts struct {
	// Shell is the new login shell path.
	Shell string
	// Home is the new home directory path.
	Home string
	// Groups is the list of supplementary group names (replaces existing).
	Groups []string
	// Lock locks or unlocks the account.
	Lock *bool
}

// userInfoCollectionFromList converts a gen.UserCollectionResponse
// to a Collection[UserInfoResult].
func userInfoCollectionFromList(
	g *gen.UserCollectionResponse,
) Collection[UserInfoResult] {
	results := make([]UserInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		result := UserInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Users != nil {
			users := make([]UserInfo, 0, len(*r.Users))
			for _, u := range *r.Users {
				users = append(users, userInfoFromGen(u))
			}

			result.Users = users
		}

		results = append(results, result)
	}

	return Collection[UserInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// userInfoCollectionFromGet converts a gen.UserCollectionResponse (from
// get-by-name) to a Collection[UserInfoResult].
func userInfoCollectionFromGet(
	g *gen.UserCollectionResponse,
) Collection[UserInfoResult] {
	return userInfoCollectionFromList(g)
}

// userMutationCollectionFromCreate converts a gen.UserMutationResponse
// to a Collection[UserMutationResult].
func userMutationCollectionFromCreate(
	g *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromGen(g)
}

// userMutationCollectionFromUpdate converts a gen.UserMutationResponse
// to a Collection[UserMutationResult].
func userMutationCollectionFromUpdate(
	g *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromGen(g)
}

// userMutationCollectionFromDelete converts a gen.UserMutationResponse
// to a Collection[UserMutationResult].
func userMutationCollectionFromDelete(
	g *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromGen(g)
}

// userMutationCollectionFromPassword converts a gen.UserMutationResponse
// to a Collection[UserMutationResult].
func userMutationCollectionFromPassword(
	g *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	return userMutationCollectionFromGen(g)
}

// userMutationCollectionFromGen is the shared converter for all user
// mutation response types.
func userMutationCollectionFromGen(
	g *gen.UserMutationResponse,
) Collection[UserMutationResult] {
	results := make([]UserMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, UserMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[UserMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// userInfoFromGen converts a gen.UserInfo to a UserInfo.
func userInfoFromGen(
	g gen.UserInfo,
) UserInfo {
	return UserInfo{
		Name:   derefString(g.Name),
		UID:    derefInt(g.Uid),
		GID:    derefInt(g.Gid),
		Home:   derefString(g.Home),
		Shell:  derefString(g.Shell),
		Groups: derefStringSlice(g.Groups),
		Locked: derefBool(g.Locked),
	}
}
