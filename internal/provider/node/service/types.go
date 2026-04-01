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

// Package service provides systemd service management operations.
// Supports listing, inspecting, and controlling systemd services, and managing
// custom unit files in /etc/systemd/system/. Delegates unit file writes to the
// file provider for SHA tracking, idempotency, and template rendering.
package service

import "context"

// Provider implements systemd service management operations.
type Provider interface {
	// List returns all systemd services.
	List(ctx context.Context) ([]Info, error)
	// Get returns a single systemd service by name.
	Get(ctx context.Context, name string) (*Info, error)
	// Create deploys a new unit file via the file provider.
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	// Update redeploys an existing unit file via the file provider.
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	// Delete undeploys a unit file via the file provider.
	Delete(ctx context.Context, name string) (*DeleteResult, error)
	// Start starts a systemd service.
	Start(ctx context.Context, name string) (*ActionResult, error)
	// Stop stops a systemd service.
	Stop(ctx context.Context, name string) (*ActionResult, error)
	// Restart restarts a systemd service.
	Restart(ctx context.Context, name string) (*ActionResult, error)
	// Enable enables a systemd service.
	Enable(ctx context.Context, name string) (*ActionResult, error)
	// Disable disables a systemd service.
	Disable(ctx context.Context, name string) (*ActionResult, error)
}

// Info represents a systemd service.
type Info struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

// Entry represents a unit file deployment request.
type Entry struct {
	Name   string `json:"name"`
	Object string `json:"object,omitempty"`
}

// CreateResult represents the outcome of a unit file creation.
type CreateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a unit file update.
type UpdateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a unit file deletion.
type DeleteResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ActionResult represents the outcome of a service control action.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
}
