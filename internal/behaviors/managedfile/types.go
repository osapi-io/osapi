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

package managedfile

import "io/fs"

// UpdateInput is a file to write and keep a record of.
type UpdateInput struct {
	Path     string            `json:"path"     validate:"required,startswith=/" doc:"Absolute path of the file"`
	Content  string            `json:"content"  validate:"required"              doc:"The file's whole content"`
	Mode     string            `json:"mode"                                      doc:"Permissions in octal, such as 0600; 0644 when empty"`
	Owner    string            `json:"owner"                                     doc:"User that owns the file; left as written when empty"`
	Group    string            `json:"group"                                     doc:"Group that owns the file; left as written when empty"`
	Secret   bool              `json:"secret"                                    doc:"The content is a secret: its record keeps no checksum of it"`
	Metadata map[string]string `json:"metadata"                                  doc:"What the caller keeps in the file's record, such as the setting it holds"`
}

// DeleteInput is a file to stop managing.
type DeleteInput struct {
	Path string `json:"path" validate:"required,startswith=/" doc:"Absolute path of the file"`
}

// GetInput is a file whose record to read.
type GetInput struct {
	Path string `json:"path" validate:"required,startswith=/" doc:"Absolute path of the file"`
}

// Result is what a change did.
type Result struct {
	Path    string `json:"path"    doc:"Absolute path of the file"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}

// Record is what osapi keeps about a file it wrote.
type Record struct {
	Path       string            `json:"path"       doc:"Absolute path of the file"`
	SHA256     string            `json:"sha256"     doc:"Checksum of the content osapi wrote"`
	Metadata   map[string]string `json:"metadata"   doc:"What the caller keeps in the record"`
	Undeployed bool              `json:"undeployed" doc:"Whether osapi has stopped managing the file"`
}

// File is what Write makes a path hold.
type File struct {
	Path    string
	Content []byte
	// Mode is the file's permissions; 0644 when zero.
	Mode fs.FileMode
	// Owner and Group own the file when set, applied with chown.
	Owner string
	Group string
	// Secret keeps a checksum of the content out of the record, where a short
	// secret's could be recovered by guessing.
	Secret   bool
	Metadata map[string]string
}
