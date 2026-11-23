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

// Package source is the source behavior: the syslog identifiers that wrote to
// the systemd journal of the host, read with journalctl on each Linux host.
package source

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the sources of the systemd journal: each syslog identifier
// that has an entry. They are what journald holds now, not a fact of the
// host, so journalctl reads them, not gohai. It runs on each Linux host;
// macOS has no journal, and reports skipped.
var Behavior = behavior.Register(behavior.Define(
	"source", "The sources that write to the systemd journal",
	behavior.Node,
	behavior.Under(behavior.GroupLog),
	behavior.API,
	behavior.Area(behavior.AreaNode),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newLinux),
	behavior.Read("list", List).
		NotAFact("the sources are what journald holds now, not a fact of the host").
		Runs("journalctl").
		Describe("List the sources of the journal").
		Explain("Runs journalctl --field=SYSLOG_IDENTIFIER: each syslog identifier "+
			"that has an entry, once, sorted. A name here is what log entry list "+
			"takes as unit when the source is a systemd unit.").
		Example("The sources", ListInput{}, []Source{{Name: "cron"}, {Name: "nginx"}, {Name: "sshd"}},
			behavior.Prints("journalctl --field=SYSLOG_IDENTIFIER", "sshd\nnginx\ncron\n\n"),
		),
))
