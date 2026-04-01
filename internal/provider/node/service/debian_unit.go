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

import (
	"context"
	"fmt"
)

// Create deploys a new unit file via the file provider.
// Stub: will be implemented in a future task.
func (d *Debian) Create(
	_ context.Context,
	_ Entry,
) (*CreateResult, error) {
	return nil, fmt.Errorf("service: create: not yet implemented")
}

// Update redeploys an existing unit file via the file provider.
// Stub: will be implemented in a future task.
func (d *Debian) Update(
	_ context.Context,
	_ Entry,
) (*UpdateResult, error) {
	return nil, fmt.Errorf("service: update: not yet implemented")
}

// Delete undeploys a unit file via the file provider.
// Stub: will be implemented in a future task.
func (d *Debian) Delete(
	_ context.Context,
	_ string,
) (*DeleteResult, error) {
	return nil, fmt.Errorf("service: delete: not yet implemented")
}
