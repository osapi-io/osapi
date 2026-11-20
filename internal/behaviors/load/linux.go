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

package load

import (
	"context"
	"fmt"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// procLoadavg is the kernel's load averages.
const procLoadavg = "/proc/loadavg"

// linux reads /proc/loadavg.
type linux struct {
	fs avfs.VFS
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{fs: h.FS}, nil
}

// Load is the first three fields of /proc/loadavg.
func (l *linux) Load(
	_ context.Context,
) (Averages, error) {
	data, err := l.fs.ReadFile(procLoadavg)
	if err != nil {
		return Averages{}, fmt.Errorf("read %s: %w", procLoadavg, err)
	}

	return averages(string(data))
}
