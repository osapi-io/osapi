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

package group

// ListInput is list's input: it takes none.
type ListInput struct{}

// GetInput names the group to read.
type GetInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The group, such as developers"`
}

// CreateInput is the group to add.
type CreateInput struct {
	Name   string `json:"name"             validate:"required,max=32,account_name" doc:"The group to add, such as developers"                       telemetry:"attr"`
	GID    int    `json:"gid,omitempty"    validate:"omitempty,min=0"              doc:"Its numeric group ID; the system assigns one when it is 0"`
	System bool   `json:"system,omitempty" validate:"omitempty"                    doc:"Add it as a system group, with an ID from the system range"`
}

// UpdateInput sets the members of a group.
type UpdateInput struct {
	Name    string   `json:"name"              validate:"required,max=32,account_name"       doc:"The group, such as developers"`
	Members []string `json:"members,omitempty" validate:"omitempty,dive,max=32,account_name" doc:"Its members, which replace the members it has; none when empty"`
}

// DeleteInput names the group to remove.
type DeleteInput struct {
	Name string `json:"name" validate:"required,max=32,account_name" doc:"The group to remove, such as developers"`
}

// Account is a group of the host.
type Account struct {
	Name    string   `json:"name"              doc:"The group"`
	GID     int      `json:"gid"               doc:"Its numeric group ID"`
	Members []string `json:"members,omitempty" doc:"The users it has as members, besides those whose primary group it is" none:"none"`
}

// Result is what a write did.
type Result struct {
	Name    string `json:"name"    doc:"The group"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
