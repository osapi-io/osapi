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

package user

import "github.com/osapi-io/osapi/internal/behavior/host"

// NewReader is newReader, for the tests of what the reader refuses.
var NewReader = func(h host.Host) (Provider, error) { return newReader(h) }

// IgnoreNotFound is ignoreNotFound, for its test.
var IgnoreNotFound = ignoreNotFound

// SameGroups is sameGroups, for its test.
var SameGroups = sameGroups

// HashCreate is create's controller step, for its test.
var HashCreate = hashCreate

// SetHashFn swaps the hash of the steps, and returns how to put it back.
func SetHashFn(fn func(string) (string, error)) func() {
	old := hashFn
	hashFn = fn

	return func() { hashFn = old }
}

// NewDebian is newDebian, for the tests of what the provider does when a
// command fails.
var NewDebian = newDebian
