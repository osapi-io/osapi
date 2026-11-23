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

package status

import (
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/disk"
	"github.com/osapi-io/osapi/internal/behaviors/hostname"
	"github.com/osapi-io/osapi/internal/behaviors/load"
	"github.com/osapi-io/osapi/internal/behaviors/memory"
	osb "github.com/osapi-io/osapi/internal/behaviors/os"
	"github.com/osapi-io/osapi/internal/behaviors/uptime"
)

// The files and commands that each example's host gives the six reads.
var (
	shared = []behavior.HostOption{
		behavior.Prints("hostname", "web-01\n"),
		behavior.Prints("hostname -s", "web-01\n"),
		behavior.Prints("sysctl -n kern.boottime",
			"{ sec = 1767229145, usec = 0 } Thu Jan  1 00:59:05 2026"),
		behavior.Prints("sysctl -n vm.loadavg", "{ 0.32 0.28 0.25 }\n"),
		behavior.Prints("sysctl -n hw.memsize", "8589934592\n"),
		behavior.Prints("vm_stat", "Mach Virtual Memory Statistics: (page size of 4096 bytes)\n"+
			"Pages free:                              524288.\n"+
			"Pages active:                            786432.\n"+
			"Pages inactive:                          524288.\n"),
		behavior.Prints("mount", "/dev/disk3s1 on / (apfs, local, journaled)\n"),
		behavior.Prints("df -ki",
			"Filesystem 1024-blocks Used Available Capacity iused ifree %iused Mounted on\n"+
				"/dev/disk3s1 51200000 20480000 30720000 40% 100 200 1% /\n"),
		behavior.Exists("/proc/uptime", "93900.42 180001.15\n"),
		behavior.Exists("/proc/stat", "btime 1767229145\n"),
		behavior.Exists("/proc/loadavg", "0.32 0.28 0.25 1/412 12345\n"),
		behavior.Exists(
			"/proc/meminfo",
			"MemTotal:        8388608 kB\nMemFree:         2097152 kB\nMemAvailable:    4194304 kB\n",
		),
		behavior.Exists("/proc/mounts", "/dev/sda1 / ext4 rw,relatime 0 0\n"),
		behavior.Prints("df -akP",
			"Filesystem 1024-blocks Used Available Capacity Mounted on\n"+
				"/dev/sda1 51200000 20480000 30720000 40% /\n"),
	}
	linux = append([]behavior.HostOption{
		behavior.OnPlatform(behavior.PlatformLinux),
		behavior.Exists("/etc/os-release", "ID=ubuntu\nVERSION_ID=\"24.04\"\n"),
	}, shared...)
	darwin = append([]behavior.HostOption{
		behavior.OnPlatform(behavior.PlatformDarwin),
		behavior.Prints("sw_vers", "ProductName:\t\tmacOS\nProductVersion:\t\t15.5\n"),
	}, shared...)
)

// summary is the status of the example host, on the platform that names
// its operating system.
func summary(
	release osb.Release,
) Entry {
	return Entry{
		Name:        "web-01",
		Uptime:      uptime.Entry{Uptime: "1 day, 2 hours, 5 minutes", Seconds: 93900},
		OSInfo:      release,
		LoadAverage: load.Averages{One: 0.32, Five: 0.28, Fifteen: 0.25},
		Memory:      memory.Usage{Total: 8589934592, Free: 2147483648, Used: 4294967296},
		Disks: []disk.Disk{
			{Name: "/", Total: 52428800000, Used: 20971520000, Free: 31457280000},
		},
	}
}

// Behavior is the status summary of a node: the hostname, uptime, os,
// load, memory and disk reads in one call.
var Behavior = behavior.Register(behavior.Define(
	"status", "The status summary of a node",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Node),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Requires(hostname.Behavior, uptime.Behavior, osb.Behavior,
		load.Behavior, memory.Behavior, disk.Behavior),
	behavior.Read("get", Get).
		Describe("Get the status summary of the host").
		Explain("Runs the get of the hostname, uptime, os, load, memory and disk "+
			"behaviors on the host, and answers with each. A read that fails is in "+
			"errors, as its name and its error, and the other fields still answer.").
		Example("An Ubuntu host", GetInput{},
			summary(osb.Release{Distribution: "ubuntu", Version: "24.04"}), linux...).
		Example("A Mac", GetInput{},
			summary(osb.Release{Distribution: "darwin", Version: "15.5"}), darwin...),
))
