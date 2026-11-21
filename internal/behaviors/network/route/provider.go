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

package route

import "context"

// Provider reads the routes of the host and writes osapi's routes of an
// interface.
type Provider interface {
	// Routes is every route of the host's main table.
	Routes(ctx context.Context) ([]Entry, error)
	// InterfaceRoutes is the routes of one interface of the host.
	InterfaceRoutes(ctx context.Context, name string) ([]Entry, error)
	// Create writes the interface's routes. It changes nothing when osapi
	// already has routes of the interface.
	Create(ctx context.Context, in Config) (bool, error)
	// Update writes the interface's routes over the ones osapi has. It fails
	// when osapi has none.
	Update(ctx context.Context, in Config) (bool, error)
	// Remove drops osapi's routes of the interface.
	Remove(ctx context.Context, name string) (bool, error)
}
