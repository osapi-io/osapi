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
// from the telemetry backends, such as the spans of a trace.
package telemetry

import (
	"errors"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

var (
	// ErrNotFound is a trace the backend does not have, or not yet.
	ErrNotFound = errors.New("trace not found")
	// ErrNoBackend is a controller with no tracing backend configured.
	ErrNoBackend = errors.New("no tracing backend is configured")
)

// Group is the telemetry group's definition.
var Group = endpoint.Register(endpoint.Define(
	"telemetry", "Read telemetry through the controller",
	endpoint.Area(area.Fleet),
	endpoint.SubDoc("trace", "Read a trace from the tracing backend"),
	endpoint.Title("Telemetry API"),
	endpoint.Tag("telemetry", "Telemetry",
		"Telemetry through the controller, such as the spans of a trace."),
	endpoint.Uses[Deps](),
	endpoint.Get[TraceInput, TraceResponse]("get").
		Sub("trace").
		Handle(Get).
		ID("GetTrace").
		Permission(types.PermTraceRead).
		Describe("Get the spans of a trace").
		Explain("Returns each span of a trace, read from the tracing backend at "+
			"telemetry.tracing.url. The backend gets the spans a few seconds after the "+
			"request, so a trace of a request that just ran can be not found yet. A token with a "+
			"target scope gets a trace only when each target that the trace records is in its scope.").
		Returns("The spans of the trace.").
		Command("Show the span tree of a trace",
			"Shows each span of a trace as a tree: the request on the controller,\n"+
				"the job, each agent that ran it with its host and status, and the\n"+
				"commands and file changes under each. The NATS calls, which record the\n"+
				"job's progress, are left out and counted; --all shows them. Each\n"+
				"command prints the ID of its trace after its job ID.\n\n"+
				"The controller reads the trace from the tracing backend at\n"+
				"`telemetry.tracing.url`, so this host needs no access to the backend.\n"+
				"The backend gets the spans a few seconds after the request: a trace\n"+
				"that is not found yet is there after a few seconds.").
		Example("--id 26e448c9dbe1c4a8d216a98c675079c9").
		Prints(sampleTrace).
		Example("--id 26e448c9dbe1c4a8d216a98c675079c9 --all").
		Example("--id 26e448c9dbe1c4a8d216a98c675079c9 --json").
		Errors(
			endpoint.NotFound(ErrNotFound,
				"The backend has no trace with this ID, or not yet, or a target of the trace is "+
					"outside the token's target scope."),
			endpoint.Unavailable(ErrNoBackend,
				"No tracing backend is configured on the controller."),
		),
))
