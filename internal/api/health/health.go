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

// Package health provides health check API handlers.
package health

import (
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/api/health/gen"
)

// ensure that we've conformed to the `StrictServerInterface` with a compile-time check
var _ gen.StrictServerInterface = (*Health)(nil)

// New factory to create a new instance.
func New(
	logger *slog.Logger,
	checker Checker,
	startTime time.Time,
	version string,
	metrics MetricsProvider,
) *Health {
	return &Health{
		Checker:   checker,
		StartTime: startTime,
		Version:   version,
		Metrics:   metrics,
		logger:    logger,
	}
}
