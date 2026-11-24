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

package user

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

// reader reads the users through gohai's users collector, which reads
// /etc/passwd, /etc/group and /etc/shadow on Linux, and the directory
// service on macOS. It changes nothing.
type reader struct {
	src source.Sources
}

// newReader is the provider on each platform but the Debian family.
func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// List reads every user of the host, by name.
func (r *reader) List(
	ctx context.Context,
) ([]Account, error) {
	out, err := users.NewWith(r.src).Collect(ctx, nil)
	list := []Account{}
	if info, ok := out.(*users.Info); ok {
		for name, u := range info.Passwd {
			list = append(list, Account{
				Name:   name,
				UID:    u.UID,
				GID:    u.GID,
				Home:   u.Dir,
				Shell:  u.Shell,
				Groups: memberOf(info.Group, name),
				Locked: u.Locked != nil && *u.Locked,
			})
		}
	}
	slices.SortFunc(list, func(a, b Account) int { return strings.Compare(a.Name, b.Name) })

	return list, err
}

// memberOf is the groups that list name as a member, by name.
func memberOf(
	groups map[string]users.GroupEntry,
	name string,
) []string {
	var out []string
	for g, entry := range groups {
		if slices.Contains(entry.Members, name) {
			out = append(out, g)
		}
	}
	slices.Sort(out)

	return out
}

// Get reads one user.
func (r *reader) Get(
	ctx context.Context,
	name string,
) (Account, error) {
	list, err := r.List(ctx)
	for _, u := range list {
		if u.Name == name {
			return u, err
		}
	}
	if err == nil {
		err = fmt.Errorf("user %q: %w", name, behavior.ErrNotFound)
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
	UpdateInput,
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
