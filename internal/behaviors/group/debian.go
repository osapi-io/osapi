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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// debian changes the groups with groupadd, gpasswd and groupdel, and reads
// them as reader does. Each write reads first, so a second run reports
// changed: false.
type debian struct {
	*reader
	exec host.Exec
}

// newDebian is the provider on the Debian family.
func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{reader: &reader{src: h.Sources()}, exec: h.Exec}, nil
}

// Create adds the group. A group that the host has already does not change.
func (d *debian) Create(
	ctx context.Context,
	in CreateInput,
) (bool, error) {
	if _, err := d.Get(ctx, in.Name); err == nil || !errors.Is(err, behavior.ErrNotFound) {
		return false, err
	}
	var args []string
	if in.GID != 0 {
		args = append(args, "-g", strconv.Itoa(in.GID))
	}
	if in.System {
		args = append(args, "-r")
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "groupadd", append(args, "--", in.Name)); err != nil {
		return false, fmt.Errorf("add group %s: %w", in.Name, err)
	}

	return true, nil
}

// Update sets the members of the group. Members that it has already, in any
// order, do not change it.
func (d *debian) Update(
	ctx context.Context,
	name string,
	members []string,
) (bool, error) {
	g, err := d.Get(ctx, name)
	if err != nil {
		return false, err
	}
	if sameMembers(g.Members, members) {
		return false, nil
	}
	if _, err := d.exec.RunPrivilegedCmd(
		ctx,
		"gpasswd",
		[]string{"-M", strings.Join(members, ","), "--", name},
	); err != nil {
		return false, fmt.Errorf("set the members of group %s: %w", name, err)
	}

	return true, nil
}

// Delete removes the group. A group that the host does not have does not
// change it.
func (d *debian) Delete(
	ctx context.Context,
	name string,
) (bool, error) {
	if _, err := d.Get(ctx, name); err != nil {
		return false, ignoreNotFound(err)
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "groupdel", []string{"--", name}); err != nil {
		return false, fmt.Errorf("remove group %s: %w", name, err)
	}

	return true, nil
}

// sameMembers reports whether a and b have the same members, in any order.
func sameMembers(
	a []string,
	b []string,
) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)

	return slices.Equal(slices.Compact(x), slices.Compact(y))
}

// ignoreNotFound is err, or nil when it says the group is not found: a
// delete of a group the host does not have changes nothing.
func ignoreNotFound(
	err error,
) error {
	if errors.Is(err, behavior.ErrNotFound) {
		return nil
	}

	return err
}
