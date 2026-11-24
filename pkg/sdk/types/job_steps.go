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

package types

import (
	"strings"
)

// Step is one call of a behavior method that a job made on a host, with the
// calls that it made in turn: dns update calls netplan update, which calls
// managedfile update. A job's response has the steps that its method made.
// The format is the same wherever a step comes from, so a behavior that an
// extension serves returns its steps in it too.
type Step struct {
	// Behavior is the behavior that the step called, such as netplan.
	Behavior string `json:"behavior"`
	// Method is the method of the behavior, such as update.
	Method string `json:"method"`
	// Changed is true when the step changed the host, and false when it did
	// not. It is nil for a method whose output has no changed field.
	Changed *bool `json:"changed,omitempty"`
	// Error is the error of a step that failed.
	Error string `json:"error,omitempty"`
	// DurationMS is how long the step ran, in milliseconds.
	DurationMS int64 `json:"duration_ms"`
	// Steps are the calls that the step made.
	Steps []Step `json:"steps,omitempty"`
}

// FormatSteps is steps on one line: netplan update (changed) → managedfile
// update (changed), with the calls of a step that called more than one in
// parentheses, and a step that failed marked failed.
func FormatSteps(
	steps []Step,
) string {
	parts := make([]string, 0, len(steps))
	for _, s := range steps {
		parts = append(parts, formatStep(s))
	}
	if len(parts) == 1 {
		return parts[0]
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

// formatStep is s and the calls that it made.
func formatStep(
	s Step,
) string {
	line := s.Behavior + " " + s.Method
	switch {
	case s.Error != "":
		line += " (failed)"
	case s.Changed != nil && *s.Changed:
		line += " (changed)"
	}
	if len(s.Steps) > 0 {
		line += " → " + FormatSteps(s.Steps)
	}

	return line
}
