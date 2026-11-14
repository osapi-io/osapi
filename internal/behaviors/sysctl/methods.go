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

package sysctl

import (
	"context"
)

// The methods. Each takes the platform's Provider and the method's input
// type, and returns its output type. They translate; the work is in the
// provider.

// List returns every parameter, or those under in.Prefix.
func List(
	ctx context.Context,
	p Provider,
	in ListInput,
) ([]Entry, error) {
	entries, err := p.List(ctx, in.Prefix)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// Get returns one parameter.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Entry, error) {
	e, err := p.Get(ctx, in.Key)
	if err != nil {
		return Entry{}, err
	}

	return *e, nil
}

// Create manages a parameter and applies it.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (MutationResult, error) {
	r, err := p.Create(ctx, Entry{Key: in.Key, Value: in.Value})
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{Key: r.Key, Changed: r.Changed}, nil
}

// Update sets a parameter, managing it if osapi did not, and applies it.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (MutationResult, error) {
	r, err := p.Update(ctx, Entry{Key: in.Key, Value: in.Value})
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{Key: r.Key, Changed: r.Changed}, nil
}

// Delete stops managing a parameter.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (MutationResult, error) {
	r, err := p.Delete(ctx, in.Key)
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{Key: r.Key, Changed: r.Changed}, nil
}
