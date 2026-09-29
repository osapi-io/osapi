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

package validation

// TargetHostname validates a `{hostname}` path parameter that addresses managed
// machines. It accepts a literal hostname, the reserved routing values `_any` and
// `_all`, or a `key:value` label selector — whatever `valid_target` admits. Returns
// the error message and false when the value is not a usable target.
//
// This exists because oapi-codegen generates no validate tags for path parameters
// in strict-server mode, an upstream limitation, so a path parameter is validated
// by hand in the handler.
//
// It lives here rather than in a handler package because two packages need the same
// check: node-targeted domains under internal/controller/api/node and the power
// domain below it. It is deliberately NOT what the agent domain uses — an agent
// endpoint addresses one enrolled agent by its literal hostname, where a routing
// value would be meaningless, so that domain validates differently and says so.
func TargetHostname(
	hostname string,
) (string, bool) {
	return Var(hostname, "required,min=1,valid_target")
}
