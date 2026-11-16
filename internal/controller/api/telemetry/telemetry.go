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

// Package telemetry provides the telemetry API handlers, such as the one that
// returns the spans of a trace.
package telemetry

import (
	"log/slog"

	"github.com/osapi-io/osapi/internal/controller/api/telemetry/gen"
)

// ensure that we've conformed to the `StrictServerInterface` with a compile-time check
var _ gen.StrictServerInterface = (*Telemetry)(nil)

// New returns the telemetry API on reader, which is nil when the controller
// has no tracing backend.
func New(
	logger *slog.Logger,
	reader TraceReader,
) *Telemetry {
	return &Telemetry{
		reader: reader,
		logger: logger.With(slog.String("subsystem", "controller.telemetry")),
	}
}
