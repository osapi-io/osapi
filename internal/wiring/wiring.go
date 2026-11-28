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

// Package wiring holds the mistakes in how osapi's parts are put together:
// a group registered twice, a section with two owners, deps of the wrong
// type. A registry reports each one here and goes on, and the program
// stops with all of them when it starts, instead of a panic at the first.
package wiring

import (
	"errors"
	"sync"
)

var (
	mu   sync.Mutex
	errs []error
)

// Report records err, a mistake in how the code is put together.
func Report(
	err error,
) {
	mu.Lock()
	defer mu.Unlock()
	errs = append(errs, err)
}

// Take returns each recorded mistake, joined, or nil when there is none,
// and clears them.
func Take() error {
	mu.Lock()
	defer mu.Unlock()
	err := errors.Join(errs...)
	errs = nil

	return err
}
