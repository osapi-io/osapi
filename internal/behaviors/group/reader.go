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

package group

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/users"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the groups through gohai's users collector, which reads
// /etc/group on Linux and on macOS. It changes nothing.
type reader struct {
	src source.Sources
}

// newReader is the provider on each platform but the Debian family.
func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// List reads every group of the host, by name.
func (r *reader) List(
	ctx context.Context,
) ([]Account, error) {
	out, err := users.NewWith(r.src).Collect(ctx, nil)
	list := []Account{}
	if info, ok := out.(*users.Info); ok {
		for name, g := range info.Group {
			list = append(list, Account{Name: name, GID: g.GID, Members: g.Members})
		}
	}
	slices.SortFunc(list, func(a, b Account) int { return strings.Compare(a.Name, b.Name) })

	return list, err
}

// Get reads one group.
func (r *reader) Get(
	ctx context.Context,
	name string,
) (Account, error) {
	list, err := r.List(ctx)
	for _, g := range list {
		if g.Name == name {
			return g, err
		}
	}
	if err == nil {
		err = fmt.Errorf("group %q: %w", name, behavior.ErrNotFound)
	}

	return Account{}, err
}

// Create is not supported off the Debian family.
func (r *reader) Create(
	context.Context,
	CreateInput,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Update is not supported off the Debian family.
func (r *reader) Update(
	context.Context,
	string,
	[]string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Delete is not supported off the Debian family.
func (r *reader) Delete(
	context.Context,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}
