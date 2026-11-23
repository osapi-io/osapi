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

// Package command is the command behavior: run a program, or a command line
// through /bin/sh -c, on the host. Both are admin only, and the agent runs
// them only where agent.command.enabled is true.
package command

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the command definition.
var Behavior = behavior.Register(behavior.Define(
	"command", "Run a command on target hosts",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaCommand),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newRunner),
	behavior.Platform(behavior.PlatformDarwin, newRunner),
	// The program is the caller's: any one, on each platform.
	behavior.Runs(behavior.AnyProgram),
	behavior.Action("exec", Exec).AdminOnly().Always().
		Describe("Run a program with its arguments").
		Explain("Runs the program with its arguments, each one word, with no shell "+
			"between them: a ; or a | in an argument is that character. It answers "+
			"with standard output and standard error apart, the exit code and the "+
			"duration. A program that exits non-zero is a result, not an error. "+
			"Only the admin role may run it, and the agent runs it only where "+
			"agent.command.enabled is true.").
		Example("Print a word", ExecInput{Command: "echo", Args: []string{"hello"}},
			Result{Stdout: "hello\n", Changed: true},
			behavior.Prints("echo hello", "hello\n"),
			behavior.Ran("echo", "hello"),
		).
		Example("A program that exits non-zero", ExecInput{Command: "false"},
			Result{ExitCode: 1, Changed: true},
			behavior.Exits("false", 1),
			behavior.Ran("false"),
		).
		Example("In a directory, for at most 5 seconds", ExecInput{Command: "true", Cwd: "/tmp", Timeout: 5},
			Result{Changed: true},
			behavior.Ran("true"),
		).
		Fails("No program", ExecInput{}, "required").
		Fails("A program that is an option", ExecInput{Command: "-x"}, "startsnotwith").
		Fails("A timeout above 300 seconds", ExecInput{Command: "true", Timeout: 301}, "max").
		Fails("A negative timeout", ExecInput{Command: "true", Timeout: -1}, "min"),
	behavior.Action("shell", Shell).Shell().Always().
		Describe("Run a command line through /bin/sh -c").
		Explain("Runs the command line through /bin/sh -c, so pipes, redirects and "+
			"variables work. It answers as exec does. Its permission is "+
			"command:shell, apart from exec's command:execute: a token that may run "+
			"a program may not open a shell. Only the admin role gets it.").
		Example("A pipe", ShellInput{Command: "echo hello | tr a-z A-Z"},
			Result{Stdout: "HELLO\n", Changed: true},
			behavior.Prints("/bin/sh -c echo hello | tr a-z A-Z", "HELLO\n"),
			behavior.Ran("/bin/sh", "-c", "echo hello | tr a-z A-Z"),
		).
		Example("A command line that exits non-zero", ShellInput{Command: "exit 3"},
			Result{ExitCode: 3, Changed: true},
			behavior.Exits("/bin/sh -c exit 3", 3),
			behavior.Ran("/bin/sh", "-c", "exit 3"),
		).
		Fails("No command line", ShellInput{}, "required").
		Fails("A command line that is an option", ShellInput{Command: "-x"}, "startsnotwith").
		Fails("A timeout above 300 seconds", ShellInput{Command: "true", Timeout: 301}, "max").
		Fails("A negative timeout", ShellInput{Command: "true", Timeout: -1}, "min"),
))
