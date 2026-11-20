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

package uptime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// procUptime is the kernel's uptime: the seconds since boot, then the
// seconds each CPU spent idle.
const procUptime = "/proc/uptime"

// linux reads the uptime the kernel keeps in /proc/uptime.
type linux struct {
	fs avfs.VFS
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{fs: h.FS}, nil
}

// Uptime is the first field of /proc/uptime.
func (l *linux) Uptime(
	_ context.Context,
) (time.Duration, error) {
	data, err := l.fs.ReadFile(procUptime)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", procUptime, err)
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("%s is empty", procUptime)
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", procUptime, err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}
