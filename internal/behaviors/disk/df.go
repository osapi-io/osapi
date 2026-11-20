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

package disk

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// df runs df -kP with args, which pick the filesystems, and is each one:
// POSIX output in kB, one line each after the header. A mount point with
// spaces is the rest of its line.
func df(
	ctx context.Context,
	exec host.Exec,
	args ...string,
) ([]Disk, error) {
	out, err := exec.RunCmd(ctx, "df", append([]string{"-kP"}, args...))
	if err != nil {
		return nil, fmt.Errorf("run df: %w", err)
	}
	disks := []Disk{}
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if i == 0 || len(fields) < 6 {
			continue
		}
		var kb [3]int
		for j := range kb {
			n, err := strconv.Atoi(fields[j+1])
			if err != nil {
				return nil, fmt.Errorf("df: %q: %w", line, err)
			}
			kb[j] = n * 1024
		}
		disks = append(disks, Disk{
			Name:  strings.Join(fields[5:], " "),
			Total: kb[0], Used: kb[1], Free: kb[2],
		})
	}

	return disks, nil
}

// linux is df of the local Linux filesystems.
type linux struct {
	exec host.Exec
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{exec: h.Exec}, nil
}

// Disks is each ext4, xfs and btrfs filesystem.
func (l *linux) Disks(
	ctx context.Context,
) ([]Disk, error) {
	return df(ctx, l.exec, "-t", "ext4", "-t", "xfs", "-t", "btrfs")
}

// darwin is df of the local macOS filesystems.
type darwin struct {
	exec host.Exec
}

// newDarwin is the provider on macOS.
func newDarwin(
	h host.Host,
) (Provider, error) {
	return &darwin{exec: h.Exec}, nil
}

// Disks is each apfs and hfs filesystem.
func (d *darwin) Disks(
	ctx context.Context,
) ([]Disk, error) {
	return df(ctx, d.exec, "-T", "apfs,hfs")
}
