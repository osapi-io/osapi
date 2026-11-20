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

// Package hostname is the hostname behavior: the name of the host, and on
// Debian, setting it.
package hostname

import (
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
)

// Behavior is the hostname definition.
var Behavior = behavior.Register(behavior.Define(
	"hostname", "Get or set the hostname of target hosts",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	// Linux reads /proc/sys/kernel/hostname; macOS reads kern.hostname.
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Platform(behavior.PlatformDarwin, newDarwin),
	behavior.Runs("sysctl"),
	behavior.Read("get", Get).
		Describe("Get the hostname of the host").
		Explain("Reads the kernel's hostname: /proc/sys/kernel/hostname on Linux, and "+
			"kern.hostname on macOS.").
		Example("A host named web-01", GetInput{}, Entry{Name: "web-01"},
			behavior.Exists("/proc/sys/kernel/hostname", "web-01\n"),
			behavior.Prints("sysctl -n kern.hostname", "web-01\n"),
		),
	behavior.Write("update", Update).
		On(behavior.PlatformDebian).
		Sudo("hostnamectl").
		Describe("Set the hostname of the host").
		Explain("Sets the static hostname with hostnamectl set-hostname, which keeps it "+
			"after a reboot. A host that already has the name does not change and "+
			"reports changed: false. It runs on the Debian family only, and elsewhere "+
			"reports skipped. In a container the runtime owns the hostname, so there "+
			"it also reports skipped.").
		// Each example names its host with After, as a live host can run it:
		// the integration test runs these on CI's runner and puts its name
		// back. Prints and PrintsAfter are the fake host's hostnamectl.
		Example("Rename a host", UpdateInput{Name: "web-02"},
			Result{Name: "web-02", Changed: true},
			append(
				named("web-01"),
				behavior.PrintsAfter(
					"hostnamectl set-hostname web-02",
					"hostnamectl hostname",
					"web-02\n",
				),
				behavior.Ran("hostnamectl", "set-hostname", "web-02"),
			)...,
		).
		Example("A host that already has the name", UpdateInput{Name: "web-01"},
			Result{Name: "web-01"},
			named("web-01")...,
		).
		Fails("A name that is not a hostname", UpdateInput{Name: "web 01;reboot"},
			"invalid hostname").
		Fails("A name longer than 253 characters", UpdateInput{Name: strings.Repeat("a", 254)},
			"invalid hostname").
		Fails("No name", UpdateInput{}, "invalid hostname"),
))

// named is a host that an earlier update named name: hostnamectl prints
// localhost until set-hostname name runs, and name after it.
func named(
	name string,
) []behavior.HostOption {
	return []behavior.HostOption{
		behavior.After(UpdateInput{Name: name}),
		behavior.Prints("hostnamectl hostname", "localhost\n"),
		behavior.PrintsAfter("hostnamectl set-hostname "+name, "hostnamectl hostname", name+"\n"),
	}
}
