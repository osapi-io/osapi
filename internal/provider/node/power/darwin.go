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

package power

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/internal/provider"
)

// Darwin implements the Provider interface for Darwin (macOS).
// All methods return ErrUnsupported as power management is not available on macOS.
type Darwin struct{}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider() *Darwin {
	return &Darwin{}
}

// Reboot returns ErrUnsupported on Darwin.
func (d *Darwin) Reboot(
	_ context.Context,
	_ Opts,
) (*Result, error) {
	return nil, fmt.Errorf("power: %w", provider.ErrUnsupported)
}

// Shutdown returns ErrUnsupported on Darwin.
func (d *Darwin) Shutdown(
	_ context.Context,
	_ Opts,
) (*Result, error) {
	return nil, fmt.Errorf("power: %w", provider.ErrUnsupported)
}
