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

package os

import (
	"context"
	"fmt"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// osRelease names the distribution: ID and VERSION_ID.
const osRelease = "/etc/os-release"

// linux reads /etc/os-release.
type linux struct {
	fs avfs.VFS
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{fs: h.FS}, nil
}

// Release is the ID and VERSION_ID of /etc/os-release, without quotes.
func (l *linux) Release(
	_ context.Context,
) (Release, error) {
	data, err := l.fs.ReadFile(osRelease)
	if err != nil {
		return Release{}, fmt.Errorf("read %s: %w", osRelease, err)
	}
	var info Release
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"'`)
		switch key {
		case "ID":
			info.Distribution = value
		case "VERSION_ID":
			info.Version = value
		}
	}
	if info.Distribution == "" {
		return Release{}, fmt.Errorf("%s has no ID", osRelease)
	}

	return info, nil
}
