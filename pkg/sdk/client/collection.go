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

package client

import (
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Collection is a generic wrapper for collection responses from node queries.
type Collection[T any] struct {
	Results []T    `json:"results"`
	JobID   string `json:"job_id"`
}

// First returns the first result and true, or the zero value and
// false if the collection is empty.
func (c Collection[T]) First() (T, bool) {
	if len(c.Results) == 0 {
		var zero T

		return zero, false
	}

	return c.Results[0], true
}

// derefString safely dereferences a string pointer, returning empty string for nil.
func derefString(
	s *string,
) string {
	if s == nil {
		return ""
	}

	return *s
}

// derefInt safely dereferences an int pointer, returning zero for nil.
func derefInt(
	i *int,
) int {
	if i == nil {
		return 0
	}

	return *i
}

// derefInt64 safely dereferences an int64 pointer, returning zero for nil.
func derefInt64(
	i *int64,
) int64 {
	if i == nil {
		return 0
	}

	return *i
}

// derefFloat64 safely dereferences a float64 pointer, returning zero for nil.
func derefFloat64(
	f *float64,
) float64 {
	if f == nil {
		return 0
	}

	return *f
}

// derefFloat32 safely dereferences a float32 pointer, returning zero for nil.
func derefFloat32(
	f *float32,
) float32 {
	if f == nil {
		return 0
	}

	return *f
}

// derefBool safely dereferences a bool pointer, returning false for nil.
func derefBool(
	b *bool,
) bool {
	if b == nil {
		return false
	}

	return *b
}

// jobIDFromGen extracts a job ID string from an optional UUID pointer.
func jobIDFromGen(
	id *openapi_types.UUID,
) string {
	if id == nil {
		return ""
	}

	return id.String()
}
