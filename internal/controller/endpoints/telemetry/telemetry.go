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
// Package telemetry is the endpoint group for what the controller reads
// from the telemetry backends, such as the spans of a trace. Until the
// group moves, the controller serves the hand-written spec in
// internal/controller/api/telemetry, and TestEndpointParity holds the two
// together.
package telemetry

import (
	"errors"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is a trace the backend does not have, or not yet.
	ErrNotFound = errors.New("trace not found")
	// ErrNoBackend is a controller with no tracing backend configured.
	ErrNoBackend = errors.New("no tracing backend is configured")
)

// Group is the telemetry group's definition.
var Group = endpoint.Register(endpoint.Define(
	"telemetry", "The controller's telemetry",
	endpoint.Title("Telemetry API"),
	endpoint.Tag("telemetry", "Telemetry",
		"Telemetry endpoints, such as the spans of a trace."),
	endpoint.Get[TraceInput, TraceResponse]("trace").
		Sub("trace").
		ID("GetTrace").
		Permission("trace:read").
		Describe("Get the spans of a trace").
		Explain("Returns each span of a trace, read from the tracing backend at "+
			"telemetry.tracing.url. The backend gets the spans a few seconds after the "+
			"request, so a trace of a request that just ran can be not found yet.").
		Returns("The spans of the trace.").
		Errors(
			endpoint.NotFound(ErrNotFound,
				"The backend has no trace with this ID, or not yet."),
			endpoint.Unavailable(ErrNoBackend,
				"No tracing backend is configured on the controller."),
		),
))
