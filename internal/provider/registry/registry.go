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

// Package registry provides a runtime registry for provider operations.
package registry

import "context"

// OperationSpec defines how to create params and run an operation.
type OperationSpec struct {
	NewParams func() any
	Run       func(ctx context.Context, params any) (any, error)
}

// Registration describes a provider and its operations.
type Registration struct {
	Name       string
	Operations map[string]OperationSpec
}

// Registry holds provider registrations.
type Registry struct {
	providers map[string]Registration
}

// New creates a new empty registry.
func New() *Registry {
	return &Registry{
		providers: make(map[string]Registration),
	}
}

// Register adds a provider registration.
func (r *Registry) Register(
	reg Registration,
) {
	r.providers[reg.Name] = reg
}

// Lookup finds an operation spec by provider and operation name.
func (r *Registry) Lookup(
	provider string,
	operation string,
) (*OperationSpec, bool) {
	reg, ok := r.providers[provider]
	if !ok {
		return nil, false
	}

	spec, ok := reg.Operations[operation]
	if !ok {
		return nil, false
	}

	return &spec, true
}
