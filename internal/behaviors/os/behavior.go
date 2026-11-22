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
	behavior.Area(behavior.AreaNode),
	behavior.CLI,
	// gohai reads /etc/os-release on Linux, and sw_vers on macOS.
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "sw_vers"),
	behavior.Read("get", Get).
		Describe("Get the operating system of the host").
		Explain("Reads the distribution and its version: the ID and VERSION_ID of "+
			"/etc/os-release on Linux, and darwin with the ProductVersion of sw_vers on "+
			"macOS, such as darwin 15.5.").
		Example("An Ubuntu host", GetInput{},
			Entry{OSInfo: Release{Distribution: "ubuntu", Version: "24.04"}},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Exists("/etc/os-release", "NAME=\"Ubuntu\"\nID=ubuntu\nVERSION_ID=\"24.04\"\n"),
		).
		Example("A Mac", GetInput{},
			Entry{OSInfo: Release{Distribution: "darwin", Version: "15.5"}},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints("sw_vers", "ProductName:\t\tmacOS\nProductVersion:\t\t15.5\nBuildVersion:\t\t24F74\n"),
		),
))
