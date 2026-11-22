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

package ntp

import (
	"context"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// Provider is what each platform implements.
type Provider interface {
	// Get returns chrony's sync status and its sources.
	Get(ctx context.Context) (Status, error)
	// Create writes the servers that osapi manages, and reports whether the
	// host changed. A file with the same servers is left alone.
	Create(ctx context.Context, servers []string) (bool, error)
	// Update replaces the servers that osapi manages. It fails with
	// behavior.ErrNotFound when osapi manages none.
	Update(ctx context.Context, servers []string) (bool, error)
	// Delete removes the servers that osapi manages, and reports whether a
	// file was removed.
	Delete(ctx context.Context) (bool, error)
}

// newDebian is the provider on Debian and its family. It writes through the
// managedfile building block.
func newDebian(
	h host.Host,
) (Provider, error) {
	// A container's clock is its host's, so a container has nothing to set.
	if h.Container {
		return nil, behavior.ErrUnsupported
	}
	files := host.Require[managedfile.Provider](h, managedfile.Behavior)
	if files == nil {
		return nil, behavior.ErrUnsupported
	}

	return &debian{logger: h.Logger, exec: h.Exec, files: files}, nil
}
