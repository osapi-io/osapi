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

package service

import "context"

// Provider reads the services of the host, and manages the unit files and
// the state of the services that osapi deploys.
type Provider interface {
	// Services is every service of the host.
	Services(ctx context.Context) ([]Entry, error)
	// Service is one service of the host, with its main process.
	Service(ctx context.Context, name string) (Entry, error)
	// Deploy writes the unit file of u. With create, it changes nothing when
	// osapi already has the unit; without it, it fails when osapi has none.
	Deploy(ctx context.Context, u Unit, create bool) (bool, error)
	// Remove stops and disables the service and drops its unit file.
	Remove(ctx context.Context, name string) (bool, error)
	// Act runs the systemctl action, such as start, on the service that
	// osapi manages.
	Act(ctx context.Context, action string, name string) (bool, error)
}
