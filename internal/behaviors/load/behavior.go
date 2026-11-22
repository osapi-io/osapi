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

// Package load is the load behavior: the load averages of the host.
package load

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the load definition.
var Behavior = behavior.Register(behavior.Define(
	"load", "Get the load averages of target hosts",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaNode),
	behavior.CLI,
	// gohai reads /proc/loadavg on Linux, and vm.loadavg from sysctl on macOS.
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "sysctl"),
	behavior.Read("get", Get).
		Describe("Get the load averages of the host").
		Explain("Reads the averages of the run queue over 1, 5 and 15 minutes.").
		Example("A host under a light load", GetInput{},
			Entry{LoadAverage: Averages{One: 0.32, Five: 0.28, Fifteen: 0.25}},
			behavior.Exists("/proc/loadavg", "0.32 0.28 0.25 1/412 12345\n"),
			behavior.Prints("sysctl -n vm.loadavg", "{ 0.32 0.28 0.25 }\n"),
		),
))
