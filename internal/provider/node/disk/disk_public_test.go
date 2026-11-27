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

package disk_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
)

type DiskPublicTestSuite struct {
	suite.Suite
}

func TestDiskPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DiskPublicTestSuite))
}

func (s *DiskPublicTestSuite) TestGetLocalUsageStats() {
	const dfHead = "Filesystem 1024-blocks Used Available Capacity Mounted on\n"
	g := &disk.Gohai{Sources: gohaitest.Linux(s.T(), map[string]string{
		"/proc/mounts": "/dev/sda1 / ext4 rw 0 0\n" +
			"tmpfs /run tmpfs rw 0 0\n" +
			"/dev/sdb1 /data xfs rw 0 0\n" +
			"server:/export /mnt/nfs nfs rw 0 0\n",
	}, map[string]string{
		// gohai runs one df for the local mounts and one for each network
		// mount, so a hung network mount cannot stall the local sizes.
		"df -akP -- / /run /data": dfHead +
			"/dev/sda1 100 40 60 40% /\n" +
			"tmpfs 10 1 9 10% /run\n",
		"df -akP -- /mnt/nfs": dfHead +
			"server:/export 1000 1 999 1% /mnt/nfs\n",
	})}
	got, err := g.GetLocalUsageStats()
	s.Require().NoError(err)
	s.Equal([]disk.Result{{Name: "/", Total: 100 * 1024, Used: 40 * 1024, Free: 60 * 1024}}, got,
		"tmpfs and nfs are not local; /data has no sizes")

	_, err = (&disk.Gohai{Sources: gohaitest.Linux(s.T(), nil, nil)}).GetLocalUsageStats()
	s.ErrorContains(err, "get disk usage")

	s.NotNil(disk.New())
}
