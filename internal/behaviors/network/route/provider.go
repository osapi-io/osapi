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

// Provider reads the routes of an interface and writes osapi's routes of
// it, one at a time.
type Provider interface {
	// Routes is the routes of one interface of the host's main table.
	Routes(ctx context.Context, name string) ([]Entry, error)
	// Set writes the route r of its interface. With create, a route of that
	// destination that osapi has is unchanged when it is the same, and
	// refused when it is not; without it, a destination osapi does not have
	// is not found.
	Set(ctx context.Context, r Route, create bool) (bool, error)
	// Remove drops osapi's route to destination of the interface.
	Remove(ctx context.Context, name string, destination string) (bool, error)
}
