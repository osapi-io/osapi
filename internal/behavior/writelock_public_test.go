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

package behavior_test

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type WriteLockPublicTestSuite struct {
	suite.Suite
}

// counting is a provider that records how many calls run at once, and the
// most that ever did. Each call takes long enough for the others to start.
type counting struct {
	active *atomic.Int32
	peak   *atomic.Int32
}

func (c counting) enter() {
	n := c.active.Add(1)
	for {
		p := c.peak.Load()
		if n <= p || c.peak.CompareAndSwap(p, n) {
			break
		}
	}
	time.Sleep(50 * time.Millisecond)
	c.active.Add(-1)
}

func countingRead(
	_ context.Context,
	c counting,
	in getIn,
) (entry, error) {
	c.enter()

	return entry{Key: in.Key}, nil
}

func countingWrite(
	_ context.Context,
	c counting,
	in getIn,
) (result, error) {
	c.enter()

	return result{Key: in.Key, Changed: true}, nil
}

// TestWrites: jobs that write one behavior run one at a time, as the agent
// runs several jobs at once; reads of it run together.
func (s *WriteLockPublicTestSuite) TestWrites() {
	b := behavior.Define("locktest", "A behavior whose calls count",
		behavior.Node,
		behavior.Read("get", countingRead).Key("key"),
		behavior.Write("update", countingWrite).Key("key"),
	)
	input, err := json.Marshal(getIn{Key: "k"})
	s.Require().NoError(err)

	for _, tc := range []struct {
		method string
		want   func(peak int32)
	}{
		{method: "update", want: func(peak int32) { s.Equal(int32(1), peak, "one write at a time") }},
		{method: "get", want: func(peak int32) { s.Greater(peak, int32(1), "reads run together") }},
	} {
		s.Run(tc.method, func() {
			c := counting{active: &atomic.Int32{}, peak: &atomic.Int32{}}
			var wg sync.WaitGroup
			for range 4 {
				wg.Go(func() {
					_, err := b.Method(tc.method).Invoke(context.Background(), c, input)
					s.NoError(err)
				})
			}
			wg.Wait()
			tc.want(c.peak.Load())
		})
	}
}

func TestWriteLockPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(WriteLockPublicTestSuite))
}
