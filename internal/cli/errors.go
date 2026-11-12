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

package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// Error codes in the --json error document. A program matches on these; the
// message beside them can change between releases.
const (
	// CodeUsage is a bad invocation: unknown command or flag, or a missing
	// required flag. The process exits 2.
	CodeUsage = "usage_error"
	// CodeFatal is a failure before the command could run, such as a config
	// file that will not load.
	CodeFatal = "fatal"
	// CodeError is any other failure the CLI cannot name more precisely.
	CodeError = "error"
	// CodeInvalidRequest and the codes below are the API's answer, by status.
	CodeInvalidRequest = "invalid_request"
	CodeUnauthorized   = "unauthorized"
	CodeForbidden      = "forbidden"
	CodeNotFound       = "not_found"
	CodeConflict       = "conflict"
	CodeServerError    = "server_error"
	CodeAPIError       = "api_error"
)

// Exit codes shared by every command.
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

// ErrorDocument is the one JSON document a failing command writes to stderr
// under --json.
type ErrorDocument struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Status  int               `json:"status,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// APICode names an API status code for the error document.
func APICode(
	status int,
) string {
	switch {
	case status == 400:
		return CodeInvalidRequest
	case status == 401:
		return CodeUnauthorized
	case status == 403:
		return CodeForbidden
	case status == 404:
		return CodeNotFound
	case status == 409:
		return CodeConflict
	case status >= 500:
		return CodeServerError
	default:
		return CodeAPIError
	}
}

// WriteErrorJSON writes doc to w as a single line of JSON.
func WriteErrorJSON(
	w io.Writer,
	doc ErrorDocument,
) {
	// The document is strings, an int and a string map, so it always marshals.
	out, _ := json.Marshal(doc)
	_, _ = fmt.Fprintln(w, string(out))
}

// ExitCode is the process exit code for err: 2 for a usage error, 1 for any
// other failure.
func ExitCode(
	err error,
) int {
	if IsUsageError(err) {
		return ExitUsage
	}

	return ExitError
}
