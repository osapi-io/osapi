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

// Package certificate provides CA certificate management operations.
// Supports listing system and custom CA certificates, and managing custom
// certificates in /usr/local/share/ca-certificates/. Delegates file writes
// to the file provider for SHA tracking, idempotency, and template rendering.
package certificate

import "context"

// Provider implements the methods to manage CA certificates.
type Provider interface {
	// List returns all system and osapi-managed CA certificates.
	List(ctx context.Context) ([]Entry, error)
	// Create deploys a new CA certificate via the file provider.
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	// Update redeploys an existing CA certificate via the file provider.
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	// Delete undeploys a CA certificate via the file provider.
	Delete(ctx context.Context, name string) (*DeleteResult, error)
}

// Entry represents a CA certificate entry.
type Entry struct {
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
	Object string `json:"object,omitempty"`
}

// CreateResult represents the outcome of a CA certificate creation.
type CreateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a CA certificate update.
type UpdateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a CA certificate deletion.
type DeleteResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
