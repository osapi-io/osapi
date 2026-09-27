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

package provider

import "errors"

// The causes a provider can report that the API answers differently. A provider
// returns one of these wrapped in an error that says which thing and where, and
// the agent turns it into a code on the response so the controller does not have
// to read the sentence back.
//
// Anything else is a failure with no further meaning, which the API answers with
// 500. Adding a sentinel is how a new distinction reaches an HTTP status; it is
// never reached by matching on words.
var (
	// ErrUnsupported means the operation is not supported on the current OS
	// family. The agent maps this to StatusSkipped instead of StatusFailed.
	ErrUnsupported = errors.New("operation not supported on this OS family")

	// ErrNotFound means the thing the request named is not on the host: a user,
	// a group, a service, a cron entry, a sysctl key, a package.
	ErrNotFound = errors.New("not found")

	// ErrNotManaged means the thing exists on the host but osapi did not deploy
	// it, so osapi will not change or remove it. Distinct from ErrNotFound
	// because the answer to the operator is different: the file is there, and
	// something else owns it.
	ErrNotManaged = errors.New("not managed by osapi")

	// ErrNotInstalled means a program an operation depends on is absent, so the
	// operation cannot be carried out on this host at all.
	ErrNotInstalled = errors.New("not installed")
)
