// Copyright (c) 2024 John Dewey

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
	"log/slog"
	"os"
)

// osExit allows tests to override os.Exit.
var osExit = os.Exit

// errOut is where errors for a person are printed. Tests replace it.
var errOut io.Writer = os.Stderr

// jsonOutput is set by --json. Errors then stay structured log lines rather
// than styled text, so a program reading them sees no color or decoration.
var jsonOutput bool

// SetJSONOutput records whether --json is in effect.
func SetJSONOutput(
	enabled bool,
) {
	jsonOutput = enabled
}

// LogFatal logs a fatal error message along with optional structured data
// and then exits the program with a status code of 1.
func LogFatal(
	logger *slog.Logger,
	message string,
	err error,
	kvPairs ...any,
) {
	if !jsonOutput {
		msg := message
		if err != nil {
			msg += ": " + err.Error()
		}
		_, _ = fmt.Fprintln(errOut, Failure(errOut, msg))
		for i := 0; i+1 < len(kvPairs); i += 2 {
			_, _ = fmt.Fprintln(
				errOut,
				Mute(errOut, fmt.Sprintf("  %v: %v", kvPairs[i], kvPairs[i+1])),
			)
		}
		osExit(1)

		return
	}

	if err != nil {
		kvPairs = append(kvPairs, slog.Any("error", err))
	}
	logger.Error(
		message,
		kvPairs...,
	)

	osExit(1)
}
