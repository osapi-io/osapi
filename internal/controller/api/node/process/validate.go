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

package process

import (
	"fmt"

	"github.com/osapi-io/osapi/internal/validation"
)

// validateHostname validates a hostname path parameter using the shared
// validator. Returns the error message and false if invalid.
//
// This exists because oapi-codegen does not generate validate tags on
// path parameters in strict-server mode (upstream limitation).
func validateHostname(
	hostname string,
) (string, bool) {
	return validation.Var(hostname, "required,min=1,valid_target")
}

// validatePid validates a pid path parameter against the bound the
// specification declares for it. Same upstream limitation as
// validateHostname.
//
// min=1 is the rule that matters, not required. A pid reaches kill(2), where a
// value below 1 does not mean "that process": 0 is every process in the
// caller's process group, -1 is every process the caller may signal, and
// anything lower is a whole process group. required only rejects 0, so without
// min=1 a negative pid turns one signal into a fleet-wide one.
func validatePid(
	pid int,
) (string, bool) {
	errMsg, ok := validation.Var(pid, "required,min=1")
	if ok {
		return "", true
	}

	// validation.Var has no field name to report and min carries no hint, so
	// the bare message is "Field validation for '' failed on the 'min' tag",
	// which tells a caller nothing. Name the value and the bound, and keep the
	// tag's own message so a test can still assert which rule rejected it.
	return fmt.Sprintf(
		"pid %d must be greater than zero: %s",
		pid, errMsg,
	), false
}
