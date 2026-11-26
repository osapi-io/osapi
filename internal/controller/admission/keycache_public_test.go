// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package admission_test

import (
	"context"
)

// TestRemovalDuringLookup: a lookup that read the record before a removal,
// and caches it after, must not keep the removed key verifying.
func (s *StorePublicTestSuite) TestRemovalDuringLookup() {
	s.Run("a removal between the read and the cache wins", func() {
		ctx := context.Background()
		s.put("accepted.m-1", map[string]any{"machine_id": "m-1", "hostname": "web-01"})
		removed := false
		s.kv.beforeGet = func(key string) {
			if key == "accepted.m-1" && !removed {
				removed = true
				// The revocation runs while this read is in flight.
				s.Require().NoError(s.watcher.RemoveAgentKey(ctx, "m-1"))
			}
		}
		_, _ = s.watcher.LookupAgentKey(ctx, "m-1")
		s.kv.beforeGet = nil

		_, err := s.watcher.LookupAgentKey(ctx, "m-1")
		s.Error(err, "the removed key must not verify")
	})
}
