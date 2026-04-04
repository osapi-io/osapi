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

package apt

import (
	"context"

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the Provider interface for Darwin (macOS).
// All methods return ErrUnsupported as apt is not available on macOS.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// List returns ErrUnsupported on Darwin.
func (d *Darwin) List(
	_ context.Context,
) ([]Package, error) {
	return nil, provider.ErrUnsupported
}

// Get returns ErrUnsupported on Darwin.
func (d *Darwin) Get(
	_ context.Context,
	_ string,
) (*Package, error) {
	return nil, provider.ErrUnsupported
}

// Install returns ErrUnsupported on Darwin.
func (d *Darwin) Install(
	_ context.Context,
	_ string,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// Remove returns ErrUnsupported on Darwin.
func (d *Darwin) Remove(
	_ context.Context,
	_ string,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// Update returns ErrUnsupported on Darwin.
func (d *Darwin) Update(
	_ context.Context,
) (*Result, error) {
	return nil, provider.ErrUnsupported
}

// ListUpdates returns ErrUnsupported on Darwin.
func (d *Darwin) ListUpdates(
	_ context.Context,
) ([]Update, error) {
	return nil, provider.ErrUnsupported
}
