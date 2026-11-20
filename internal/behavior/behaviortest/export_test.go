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

package behaviortest

import (
	"time"

	"github.com/osapi-io/osapi/internal/behavior"
)

// ExportReport calls the mock reporter of a fake host, as the mock
// controller would for a call that the host has no answer for.
func ExportReport(
	fatal bool,
) {
	h := &fakeHost{}
	if fatal {
		h.Fatalf("Exec.%s", "Unknown")
	}
	h.Errorf("Exec.%s", "Unknown")
}

// SetSourceDir sets the directory Run reads Go files from, and returns a
// function that puts it back.
func SetSourceDir(
	dir string,
) func() {
	previous := sourceDir
	sourceDir = dir

	return func() { sourceDir = previous }
}

// ExportClock is the time the fake host's clock gives.
func ExportClock() time.Time {
	return newHost(&behavior.Behavior{Name: "clock"}, "linux").Time()
}
