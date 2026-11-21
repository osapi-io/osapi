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

package memory_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
)

type MemoryPublicTestSuite struct {
	suite.Suite
}

func TestMemoryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MemoryPublicTestSuite))
}

func (s *MemoryPublicTestSuite) TestGetStats() {
	g := &memory.Gohai{Sources: gohaitest.Linux(s.T(), map[string]string{
		"/proc/meminfo": "MemTotal: 4096 kB\nMemFree: 1024 kB\nMemAvailable: 2048 kB\nCached: 512 kB\nSReclaimable: 0 kB\n",
	}, nil)}
	got, err := g.GetStats()
	s.Require().NoError(err)
	s.Equal(
		&memory.Result{
			Total:     4096 * 1024,
			Available: 2048 * 1024,
			Free:      1024 * 1024,
			Cached:    512 * 1024,
		},
		got,
	)

	_, err = (&memory.Gohai{Sources: gohaitest.Linux(s.T(), nil, nil)}).GetStats()
	s.ErrorContains(err, "get memory")

	s.NotNil(memory.New())
}
