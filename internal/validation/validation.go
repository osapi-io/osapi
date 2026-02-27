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

// Package validation provides a shared validator instance.
package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var instance = validator.New()

// customHints maps validator tags to a hint appended to the default error.
var customHints = map[string]func(fe validator.FieldError) string{
	"valid_target": func(fe validator.FieldError) string {
		return fmt.Sprintf("target agent %q not found", fe.Value())
	},
}

// Struct validates a struct and returns the error message and false if invalid.
func Struct(
	v any,
) (string, bool) {
	if err := instance.Struct(v); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return formatErrors(validationErrors), false
	}

	return "", true
}

// formatErrors builds the error string, appending a custom hint for known
// tags while keeping the standard validator prefix.
func formatErrors(
	errs validator.ValidationErrors,
) string {
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		msg := fe.Error()
		if fn, ok := customHints[fe.Tag()]; ok {
			msg = fmt.Sprintf("%s: %s", msg, fn(fe))
		}
		msgs = append(msgs, msg)
	}

	return strings.Join(msgs, "; ")
}

// Instance returns the shared validator for registering custom validators.
func Instance() *validator.Validate {
	return instance
}
