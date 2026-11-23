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

package docker

import "context"

// List reads the containers of the host.
func List(
	ctx context.Context,
	p Provider,
	in ListInput,
) ([]Entry, error) {
	return p.List(ctx, in.State)
}

// Get reads one container.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Entry, error) {
	return p.Get(ctx, in.Name)
}

// Create makes a container.
func Create(
	ctx context.Context,
	p Provider,
	in CreateInput,
) (Result, error) {
	return p.Create(ctx, in)
}

// Delete removes a container.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Delete(ctx, in.Name)

	return Result{Name: in.Name, Changed: changed}, err
}

// Start starts a container.
func Start(
	ctx context.Context,
	p Provider,
	in StartInput,
) (Result, error) {
	changed, err := p.Start(ctx, in.Name)

	return Result{Name: in.Name, Changed: changed}, err
}

// Stop stops a container.
func Stop(
	ctx context.Context,
	p Provider,
	in StopInput,
) (Result, error) {
	changed, err := p.Stop(ctx, in.Name, in.Timeout)

	return Result{Name: in.Name, Changed: changed}, err
}

// Exec runs a command in a container.
func Exec(
	ctx context.Context,
	p Provider,
	in ExecInput,
) (ExecResult, error) {
	return p.Exec(ctx, in)
}
