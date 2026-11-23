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

package command

// ExecInput is a program to run, with no shell between it and its
// arguments.
type ExecInput struct {
	Command string   `json:"command"           validate:"required,min=1,startsnotwith=-" doc:"The program, a name on the agent's PATH or a path, such as ls"    telemetry:"attr"`
	Args    []string `json:"args,omitempty"    validate:"omitempty"                      doc:"Its arguments, each one word, such as -la /tmp"`
	Cwd     string   `json:"cwd,omitempty"     validate:"omitempty"                      doc:"The directory to run it in; the agent's own when empty"`
	Timeout int      `json:"timeout,omitempty" validate:"omitempty,min=1,max=300"        doc:"Seconds to let it run before it is stopped, up to 300; 30 when 0"`
}

// ShellInput is a command line to run through /bin/sh -c.
type ShellInput struct {
	Command string `json:"command"           validate:"required,min=1,startsnotwith=-" doc:"The command line, which /bin/sh -c reads, such as ls -la /tmp | grep log" telemetry:"attr"`
	Cwd     string `json:"cwd,omitempty"     validate:"omitempty"                      doc:"The directory to run it in; the agent's own when empty"`
	Timeout int    `json:"timeout,omitempty" validate:"omitempty,min=1,max=300"        doc:"Seconds to let it run before it is stopped, up to 300; 30 when 0"`
}

// Result is what the command did: its output, apart by stream, its exit
// code and how long it ran. A command that exits non-zero is a result, not
// a failure.
type Result struct {
	Stdout     string `json:"stdout"      doc:"What it wrote to standard output"`
	Stderr     string `json:"stderr"      doc:"What it wrote to standard error"`
	ExitCode   int    `json:"exit_code"   doc:"Its exit code"`
	DurationMs int64  `json:"duration_ms" doc:"How long it ran, in milliseconds"`
	Changed    bool   `json:"changed"     doc:"Whether the host changed: true, since a command can change anything"`
}
