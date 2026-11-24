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

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Reasons in the --json error document: why a command failed, as a word that
// a program matches on; the message beside them can change between releases.
// The ReasonInvalidRequest group is the API's, which the API names itself.
const (
	// ReasonUsage is a bad invocation: unknown command or flag, or a missing
	// required flag. The process exits 2.
	ReasonUsage = "usage_error"
	// ReasonFatal is a failure before the command could run, such as a config
	// file that will not load.
	ReasonFatal = "fatal"
	// ReasonError is any other failure the CLI cannot name more precisely.
	ReasonError = "error"
	// ReasonAPIError is an API answer with no reason of types.Reasons. The
	// API's own reasons are the words of types.Reason.
	ReasonAPIError = "api_error"
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
	Reason  string            `json:"reason"`
	Status  int               `json:"status,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// APIReason is the reason of an API answer that names none, by its status.
func APIReason(
	status int,
) string {
	switch {
	case status == 400:
		return string(types.ReasonInvalidRequest)
	case status == 401:
		return string(types.ReasonUnauthorized)
	case status == 403:
		return string(types.ReasonForbidden)
	case status == 404:
		return string(types.ReasonNotFound)
	case status == 409:
		return string(types.ReasonConflict)
	case status >= 500:
		return string(types.ReasonServerError)
	default:
		return ReasonAPIError
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
