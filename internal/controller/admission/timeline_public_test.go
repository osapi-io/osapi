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
package admission_test

import (
	"context"
	"errors"

	"go.uber.org/mock/gomock"
)

// timelineLog keeps the events that the watcher writes.
type timelineLog struct {
	events []string
	err    error
}

func (l *timelineLog) WriteAgentTimelineEvent(
	_ context.Context,
	machineID, hostname, event, message string,
) error {
	l.events = append(l.events, machineID+" "+hostname+" "+event+": "+message)

	return l.err
}

// TestTimeline: each decision of admission writes one event to the
// machine's timeline, with its hostname (osapi-io/osapi#1424).
func (s *StorePublicTestSuite) TestTimeline() {
	log := &timelineLog{}
	s.watcher.SetTimeline(log)
	defer s.watcher.SetTimeline(nil)
	ctx := context.Background()

	s.pending("m-1", "web-01")
	s.Require().NoError(s.watcher.Admit(ctx, "m-1", s.fp))
	s.Require().NoError(s.watcher.Revoke(ctx, "m-1"))
	s.pending("m-2", "web-02")
	s.Require().NoError(s.watcher.Reject(ctx, "m-2", "unknown host"))
	s.pending("m-3", "web-03")
	s.Require().NoError(s.watcher.Delete(ctx, "m-3"))
	s.watcher.SetJoinTokens(newTokens(gomock.NewController(s.T()), nil))
	_, err := s.watcher.Join(ctx, joinRequest("osapi_jt_t1.secret"))
	s.Require().NoError(err)

	s.Equal([]string{
		"m-1 web-01 admitted: admitted by an operator",
		"m-1 web-01 revoked: revoked by an operator",
		"m-2 web-02 rejected: rejected by an operator: unknown host",
		"m-3 web-03 deleted: deleted by an operator",
		"m-1 web-01 admitted: admitted by join token t1",
	}, log.events)

	// A decision that fails writes no event.
	log.events = nil
	s.Error(s.watcher.Admit(ctx, "m-none", s.fp))
	s.Error(s.watcher.Reject(ctx, "m-none", "x"))
	s.Error(s.watcher.Revoke(ctx, "m-none"))
	s.Error(s.watcher.Delete(ctx, "m-none"))
	s.Empty(log.events)

	// A timeline that fails is logged, and the decision stands.
	log.err = errors.New("kv down")
	s.pending("m-4", "web-04")
	s.NoError(s.watcher.Admit(ctx, "m-4", s.fp))
	s.Len(log.events, 1)
}
