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

package timezone

import (
	"context"

	"github.com/retr0h/osapi/internal/provider"
)

// Linux implements the Provider interface for generic Linux.
// All methods return ErrUnsupported as this is a generic Linux stub.
type Linux struct{}

// NewLinuxProvider factory to create a new Linux instance.
func NewLinuxProvider() *Linux {
	return &Linux{}
}

// Get returns ErrUnsupported on generic Linux.
func (l *Linux) Get(
	_ context.Context,
) (*Info, error) {
	return nil, provider.ErrUnsupported
}

// Update returns ErrUnsupported on generic Linux.
func (l *Linux) Update(
	_ context.Context,
	_ string,
) (*UpdateResult, error) {
	return nil, provider.ErrUnsupported
}
