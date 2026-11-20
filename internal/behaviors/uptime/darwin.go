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
	"regexp"
	"strconv"
	"time"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// bootSec is the seconds of the boot time that sysctl -n kern.boottime
// prints: { sec = 1767229145, usec = 0 } Thu Jan  1 00:59:05 2026.
var bootSec = regexp.MustCompile(`sec = (\d+)`)

// darwin reads the boot time from sysctl, and the uptime is the host's time
// since it.
type darwin struct {
	exec host.Exec
	now  func() time.Time
}

// newDarwin is the provider on macOS.
func newDarwin(
	h host.Host,
) (Provider, error) {
	return &darwin{exec: h.Exec, now: h.Time}, nil
}

// Uptime is the host's time since its boot time.
func (d *darwin) Uptime(
	ctx context.Context,
) (time.Duration, error) {
	out, err := d.exec.RunCmd(ctx, "sysctl", []string{"-n", "kern.boottime"})
	if err != nil {
		return 0, fmt.Errorf("read the boot time: %w", err)
	}
	m := bootSec.FindStringSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("no boot time in %q", out)
	}
	sec, _ := strconv.ParseInt(m[1], 10, 64)

	return d.now().Sub(time.Unix(sec, 0)), nil
}
