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

import "sync"

// Rule is a validate tag a behavior declares for its own input fields: the
// check the server runs, the JSON Schema the spec says it means, and what a
// caller is told when a value fails it.
type Rule struct {
	// Tag is the name used in validate tags, such as sysctl_key.
	Tag string
	// Check reports whether a field's value, as a string, passes.
	Check func(value string) bool
	// Schema is what the tag means in the generated spec.
	Schema func(s *Schema)
	// Message follows the failing value in the error, such as "is not a
	// valid sysctl key".
	Message string
}

var (
	rulesMu sync.Mutex
	rules   []Rule
	onRule  func(r Rule)
)

// Validator declares a validate tag for the behavior's input fields, so a
// behavior that needs its own rule keeps it in its own directory. The server
// checks it before a job is queued, the spec carries schema, and a value that
// fails reads "<value> <message>". The tag is usable by any behavior once
// declared.
func Validator(
	tag string,
	check func(value string) bool,
	schema func(s *Schema),
	message string,
) Option {
	return optionFunc(func(*Behavior) {
		r := Rule{Tag: tag, Check: check, Schema: schema, Message: message}
		ValidatorSchema(tag, schema)

		rulesMu.Lock()
		rules = append(rules, r)
		hook := onRule
		rulesMu.Unlock()

		if hook != nil {
			hook(r)
		}
	})
}

// OnRule has fn called for every rule a behavior declares: those declared
// already, then each one after. The validation package calls it once, to
// register each rule with the validator the server uses.
func OnRule(
	fn func(r Rule),
) {
	rulesMu.Lock()
	onRule = fn
	declared := append([]Rule(nil), rules...)
	rulesMu.Unlock()

	for _, r := range declared {
		fn(r)
	}
}
