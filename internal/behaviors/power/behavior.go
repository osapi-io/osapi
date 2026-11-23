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

// Package power is the power behavior: reboot or shut down the host. Each
// method is an action with the execute permission, power:execute, so the
// operator role has it and the automation role does not. The host goes
// down, so no example runs on a live host.
package power

import (
	"context"

	"github.com/osapi-io/osapi/internal/behavior"
)

// Behavior is the power definition.
var Behavior = behavior.Register(behavior.Define(
	"power", "Reboot or shut down the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaNode),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	action("reboot", Reboot, "Reboot the host", "-r"),
	action("shutdown", Shutdown, "Shut the host down", "-h"),
))

// action is reboot or shutdown: shutdown with flag, after the delay in whole
// minutes, at least one, so the agent can answer before the host goes down.
func action(
	name string,
	fn func(context.Context, Provider, Input) (Result, error),
	describe string,
	flag string,
) *behavior.Method {
	return behavior.Action(name, fn).
		Sudo("shutdown").
		Always().
		Disrupts("the host goes down, and the test with it").
		Describe(describe).
		Explain("Runs shutdown "+flag+" -- +N, with N the delay in whole minutes and at "+
			"least one, so the agent can answer before the host goes down. A message "+
			"goes to the logged-in users. Each run changes the host.").
		Example("In one minute", Input{}, Result{Action: name, Delay: 60, Changed: true},
			behavior.Ran("shutdown", flag, "--", "+1"),
		).
		Example("In five minutes, with a message", Input{Delay: 300, Message: "maintenance"},
			Result{Action: name, Delay: 300, Changed: true},
			behavior.Ran("shutdown", flag, "--", "+5", "maintenance"),
		).
		Fails("A negative delay", Input{Delay: -1}, "Delay")
}
