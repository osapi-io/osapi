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

package types

import "slices"

// Reason is the reason of an error answer, one of its Values. A program
// branches on it, so a reason outside the list breaks the contract of the
// spec's ErrorResponse.
type Reason string

// The reasons of an error answer: why the request failed, as a word that a
// program branches on. The HTTP status is the class of the problem, the
// reason its cause, and the message is for a person. The ErrorResponse
// schema of the spec lists the same words.
const (
	ReasonInvalidRequest Reason = "invalid_request"
	ReasonUnauthorized   Reason = "unauthorized"
	ReasonForbidden      Reason = "forbidden"
	ReasonNotFound       Reason = "not_found"
	ReasonNotManaged     Reason = "not_managed"
	ReasonNotInstalled   Reason = "not_installed"
	ReasonConflict       Reason = "conflict"
	ReasonUnavailable    Reason = "unavailable"
	ReasonJobStoreFull   Reason = "job_store_full"
	ReasonServerError    Reason = "server_error"
)

// Values are each reason, in the order of the schema.
func (Reason) Values() []Reason {
	return []Reason{
		ReasonInvalidRequest, ReasonUnauthorized, ReasonForbidden, ReasonNotFound,
		ReasonNotManaged, ReasonNotInstalled, ReasonConflict, ReasonUnavailable,
		ReasonJobStoreFull, ReasonServerError,
	}
}

// Valid reports whether r is one of the reasons.
func (r Reason) Valid() bool {
	return slices.Contains(r.Values(), r)
}
