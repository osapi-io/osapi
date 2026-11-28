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
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/notify"
)

type LogNotifierPublicTestSuite struct {
	suite.Suite

	notifier *notify.LogNotifier
}

func (s *LogNotifierPublicTestSuite) SetupTest() {
	s.notifier = notify.NewLogNotifier(slog.Default())
}

func (s *LogNotifierPublicTestSuite) TestNotify() {
	tests := []struct {
		name         string
		event        notify.ConditionEvent
		validateFunc func(error)
	}{
		{
			name: "logs fired event",
			event: notify.ConditionEvent{
				ComponentType: "agent",
				Hostname:      "web-01",
				Condition:     "MemoryPressure",
				Active:        true,
				Reason:        "memory usage above threshold",
				Timestamp:     time.Now(),
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "logs resolved event",
			event: notify.ConditionEvent{
				ComponentType: "agent",
				Hostname:      "web-01",
				Condition:     "MemoryPressure",
				Active:        false,
				Reason:        "memory usage returned to normal",
				Timestamp:     time.Now(),
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(s.notifier.Notify(context.Background(), tt.event))
		})
	}
}

func (s *LogNotifierPublicTestSuite) TestNotifyKeys() {
	tests := []struct {
		name   string
		active bool
	}{
		{
			name:   "fired event names the component of the condition as source",
			active: true,
		},
		{
			name:   "resolved event names the component of the condition as source",
			active: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var out bytes.Buffer
			// The logger of the controller carries component=controller.
			logger := slog.New(slog.NewJSONHandler(&out, nil)).
				With(slog.String("component", "controller"))

			err := notify.NewLogNotifier(logger).Notify(
				context.Background(),
				notify.ConditionEvent{
					ComponentType: "agent",
					Hostname:      "web-01",
					Condition:     "DiskPressure",
					Active:        tt.active,
				},
			)
			s.Require().NoError(err)

			s.Equal(1, strings.Count(out.String(), `"component":`))
			s.Contains(out.String(), `"component":"controller"`)
			s.Contains(out.String(), `"source":"agent"`)
			s.Contains(out.String(), `"subsystem":"controller.notify.log"`)
		})
	}
}

func TestLogNotifierPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LogNotifierPublicTestSuite))
}
