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

// Package os is the os behavior: the operating system of the host, its
// distribution and version.
package os

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the os definition.
var Behavior = behavior.Register(behavior.Define(
	"os", "Get the operating system of target hosts",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	// Linux reads /etc/os-release; macOS reads its version from sysctl.
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Platform(behavior.PlatformDarwin, newDarwin),
	behavior.Runs("sysctl"),
	behavior.Read("get", Get).
		Describe("Get the operating system of the host").
		Explain("Reads the distribution and its version: the ID and VERSION_ID of "+
			"/etc/os-release on Linux, and the system type and product version that "+
			"sysctl gives on macOS, such as darwin 15.5.").
		Example("An Ubuntu host", GetInput{},
			Entry{OSInfo: Release{Distribution: "ubuntu", Version: "24.04"}},
			behavior.Exists("/etc/os-release", "NAME=\"Ubuntu\"\nID=ubuntu\nVERSION_ID=\"24.04\"\n"),
			behavior.Prints("sysctl -n kern.ostype", "Ubuntu\n"),
			behavior.Prints("sysctl -n kern.osproductversion", "24.04\n"),
		),
))
