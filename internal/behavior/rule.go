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

package behavior

import "github.com/osapi-io/osapi/internal/validation/rule"

// Validator declares a validate tag for the behavior's input fields, so a
// behavior that needs its own rule keeps it in its own directory. The server
// checks it before a job is queued, the spec carries schema, and a value that
// fails reads "<value> <message>". The tag is usable by any behavior once
// declared.
func Validator(
	tag string,
	check func(value string) bool,
	schema func(t rule.Target),
	message string,
) Option {
	return optionFunc(func(*Behavior) {
		rule.Declare(rule.Custom{Tag: tag, Check: check, Schema: schema, Message: message})
	})
}
