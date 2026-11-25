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

// Package entry is the log behavior: the entries of the systemd journal of
// the host, read with journalctl on each Linux host. The Go package keeps the
// name entry, so it does not hide the standard library's log.
package entry

import (
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The journal of the examples: two entries, as journalctl --output=json
// prints them, one line each.
const (
	nginx = `{"__REALTIME_TIMESTAMP":"1711929045123456","SYSLOG_IDENTIFIER":"nginx",` +
		`"PRIORITY":"6","MESSAGE":"Started nginx","_PID":"1234","_HOSTNAME":"web-01"}`
	sshd = `{"__REALTIME_TIMESTAMP":"1711929046000000","SYSLOG_IDENTIFIER":"sshd",` +
		`"PRIORITY":"5","MESSAGE":"Accepted key","_PID":"5678","_HOSTNAME":"web-01"}`
)

var (
	nginxEntry = Entry{
		Timestamp: "2024-03-31T23:50:45.123456Z", Unit: "nginx", Priority: "info",
		Message: "Started nginx", PID: 1234,
	}
	sshdEntry = Entry{
		Timestamp: "2024-03-31T23:50:46Z", Unit: "sshd", Priority: "notice",
		Message: "Accepted key", PID: 5678,
	}
)

// Behavior is the entries of the systemd journal. They are what journald
// holds now, not a fact of the host, so journalctl reads them, not gohai. It
// runs on each Linux host; macOS has no journal, and reports skipped.
var Behavior = behavior.Register(behavior.Define(
	"log", "The entries of the systemd journal",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Logs),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Read("list", List).
		NotAFact("the entries are what journald holds now, not a fact of the host").
		Runs("journalctl").
		Describe("List the entries of the journal").
		Explain("Runs journalctl --output=json, the newest entries last: the last "+
			"100 unless lines says, of one unit when unit says, since a time when "+
			"since says, and at a priority or above when priority says. A priority "+
			"is its name, such as err or info. The hand-written log domain read the "+
			"Debian family only; this reads each Linux host that has journald.").
		Example("The newest entries", ListInput{}, []Entry{nginxEntry, sshdEntry},
			behavior.Prints("journalctl --output=json -n 100", nginx+"\n"+sshd+"\n"),
		).
		Example("The last entry of one unit at info or above",
			ListInput{Unit: "nginx.service", Lines: 1, Since: "1 hour ago", Priority: "info"},
			[]Entry{nginxEntry},
			behavior.Prints(
				"journalctl --output=json --unit=nginx.service --since=1 hour ago --priority=info -n 1",
				nginx+"\n",
			),
		).
		Fails("Lines below one", ListInput{Lines: -1}, "min").
		Fails("More than 10000 lines", ListInput{Lines: 10001}, "max").
		Fails("A priority that journald does not have", ListInput{Priority: "loud"}, "enum").
		Fails("A unit with a space in it", ListInput{Unit: "nginx service"}, "unit_name"),
))
