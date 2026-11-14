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

import (
	"crypto/sha256"
	"encoding/hex"
)

// State is the record osapi keeps of a file it wrote, in the file-state
// store: what it wrote, where, with what checksum and when. Every part of
// osapi that writes files on a host records them this way, and the
// controller reads the records to report drift, so the format lives with the
// behavior that owns managed files.
type State struct {
	ObjectName   string            `json:"object_name"`
	Path         string            `json:"path"`
	SHA256       string            `json:"sha256"`
	Mode         string            `json:"mode,omitempty"`
	Owner        string            `json:"owner,omitempty"`
	Group        string            `json:"group,omitempty"`
	DeployedAt   string            `json:"deployed_at"`
	ContentType  string            `json:"content_type"`
	UndeployedAt string            `json:"undeployed_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StateKey is where a file's State is kept: the host's name and the SHA-256
// of the file's path, so a path's characters never reach the key.
func StateKey(
	hostname string,
	path string,
) string {
	sum := sha256.Sum256([]byte(path))

	return hostname + "." + hex.EncodeToString(sum[:])
}
