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

package process_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/telemetry/process"
)

type ProcessPublicTestSuite struct {
	suite.Suite
}

// statLine is /proc/7/stat of a process 100 ticks after boot with utime
// and stime in ticks of 1/100 s.
func statLine(
	utime, stime int,
) string {
	return "7 (osapi) S 1 7 7 0 -1 4194560 1 2 3 4 " +
		itoa(utime) + " " + itoa(stime) + " 7 8 20 0 1 0 100 9 10\n"
}

func itoa(
	n int,
) string {
	return strconv.Itoa(n)
}

func (suite *ProcessPublicTestSuite) writeHost(
	fs avfs.VFS,
	utime int,
) {
	suite.Require().NoError(fs.MkdirAll("/proc/7", 0o755))
	suite.Require().NoError(fs.WriteFile("/proc/stat", []byte("btime 1700000000\n"), 0o644))
	suite.Require().NoError(fs.WriteFile("/proc/7/stat", []byte(statLine(utime, 0)), 0o644))
	suite.Require().NoError(fs.WriteFile("/proc/7/status", []byte("VmRSS:\t2048 kB\n"), 0o644))
}

func (suite *ProcessPublicTestSuite) TestGetMetrics() {
	suite.Run("the real process", func() {
		got, err := process.New().GetMetrics()
		suite.Require().NoError(err)
		suite.Greater(got.Goroutines, 0)
		suite.Positive(got.RSSBytes)
	})

	suite.Run("the CPU share between two samples", func() {
		fs := memfs.New()
		suite.writeHost(fs, 200) // 2 s of CPU, 9 s after start
		now := time.Unix(1700000010, 0)
		p := process.ExportNewProvider(7, source.Sources{
			FS: fs, Family: "debian", Now: func() time.Time { return now },
		}, func() time.Time { return now })

		first, err := p.GetMetrics()
		suite.Require().NoError(err)
		suite.InDelta(
			2.0/9*100,
			first.CPUPercent,
			0.001,
			"the first sample is the average since start",
		)
		suite.Equal(int64(2048*1024), first.RSSBytes)

		suite.writeHost(fs, 300) // 1 s more CPU in the next 4 s
		now = now.Add(4 * time.Second)
		second, err := p.GetMetrics()
		suite.Require().NoError(err)
		suite.InDelta(25.0, second.CPUPercent, 0.001)
	})

	suite.Run("no such process", func() {
		p := process.ExportNewProvider(7, source.Sources{
			FS: memfs.New(), Family: "debian", Now: time.Now,
		}, time.Now)
		_, err := p.GetMetrics()
		suite.ErrorContains(err, "get process")
	})
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestProcessPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessPublicTestSuite))
}
