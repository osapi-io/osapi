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

// Package cron provides management of cron drop-in files and periodic scripts.
// Supports /etc/cron.d/ (custom schedules) and /etc/cron.{hourly,daily,weekly,monthly}/
// (interval-based scripts).
package cron

// Provider implements the methods to manage cron entries.
type Provider interface {
	// List returns all osapi-managed cron entries.
	List() ([]Entry, error)
	// Get returns a single cron entry by name.
	Get(name string) (*Entry, error)
	// Create writes a new cron entry.
	Create(entry Entry) (*CreateResult, error)
	// Update overwrites an existing cron entry.
	Update(entry Entry) (*UpdateResult, error)
	// Delete removes a cron entry.
	Delete(name string) (*DeleteResult, error)
}

// Entry represents a cron entry — either a /etc/cron.d/ drop-in file
// with a custom schedule or a /etc/cron.{interval}/ periodic script.
type Entry struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule,omitempty"`
	Interval string `json:"interval,omitempty"`
	Source   string `json:"source,omitempty"`
	User     string `json:"user,omitempty"`
	Command  string `json:"command"`
}

// CreateResult represents the outcome of a cron entry creation.
type CreateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UpdateResult represents the outcome of a cron entry update.
type UpdateResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult represents the outcome of a cron entry deletion.
type DeleteResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
