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

// Package apt provides package management operations via apt.
package apt

import "context"

// Provider implements the methods to interact with apt package management.
type Provider interface {
	// List returns all installed packages.
	List(ctx context.Context) ([]Package, error)
	// Get returns details for a single installed package.
	Get(ctx context.Context, name string) (*Package, error)
	// Install installs a package by name.
	Install(ctx context.Context, name string) (*Result, error)
	// Remove removes a package by name.
	Remove(ctx context.Context, name string) (*Result, error)
	// Update refreshes the package index.
	Update(ctx context.Context) (*Result, error)
	// ListUpdates returns packages with available updates.
	ListUpdates(ctx context.Context) ([]Update, error)
}

// Package represents an installed apt package.
type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Size        int64  `json:"size,omitempty"`
}

// Update represents a package with an available update.
type Update struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version"`
}

// Result represents the outcome of a package mutation operation.
type Result struct {
	Name    string `json:"name,omitempty"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
