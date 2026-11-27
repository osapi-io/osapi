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

package notify_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/notify"
	notifymocks "github.com/osapi-io/osapi/internal/controller/notify/mocks"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type WatcherPublicTestSuite struct {
	suite.Suite

	ctrl         *gomock.Controller
	mockNotifier *notifymocks.MockNotifier
	capturedEvts []notify.ConditionEvent
	watcher      *notify.Watcher
}

func (s *WatcherPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockNotifier = notifymocks.NewMockNotifier(s.ctrl)
	s.capturedEvts = nil
	s.watcher = notify.NewWatcher(nil, "registry", s.mockNotifier, slog.Default(), 0)
}

func (s *WatcherPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *WatcherPublicTestSuite) TestDetectTransitions() {
	tests := []struct {
		name          string
		key           string
		componentType string
		hostname      string
		conditions    []job.Condition
		setupPrev     func()
		wantEvents    int
		validateFunc  func(events []notify.ConditionEvent)
	}{
		{
			name:          "detects new condition",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions: []job.Condition{
				{
					Type:               types.ConditionMemoryPressure,
					Status:             true,
					Reason:             "memory above threshold",
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev:  func() {},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(types.ConditionMemoryPressure, events[0].Condition)
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
				notify.WatcherSetPrevEntry(
					s.watcher,
					"agents.web_01",
					map[types.ConditionType]*notify.ExportConditionState{
						types.ConditionMemoryPressure: notify.NewConditionState(true, time.Now()),
					},
				)
			},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(types.ConditionMemoryPressure, events[0].Condition)
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
					Type:               types.ConditionMemoryPressure,
					Status:             true,
					Reason:             "memory above threshold",
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev: func() {
				notify.WatcherSetPrevEntry(
					s.watcher,
					"agents.web_01",
					map[types.ConditionType]*notify.ExportConditionState{
						types.ConditionMemoryPressure: notify.NewConditionState(true, time.Now()),
					},
				)
			},
			wantEvents: 0,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Empty(events)
			},
		},
		{
			name:          "re-notifies after renotify interval",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions: []job.Condition{
				{
					Type:               types.ConditionMemoryPressure,
					Status:             true,
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev: func() {
				notify.WatcherSetRenotifyInterval(s.watcher, 1*time.Millisecond)
				notify.WatcherSetPrevEntry(
					s.watcher,
					"agents.web_01",
					map[types.ConditionType]*notify.ExportConditionState{
						types.ConditionMemoryPressure: notify.NewConditionState(
							true,
							time.Now().Add(-1*time.Second),
						),
					},
				)
			},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.True(events[0].Active)
				s.Equal(types.ConditionMemoryPressure, events[0].Condition)
			},
		},
		{
			name:          "does not re-notify before interval elapses",
			key:           "agents.web_01",
			componentType: "agent",
			hostname:      "web-01",
			conditions: []job.Condition{
				{
					Type:               types.ConditionMemoryPressure,
					Status:             true,
					LastTransitionTime: time.Now(),
				},
			},
			setupPrev: func() {
				notify.WatcherSetRenotifyInterval(s.watcher, 1*time.Hour)
				notify.WatcherSetPrevEntry(
					s.watcher,
					"agents.web_01",
					map[types.ConditionType]*notify.ExportConditionState{
						types.ConditionMemoryPressure: notify.NewConditionState(true, time.Now()),
					},
				)
			},
			wantEvents: 0,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Empty(events)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset captured events and watcher state between sub-tests.
			s.capturedEvts = nil
			notify.WatcherSetPrev(
				s.watcher,
				make(map[string]map[types.ConditionType]*notify.ExportConditionState),
			)

			if tt.wantEvents > 0 {
				s.mockNotifier.EXPECT().
					Notify(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						event notify.ConditionEvent,
					) error {
						s.capturedEvts = append(s.capturedEvts, event)
						return nil
					}).
					Times(tt.wantEvents)
			}

			tt.setupPrev()

			notify.WatcherDetectTransitions(
				context.Background(),
				s.watcher,
				tt.key,
				tt.componentType,
				tt.hostname,
				tt.conditions,
			)

			tt.validateFunc(s.capturedEvts)
		})
	}
}

func (s *WatcherPublicTestSuite) TestParseRegistryKey() {
	tests := []struct {
		name              string
		key               string
		wantComponentType string
		wantHostname      string
		validateFunc      func(bool)
	}{
		{
			name:              "agents prefix returns agent type",
			key:               "agents.web-01",
			wantComponentType: "agent",
			wantHostname:      "web-01",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:              "api prefix returns api type",
			key:               "api.api-server-01",
			wantComponentType: "api",
			wantHostname:      "api-server-01",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:              "nats prefix returns nats type",
			key:               "nats.nats-01",
			wantComponentType: "nats",
			wantHostname:      "nats-01",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:              "controller prefix returns controller type",
			key:               "controller.ctrl-01",
			wantComponentType: "controller",
			wantHostname:      "ctrl-01",
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:              "unknown prefix returns false",
			key:               "unknown.host-01",
			wantComponentType: "",
			wantHostname:      "",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:              "key without dot returns false",
			key:               "invalid",
			wantComponentType: "",
			wantHostname:      "",
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			componentType, hostname, ok := notify.ParseRegistryKey(tt.key)

			s.Equal(tt.wantComponentType, componentType)
			s.Equal(tt.wantHostname, hostname)
			tt.validateFunc(ok)
		})
	}
}

func (s *WatcherPublicTestSuite) TestResolveDisplayName() {
	tests := []struct {
		name          string
		componentType string
		identifier    string
		value         []byte
		validateFunc  func(string)
	}{
		{
			name:          "non-agent returns identifier as-is",
			componentType: "controller",
			identifier:    "ctrl-01",
			value:         nil,
			validateFunc: func(got string) {
				s.Equal("ctrl-01", got)
			},
		},
		{
			name:          "agent with valid hostname returns hostname",
			componentType: "agent",
			identifier:    "abc123",
			value:         []byte(`{"hostname":"web-01"}`),
			validateFunc: func(got string) {
				s.Equal("web-01", got)
			},
		},
		{
			name:          "agent with invalid JSON returns identifier",
			componentType: "agent",
			identifier:    "abc123",
			value:         []byte("invalid"),
			validateFunc: func(got string) {
				s.Equal("abc123", got)
			},
		},
		{
			name:          "agent with empty hostname returns identifier",
			componentType: "agent",
			identifier:    "abc123",
			value:         []byte(`{"hostname":""}`),
			validateFunc: func(got string) {
				s.Equal("abc123", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := notify.ExportResolveDisplayName(
				tt.componentType,
				tt.identifier,
				tt.value,
			)
			tt.validateFunc(got)
		})
	}
}

func (s *WatcherPublicTestSuite) TestExtractConditions() {
	tests := []struct {
		name          string
		key           string
		componentType string
		value         func() []byte
		wantOK        bool
		validateFunc  func(conditions []job.Condition)
	}{
		{
			name:          "extracts conditions from agent registration",
			key:           "agents.web-01",
			componentType: "agent",
			value: func() []byte {
				reg := job.AgentRegistration{
					Hostname: "web-01",
					Conditions: []job.Condition{
						{
							Type:   types.ConditionMemoryPressure,
							Status: true,
						},
					},
				}
				b, _ := json.Marshal(reg)
				return b
			},
			wantOK: true,
			validateFunc: func(conditions []job.Condition) {
				s.Require().Len(conditions, 1)
				s.Equal(types.ConditionMemoryPressure, conditions[0].Type)
				s.True(conditions[0].Status)
			},
		},
		{
			name:          "extracts conditions from api component registration",
			key:           "api.api-server-01",
			componentType: "api",
			value: func() []byte {
				reg := job.ComponentRegistration{
					Type:     "api",
					Hostname: "api-server-01",
					Conditions: []job.Condition{
						{
							Type:   types.ConditionHighLoad,
							Status: true,
						},
					},
				}
				b, _ := json.Marshal(reg)
				return b
			},
			wantOK: true,
			validateFunc: func(conditions []job.Condition) {
				s.Require().Len(conditions, 1)
				s.Equal(types.ConditionHighLoad, conditions[0].Type)
			},
		},
		{
			name:          "extracts conditions from nats component registration",
			key:           "nats.nats-01",
			componentType: "nats",
			value: func() []byte {
				reg := job.ComponentRegistration{
					Type:       "nats",
					Hostname:   "nats-01",
					Conditions: []job.Condition{},
				}
				b, _ := json.Marshal(reg)
				return b
			},
			wantOK: true,
			validateFunc: func(conditions []job.Condition) {
				s.Empty(conditions)
			},
		},
		{
			name:          "returns false for invalid agent json",
			key:           "agents.web-01",
			componentType: "agent",
			value: func() []byte {
				return []byte("not-valid-json")
			},
			wantOK: false,
			validateFunc: func(conditions []job.Condition) {
				s.Nil(conditions)
			},
		},
		{
			name:          "returns false for invalid component json",
			key:           "api.api-01",
			componentType: "api",
			value: func() []byte {
				return []byte("not-valid-json")
			},
			wantOK: false,
			validateFunc: func(conditions []job.Condition) {
				s.Nil(conditions)
			},
		},
		{
			name:          "returns false for unknown component type",
			key:           "other.host-01",
			componentType: "other",
			value: func() []byte {
				return []byte("{}")
			},
			wantOK: false,
			validateFunc: func(conditions []job.Condition) {
				s.Nil(conditions)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			conditions, ok := notify.WatcherExtractConditions(
				s.watcher,
				tt.key,
				tt.componentType,
				tt.value(),
			)

			s.Equal(tt.wantOK, ok)
			tt.validateFunc(conditions)
		})
	}
}

func (s *WatcherPublicTestSuite) TestHandleEntry() {
	tests := []struct {
		name         string
		setupEntry   func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry
		wantEvents   int
		validateFunc func(events []notify.ConditionEvent)
	}{
		{
			name: "ignores entry with unrecognized key",
			setupEntry: func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
				entry := jobmocks.NewMockKeyValueEntry(ctrl)
				entry.EXPECT().Key().Return("unknown.host")
				return entry
			},
			wantEvents: 0,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Empty(events)
			},
		},
		{
			name: "handles delete operation and emits ComponentUnreachable",
			setupEntry: func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
				entry := jobmocks.NewMockKeyValueEntry(ctrl)
				entry.EXPECT().Key().Return("agents.web-01")
				entry.EXPECT().Operation().Return(jetstream.KeyValueDelete)
				return entry
			},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(notify.ConditionComponentUnreachable, events[0].Condition)
				s.True(events[0].Active)
				s.Equal("heartbeat expired", events[0].Reason)
			},
		},
		{
			name: "handles purge operation and emits ComponentUnreachable",
			setupEntry: func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
				// Operation() is called twice: first checks KeyValueDelete (false),
				// then checks KeyValuePurge (true).
				entry := jobmocks.NewMockKeyValueEntry(ctrl)
				entry.EXPECT().Key().Return("nats.nats-01")
				entry.EXPECT().Operation().Return(jetstream.KeyValuePurge).Times(2)
				return entry
			},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("nats", events[0].ComponentType)
				s.Equal("nats-01", events[0].Hostname)
				s.Equal(notify.ConditionComponentUnreachable, events[0].Condition)
				s.True(events[0].Active)
			},
		},
		{
			name: "handles put operation with valid agent registration",
			setupEntry: func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
				reg := job.AgentRegistration{
					Hostname: "web-01",
					Conditions: []job.Condition{
						{Type: types.ConditionMemoryPressure, Status: true},
					},
				}
				b, _ := json.Marshal(reg)

				// Operation() is called twice: first checks KeyValueDelete (false),
				// then checks KeyValuePurge (false), so Put proceeds to Value().
				entry := jobmocks.NewMockKeyValueEntry(ctrl)
				entry.EXPECT().Key().Return("agents.web-01")
				entry.EXPECT().Operation().Return(jetstream.KeyValuePut).Times(2)
				entry.EXPECT().Value().Return(b).AnyTimes()
				return entry
			},
			wantEvents: 1,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(types.ConditionMemoryPressure, events[0].Condition)
				s.True(events[0].Active)
			},
		},
		{
			name: "ignores put operation with invalid json",
			setupEntry: func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
				// Operation() is called twice: checks KeyValueDelete and KeyValuePurge.
				entry := jobmocks.NewMockKeyValueEntry(ctrl)
				entry.EXPECT().Key().Return("agents.web-01")
				entry.EXPECT().Operation().Return(jetstream.KeyValuePut).Times(2)
				entry.EXPECT().Value().Return([]byte("invalid")).AnyTimes()
				return entry
			},
			wantEvents: 0,
			validateFunc: func(events []notify.ConditionEvent) {
				s.Empty(events)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			s.capturedEvts = nil
			notify.WatcherSetPrev(
				s.watcher,
				make(map[string]map[types.ConditionType]*notify.ExportConditionState),
			)

			if tt.wantEvents > 0 {
				s.mockNotifier.EXPECT().
					Notify(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						event notify.ConditionEvent,
					) error {
						s.capturedEvts = append(s.capturedEvts, event)
						return nil
					}).
					Times(tt.wantEvents)
			}

			entry := tt.setupEntry(ctrl)
			notify.WatcherHandleEntry(context.Background(), s.watcher, entry)

			tt.validateFunc(s.capturedEvts)
		})
	}
}

func (s *WatcherPublicTestSuite) TestHandleDelete() {
	tests := []struct {
		name          string
		key           string
		componentType string
		hostname      string
		setupPrev     func()
		validateFunc  func(events []notify.ConditionEvent)
	}{
		{
			name:          "emits ComponentUnreachable event",
			key:           "agents.web-01",
			componentType: "agent",
			hostname:      "web-01",
			setupPrev:     func() {},
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("agent", events[0].ComponentType)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(notify.ConditionComponentUnreachable, events[0].Condition)
				s.True(events[0].Active)
				s.Equal("heartbeat expired", events[0].Reason)
			},
		},
		{
			name:          "removes key from prev state",
			key:           "agents.web-01",
			componentType: "agent",
			hostname:      "web-01",
			setupPrev: func() {
				notify.WatcherSetPrevEntry(
					s.watcher,
					"agents.web-01",
					map[types.ConditionType]*notify.ExportConditionState{
						types.ConditionMemoryPressure: notify.NewConditionState(true, time.Now()),
					},
				)
			},
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal(notify.ConditionComponentUnreachable, events[0].Condition)
				prev := notify.WatcherGetPrev(s.watcher)
				s.Nil(prev["agents.web-01"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.capturedEvts = nil
			notify.WatcherSetPrev(
				s.watcher,
				make(map[string]map[types.ConditionType]*notify.ExportConditionState),
			)

			// HandleDelete always emits exactly one ComponentUnreachable event.
			s.mockNotifier.EXPECT().
				Notify(gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					event notify.ConditionEvent,
				) error {
					s.capturedEvts = append(s.capturedEvts, event)
					return nil
				}).
				Times(1)

			tt.setupPrev()

			notify.WatcherHandleDelete(
				context.Background(),
				s.watcher,
				tt.key,
				tt.componentType,
				tt.hostname,
			)

			tt.validateFunc(s.capturedEvts)
		})
	}
}

func (s *WatcherPublicTestSuite) TestStart() {
	// watch returns a Registry whose KVWatch gives ch, or fails with err.
	watch := func(
		ctrl *gomock.Controller,
		ch chan jetstream.KeyValueEntry,
		err error,
	) *notifymocks.MockRegistry {
		kv := notifymocks.NewMockRegistry(ctrl)
		kv.EXPECT().
			KVWatch(gomock.Any(), "registry", ">").
			Return((<-chan jetstream.KeyValueEntry)(ch), err)

		return kv
	}

	tests := []struct {
		name         string
		setup        func(ctrl *gomock.Controller, ctx context.Context) *notifymocks.MockRegistry
		validateFunc func(err error)
	}{
		{
			name: "returns error when the watch fails",
			setup: func(ctrl *gomock.Controller, _ context.Context) *notifymocks.MockRegistry {
				return watch(ctrl, nil, errors.New("watch failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "watch failed")
			},
		},
		{
			name: "returns nil when context is cancelled",
			setup: func(ctrl *gomock.Controller, _ context.Context) *notifymocks.MockRegistry {
				return watch(ctrl, make(chan jetstream.KeyValueEntry, 1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			s.capturedEvts = nil
			notify.WatcherSetPrev(
				s.watcher,
				make(map[string]map[types.ConditionType]*notify.ExportConditionState),
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			kv := tt.setup(ctrl, ctx)
			notify.WatcherSetKV(s.watcher, kv, "registry")

			// Cancel context immediately for cases that need context cancellation.
			// For cases with closed channels, the channel close drives exit.
			switch tt.name {
			case "returns nil when context is cancelled":
				cancel()
			}

			doneCh := make(chan error, 1)
			go func() {
				doneCh <- s.watcher.Start(ctx)
			}()

			select {
			case err := <-doneCh:
				tt.validateFunc(err)
			case <-time.After(2 * time.Second):
				s.Fail("Start did not return within timeout")
			}
		})
	}
}

func TestWatcherPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(WatcherPublicTestSuite))
}

func (s *WatcherPublicTestSuite) TestStartWatchesAgain() {
	tests := []struct {
		name string
		// second is the error of the second watch; a third watch succeeds.
		second  error
		watches int
	}{
		{name: "when the watch ends, it watches again", watches: 2},
		{
			name:    "when a new watch fails, it tries again",
			second:  errors.New("nats: no responders"),
			watches: 3,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			closed := make(chan jetstream.KeyValueEntry)
			close(closed)
			// The last watch sends one entry, and the test ends when the
			// watcher handles it.
			entry := jobmocks.NewMockKeyValueEntry(ctrl)
			entry.EXPECT().Key().DoAndReturn(func() string {
				cancel()

				return "unknown.host"
			})
			last := make(chan jetstream.KeyValueEntry, 1)
			last <- entry

			kv := notifymocks.NewMockRegistry(ctrl)
			calls := []any{
				kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
					Return((<-chan jetstream.KeyValueEntry)(closed), nil),
			}
			if tt.second != nil {
				calls = append(calls, kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
					Return(nil, tt.second))
			}
			calls = append(calls, kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
				Return((<-chan jetstream.KeyValueEntry)(last), nil))
			gomock.InOrder(calls...)
			s.Len(calls, tt.watches)

			w := notify.NewWatcher(kv, "registry", s.mockNotifier, slog.Default(), 0)
			notify.SetRewatchDelay(w, time.Millisecond)

			done := make(chan error, 1)
			go func() { done <- w.Start(ctx) }()
			select {
			case err := <-done:
				s.NoError(err)
			case <-time.After(2 * time.Second):
				s.Fail("Start did not return within timeout")
			}
		})
	}
	s.Run("when ctx ends after a new watch failed, it returns nil", func() {
		ctrl := gomock.NewController(s.T())
		defer ctrl.Finish()
		ctx, cancel := context.WithCancel(context.Background())
		closed := make(chan jetstream.KeyValueEntry)
		close(closed)
		kv := notifymocks.NewMockRegistry(ctrl)
		gomock.InOrder(
			kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
				Return((<-chan jetstream.KeyValueEntry)(closed), nil),
			kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
				DoAndReturn(func(context.Context, string, string, ...jetstream.WatchOpt) (<-chan jetstream.KeyValueEntry, error) {
					cancel()

					return nil, errors.New("nats: connection closed")
				}),
		)
		w := notify.NewWatcher(kv, "registry", s.mockNotifier, slog.Default(), 0)
		notify.SetRewatchDelay(w, time.Millisecond)
		s.NoError(w.Start(ctx))
	})
	s.Run("when ctx ends while it waits to watch again, it returns nil", func() {
		ctrl := gomock.NewController(s.T())
		defer ctrl.Finish()
		ctx, cancel := context.WithCancel(context.Background())
		closed := make(chan jetstream.KeyValueEntry)
		close(closed)
		kv := notifymocks.NewMockRegistry(ctrl)
		kv.EXPECT().KVWatch(gomock.Any(), "registry", ">").
			DoAndReturn(func(context.Context, string, string, ...jetstream.WatchOpt) (<-chan jetstream.KeyValueEntry, error) {
				cancel()

				return closed, nil
			})
		w := notify.NewWatcher(kv, "registry", s.mockNotifier, slog.Default(), 0)
		notify.SetRewatchDelay(w, time.Hour)
		s.NoError(w.Start(ctx))
	})
}

func (s *WatcherPublicTestSuite) TestUnreachable() {
	put := func(ctrl *gomock.Controller, hostname string) *jobmocks.MockKeyValueEntry {
		b, _ := json.Marshal(job.AgentRegistration{Hostname: hostname})
		entry := jobmocks.NewMockKeyValueEntry(ctrl)
		entry.EXPECT().Key().Return("agents.m-1").AnyTimes()
		entry.EXPECT().Operation().Return(jetstream.KeyValuePut).AnyTimes()
		entry.EXPECT().Value().Return(b).AnyTimes()

		return entry
	}
	expire := func(ctrl *gomock.Controller) *jobmocks.MockKeyValueEntry {
		entry := jobmocks.NewMockKeyValueEntry(ctrl)
		entry.EXPECT().Key().Return("agents.m-1").AnyTimes()
		entry.EXPECT().Operation().Return(jetstream.KeyValueDelete).AnyTimes()

		return entry
	}
	tests := []struct {
		name         string
		steps        func(ctrl *gomock.Controller) []*jobmocks.MockKeyValueEntry
		notifyErr    error
		validateFunc func(events []notify.ConditionEvent)
	}{
		{
			name: "an expired agent is named by the hostname of its last heartbeat",
			steps: func(ctrl *gomock.Controller) []*jobmocks.MockKeyValueEntry {
				return []*jobmocks.MockKeyValueEntry{put(ctrl, "web-01"), expire(ctrl)}
			},
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("web-01", events[0].Hostname)
				s.Equal(notify.ConditionComponentUnreachable, events[0].Condition)
				s.True(events[0].Active)
			},
		},
		{
			name: "an agent that expires before any heartbeat is named by its machine ID",
			steps: func(ctrl *gomock.Controller) []*jobmocks.MockKeyValueEntry {
				return []*jobmocks.MockKeyValueEntry{expire(ctrl)}
			},
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 1)
				s.Equal("m-1", events[0].Hostname)
			},
		},
		{
			name: "the next heartbeat resolves the unreachable condition, once",
			steps: func(ctrl *gomock.Controller) []*jobmocks.MockKeyValueEntry {
				return []*jobmocks.MockKeyValueEntry{
					put(ctrl, "web-01"), expire(ctrl), put(ctrl, "web-01"), put(ctrl, "web-01"),
				}
			},
			validateFunc: func(events []notify.ConditionEvent) {
				s.Require().Len(events, 2)
				s.True(events[0].Active)
				s.Equal(notify.ConditionComponentUnreachable, events[1].Condition)
				s.False(events[1].Active)
				s.Equal("heartbeat resumed", events[1].Reason)
			},
		},
		{
			name: "a notifier that fails is logged, and the watcher goes on",
			steps: func(ctrl *gomock.Controller) []*jobmocks.MockKeyValueEntry {
				return []*jobmocks.MockKeyValueEntry{expire(ctrl), put(ctrl, "web-01")}
			},
			notifyErr: errors.New("webhook down"),
			validateFunc: func(events []notify.ConditionEvent) {
				s.Len(events, 2)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			var events []notify.ConditionEvent
			notifier := notifymocks.NewMockNotifier(ctrl)
			notifier.EXPECT().
				Notify(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, event notify.ConditionEvent) error {
					events = append(events, event)

					return tt.notifyErr
				}).
				AnyTimes()
			w := notify.NewWatcher(nil, "registry", notifier, slog.Default(), 0)
			for _, entry := range tt.steps(ctrl) {
				notify.WatcherHandleEntry(context.Background(), w, entry)
			}
			tt.validateFunc(events)
		})
	}
}
