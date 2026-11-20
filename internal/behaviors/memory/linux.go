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

package memory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// procMeminfo is the kernel's memory, in kB.
const procMeminfo = "/proc/meminfo"

// linux reads /proc/meminfo.
type linux struct {
	fs avfs.VFS
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{fs: h.FS}, nil
}

// Memory is MemTotal and MemFree, and the total less MemAvailable.
func (l *linux) Memory(
	_ context.Context,
) (Usage, error) {
	data, err := l.fs.ReadFile(procMeminfo)
	if err != nil {
		return Usage{}, fmt.Errorf("read %s: %w", procMeminfo, err)
	}
	kb := map[string]int{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		kb[strings.TrimSuffix(fields[0], ":")] = n
	}
	total, ok := kb["MemTotal"]
	if !ok {
		return Usage{}, fmt.Errorf("%s has no MemTotal", procMeminfo)
	}

	return Usage{
		Total: total * 1024,
		Free:  kb["MemFree"] * 1024,
		Used:  (total - kb["MemAvailable"]) * 1024,
	}, nil
}
