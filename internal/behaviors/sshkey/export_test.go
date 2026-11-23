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

package sshkey

import (
	"context"

	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// NewReader is newReader, for the tests of what the reader refuses.
var NewReader = func(h host.Host) (Provider, error) { return newReader(h) }

// Fingerprint is fingerprint, for its test.
var Fingerprint = fingerprint

// SetCollectSSH replaces gohai's ssh collector, and returns a func that puts
// it back.
func SetCollectSSH(
	fn func(context.Context, source.Sources) (any, error),
) func() {
	old := collectSSH
	collectSSH = fn

	return func() { collectSSH = old }
}

// KeysOf is keysOf, for its test.
var KeysOf = keysOf
