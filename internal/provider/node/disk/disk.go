// Copyright (c) 2024 John Dewey

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

// Package disk reads the local disks of the host through gohai.
package disk

import (
	"context"
	"fmt"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/filesystem"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

var _ Provider = (*Gohai)(nil)

// localTypes are the filesystem types of local disks. Network, virtual
// and container mounts, such as nfs, tmpfs and overlay, are left out.
var localTypes = map[string]bool{
	"ext4": true, "xfs": true, "btrfs": true, "apfs": true, "hfs": true,
}

// Gohai reads the usage of the host's local disks with gohai's filesystem
// collector, on every Linux family and on macOS.
type Gohai struct {
	// Sources are what the collector reads. A zero field is the real
	// host's.
	Sources source.Sources
}

// New returns a provider that reads the real host.
func New() *Gohai {
	return &Gohai{}
}

// GetLocalUsageStats is the total, used and free bytes of each local
// disk's mount. A mount without sizes, such as one that df cannot read,
// is left out.
func (g *Gohai) GetLocalUsageStats() ([]Result, error) {
	got, err := filesystem.NewWith(g.Sources).Collect(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("get disk usage: %w", err)
	}
	out := []Result{}
	for _, m := range got.(*filesystem.Info).Mounts {
		if m.Device == "" || !localTypes[m.Type] || m.Total == 0 {
			continue
		}
		out = append(out, Result{Name: m.Mountpoint, Total: m.Total, Used: m.Used, Free: m.Free})
	}

	return out, nil
}
