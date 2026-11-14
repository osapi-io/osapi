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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// Provider is what each platform implements: debian.go on Debian and its
// family, darwin.go and linux.go elsewhere. The hand-written sysctl domain
// uses the same implementations until its cutover.
type Provider interface {
	// List returns every parameter with its current value and whether osapi
	// manages it, narrowed to prefix and the parameters under it if set.
	List(ctx context.Context, prefix string) ([]Entry, error)
	// Get returns one parameter's current value and whether osapi manages it.
	Get(ctx context.Context, key string) (*Entry, error)
	// Create deploys a new sysctl conf file and applies it. Idempotent: returns Changed: false if already managed.
	Create(ctx context.Context, entry Entry) (*CreateResult, error)
	// Update sets a parameter and manages it, as Create does.
	Update(ctx context.Context, entry Entry) (*UpdateResult, error)
	// Delete removes a managed sysctl conf file and reloads defaults.
	Delete(ctx context.Context, key string) (*DeleteResult, error)
}

// newDebian is the provider on Debian and its family: it writes one file per
// parameter under behaviors.sysctl.dir through the managedfile behavior it
// requires, so it declines a host where managedfile has no provider.
func newDebian(
	h host.Host,
) (Provider, error) {
	files := host.Require[managedfile.Provider](h, managedfile.Behavior)
	if files == nil {
		return nil, behavior.ErrUnsupported
	}

	return NewDebianProvider(h.Logger, h.FS, h.FileState, h.Exec, h.Hostname).
		WithDir(host.Settings[Settings](h).Dir).
		WithFiles(files), nil
}
