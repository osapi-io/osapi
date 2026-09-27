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

package job

import (
	"errors"
	"fmt"
)

// ErrorCode names the cause of a failed job in a form that survives the trip
// from the agent to the controller. The message beside it is for a person; the
// code is what the API switches on.
type ErrorCode string

const (
	// ErrorCodeNotFound means the thing the request named is not on the host.
	ErrorCodeNotFound ErrorCode = "not_found"
	// ErrorCodeNotManaged means it exists but osapi did not deploy it.
	ErrorCodeNotManaged ErrorCode = "not_managed"
	// ErrorCodeNotInstalled means a program the operation depends on is absent.
	ErrorCodeNotInstalled ErrorCode = "not_installed"
	// ErrorCodeUnsupported means the operation does not apply to this OS family.
	ErrorCodeUnsupported ErrorCode = "unsupported"
	// ErrorCodeTimeout means the agent never answered.
	ErrorCodeTimeout ErrorCode = "timeout"
)

// The causes a handler can distinguish, as errors, because that is the shape Go
// code already knows how to test. A failed job's error wraps whichever one its
// code names, so a handler asks errors.Is rather than reading the message.
var (
	// ErrNotFound corresponds to ErrorCodeNotFound.
	ErrNotFound = errors.New("not found")
	// ErrNotManaged corresponds to ErrorCodeNotManaged.
	ErrNotManaged = errors.New("not managed by osapi")
	// ErrNotInstalled corresponds to ErrorCodeNotInstalled.
	ErrNotInstalled = errors.New("not installed")
	// ErrUnsupported corresponds to ErrorCodeUnsupported.
	ErrUnsupported = errors.New("operation not supported on this OS family")

	// ErrTimeout corresponds to ErrorCodeTimeout: the agent never answered,
	// which is not the same as an operation that ran and failed.
	ErrTimeout = errors.New("agent did not respond")

	// ErrNoOperationData means a stored job carries nothing to re-dispatch, so
	// it cannot be retried. Not a missing job: the record is there and unusable,
	// which the API answers with 400 rather than 404.
	ErrNoOperationData = errors.New("job has no operation data")
)

// sentinelFor maps a code to the error a handler tests against. A code this
// version does not know maps to nothing, so it reads as an ordinary failure
// rather than as the wrong cause: an agent newer than its controller is a
// rollout, not a reason to answer 404.
func sentinelFor(
	code ErrorCode,
) error {
	switch code {
	case ErrorCodeNotFound:
		return ErrNotFound
	case ErrorCodeNotManaged:
		return ErrNotManaged
	case ErrorCodeNotInstalled:
		return ErrNotInstalled
	case ErrorCodeUnsupported:
		return ErrUnsupported
	case ErrorCodeTimeout:
		return ErrTimeout
	}

	return nil
}

// ResponseError is the error a failed job becomes on the controller. It carries
// the agent's message for the reader and the code for the handler.
type ResponseError struct {
	// Code names the cause, empty when the agent reported none.
	Code ErrorCode
	// Message is what the provider said, unchanged.
	Message string
	// Hostname is the agent that answered.
	Hostname string
}

// NewResponseError builds the error for a failed response.
func NewResponseError(
	resp *Response,
) *ResponseError {
	if resp == nil {
		return nil
	}

	return &ResponseError{
		Code:     resp.ErrorCode,
		Message:  resp.Error,
		Hostname: resp.Hostname,
	}
}

// Error reads the way the string it replaces did, so an operator sees no change
// in what a failed job says.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("job failed: %s", e.Message)
}

// Is reports whether this failure has the cause the caller is asking about,
// which is how a handler chooses a status code.
func (e *ResponseError) Is(
	target error,
) bool {
	if e == nil {
		return false
	}

	sentinel := sentinelFor(e.Code)

	return sentinel != nil && errors.Is(sentinel, target)
}
