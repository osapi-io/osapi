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

// Package route provides network route configuration management via Netplan.
package route

import "context"

// Provider manages route configuration via Netplan.
type Provider interface {
	// List returns all routes from the system routing table.
	List(ctx context.Context) ([]ListEntry, error)
	// Get returns the managed routes for a specific interface.
	Get(ctx context.Context, interfaceName string) (*Entry, error)
	// Create deploys new routes for an interface via Netplan.
	Create(ctx context.Context, entry Entry) (*Result, error)
	// Update redeploys routes for an existing interface via Netplan.
	Update(ctx context.Context, entry Entry) (*Result, error)
	// Delete removes managed routes for an interface via Netplan.
	Delete(ctx context.Context, interfaceName string) (*Result, error)
}

// Entry represents managed routes for an interface.
type Entry struct {
	Interface string  `json:"interface"`
	Routes    []Route `json:"routes"`
}

// Route is a single route definition.
type Route struct {
	To     string `json:"to"`
	Via    string `json:"via"`
	Metric int    `json:"metric,omitempty"`
}

// ListEntry is a route from the system routing table.
type ListEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Mask        string `json:"mask,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Flags       string `json:"flags,omitempty"`
}

// Result is the outcome of a route create/update/delete operation.
type Result struct {
	Interface string `json:"interface"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}
