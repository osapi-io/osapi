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

package notify

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
)

// recordingNotifier is a hand-rolled test double for Notifier that records
// all Notify calls for assertion in tests.
type recordingNotifier struct {
	events []ConditionEvent
}

func (r *recordingNotifier) Notify(
	_ context.Context,
	event ConditionEvent,
) error {
	r.events = append(r.events, event)
	return nil
}

type WatcherTestSuite struct {
	suite.Suite

	notifier *recordingNotifier
	watcher  *Watcher
}

func (s *WatcherTestSuite) SetupTest() {
	s.notifier = &recordingNotifier{}
	s.watcher = NewWatcher(nil, s.notifier, slog.Default())
}

func (s *WatcherTestSuite) TearDownTest() {}

func (s *WatcherTestSuite) TestDetectTransitions() {
	tests := []struct {
		name          string
		key           string
		componentType string
		hostname      string
		conditions    []job.Condition
		setupPrev     func()
		validateFunc  func(events []ConditionEvent)
	}{
		{
			name:          "detects new condition",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions: []job.Condition{
				{
					Type:               job.ConditionMemoryPressure,
					Status:             true,
					Reason:             "memory above threshold",
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev: func() {},
			validateFunc: func(events []ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(job.ConditionMemoryPressure, events[0].Condition)
				s.True(events[0].Active)
			},
		},
		{
			name:          "detects resolved condition",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions:    []job.Condition{},
			setupPrev: func() {
				// Seed previous state with an active condition.
				s.watcher.prev["agents.web_01"] = map[string]bool{
					job.ConditionMemoryPressure: true,
				}
			},
			validateFunc: func(events []ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(job.ConditionMemoryPressure, events[0].Condition)
				s.False(events[0].Active)
			},
		},
		{
			name:          "no notification when no change",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions: []job.Condition{
				{
					Type:               job.ConditionMemoryPressure,
					Status:             true,
					Reason:             "memory above threshold",
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev: func() {
				// Seed previous state identical to incoming conditions.
				s.watcher.prev["agents.web_01"] = map[string]bool{
					job.ConditionMemoryPressure: true,
				}
			},
			validateFunc: func(events []ConditionEvent) {
				s.Empty(events)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset watcher and notifier state between sub-tests.
			s.notifier.events = nil
			s.watcher.prev = make(map[string]map[string]bool)

			tt.setupPrev()

			s.watcher.detectTransitions(
				tt.key,
				tt.componentType,
				tt.hostname,
				tt.conditions,
			)

			tt.validateFunc(s.notifier.events)
		})
	}
}

func TestWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(WatcherTestSuite))
}
