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

	"github.com/osapi-io/gohai/pkg/gohai/collectors/filesystem"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// localTypes are the filesystem types of local disks. Network, virtual and
// container mounts, such as nfs, tmpfs and overlay, are left out.
var localTypes = map[string]bool{
	"ext4": true, "xfs": true, "btrfs": true, "apfs": true, "hfs": true,
}

// reader reads the local filesystems with gohai's filesystem collector,
// on Linux and on macOS, through the host's sources.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Disks is the size, the used and the free bytes of each local
// filesystem. A mount without sizes is left out.
func (r *reader) Disks(
	ctx context.Context,
) ([]Disk, error) {
	got, err := filesystem.NewWith(r.src).Collect(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("read filesystems: %w", err)
	}
	disks := []Disk{}
	for _, m := range got.(*filesystem.Info).Mounts {
		if m.Device == "" || !localTypes[m.Type] || m.Total == 0 {
			continue
		}
		disks = append(
			disks,
			Disk{Name: m.Mountpoint, Total: int(m.Total), Used: int(m.Used), Free: int(m.Free)},
		)
	}

	return disks, nil
}
