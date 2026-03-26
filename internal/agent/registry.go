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

package agent

import (
	"encoding/json"
	"fmt"

	"github.com/retr0h/osapi/internal/job"
)

// ProcessorFunc is a function that handles a job request and returns
// a JSON-encoded result.
type ProcessorFunc func(job.Request) (json.RawMessage, error)

// ProviderRegistry maps job categories to processor functions and
// tracks all registered providers for lifecycle wiring (e.g., facts).
type ProviderRegistry struct {
	processors map[string]ProcessorFunc
	providers  []any
}

// NewProviderRegistry creates an empty ProviderRegistry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		processors: make(map[string]ProcessorFunc),
	}
}

// Register associates a category with a processor function and records
// any providers for WireProviderFacts.
func (r *ProviderRegistry) Register(
	category string,
	processFn ProcessorFunc,
	providers ...any,
) {
	r.processors[category] = processFn
	r.providers = append(r.providers, providers...)
}

// Dispatch routes a job request to the registered processor for its category.
func (r *ProviderRegistry) Dispatch(
	req job.Request,
) (json.RawMessage, error) {
	fn, ok := r.processors[req.Category]
	if !ok {
		return nil, fmt.Errorf("unsupported job category: %s", req.Category)
	}

	return fn(req)
}

// AllProviders returns all providers registered across all categories.
func (r *ProviderRegistry) AllProviders() []any {
	return r.providers
}
