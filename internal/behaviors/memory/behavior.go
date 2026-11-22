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

// Package memory is the memory behavior: the memory of the host, in bytes.
package memory

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the memory definition.
var Behavior = behavior.Register(behavior.Define(
	"memory", "Get the memory usage of target hosts",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaHardware),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	// gohai reads /proc/meminfo on Linux, and sysctl and vm_stat on macOS.
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "sysctl", "vm_stat"),
	behavior.Read("get", Get).
		Describe("Get the memory usage of the host").
		Explain("Reads the total, the free and the used memory, in bytes. Used is the "+
			"total less what is available, which includes the cache, as free(1) "+
			"reports it.").
		Example("A host with 8 GiB, 2 GiB of it free", GetInput{},
			Entry{Memory: Usage{Total: 8589934592, Free: 2147483648, Used: 4294967296}},
			behavior.Exists(
				"/proc/meminfo",
				"MemTotal:        8388608 kB\nMemFree:         2097152 kB\nMemAvailable:    4194304 kB\n",
			),
			behavior.Prints("sysctl -n hw.memsize", "8589934592\n"),
			behavior.Prints("vm_stat", "Mach Virtual Memory Statistics: (page size of 4096 bytes)\n"+
				"Pages free:                              524288.\n"+
				"Pages active:                            786432.\n"+
				"Pages inactive:                          524288.\n"),
		),
))
