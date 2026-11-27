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

package exec

import "context"

// sudoProgram is the program that RunPrivilegedCmdFull and
// RunPrivilegedCmdInteractive run a command through; a test replaces it.
var sudoProgram = "sudo"

// RunPrivilegedCmdFull runs name with args as RunPrivilegedCmd does, through
// sudo when privilege escalation is on, and gives the standard output, the
// standard error and the exit code apart. An exit code other than 0 is not
// an error.
//
// stdin, when it is not empty, goes to the command's standard input: how a
// command gets a secret, which must never be in its arguments.
func (e *Exec) RunPrivilegedCmdFull(
	ctx context.Context,
	name string,
	args []string,
	stdin string,
) (string, string, int, error) {
	if e.sudo {
		args = sudoArgs(name, args)
		name = sudoProgram
	}
	// runFull gives a result also with an error.
	result, err := e.runFull(ctx, name, args, "", 0, stdin)

	return result.Stdout, result.Stderr, result.ExitCode, err
}
