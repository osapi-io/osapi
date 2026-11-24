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

// RowStatus is the status of one host in the answer of a behavior method:
// what happened on that host.
type RowStatus string

// The statuses of a host's row.
const (
	// RowStatusOK is a host where the operation ran and succeeded.
	RowStatusOK RowStatus = "ok"
	// RowStatusFailed is a host where the operation ran and did not succeed.
	RowStatusFailed RowStatus = "failed"
	// RowStatusSkipped is a host whose OS family the operation does not
	// support, so it did not run there.
	RowStatusSkipped RowStatus = "skipped"
	// RowStatusTimeout is a host that did not answer. It is not failed: a
	// timeout says nothing about whether the operation ran, so an operator
	// has to go and look.
	RowStatusTimeout RowStatus = "timeout"
)

// Values are each status of a row, in the order of the schema.
func (RowStatus) Values() []RowStatus {
	return []RowStatus{RowStatusOK, RowStatusFailed, RowStatusSkipped, RowStatusTimeout}
}

// Valid is whether s is one of Values.
func (s RowStatus) Valid() bool {
	return slices.Contains(s.Values(), s)
}
