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

// UserInfoResult represents a user listing result from a query operation.
type UserInfoResult struct {
	Hostname string   `json:"hostname"`
	Status   string   `json:"status"`
	Name     string   `json:"name,omitempty"`
	UID      int      `json:"uid,omitempty"`
	GID      int      `json:"gid,omitempty"`
	Home     string   `json:"home,omitempty"`
	Shell    string   `json:"shell,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Locked   bool     `json:"locked,omitempty"`
	Error    string   `json:"error,omitempty"`
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
	// Password is the initial password, hashed by the controller before
	// storage.
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

// SSHKeyInfoResult represents SSH key list result for one host.
type SSHKeyInfoResult struct {
	Hostname string       `json:"hostname"`
	Status   string       `json:"status"`
	Keys     []SSHKeyInfo `json:"keys,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// SSHKeyInfo represents a single SSH authorized key.
type SSHKeyInfo struct {
	Type        string `json:"type,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// SSHKeyMutationResult represents SSH key add/remove result for one host.
type SSHKeyMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// SSHKeyAddOpts contains options for adding an SSH key.
type SSHKeyAddOpts struct {
	// Key is the full SSH public key line (e.g., "ssh-ed25519 AAAA... user@host").
	Key string
}

// sshKeyCollectionFromGen converts a gen.SSHKeyCollectionResponse
// to a Collection[SSHKeyInfoResult].
func sshKeyCollectionFromGen(
	g *gen.SSHKeyCollectionResponse,
) Collection[SSHKeyInfoResult] {
	results := make([]SSHKeyInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, sshKeyInfoResultFromGen(r))
	}

	return Collection[SSHKeyInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sshKeyInfoResultFromGen converts a gen.SSHKeyEntry to an SSHKeyInfoResult.
func sshKeyInfoResultFromGen(
	r gen.SSHKeyEntry,
) SSHKeyInfoResult {
	result := SSHKeyInfoResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Error:    derefString(r.Error),
	}

	if r.Keys != nil {
		keys := make([]SSHKeyInfo, 0, len(*r.Keys))
		for _, k := range *r.Keys {
			keys = append(keys, sshKeyInfoFromGen(k))
		}

		result.Keys = keys
	}

	return result
}

// sshKeyInfoFromGen converts a gen.SSHKeyInfo to an SSHKeyInfo.
func sshKeyInfoFromGen(
	k gen.SSHKeyInfo,
) SSHKeyInfo {
	return SSHKeyInfo{
		Type:        derefString(k.Type),
		Fingerprint: derefString(k.Fingerprint),
		Comment:     derefString(k.Comment),
	}
}

// sshKeyMutationCollectionFromGen converts a gen.SSHKeyMutationResponse
// to a Collection[SSHKeyMutationResult].
func sshKeyMutationCollectionFromGen(
	g *gen.SSHKeyMutationResponse,
) Collection[SSHKeyMutationResult] {
	results := make([]SSHKeyMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, sshKeyMutationResultFromGen(r))
	}

	return Collection[SSHKeyMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// sshKeyMutationResultFromGen converts a gen.SSHKeyMutationEntry
// to an SSHKeyMutationResult.
func sshKeyMutationResultFromGen(
	r gen.SSHKeyMutationEntry,
) SSHKeyMutationResult {
	return SSHKeyMutationResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Changed:  derefBool(r.Changed),
		Error:    derefString(r.Error),
	}
}

func userInfoCollectionFromGen(
	rows []gen.UserAccount,
	jobID *openapi_types.UUID,
) Collection[UserInfoResult] {
	results := make([]UserInfoResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, UserInfoResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			UID:      derefInt(r.Uid),
			GID:      derefInt(r.Gid),
			Home:     derefString(r.Home),
			Shell:    derefString(r.Shell),
			Groups:   derefStringSlice(r.Groups),
			Locked:   derefBool(r.Locked),
			Error:    derefString(r.Error),
		})
	}

	return Collection[UserInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}

func userMutationCollectionFromGen(
	rows []gen.UserResult,
	jobID *openapi_types.UUID,
) Collection[UserMutationResult] {
	results := make([]UserMutationResult, 0, len(rows))
	for _, r := range rows {
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
		JobID:   jobIDFromGen(jobID),
	}
}
