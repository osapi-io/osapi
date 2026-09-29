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

package agent

import "github.com/osapi-io/osapi/internal/validation"

// validateAgentHostname validates a `{hostname}` path parameter that names one
// enrolled agent. It accepts a literal hostname only: an agent endpoint addresses a
// specific machine, so the reserved routing values `_any` and `_all` and label
// selectors are not meaningful here and are not admitted.
//
// This is deliberately a different check from validation.TargetHostname, which the
// node-targeted domains use. The two were once identically named functions with
// identical comments and different rules, which read as duplication and was not.
//
// It exists at all because oapi-codegen generates no validate tags for path
// parameters in strict-server mode, an upstream limitation.
func validateAgentHostname(
	hostname string,
) (string, bool) {
	return validation.Var(hostname, "required,min=1,max=255")
}
