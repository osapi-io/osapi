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

	components "github.com/osapi-io/osapi/internal/controller/api/components/gen"
	"github.com/osapi-io/osapi/internal/job"
)

// IsMissing reports whether a failed job means the thing the endpoint addresses
// is not there as far as osapi is concerned: the host has no such item, the
// item is there but osapi does not manage it, or the program the operation
// needs is not installed. Reason tells them apart in the answer, so a program
// does not read the message.
func IsMissing(
	err error,
) bool {
	return errors.Is(err, job.ErrNotFound) ||
		errors.Is(err, job.ErrNotManaged) ||
		errors.Is(err, job.ErrNotInstalled)
}

// Reason is the reason of the answer to a failed job: not_managed or
// not_installed when the job says so, and else not_found.
func Reason(
	err error,
) *components.ErrorResponseReason {
	switch {
	case errors.Is(err, job.ErrNotManaged):
		return reason(components.NotManaged)
	case errors.Is(err, job.ErrNotInstalled):
		return reason(components.NotInstalled)
	}

	return reason(components.NotFound)
}

// Invalid is the reason of a request that the handler refuses, a 400.
func Invalid() *components.ErrorResponseReason {
	return reason(components.InvalidRequest)
}

// ServerError is the reason of a failure of the controller, a 500.
func ServerError() *components.ErrorResponseReason {
	return reason(components.ServerError)
}

// Unauthorized and Forbidden are the reasons of a request that the token
// does not allow.
func Unauthorized() *components.ErrorResponseReason {
	return reason(components.Unauthorized)
}

// Forbidden is the reason of a token that lacks the permission.
func Forbidden() *components.ErrorResponseReason {
	return reason(components.Forbidden)
}

func reason(
	r components.ErrorResponseReason,
) *components.ErrorResponseReason {
	return &r
}
