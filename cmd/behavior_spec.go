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

package cmd

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/job"
)

var behaviorSpecCmd = newSpecCommand(
	func() guide.Spec { return behavior.CurrentSpec(job.FactKeys()...) },
	"Print the behavior DSL: each option, method, example part and field tag",
	"Print each part of the behavior DSL, what it makes osapi do and when to use\n"+
		"it. The list comes from this binary. Thus it agrees with the build that you\n"+
		"run. Read it before you write a behavior, then check the result with\n"+
		"`osapi behavior validate`.",
	`  osapi behavior spec
  osapi behavior spec --json`,
)

func init() {
	behaviorCmd.AddCommand(behaviorSpecCmd)
}
