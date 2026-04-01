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

package log

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// journalEntry represents the raw JSON structure of a journalctl --output=json line.
type journalEntry struct {
	Timestamp string `json:"__REALTIME_TIMESTAMP"`
	Unit      string `json:"SYSLOG_IDENTIFIER"`
	Priority  string `json:"PRIORITY"`
	Message   string `json:"MESSAGE"`
	PID       string `json:"_PID"`
	Hostname  string `json:"_HOSTNAME"`
}

// priorityNames maps journald priority numbers to human-readable names.
var priorityNames = map[string]string{
	"0": "emerg",
	"1": "alert",
	"2": "crit",
	"3": "err",
	"4": "warning",
	"5": "notice",
	"6": "info",
	"7": "debug",
}

// buildArgs constructs journalctl arguments for a general query.
// Always includes --output=json. Defaults to 100 lines if opts.Lines <= 0.
func buildArgs(
	opts QueryOpts,
) []string {
	lines := opts.Lines
	if lines <= 0 {
		lines = 100
	}

	args := []string{"--output=json"}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	if opts.Priority != "" {
		args = append(args, "--priority", opts.Priority)
	}

	args = append(args, "-n", strconv.Itoa(lines))

	return args
}

// buildUnitArgs constructs journalctl arguments for a unit-specific query.
// Adds -u <unit> before the line count argument.
func buildUnitArgs(
	unit string,
	opts QueryOpts,
) []string {
	lines := opts.Lines
	if lines <= 0 {
		lines = 100
	}

	args := []string{"--output=json", "-u", unit}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	if opts.Priority != "" {
		args = append(args, "--priority", opts.Priority)
	}

	args = append(args, "-n", strconv.Itoa(lines))

	return args
}

// parseJournalOutput parses newline-delimited JSON output from journalctl.
// Malformed or empty lines are skipped with a debug log entry.
func parseJournalOutput(
	output string,
	logger *slog.Logger,
) []Entry {
	lines := strings.Split(output, "\n")
	entries := make([]Entry, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var je journalEntry
		if err := json.Unmarshal([]byte(line), &je); err != nil {
			logger.Debug(
				"skipping malformed journal line",
				slog.String("error", err.Error()),
			)

			continue
		}

		entries = append(entries, journalEntryToEntry(je))
	}

	return entries
}

// journalEntryToEntry converts a raw journalEntry to an Entry.
func journalEntryToEntry(
	je journalEntry,
) Entry {
	ts := parseTimestamp(je.Timestamp)

	priority := je.Priority
	if name, ok := priorityNames[je.Priority]; ok {
		priority = name
	}

	pid := 0

	if je.PID != "" {
		if p, err := strconv.Atoi(je.PID); err == nil {
			pid = p
		}
	}

	return Entry{
		Timestamp: ts,
		Unit:      je.Unit,
		Priority:  priority,
		Message:   je.Message,
		PID:       pid,
		Hostname:  je.Hostname,
	}
}

// parseTimestamp converts a journald microsecond timestamp string to RFC3339Nano.
// Returns the original string if parsing fails.
func parseTimestamp(
	usec string,
) string {
	if usec == "" {
		return ""
	}

	micros, err := strconv.ParseInt(usec, 10, 64)
	if err != nil {
		return usec
	}

	return time.UnixMicro(micros).UTC().Format(time.RFC3339Nano)
}
