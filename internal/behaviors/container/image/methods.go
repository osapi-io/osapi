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

package image

import "context"

// List reads the images of the host.
func List(
	ctx context.Context,
	p Provider,
	_ ListInput,
) ([]Entry, error) {
	return p.List(ctx)
}

// Pull pulls an image.
func Pull(
	ctx context.Context,
	p Provider,
	in PullInput,
) (PullResult, error) {
	return p.Pull(ctx, in.Image)
}

// Delete removes an image.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	changed, err := p.Delete(ctx, in.Image, in.Force)

	return Result{Image: in.Image, Changed: changed}, err
}
