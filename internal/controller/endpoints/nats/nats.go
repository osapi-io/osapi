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

// Package nats is the nats endpoint group: the NATS server of the fleet.
package nats

import (
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
)

// Group is the nats group's definition. The routes, the spec, the CLI
// commands and their pages come from it. osapi controller endpoint spec
// lists every option.
var Group = endpoint.Register(endpoint.Define(
	"nats", "The NATS server of the fleet",
	endpoint.Title("NATS API"),
	endpoint.Tag("nats", "NATS", "The NATS server of the fleet."),
	endpoint.Uses[Deps](),
	endpoint.Get[StatusInput, health.ComponentList]("status").
		Handle(Status).
		Sub("status").
		ID("GetNATSStatus").
		Describe("Show the health of the NATS server").
		Explain("The NATS server, with each of its sub-components, from the health status of the controller.").
		Returns("Each nats component, with its sub-components.").
		Permission("health:read").
		Command("Show the health of the NATS server", "The NATS server, with each of its sub-components, from the health status of the controller.").
		Example("").
		Example("--json"),
))
