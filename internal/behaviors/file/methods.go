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

package file

import "context"

// Deploy writes an object of the object store to the host.
func Deploy(
	ctx context.Context,
	p Provider,
	in DeployInput,
) (DeployResult, error) {
	return p.Deploy(ctx, in)
}

// Undeploy removes a file that osapi deployed.
func Undeploy(
	ctx context.Context,
	p Provider,
	in UndeployInput,
) (UndeployResult, error) {
	changed, err := p.Undeploy(ctx, in.Path)
	if err != nil {
		return UndeployResult{}, err
	}

	return UndeployResult{Path: in.Path, Changed: changed}, nil
}

// GetStatus is the file at the path against what osapi deployed there.
func GetStatus(
	ctx context.Context,
	p Provider,
	in StatusInput,
) (Status, error) {
	return p.Status(ctx, in.Path)
}
