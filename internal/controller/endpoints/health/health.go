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
// Package health is the endpoint group for the controller's health: the
// liveness and readiness probes, and the status of each component.
package health

import (
	"net/http"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// Group is the health group's definition.
var Group = endpoint.Register(endpoint.Define(
	"health", "The controller's health",
	endpoint.Title("Health Check API"),
	endpoint.Tag("health", "Health",
		"Health check endpoints for liveness, readiness, and detailed status."),
	endpoint.Uses[Deps](),
	endpoint.Get[LivenessInput, HealthResponse]("liveness").
		Handle(Liveness).
		ID("GetHealth").
		Public().
		Describe("Liveness probe").
		Explain("Returns 200 if the process is alive. No checks performed.").
		Returns("Process is alive.").
		Command("Liveness probe", "").
		Example("").
		Example("--json"),
	endpoint.Get[ReadyInput, ReadyResponse]("ready").
		Handle(Ready).
		Sub("ready").
		ID("GetHealthReady").
		Public().
		Describe("Readiness probe").
		Explain("Returns 200 when the service is ready to accept traffic.").
		Returns("Service is ready.").
		Answers(http.StatusServiceUnavailable, "Service is not ready.").
		Command("Readiness probe",
			"The command checks that the controller can reach NATS and the KV buckets.").
		Example("").
		Example("--json"),
	endpoint.Get[StatusInput, StatusResponse]("status").
		Handle(Status).
		Sub("status").
		ID("GetHealthStatus").
		Permission("health:read").
		Describe("System status and component health").
		Explain("Returns per-component health status with system metrics. "+
			"Requires authentication.").
		Returns("The status of each component. A component that is not healthy makes the overall status degraded; the request still succeeds.").
		Command("System status and component health",
			"The table has each component that registered: its status, its active\n"+
				"conditions, its uptime, and the CPU and memory of its process.").
		Notes("Each sub-component is under its component, with its address and its URL.\n"+
			"The connectivity rows tell if the controller can reach NATS and the KV\n"+
			"buckets. The output also has the NATS connection, the streams, the KV\n"+
			"buckets, the consumers and the counts of the job queue. The command needs\n"+
			"a token. Use `osapi client agent list` for the labels and the OS of each\n"+
			"agent. When a component is not healthy, the overall status is `degraded`,\n"+
			"and the component shows the error.").
		Example("").
		Example("--json"),
))
