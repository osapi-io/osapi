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

// Priority is the priority of a journal entry, by the name that journald
// gives it.
type Priority string

// The priorities of journald, from the highest.
const (
	PriorityEmerg   Priority = "emerg"
	PriorityAlert   Priority = "alert"
	PriorityCrit    Priority = "crit"
	PriorityErr     Priority = "err"
	PriorityWarning Priority = "warning"
	PriorityNotice  Priority = "notice"
	PriorityInfo    Priority = "info"
	PriorityDebug   Priority = "debug"
)

// Priorities are the priorities in journald's order: the index of each is
// its number, 0 for emerg to 7 for debug.
func Priorities() []Priority {
	return []Priority{
		PriorityEmerg, PriorityAlert, PriorityCrit, PriorityErr,
		PriorityWarning, PriorityNotice, PriorityInfo, PriorityDebug,
	}
}

// ListInput selects the entries: each field is a query parameter and a flag.
type ListInput struct {
	Unit     string   `json:"unit,omitempty"     validate:"omitempty,unit_name"                                            doc:"Only the entries of this systemd unit, such as nginx.service"`
	Lines    int      `json:"lines,omitempty"    validate:"omitempty,min=1,max=10000"                                      doc:"How many of the newest entries; 100 when it is not set"`
	Since    string   `json:"since,omitempty"    validate:"omitempty"                                                      doc:"Only the entries since this time, in any form that journalctl --since takes, such as 1 hour ago"`
	Priority Priority `json:"priority,omitempty" validate:"omitempty,oneof=emerg alert crit err warning notice info debug" doc:"Only the entries at this priority or above"`
}

// Entry is one entry of the journal.
type Entry struct {
	Timestamp string   `json:"timestamp"      doc:"When journald recorded it, in RFC 3339"`
	Unit      string   `json:"unit,omitempty" doc:"The syslog identifier that wrote it"`
	Priority  Priority `json:"priority"       doc:"Its priority, by name"`
	Message   string   `json:"message"        doc:"The message"`
	PID       int      `json:"pid,omitempty"  doc:"The process that wrote it"`
}
