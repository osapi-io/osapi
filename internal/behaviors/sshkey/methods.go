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

import "context"

// List is the keys of a user.
func List(
	ctx context.Context,
	p Provider,
	in ListInput,
) ([]Key, error) {
	return p.Keys(ctx, in.Name)
}

// Create adds a key to a user.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (Result, error) {
	fp, changed, err := p.Add(ctx, in.Name, in.Key)

	return Result{Fingerprint: fp, Changed: changed}, err
}

// Delete removes a key of a user.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Remove(ctx, in.Name, in.Fingerprint)

	return Result{Fingerprint: in.Fingerprint, Changed: changed}, err
}
