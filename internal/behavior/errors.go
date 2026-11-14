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

import "errors"

// The causes a provider can report that the API answers differently. A
// provider returns one of these wrapped in an error that says which thing and
// where; the agent turns it into a code on the response, and the controller
// into a status, so nothing has to read the sentence back.
//
// They are here, in the package every behavior imports, because every
// behavior's provider returns them. internal/provider gives the hand-written
// providers the same values under its own names until they migrate.
var (
	// ErrUnsupported means the operation is not supported on this host's
	// platform. The method reports skipped rather than failed.
	ErrUnsupported = errors.New("operation not supported on this OS family")

	// ErrNotFound means the thing the request named is not on the host: a
	// user, a group, a service, a cron entry, a sysctl key, a package.
	ErrNotFound = errors.New("not found")

	// ErrNotManaged means the thing exists on the host but osapi did not
	// deploy it, so osapi will not change or remove it.
	ErrNotManaged = errors.New("not managed by osapi")

	// ErrNotInstalled means a program the operation depends on is absent.
	ErrNotInstalled = errors.New("not installed")
)
