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

// Package controllercmd is the controller group: osapi controller start,
// which runs the REST API. The package is not named controller, which is
// the controller itself.
package controllercmd

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the controller group's definition. It is flat, Under controller,
// so its method is osapi controller start.
var Group = command.Register(command.Define(
	"controller", "Run the controller",
	command.Flat(),
	command.Under("controller"),
	command.Uses[Deps](),
	command.Process("start").
		Short("Start the controller").
		Long("The controller serves the REST API, writes its heartbeat, and watches\n"+
			"conditions to send notifications.").
		Example("").
		Example("--controller-api-port 9000").
		Flag("controller.api.port").
		Setup(Start),
))
