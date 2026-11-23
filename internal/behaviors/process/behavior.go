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

// Package process is the process behavior: the processes of the host. It
// reads them through gohai's process collector, from /proc on Linux and
// from ps on macOS, and sends a signal to one with kill.
package process

import "github.com/osapi-io/osapi/internal/behavior"

// The Linux host of the examples booted 1000 seconds before the fake
// host's time. Process 1234, sleep 300 of alice, started 500 seconds after
// the boot and used 25 seconds of CPU, 5 percent, and 20 MiB of the
// host's 2000 MiB, 1 percent.
var linux = []behavior.HostOption{
	behavior.Exists("/proc/stat", "cpu  1 2 3 4\nbtime 1767322045\n"),
	behavior.Exists("/proc/meminfo", "MemTotal:        2048000 kB\n"),
	behavior.Exists(
		"/etc/passwd",
		"root:x:0:0::/root:/bin/sh\nalice:x:1000:1000::/home/alice:/bin/bash\n",
	),
	behavior.Exists("/proc/1234/stat",
		"1234 (sleep) S 1 1234 1234 0 -1 4194304 100 0 0 0 1500 1000 0 0 20 0 1 0 50000 0 0\n"),
	behavior.Exists(
		"/proc/1234/status",
		"Name:\tsleep\nUid:\t1000\t1000\t1000\t1000\nVmRSS:\t   20480 kB\n",
	),
	behavior.Exists("/proc/1234/cmdline", "sleep\x00300\x00"),
	behavior.OnPlatform(behavior.PlatformLinux),
}

var sleeping = Entry{
	PID: 1234, Name: "sleep", User: "alice", State: "sleep",
	CPUPercent: 5, MemPercent: 1, MemRSS: 20971520,
	Command: "sleep 300", StartTime: "2026-01-02T02:55:45Z",
}

// ps is what ps prints on the Mac of the examples, for the selector sel:
// -ax for each process, -p 1234 for one. The examples leave out lstart,
// which ps prints in the host's time zone.
func ps(
	sel string,
) []behavior.HostOption {
	return []behavior.HostOption{
		behavior.Prints("ps "+sel+" -o pid=,ppid=,user=,state=,pcpu=,pmem=,rss=,time=,lstart=",
			"1234     1 alice    S      5.0  1.0  20480   0:25.00\n"),
		behavior.Prints("ps "+sel+" -o pid=,ucomm=", " 1234 sleep\n"),
		behavior.Prints("ps "+sel+" -ww -o pid=,args=", " 1234 sleep 300\n"),
		behavior.OnPlatform(behavior.PlatformDarwin),
	}
}

var sleepingMac = Entry{
	PID: 1234, Name: "sleep", User: "alice", State: "sleep",
	CPUPercent: 5, MemPercent: 1, MemRSS: 20971520, Command: "sleep 300",
}

// both is the host of the signal examples: process 1234 on Linux and on
// macOS.
var both = []behavior.HostOption{
	behavior.Exists("/proc/1234/stat",
		"1234 (sleep) S 1 1234 1234 0 -1 4194304 100 0 0 0 1500 1000 0 0 20 0 1 0 50000 0 0\n"),
	behavior.Prints("ps -p 1234 -o pid=,ppid=,user=,state=,pcpu=,pmem=,rss=,time=,lstart=",
		"1234     1 alice    S      5.0  1.0  20480   0:25.00\n"),
}

// The integration test cannot give a live host a process to signal, since
// no method of this behavior starts one, so the signal example runs on the
// fake host only. The lab tests it by hand: start sleep 300 over ssh, then
// signal it.
var Behavior = behavior.Register(behavior.Define(
	"process", "The processes of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaNode),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newGohai),
	behavior.Platform(behavior.PlatformDarwin, newGohai),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "ps"),
	behavior.Read("list", List).
		Describe("List the processes of the host").
		Explain("Lists each process with its user, state, CPU and memory, its command line "+
			"and when it started, through gohai: from /proc on Linux and from ps on macOS. "+
			"A process that stops while the list is read is not in it.").
		Example("A Linux host that runs sleep", ListInput{}, []Entry{sleeping}, linux...).
		Example("A Mac that runs sleep", ListInput{}, []Entry{sleepingMac}, ps("-ax")...),
	behavior.Read("get", Get).Key("pid").
		Describe("Get one process").
		Explain("Reads one process by its ID, without the others. A process that the host "+
			"does not have is not found.").
		Example("A process of a Linux host", GetInput{PID: 1234}, sleeping, linux...).
		Example("A process of a Mac", GetInput{PID: 1234}, sleepingMac, ps("-p 1234")...).
		Fails("A process that the host does not have", GetInput{PID: 99999}, "not found").
		Fails("A process ID below 1", GetInput{PID: -1}, "min"),
	behavior.Action("signal", Signal).Key("pid").
		Sudo("kill").
		Always().
		Describe("Send a signal to a process").
		Explain("Sends TERM, KILL, HUP, INT, USR1 or USR2 to one process with kill, and "+
			"reports changed: true. A process that the host does not have is not found. "+
			"A process ID below 1 is refused, because kill sends to a group of processes "+
			"for 0 and for a negative ID.").
		Example("Stop a process", SignalInput{PID: 1234, Signal: "TERM"},
			Result{PID: 1234, Signal: "TERM", Changed: true},
			append(both,
				behavior.Ran("kill", "-TERM", "1234"),
				behavior.NotLive("no method of process starts a process for the signal to stop"),
			)...,
		).
		Fails("A process that the host does not have",
			SignalInput{PID: 99999, Signal: "TERM"}, "not found").
		Fails("A process ID below 1", SignalInput{PID: -1, Signal: "TERM"}, "min").
		Fails("A signal that osapi does not send", SignalInput{PID: 1234, Signal: "STOP"}, "oneof"),
))
