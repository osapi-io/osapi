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

// Package provider defines shared types for all provider implementations.
package provider

// FactsFunc returns the current agent facts for use by providers.
// Called at execution time so providers always get the latest facts.
type FactsFunc func() map[string]any

// FactsAware gives the file provider in internal/provider/file the facts
// that its templates read. A behavior's provider does not embed it: it reads
// the agent's facts from the host it gets, as h.Facts. The agent does not use
// the file provider, thus osapi-io/osapi#1151 deletes both.
type FactsAware struct {
	factsFn FactsFunc
}

// SetFactsFunc sets the facts getter.
func (f *FactsAware) SetFactsFunc(
	fn FactsFunc,
) {
	f.factsFn = fn
}

// Facts returns the current agent facts, or nil if not available.
func (f *FactsAware) Facts() map[string]any {
	if f.factsFn == nil {
		return nil
	}

	return f.factsFn()
}
