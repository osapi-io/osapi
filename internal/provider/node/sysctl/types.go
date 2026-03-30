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

// Package sysctl provides kernel parameter management via /etc/sysctl.d/.
// It is a meta-provider that delegates file writes to the file provider
// for SHA tracking, idempotency, and drift detection.
package sysctl

import "context"

// Provider implements the methods to manage sysctl entries.
type Provider interface {
	// List returns all osapi-managed sysctl entries with current runtime values.
	List(ctx context.Context) ([]Entry, error)
	// Get returns a single sysctl entry by key with current runtime value.
	Get(ctx context.Context, key string) (*Entry, error)
	// Set deploys a sysctl conf file and applies it. Idempotent.
	Set(ctx context.Context, entry Entry) (*SetResult, error)
	// Delete removes a managed sysctl conf file and reloads defaults.
	Delete(ctx context.Context, key string) (*DeleteResult, error)
}

// Entry represents a sysctl kernel parameter.
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetResult represents the outcome of a sysctl set operation.
type SetResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a sysctl delete operation.
type DeleteResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
