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

package controller_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type NotifyPublicTestSuite struct {
	suite.Suite
}

func TestNotifyPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NotifyPublicTestSuite))
}

func (s *NotifyPublicTestSuite) TestStartConditionWatcher() {
	tests := []struct {
		name      string
		cfg       config.NotificationsConfig
		registry  bool
		wantWatch bool
	}{
		{name: "when notifications are off starts nothing", registry: true},
		{
			name: "when there is no registry starts nothing",
			cfg:  config.NotificationsConfig{Enabled: true},
		},
		{
			name:      "when the renotify interval is valid starts the watcher",
			cfg:       config.NotificationsConfig{Enabled: true, RenotifyInterval: "1m"},
			registry:  true,
			wantWatch: true,
		},
		{
			name:      "when the renotify interval is not valid starts the watcher without it",
			cfg:       config.NotificationsConfig{Enabled: true, RenotifyInterval: "soon"},
			registry:  true,
			wantWatch: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var registry jetstream.KeyValue
			watched := make(chan struct{})
			if tc.registry {
				kv := jobmocks.NewMockKeyValue(gomock.NewController(s.T()))
				if tc.wantWatch {
					// The watcher stops at once with this error, and
					// background logs it.
					kv.EXPECT().WatchAll(gomock.Any()).DoAndReturn(
						func(context.Context, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error) {
							close(watched)
							return nil, errBoom
						},
					)
				}
				registry = kv
			}

			controller.ExportStartConditionWatcher(ctx, slog.Default(), registry, tc.cfg)

			if tc.wantWatch {
				select {
				case <-watched:
				case <-time.After(5 * time.Second):
					s.Fail("the watcher did not start")
				}
			}
		})
	}
}
