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

package osapi

import "fmt"

// APIError is the base error type for OSAPI API errors.
type APIError struct {
	StatusCode int
	Message    string
}

// Error returns a formatted error string.
func (e *APIError) Error() string {
	return fmt.Sprintf(
		"api error (status %d): %s",
		e.StatusCode,
		e.Message,
	)
}

// AuthError represents authentication/authorization errors (401, 403).
type AuthError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *AuthError) Unwrap() error {
	return &e.APIError
}

// NotFoundError represents resource not found errors (404).
type NotFoundError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *NotFoundError) Unwrap() error {
	return &e.APIError
}

// ValidationError represents validation errors (400).
type ValidationError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *ValidationError) Unwrap() error {
	return &e.APIError
}

// ServerError represents internal server errors (500).
type ServerError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *ServerError) Unwrap() error {
	return &e.APIError
}

// ConflictError represents conflict errors (409).
type ConflictError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *ConflictError) Unwrap() error {
	return &e.APIError
}

// UnexpectedStatusError represents unexpected HTTP status codes.
type UnexpectedStatusError struct {
	APIError
}

// Unwrap returns the underlying APIError.
func (e *UnexpectedStatusError) Unwrap() error {
	return &e.APIError
}
