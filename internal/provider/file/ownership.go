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

package file

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// ownershipOf reports the numeric owner and group of the file at path.
//
// It reads the real filesystem rather than the provider's avfs handle, because
// the uid and gid live in the stat structure that abstraction does not carry.
// Overridable in tests via export_test.go.
var ownershipOf = func(
	path string,
) (uid int, gid int, err error) {
	info, err := statFn(path)
	if err != nil {
		return 0, 0, fmt.Errorf("stat %q: %w", path, err)
	}

	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, errOwnershipUnreadable
	}

	return int(st.Uid), int(st.Gid), nil
}

// statFn stats a path on the real filesystem. Overridable in tests, which is
// how the platform that cannot report ownership is described on one that can.
var statFn = os.Stat

// lookupUID resolves a user name to its uid. Overridable in tests.
var lookupUID = func(
	name string,
) (int, error) {
	u, err := user.Lookup(name)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(u.Uid)
}

// lookupGID resolves a group name to its gid. Overridable in tests.
var lookupGID = func(
	name string,
) (int, error) {
	g, err := user.LookupGroup(name)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(g.Gid)
}

// errOwnershipUnreadable means this platform does not expose a file's uid and
// gid. The caller applies the requested ownership rather than skipping it, since
// the alternative is leaving a declared owner unenforced.
var errOwnershipUnreadable = fmt.Errorf("file ownership is not readable on this platform")

// unsetID marks an owner or group the request did not name, which is therefore
// not compared and not changed.
const unsetID = -1

// resolveID turns a name from a deploy request into a numeric id. A name that
// the host does not know is an error rather than a silent no-op, because the
// deploy asked for an owner that does not exist. A numeric id is accepted as
// itself, which is what a container image without a passwd entry needs.
func resolveID(
	name string,
	lookup func(string) (int, error),
) (int, error) {
	if name == "" {
		return unsetID, nil
	}

	if id, err := lookup(name); err == nil {
		return id, nil
	}

	if id, err := strconv.Atoi(name); err == nil {
		return id, nil
	}

	return unsetID, fmt.Errorf("unknown owner or group %q", name)
}

// ownershipMatches reports whether the file already has the requested owner and
// group. An owner or group the request did not name is not compared.
//
// When the platform cannot report a file's ownership, it answers false: the
// requested ownership is then applied rather than assumed.
func ownershipMatches(
	path string,
	wantUID int,
	wantGID int,
) (bool, error) {
	uid, gid, err := ownershipOf(path)
	if err != nil {
		if errors.Is(err, errOwnershipUnreadable) {
			return false, nil
		}

		return false, err
	}

	if wantUID != unsetID && uid != wantUID {
		return false, nil
	}

	if wantGID != unsetID && gid != wantGID {
		return false, nil
	}

	return true, nil
}
