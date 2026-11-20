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

package hostname

import (
	"context"
	"fmt"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// procHostname is the kernel's hostname.
const procHostname = "/proc/sys/kernel/hostname"

// linux reads the kernel's hostname and sets it with hostnamectl.
type linux struct {
	fs        avfs.VFS
	exec      host.Exec
	container bool
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{fs: h.FS, exec: h.Exec, container: h.Container}, nil
}

// Hostname is /proc/sys/kernel/hostname.
func (l *linux) Hostname(
	_ context.Context,
) (string, error) {
	data, err := l.fs.ReadFile(procHostname)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", procHostname, err)
	}

	return strings.TrimSpace(string(data)), nil
}

// SetHostname runs hostnamectl set-hostname when hostnamectl reports a
// different name. A container's runtime owns its hostname, so there it is
// unsupported.
func (l *linux) SetHostname(
	ctx context.Context,
	name string,
) (bool, error) {
	if l.container {
		return false, behavior.ErrUnsupported
	}
	current, err := l.exec.RunCmd(ctx, "hostnamectl", []string{"hostname"})
	if err != nil {
		return false, fmt.Errorf("read the hostname: %w", err)
	}
	if strings.TrimSpace(current) == name {
		return false, nil
	}
	if _, err := l.exec.RunPrivilegedCmd(
		ctx,
		"hostnamectl",
		[]string{"set-hostname", name},
	); err != nil {
		return false, fmt.Errorf("set the hostname: %w", err)
	}

	return true, nil
}
