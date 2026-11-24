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

package user

import "github.com/osapi-io/osapi/internal/secret"

// ListInput is list's input: it takes none.
type ListInput struct{}

// GetInput names the user to read.
type GetInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The user, such as alice"`
}

// CreateInput is the user to add. The caller sends password; the
// controller's step hashes it into password_hash, so the agent, the job data
// and NATS have only the hash.
type CreateInput struct {
	Name         string       `json:"name"                    validate:"required,max=32,account_name"        doc:"The user to add, such as alice"                                      telemetry:"attr"`
	UID          int          `json:"uid,omitempty"           validate:"omitempty,min=0"                     doc:"Its numeric user ID; the system assigns one when it is 0"`
	GID          int          `json:"gid,omitempty"           validate:"omitempty,min=0"                     doc:"Its primary group ID; a group of its own when it is 0"`
	Home         string       `json:"home,omitempty"          validate:"omitempty,startswith=/,no_linebreak" doc:"Its home directory; /home/<name> when empty"`
	Shell        string       `json:"shell,omitempty"         validate:"omitempty,startswith=/,no_linebreak" doc:"Its login shell, such as /bin/bash; the system's default when empty"`
	Groups       []string     `json:"groups,omitempty"        validate:"omitempty,dive,max=32,account_name"  doc:"The groups it is a member of, besides its primary group"`
	System       bool         `json:"system,omitempty"        validate:"omitempty"                           doc:"Add it as a system user, with an ID from the system range"`
	Password     secret.Value `json:"password,omitempty"      validate:"omitempty,no_linebreak"              doc:"Its password, which osapi hashes on the controller; none when empty"`
	PasswordHash string       `json:"password_hash,omitempty" validate:"omitempty"                           doc:"The sha512-crypt hash of the password, which the controller sets"                     before:"set"`
}

// UpdateInput changes a user. An empty field keeps what the user has.
type UpdateInput struct {
	Name   string   `json:"name"             validate:"required,max=32,account_name"        doc:"The user, such as alice"`
	Shell  string   `json:"shell,omitempty"  validate:"omitempty,startswith=/,no_linebreak" doc:"Its login shell, such as /bin/bash"`
	Home   string   `json:"home,omitempty"   validate:"omitempty,startswith=/,no_linebreak" doc:"Its home directory, which usermod moves the files to"`
	Groups []string `json:"groups,omitempty" validate:"omitempty,dive,max=32,account_name"  doc:"The groups it is a member of, which replace the groups it has"`
	Lock   string   `json:"lock,omitempty"   validate:"omitempty,oneof=locked unlocked"     doc:"Whether to refuse logins with its password: locked refuses them, unlocked allows them"`
}

// DeleteInput names the user to remove.
type DeleteInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The user to remove, with its home directory, such as alice"`
}

// Account is a user of the host.
type Account struct {
	Name   string   `json:"name"             doc:"The user"`
	UID    int      `json:"uid"              doc:"Its numeric user ID"`
	GID    int      `json:"gid"              doc:"Its primary group ID"`
	Home   string   `json:"home"             doc:"Its home directory"`
	Shell  string   `json:"shell"            doc:"Its login shell"`
	Groups []string `json:"groups,omitempty" doc:"The groups it is a member of, besides its primary group" none:"none"`
	Locked bool     `json:"locked"           doc:"Whether a login with its password is refused"`
}

// Result is what a write did.
type Result struct {
	Name    string `json:"name"    doc:"The user"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
