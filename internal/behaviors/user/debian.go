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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// defaultHome is the directory under which useradd makes the home of a user
// that names none, HOME of /etc/default/useradd as each supported Linux
// ships it.
const defaultHome = "/home"

// debian changes the users with useradd, usermod, userdel and chpasswd, and
// reads them as reader does. Each write reads first, so a second run reports
// changed: false.
type debian struct {
	*reader
	exec host.Exec
	fs   avfs.VFS
}

// newDebian is the provider on the Debian family.
func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{reader: &reader{src: h.Sources()}, exec: h.Exec, fs: h.FS}, nil
}

// Create adds the user, with its home directory, and sets its password when
// in has its hash. A user that the host has already does not change.
func (d *debian) Create(
	ctx context.Context,
	in CreateInput,
) (bool, error) {
	if _, err := d.Get(ctx, in.Name); err == nil || !errors.Is(err, behavior.ErrNotFound) {
		return false, err
	}
	if in.PasswordHash != "" {
		if err := checkHash(in.PasswordHash); err != nil {
			return false, err
		}
	}
	args := []string{"--create-home"}
	if in.UID != 0 {
		args = append(args, "-u", strconv.Itoa(in.UID))
	}
	if in.GID != 0 {
		args = append(args, "-g", strconv.Itoa(in.GID))
	}
	if in.Home != "" {
		args = append(args, "-d", in.Home)
	}
	if in.Shell != "" {
		args = append(args, "-s", in.Shell)
	}
	if len(in.Groups) > 0 {
		args = append(args, "-G", strings.Join(in.Groups, ","))
	}
	if in.System {
		args = append(args, "-r")
	}
	// useradd keeps a home directory that is there already, so the clean-up
	// below removes the home only when useradd made it.
	home := in.Home
	if home == "" {
		home = defaultHome + "/" + in.Name
	}
	_, statErr := d.fs.Stat(home)
	undo := []string{"-r", "--", in.Name}
	if statErr == nil {
		undo = []string{"--", in.Name}
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "useradd", append(args, "--", in.Name)); err != nil {
		return false, fmt.Errorf("add user %s: %w", in.Name, err)
	}
	if in.PasswordHash != "" {
		if err := d.chpasswd(ctx, in.Name, in.PasswordHash); err != nil {
			// A user with no password would stay: a second run finds it and
			// reports changed: false, and nothing sets the password. Thus
			// the user goes, and a second run adds it again. Its home goes
			// with it only when useradd made it.
			if _, delErr := d.exec.RunPrivilegedCmd(ctx, "userdel", undo); delErr != nil {
				return true, errors.Join(
					err, fmt.Errorf("remove user %s after it: %w", in.Name, delErr),
				)
			}

			return false, err
		}
	}

	return true, nil
}

// Update changes what in gives and the user does not have already: its
// shell, its home, its groups and whether it is locked.
func (d *debian) Update(
	ctx context.Context,
	in UpdateInput,
) (bool, error) {
	u, err := d.Get(ctx, in.Name)
	if err != nil {
		return false, err
	}
	var args []string
	if in.Shell != "" && in.Shell != u.Shell {
		args = append(args, "-s", in.Shell)
	}
	if in.Home != "" && in.Home != u.Home {
		args = append(args, "-d", in.Home, "-m")
	}
	if len(in.Groups) > 0 && !sameGroups(u.Groups, in.Groups) {
		args = append(args, "-G", strings.Join(in.Groups, ","))
	}
	switch {
	case in.Lock == lockLocked && !u.Locked:
		args = append(args, "-L")
	case in.Lock == lockUnlocked && u.Locked:
		args = append(args, "-U")
	}
	if len(args) == 0 {
		return false, nil
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "usermod", append(args, "--", in.Name)); err != nil {
		return false, fmt.Errorf("change user %s: %w", in.Name, err)
	}

	return true, nil
}

// The values of UpdateInput.Lock.
const (
	lockLocked   = "locked"
	lockUnlocked = "unlocked"
)

// Delete removes the user and its home directory. A user that the host does
// not have does not change it.
func (d *debian) Delete(
	ctx context.Context,
	name string,
) (bool, error) {
	if _, err := d.Get(ctx, name); err != nil {
		return false, ignoreNotFound(err)
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "userdel", []string{"-r", "--", name}); err != nil {
		return false, fmt.Errorf("remove user %s: %w", name, err)
	}

	return true, nil
}

// chpasswd sets the password of name from hash, on standard input, so the
// hash is in no argument list.
func (d *debian) chpasswd(
	ctx context.Context,
	name string,
	hash string,
) error {
	if _, err := d.exec.RunPrivilegedCmdWithStdin(
		ctx, "chpasswd", []string{"-e"}, name+":"+hash+"\n",
	); err != nil {
		return fmt.Errorf("set the password of user %s: %w", name, err)
	}

	return nil
}

// checkHash refuses a hash that is not a crypt hash, or that would break
// the line that chpasswd reads.
func checkHash(
	hash string,
) error {
	if !strings.HasPrefix(hash, "$") || strings.ContainsAny(hash, ":\r\n") {
		return errors.New("the password hash is not a crypt hash")
	}

	return nil
}

// sameGroups reports whether a and b have the same groups, in any order.
func sameGroups(
	a []string,
	b []string,
) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)

	return slices.Equal(slices.Compact(x), slices.Compact(y))
}

// ignoreNotFound is err, or nil when it says the user is not found: a
// delete of a user the host does not have changes nothing.
func ignoreNotFound(
	err error,
) error {
	if errors.Is(err, behavior.ErrNotFound) {
		return nil
	}

	return err
}
