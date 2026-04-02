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

// Package user provides user and group management operations.
//
// Password and SSH key operations accept values inline in the request
// body rather than via Object Store + file.Deployer (which cron and
// certificate providers use). These are small values attached to a
// user account, not file deployments — Object Store would add
// unnecessary ceremony. SSH key fingerprints serve as the natural
// identity, so no SHA state tracking is needed.
package user

import "context"

// Provider implements the methods to manage users and groups.
type Provider interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, name string) (*User, error)
	CreateUser(ctx context.Context, opts CreateUserOpts) (*Result, error)
	UpdateUser(ctx context.Context, name string, opts UpdateUserOpts) (*Result, error)
	DeleteUser(ctx context.Context, name string) (*Result, error)
	ChangePassword(ctx context.Context, name string, password string) (*Result, error)
	ListGroups(ctx context.Context) ([]Group, error)
	GetGroup(ctx context.Context, name string) (*Group, error)
	CreateGroup(ctx context.Context, opts CreateGroupOpts) (*GroupResult, error)
	UpdateGroup(ctx context.Context, name string, opts UpdateGroupOpts) (*GroupResult, error)
	DeleteGroup(ctx context.Context, name string) (*GroupResult, error)
	ListKeys(ctx context.Context, username string) ([]SSHKey, error)
	AddKey(ctx context.Context, username string, key SSHKey) (*SSHKeyResult, error)
	RemoveKey(ctx context.Context, username string, fingerprint string) (*SSHKeyResult, error)
}

// User represents a system user account.
type User struct {
	Name   string   `json:"name"`
	UID    int      `json:"uid"`
	GID    int      `json:"gid"`
	Home   string   `json:"home"`
	Shell  string   `json:"shell"`
	Groups []string `json:"groups,omitempty"`
	Locked bool     `json:"locked"`
}

// CreateUserOpts contains options for creating a new user.
type CreateUserOpts struct {
	Name     string   `json:"name"`
	UID      int      `json:"uid,omitempty"`
	GID      int      `json:"gid,omitempty"`
	Home     string   `json:"home,omitempty"`
	Shell    string   `json:"shell,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Password string   `json:"password,omitempty"`
	System   bool     `json:"system,omitempty"`
}

// UpdateUserOpts contains options for updating an existing user.
type UpdateUserOpts struct {
	Shell  string   `json:"shell,omitempty"`
	Home   string   `json:"home,omitempty"`
	Groups []string `json:"groups,omitempty"`
	Lock   *bool    `json:"lock,omitempty"`
}

// Group represents a system group.
type Group struct {
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members,omitempty"`
}

// CreateGroupOpts contains options for creating a new group.
type CreateGroupOpts struct {
	Name   string `json:"name"`
	GID    int    `json:"gid,omitempty"`
	System bool   `json:"system,omitempty"`
}

// UpdateGroupOpts contains options for updating an existing group.
type UpdateGroupOpts struct {
	Members []string `json:"members,omitempty"`
}

// Result represents the result of a user mutation operation.
type Result struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// GroupResult represents the result of a group mutation operation.
type GroupResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// SSHKey represents an SSH public key from authorized_keys.
type SSHKey struct {
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
	RawLine     string `json:"raw_line,omitempty"`
}

// SSHKeyResult represents the result of an SSH key mutation operation.
type SSHKeyResult struct {
	Changed bool `json:"changed"`
}
