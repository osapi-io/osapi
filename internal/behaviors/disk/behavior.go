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

// Package disk is the disk behavior: the usage of the host's local
// filesystems.
package disk

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the disk definition.
var Behavior = behavior.Register(behavior.Define(
	"disk", "Get the disk usage of target hosts",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Platform(behavior.PlatformDarwin, newDarwin),
	behavior.Runs("df"),
	behavior.Read("get", Get).
		Describe("Get the disk usage of the host").
		Explain("Reads the size, the used space and the free space of each local "+
			"filesystem, in bytes, from df: ext4, xfs and btrfs on Linux, and apfs "+
			"and hfs on macOS. Network and special filesystems are left out.").
		Example("A host with one local filesystem", GetInput{},
			Entry{
				Disks: []Disk{
					{Name: "/", Total: 52428800000, Used: 20971520000, Free: 31457280000},
				},
			},
			behavior.Prints("df -kP -t ext4 -t xfs -t btrfs",
				"Filesystem 1024-blocks Used Available Capacity Mounted on\n"+
					"/dev/sda1 51200000 20480000 30720000 40% /\n"),
			behavior.Prints("df -kP -T apfs,hfs",
				"Filesystem 1024-blocks Used Available Capacity Mounted on\n"+
					"/dev/disk3s1 51200000 20480000 30720000 40% /\n"),
		),
))
