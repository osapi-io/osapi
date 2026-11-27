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

package exec

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// waitDelay is how long Wait waits for a command's output after its context
// ended and its process group was killed. A process that left the group, or
// that the agent may not signal, such as a child of sudo, can still hold the
// output open. After waitDelay, Wait closes it and returns.
var waitDelay = 5 * time.Second

// killGroup signals a process group; a test replaces it.
var killGroup = syscall.Kill

// sudoArgs gives the arguments of sudo to run name with args. -n makes sudo
// fail at once when it needs a password, as the preflight's sudo -n -l
// does, rather than wait on a terminal for one until the timeout.
func sudoArgs(
	name string,
	args []string,
) []string {
	return append([]string{"-n", name}, args...)
}

// newCommand is exec.CommandContext for name with args, in a process group of
// its own. When ctx ends, the whole group is killed, not only name: a
// command's child, such as the background job of a shell, that holds the
// output open would otherwise keep Wait blocked past the timeout.
func newCommand(
	ctx context.Context,
	name string,
	args []string,
) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := killGroup(-cmd.Process.Pid, syscall.SIGKILL)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, syscall.ESRCH):
			return os.ErrProcessDone
		default:
			// The agent may not signal the group, such as one that sudo
			// leads as root. Kill what it can. waitDelay bounds the rest.
			return cmd.Process.Kill()
		}
	}
	cmd.WaitDelay = waitDelay

	return cmd
}
