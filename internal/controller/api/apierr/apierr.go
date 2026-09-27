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

// Package apierr answers, in one place, which HTTP status a failed job deserves.
//
// It exists so that thirty-odd handlers do not each decide for themselves. They
// used to, by matching on the words a provider happened to use, and the answers
// drifted: one endpoint checked for "does not exist" while its neighbour checked
// "not found", a third checked neither and answered 500 where its own
// specification promised 404, and two checked for a phrase the code they call can
// never produce.
package apierr

import (
	"errors"

	"github.com/osapi-io/osapi/internal/job"
)

// IsMissing reports whether a failed job means the thing the endpoint addresses
// is not there as far as osapi is concerned, which the API answers with 404.
//
// Three causes qualify, and they stay distinct behind this one question because
// they are not the same thing to an operator: the host has no such user, the file
// is there but osapi does not manage it, and the program the operation needs is
// not installed. An endpoint that ever needs to answer them differently asks
// errors.Is directly.
func IsMissing(
	err error,
) bool {
	return errors.Is(err, job.ErrNotFound) ||
		errors.Is(err, job.ErrNotManaged) ||
		errors.Is(err, job.ErrNotInstalled)
}
