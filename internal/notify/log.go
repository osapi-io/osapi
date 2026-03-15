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
)

// LogNotifier logs condition events at INFO level.
type LogNotifier struct {
	logger *slog.Logger
}

// NewLogNotifier creates a new LogNotifier that logs condition events using
// the provided logger.
func NewLogNotifier(
	logger *slog.Logger,
) *LogNotifier {
	return &LogNotifier{logger: logger}
}

// Notify logs the condition event at INFO level and returns nil.
func (n *LogNotifier) Notify(
	_ context.Context,
	event ConditionEvent,
) error {
	if event.Active {
		n.logger.Warn(
			"condition fired",
			slog.String("component", event.ComponentType),
			slog.String("hostname", event.Hostname),
			slog.String("condition", event.Condition),
			slog.Bool("active", event.Active),
			slog.String("reason", event.Reason),
		)

		return nil
	}

	n.logger.Info(
		"condition resolved",
		slog.String("component", event.ComponentType),
		slog.String("hostname", event.Hostname),
		slog.String("condition", event.Condition),
		slog.Bool("active", event.Active),
		slog.String("reason", event.Reason),
	)

	return nil
}
