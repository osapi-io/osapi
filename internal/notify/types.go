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

// Package notify provides a pluggable condition notification system that
// watches the registry KV bucket for component condition transitions and
// dispatches events via configurable notifiers.
package notify

import (
	"context"
	"time"
)

// ConditionEvent represents a condition state transition on a component.
type ConditionEvent struct {
	// ComponentType is "agent", "api", or "nats".
	ComponentType string
	// Hostname identifies the component.
	Hostname string
	// Condition is the condition type (e.g., "MemoryPressure").
	Condition string
	// Active is true when the condition fires, false when resolved.
	Active bool
	// Reason describes why the condition triggered or resolved.
	Reason string
	// Timestamp is when the transition occurred.
	Timestamp time.Time
}

// Notifier sends notifications when component conditions change.
type Notifier interface {
	Notify(ctx context.Context, event ConditionEvent) error
}
