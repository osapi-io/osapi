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
// liveness and readiness probes, and the status of each component. Until
// the group moves, the controller serves the hand-written spec in
// internal/controller/api/health, and TestEndpointParity holds the two
// together.
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
	endpoint.Get[LivenessInput, HealthResponse]("liveness").
		ID("GetHealth").
		Public().
		Describe("Liveness probe").
		Explain("Returns 200 if the process is alive. No checks performed.").
		Returns("Process is alive."),
	endpoint.Get[ReadyInput, ReadyResponse]("ready").
		Sub("ready").
		ID("GetHealthReady").
		Public().
		Describe("Readiness probe").
		Explain("Returns 200 when the service is ready to accept traffic.").
		Returns("Service is ready.").
		Answers(http.StatusServiceUnavailable, "Service is not ready."),
	endpoint.Get[StatusInput, StatusResponse]("status").
		Sub("status").
		ID("GetHealthStatus").
		Permission("health:read").
		Describe("System status and component health").
		Explain("Returns per-component health status with system metrics. "+
			"Requires authentication.").
		Returns("All components healthy.").
		Answers(http.StatusServiceUnavailable, "One or more components unhealthy."),
))
