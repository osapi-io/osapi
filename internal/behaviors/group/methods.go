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

package group

import "context"

// List reads the groups of the host.
func List(
	ctx context.Context,
	p Provider,
	_ ListInput,
) ([]Account, error) {
	return p.List(ctx)
}

// Get reads one group.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Account, error) {
	return p.Get(ctx, in.Name)
}

// Create adds a group.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (Result, error) {
	changed, err := p.Create(ctx, in)

	return Result{Name: in.Name, Changed: changed}, err
}

// Update sets the members of a group.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	changed, err := p.Update(ctx, in.Name, in.Members)

	return Result{Name: in.Name, Changed: changed}, err
}

// Delete removes a group.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Delete(ctx, in.Name)

	return Result{Name: in.Name, Changed: changed}, err
}
