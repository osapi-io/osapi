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

package node

import "github.com/retr0h/osapi/internal/validation"

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

// validateInterfaceName validates an interfaceName path parameter.
// Same upstream limitation as validateHostname.
func validateInterfaceName(
	name string,
) (string, bool) {
	return validation.Var(name, "required,alphanum")
}
