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

package dns

import "context"

// Get is the resolvers of an interface.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Entry, error) {
	return p.Resolvers(ctx, in.InterfaceName)
}

// Update sets the resolvers of an interface.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	changed, err := p.Set(ctx, in)

	return Result{InterfaceName: in.InterfaceName, Changed: changed}, err
}

// Delete drops the resolvers osapi set on an interface.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Remove(ctx, in.InterfaceName)

	return Result{InterfaceName: in.InterfaceName, Changed: changed}, err
}
