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

// Package ntp provides NTP server management via chrony.
package ntp

import "context"

// Provider implements the methods to manage NTP configuration.
type Provider interface {
	// Get returns current NTP sync status and configured servers.
	Get(ctx context.Context) (*Status, error)
	// Create deploys a managed NTP server configuration. Idempotent: returns Changed: false if already managed.
	Create(ctx context.Context, config Config) (*CreateResult, error)
	// Update replaces the managed NTP server configuration. Fails if not managed.
	Update(ctx context.Context, config Config) (*UpdateResult, error)
	// Delete removes the managed NTP server configuration.
	Delete(ctx context.Context) (*DeleteResult, error)
}

// Config represents an NTP server configuration to deploy.
type Config struct {
	Servers []string `json:"servers"`
}

// Status represents the current NTP sync state and configured servers.
type Status struct {
	Synchronized  bool     `json:"synchronized"`
	Stratum       int      `json:"stratum,omitempty"`
	Offset        string   `json:"offset,omitempty"`
	CurrentSource string   `json:"current_source,omitempty"`
	Servers       []string `json:"servers,omitempty"`
}

// CreateResult represents the outcome of an NTP config create operation.
type CreateResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of an NTP config update operation.
type UpdateResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of an NTP config delete operation.
type DeleteResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
