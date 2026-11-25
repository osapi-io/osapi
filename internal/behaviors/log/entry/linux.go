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

package entry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// linux reads the journal with journalctl. The journal is what journald
// holds now, not a fact of the host, so it is not gohai's.
type linux struct {
	exec host.Exec
}

// newLinux is the provider on each Linux host. A container has no journald
// of its own, so its methods report skipped there, as the hand-written log
// domain did.
func newLinux(
	h host.Host,
) (Provider, error) {
	if h.Container {
		return nil, behavior.ErrUnsupported
	}

	return &linux{exec: h.Exec}, nil
}

// Entries runs journalctl --output=json with the filters of opts.
func (l *linux) Entries(
	ctx context.Context,
	opts ListInput,
) ([]Entry, error) {
	out, err := l.exec.RunCmd(ctx, "journalctl", args(opts))
	if err != nil {
		return nil, fmt.Errorf("read the journal: %w", err)
	}

	return parse(out), nil
}

// args is journalctl's arguments for opts: the last 100 entries unless
// Lines says.
func args(
	opts ListInput,
) []string {
	lines := opts.Lines
	if lines <= 0 {
		lines = 100
	}
	out := []string{"--output=json"}
	// Each value is in the same argument as its option, so a value that
	// starts with - is never read as an option.
	if opts.Unit != "" {
		out = append(out, "--unit="+opts.Unit)
	}
	if opts.Since != "" {
		out = append(out, "--since="+opts.Since)
	}
	if opts.Priority != "" {
		out = append(out, "--priority="+string(opts.Priority))
	}

	return append(out, "-n", strconv.Itoa(lines))
}

// journalEntry is one line of journalctl --output=json.
type journalEntry struct {
	Timestamp string `json:"__REALTIME_TIMESTAMP"`
	Unit      string `json:"SYSLOG_IDENTIFIER"`
	Priority  string `json:"PRIORITY"`
	Message   string `json:"MESSAGE"`
	PID       string `json:"_PID"`
}

// parse reads journalctl --output=json, one entry a line. A line that is
// not an entry is left out.
func parse(
	out string,
) []Entry {
	entries := []Entry{}
	for _, line := range strings.Split(out, "\n") {
		var je journalEntry
		if json.Unmarshal([]byte(strings.TrimSpace(line)), &je) != nil {
			continue
		}
		// journald gives the number of a syslog priority. An entry with no
		// PRIORITY, or one that is not 0 to 7, has none: the name is one
		// of journald's or nothing (osapi-io/osapi#1434).
		var priority Priority
		if n, err := strconv.Atoi(je.Priority); err == nil && n >= 0 && n < len(Priorities()) {
			priority = Priorities()[n]
		}
		pid, _ := strconv.Atoi(je.PID)
		entries = append(entries, Entry{
			Timestamp: timestamp(je.Timestamp),
			Unit:      je.Unit,
			Priority:  priority,
			Message:   je.Message,
			PID:       pid,
		})
	}

	return entries
}

// timestamp is journald's time in microseconds as RFC 3339, or the time as
// journald gave it when it is not a number.
func timestamp(
	usec string,
) string {
	micros, err := strconv.ParseInt(usec, 10, 64)
	if err != nil {
		return usec
	}

	return time.UnixMicro(micros).UTC().Format(time.RFC3339Nano)
}
