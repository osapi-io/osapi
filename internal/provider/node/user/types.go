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

// Package user provides the SSH key operations of a user's account.
//
// SSH key operations accept values inline in the request
// body rather than via Object Store + file.Deployer (which cron and
// certificate providers use). These are small values attached to a
// user account, not file deployments — Object Store would add
// unnecessary ceremony. SSH key fingerprints serve as the natural
// identity, so no SHA state tracking is needed.
package user

import "context"

// Provider manages the authorized SSH keys of a user.
type Provider interface {
	ListKeys(ctx context.Context, username string) ([]SSHKey, error)
	AddKey(ctx context.Context, username string, key SSHKey) (*SSHKeyResult, error)
	RemoveKey(ctx context.Context, username string, fingerprint string) (*SSHKeyResult, error)
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
