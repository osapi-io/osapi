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

import "context"

// DeployRequest contains parameters for deploying a file to disk.
type DeployRequest struct {
	// ObjectName is the name of the object in the NATS object store.
	ObjectName string `json:"object_name"`
	// Path is the destination path on the target filesystem.
	Path string `json:"path"`
	// Mode is the file permission mode (e.g., "0644").
	Mode string `json:"mode,omitempty"`
	// Owner is the file owner user.
	Owner string `json:"owner,omitempty"`
	// Group is the file owner group.
	Group string `json:"group,omitempty"`
	// ContentType specifies whether the content is "raw" or "template".
	ContentType string `json:"content_type"`
	// Vars contains template variables when ContentType is "template".
	Vars map[string]any `json:"vars,omitempty"`
}

// DeployResult contains the result of a file deploy operation.
type DeployResult struct {
	// Changed indicates whether the file was written (false if SHA matched).
	Changed bool `json:"changed"`
	// SHA256 is the SHA-256 hash of the deployed file content.
	SHA256 string `json:"sha256"`
	// Path is the destination path where the file was deployed.
	Path string `json:"path"`
}

// StatusRequest contains parameters for checking file status.
type StatusRequest struct {
	// Path is the filesystem path to check.
	Path string `json:"path"`
}

// StatusResult contains the result of a file status check.
type StatusResult struct {
	// Path is the filesystem path that was checked.
	Path string `json:"path"`
	// Status indicates the file state: "in-sync", "drifted", or "missing".
	Status string `json:"status"`
	// SHA256 is the current SHA-256 hash of the file on disk, if present.
	SHA256 string `json:"sha256,omitempty"`
}

// Provider defines the interface for file operations.
type Provider interface {
	// Deploy writes file content to the target path with the specified
	// permissions. Returns whether the file was changed and its SHA-256.
	Deploy(
		ctx context.Context,
		req DeployRequest,
	) (*DeployResult, error)
	// Status checks the current state of a deployed file against its
	// expected SHA-256 from the file-state KV.
	Status(
		ctx context.Context,
		req StatusRequest,
	) (*StatusResult, error)
}
