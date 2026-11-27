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
	"fmt"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// checkError inspects the HTTP status code and returns the appropriate
// typed error. For success codes (200, 201, 202, 204) it returns nil.
// body is the error body of the answer, or nil when it has none.
func checkError(
	statusCode int,
	body *ErrorResponse,
) error {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return nil
	}

	e := APIError{
		StatusCode: statusCode,
		Message:    errorMessage(statusCode, body),
		Reason:     errorReason(body),
	}

	switch {
	case statusCode == 400:
		return &ValidationError{e}
	case statusCode == 401 || statusCode == 403:
		return &AuthError{e}
	case statusCode == 404:
		return &NotFoundError{e}
	case statusCode == 409:
		return &ConflictError{e}
	case statusCode >= 500:
		return &ServerError{e}
	default:
		return &UnexpectedStatusError{e}
	}
}

// errorMessage is the message of body, or a message of the status code
// when body has none.
func errorMessage(
	statusCode int,
	body *ErrorResponse,
) string {
	if body != nil && body.Error != "" {
		return body.Error
	}

	return fmt.Sprintf("unexpected status %d", statusCode)
}

// errorReason is the reason of body, or "".
func errorReason(
	body *ErrorResponse,
) types.Reason {
	if body == nil {
		return ""
	}

	return types.Reason(body.Reason)
}
