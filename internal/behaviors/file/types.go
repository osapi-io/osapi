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

package file

// UpdateInput is what update writes: an object of the object store, at a
// path, with its mode and ownership.
type UpdateInput struct {
	ObjectName  string            `json:"object_name"            validate:"required,min=1,max=255"       doc:"The object of the object store, from osapi client file list"`
	Path        string            `json:"path"                   validate:"required,startswith=/"        doc:"Absolute path of the file on the host"                                                                                       telemetry:"attr"`
	Mode        string            `json:"mode,omitempty"         validate:"omitempty,file_mode"          doc:"Permissions in octal, such as 0644; when empty, a file keeps its mode and a new file gets 0644"`
	Owner       string            `json:"owner,omitempty"        validate:"omitempty,account_name"       doc:"User that owns the file; left as written when empty"`
	Group       string            `json:"group,omitempty"        validate:"omitempty,account_name"       doc:"Group that owns the file; left as written when empty"`
	ContentType string            `json:"content_type,omitempty" validate:"omitempty,oneof=raw template" doc:"How the object is written: raw as it is, or template rendered with .Facts, .Vars and .Hostname; the object's own when empty"`
	Vars        map[string]string `json:"vars,omitempty"         validate:"omitempty"                    doc:"Variables of a template, as .Vars, each as key=value"`
}

// UpdateResult is the file that update wrote.
type UpdateResult struct {
	Path    string `json:"path"    doc:"Absolute path of the file"`
	SHA256  string `json:"sha256"  doc:"SHA-256 of the content that the file holds"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}

// DeleteInput names the file that delete removes.
type DeleteInput struct {
	Path string `json:"path" validate:"required,startswith=/" doc:"Absolute path of the file on the host" telemetry:"attr"`
}

// DeleteResult is the file that delete removed.
type DeleteResult struct {
	Path    string `json:"path"    doc:"Absolute path of the file"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}

// StatusInput names the file that status reads.
type StatusInput struct {
	Path string `json:"path" validate:"required,startswith=/" doc:"Absolute path of the file on the host"`
}

// The states of a deployed file.
const (
	// InSync is a file that holds what osapi deployed.
	InSync = "in-sync"
	// Drifted is a file that something changed after osapi deployed it.
	Drifted = "drifted"
	// Missing is a file that osapi did not deploy, or that is gone.
	Missing = "missing"
)

// Status is how a file is against what osapi deployed.
type Status struct {
	Path   string `json:"path"             doc:"Absolute path of the file"`
	State  string `json:"state"            doc:"in-sync, drifted or missing"`
	SHA256 string `json:"sha256,omitempty" doc:"SHA-256 of the file on the host, when it is there"`
}
