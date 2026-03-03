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

package cli

import (
	"fmt"
	"io"
	"strings"
)

// RawResult holds raw command output for a single host.
type RawResult struct {
	Hostname string
	Stdout   string
	Stderr   string
	ExitCode int
}

// PrintRawOutput writes raw command output to the given writers.
// For single results, output is printed without hostname prefix.
// For multiple results, each line is prefixed with a styled hostname.
// showStdout/showStderr control which streams are printed.
func PrintRawOutput(
	stdout io.Writer,
	stderr io.Writer,
	results []RawResult,
	showStdout bool,
	showStderr bool,
) {
	printRaw(stdout, stderr, results, showStdout, showStderr, true)
}

// PrintRawOutputPlain writes raw output without lipgloss styling.
// Used for testing and non-TTY output.
func PrintRawOutputPlain(
	stdout io.Writer,
	stderr io.Writer,
	results []RawResult,
	showStdout bool,
	showStderr bool,
) {
	printRaw(stdout, stderr, results, showStdout, showStderr, false)
}

// MaxExitCode returns the highest exit code from a slice of RawResults.
func MaxExitCode(
	results []RawResult,
) int {
	max := 0
	for _, r := range results {
		if r.ExitCode > max {
			max = r.ExitCode
		}
	}
	return max
}

func printRaw(
	stdout io.Writer,
	stderr io.Writer,
	results []RawResult,
	showStdout bool,
	showStderr bool,
	styled bool,
) {
	for _, r := range results {
		if showStdout && r.Stdout != "" {
			writeLines(stdout, r.Hostname, r.Stdout, styled)
		}
		if showStderr && r.Stderr != "" {
			writeLines(stderr, r.Hostname, r.Stderr, styled)
		}
	}
}

func writeLines(
	w io.Writer,
	hostname string,
	content string,
	styled bool,
) {
	lines := strings.Split(content, "\n")

	// strings.Split on newline-terminated content produces a trailing empty
	// string. Trim it so we don't emit a spurious blank line.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for _, line := range lines {
		prefix := "[" + hostname + "]"
		if styled {
			prefix = DimStyle.Render(prefix)
		}
		fmt.Fprintf(w, "%s %s\n", prefix, line)
	}
}
