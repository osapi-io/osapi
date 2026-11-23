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

package sshkey

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

// The modes that sshd's StrictModes accepts.
const (
	dirMode  = 0o700
	fileMode = 0o600
)

// debian is the provider on the Debian family: it reads as the reader
// does, and writes ~/.ssh/authorized_keys of the user.
type debian struct {
	*reader
	fs   avfs.VFS
	exec host.Exec
}

// newDebian is the provider on the Debian family.
func newDebian(
	h host.Host,
) (Provider, error) {
	r, _ := newReader(h)

	return &debian{reader: r.(*reader), fs: h.FS, exec: h.Exec}, nil
}

// file is the authorized_keys of the user, read through the user behavior,
// with its content; none when the file does not exist.
func (d *debian) file(
	ctx context.Context,
	name string,
) (user.Account, string, string, error) {
	u, err := d.user(ctx, name)
	if err != nil {
		return u, "", "", err
	}
	p := path.Join(u.Home, ".ssh", "authorized_keys")
	b, err := d.fs.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return u, p, "", nil
	}

	return u, p, string(b), err
}

// Add appends the line, after the user's other keys. The directory and the
// file belong to the user, so that sshd reads them.
func (d *debian) Add(
	ctx context.Context,
	name string,
	line string,
) (string, bool, error) {
	fp, err := fingerprint(line)
	if err != nil {
		return "", false, err
	}
	u, p, content, err := d.file(ctx, name)
	if err != nil {
		return fp, false, err
	}
	for _, l := range strings.Split(content, "\n") {
		if have, err := fingerprint(l); err == nil && have == fp {
			return fp, false, nil
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	dir := path.Dir(p)
	if err := d.fs.MkdirAll(dir, dirMode); err != nil {
		return fp, false, err
	}
	if err := atomicfile.Write(
		ctx,
		d.fs,
		p,
		[]byte(content+strings.TrimSpace(line)+"\n"),
		fileMode,
	); err != nil {
		return fp, false, err
	}
	owner := strconv.Itoa(u.UID) + ":" + strconv.Itoa(u.GID)
	if _, err := d.exec.RunPrivilegedCmd(ctx, "chown", []string{"-R", owner, dir}); err != nil {
		return fp, false, err
	}

	return fp, true, nil
}

// Remove drops each line of the key with the fingerprint, and keeps the
// others as they are.
func (d *debian) Remove(
	ctx context.Context,
	name string,
	fp string,
) (bool, error) {
	_, p, content, err := d.file(ctx, name)
	if errors.Is(err, behavior.ErrNotFound) {
		// A user that the host does not have has no key to remove.
		return false, nil
	}
	if err != nil || content == "" {
		return false, err
	}
	var keep []string
	for _, l := range strings.Split(content, "\n") {
		if have, err := fingerprint(l); err == nil && have == fp {
			continue
		}
		keep = append(keep, l)
	}
	out := strings.Join(keep, "\n")
	if out == content {
		return false, nil
	}

	return true, atomicfile.Write(ctx, d.fs, p, []byte(out), fileMode)
}
