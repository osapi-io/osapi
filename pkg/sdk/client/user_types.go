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

// SSHKeyInfoResult is one authorized key of a user on one host.
type SSHKeyInfoResult struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Type        string `json:"type,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Options     string `json:"options,omitempty"`
	Error       string `json:"error,omitempty"`
}

// SSHKeyMutationResult is what an add or a remove of a key did on one host.
type SSHKeyMutationResult struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Changed     bool   `json:"changed"`
	Error       string `json:"error,omitempty"`
}

// SSHKeyAddOpts contains options for adding an SSH key.
type SSHKeyAddOpts struct {
	// Key is the full public key line, such as ssh-ed25519 AAAA... alice@laptop.
	Key string
}

func sshKeyCollectionFromGen(
	rows []gen.UserSshkeyKey,
	jobID *openapi_types.UUID,
) Collection[SSHKeyInfoResult] {
	results := make([]SSHKeyInfoResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, SSHKeyInfoResult{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Type:        derefString(r.Type),
			Fingerprint: derefString(r.Fingerprint),
			Comment:     derefString(r.Comment),
			Options:     derefString(r.Options),
			Error:       derefString(r.Error),
		})
	}

	return Collection[SSHKeyInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}

func sshKeyMutationCollectionFromGen(
	rows []gen.UserSshkeyResult,
	jobID *openapi_types.UUID,
) Collection[SSHKeyMutationResult] {
	results := make([]SSHKeyMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, SSHKeyMutationResult{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Fingerprint: derefString(r.Fingerprint),
			Changed:     derefBool(r.Changed),
			Error:       derefString(r.Error),
		})
	}

	return Collection[SSHKeyMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
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
