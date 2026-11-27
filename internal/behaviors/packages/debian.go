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

package packages

import (
	"context"
	"errors"
	"fmt"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/validation"
)

// debian reads through gohai, and installs and removes with apt-get.
type debian struct {
	*reader
	exec host.Exec
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{reader: &reader{src: h.Sources()}, exec: h.Exec}, nil
}

// Install runs apt-get install, unless the package is installed, at
// version when it is not empty.
func (d *debian) Install(
	ctx context.Context,
	name string,
	version string,
) (bool, error) {
	if version != "" &&
		(len(version) > 255 || !validation.PackageVersionPattern.MatchString(version)) {
		return false, fmt.Errorf("%q is not a valid package version", version)
	}
	if installed, err := d.installed(ctx, name, version); err != nil || installed {
		return false, err
	}
	target := name
	if version != "" {
		target = name + "=" + version
	}
	if _, err := d.exec.RunPrivilegedCmd(
		ctx,
		"apt-get",
		[]string{"install", "-y", "-o", lockTimeout, "--", target},
	); err != nil {
		return false, fmt.Errorf("install %s: %w", target, err)
	}

	return true, nil
}

// lockTimeout makes apt-get wait for the lock of dpkg, which another apt run
// holds, such as unattended-upgrades or a refresh of the index from another
// job, before it fails: without it, apt-get fails at once with "Could not get
// lock".
const lockTimeout = "DPkg::Lock::Timeout=300"

// Remove runs apt-get remove, unless the package is not installed.
func (d *debian) Remove(
	ctx context.Context,
	name string,
) (bool, error) {
	if installed, err := d.installed(ctx, name, ""); err != nil || !installed {
		return false, err
	}
	if _, err := d.exec.RunPrivilegedCmd(
		ctx,
		"apt-get",
		[]string{"remove", "-y", "-o", lockTimeout, "--", name},
	); err != nil {
		return false, fmt.Errorf("remove %s: %w", name, err)
	}

	return true, nil
}

// installed reports whether gohai finds name installed, at version when
// it is not empty.
func (d *debian) installed(
	ctx context.Context,
	name string,
	version string,
) (bool, error) {
	p, err := d.Get(ctx, name)
	if errors.Is(err, behavior.ErrNotFound) {
		return false, nil
	}

	return err == nil && (version == "" || p.Version == version), err
}
