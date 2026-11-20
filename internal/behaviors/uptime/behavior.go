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

// Package uptime is the uptime behavior: how long the host has run since it
// booted.
package uptime

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the uptime definition. The routes, the agent's dispatch, the
// CLI commands, the spec, the docs pages and the tests of every layer above
// the provider are built from it. osapi behavior spec get lists every option.
var Behavior = behavior.Register(behavior.Define(
	"uptime", "Get the uptime of target hosts",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	// Linux reads /proc/uptime; macOS reads its boot time from sysctl.
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Platform(behavior.PlatformDarwin, newDarwin),
	behavior.Runs("sysctl"),
	behavior.Read("get", Get).
		Describe("Get the uptime of the host").
		Explain("Reads how long the host has run since it booted: /proc/uptime on Linux, "+
			"and the boot time that sysctl gives on macOS. The uptime is in days, hours "+
			"and minutes, and in seconds.").
		Example("A host up for a day, two hours and five minutes", GetInput{},
			Entry{Uptime: "1 day, 2 hours, 5 minutes", Seconds: 93900},
			behavior.Exists("/proc/uptime", "93900.42 180001.15\n"),
			behavior.Prints("sysctl -n kern.boottime",
				"{ sec = 1767229145, usec = 0 } Thu Jan  1 00:59:05 2026"),
		),
))
