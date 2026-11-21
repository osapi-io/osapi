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

package netplan

// Dir is where netplan reads its config, and where osapi writes its own
// files, each named osapi-<something>.yaml.
const Dir = "/etc/netplan"

// UpdateInput is a netplan file to write and apply.
type UpdateInput struct {
	Path     string            `json:"path"               validate:"required,startswith=/etc/netplan/osapi-" doc:"The file, under /etc/netplan, named osapi-<something>.yaml" telemetry:"attr"`
	Content  string            `json:"content"            validate:"required"                                doc:"The file's whole YAML"`
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty"                               doc:"What the caller keeps in the file's record"`
}

// DeleteInput is a netplan file to remove.
type DeleteInput struct {
	Path string `json:"path" validate:"required,startswith=/etc/netplan/osapi-" doc:"The file, under /etc/netplan, named osapi-<something>.yaml" telemetry:"attr"`
}

// GetInput names an interface.
type GetInput struct {
	Name string `json:"name" validate:"required,min=1" doc:"The interface, such as eth0"`
}

// Interface is an interface as netplan status reports it.
type Interface struct {
	Name    string   `json:"name"    doc:"The interface"`
	Type    string   `json:"type"    doc:"Its type, such as ethernet, bond or wifi; empty when netplan does not report it"`
	Section string   `json:"section" doc:"The netplan section that configures it, such as ethernets or bonds"`
	ID      string   `json:"id"      doc:"The netplan ID of the definition that matches it, such as all for a match on en*; the interface's name when netplan does not define it"`
	Matches []string `json:"matches" doc:"Each interface that the definition matches, this one too, in name order"`
}

// Result is a netplan file after a write or a removal.
type Result struct {
	Path    string `json:"path"    doc:"The file"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
