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

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the Provider interface for Darwin (macOS).
// All methods return ErrUnsupported as service management is not available on macOS.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// List returns ErrUnsupported on Darwin.
func (d *Darwin) List(
	_ context.Context,
) ([]Info, error) {
	return nil, provider.ErrUnsupported
}

// Get returns ErrUnsupported on Darwin.
func (d *Darwin) Get(
	_ context.Context,
	_ string,
) (*Info, error) {
	return nil, provider.ErrUnsupported
}

// Create returns ErrUnsupported on Darwin.
func (d *Darwin) Create(
	_ context.Context,
	_ Entry,
) (*CreateResult, error) {
	return nil, provider.ErrUnsupported
}

// Update returns ErrUnsupported on Darwin.
func (d *Darwin) Update(
	_ context.Context,
	_ Entry,
) (*UpdateResult, error) {
	return nil, provider.ErrUnsupported
}

// Delete returns ErrUnsupported on Darwin.
func (d *Darwin) Delete(
	_ context.Context,
	_ string,
) (*DeleteResult, error) {
	return nil, provider.ErrUnsupported
}

// Start returns ErrUnsupported on Darwin.
func (d *Darwin) Start(
	_ context.Context,
	_ string,
) (*ActionResult, error) {
	return nil, provider.ErrUnsupported
}

// Stop returns ErrUnsupported on Darwin.
func (d *Darwin) Stop(
	_ context.Context,
	_ string,
) (*ActionResult, error) {
	return nil, provider.ErrUnsupported
}

// Restart returns ErrUnsupported on Darwin.
func (d *Darwin) Restart(
	_ context.Context,
	_ string,
) (*ActionResult, error) {
	return nil, provider.ErrUnsupported
}

// Enable returns ErrUnsupported on Darwin.
func (d *Darwin) Enable(
	_ context.Context,
	_ string,
) (*ActionResult, error) {
	return nil, provider.ErrUnsupported
}

// Disable returns ErrUnsupported on Darwin.
func (d *Darwin) Disable(
	_ context.Context,
	_ string,
) (*ActionResult, error) {
	return nil, provider.ErrUnsupported
}
