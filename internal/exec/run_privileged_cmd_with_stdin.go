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

// RunPrivilegedCmdWithStdin executes the provided command with arguments and
// writes stdin to the command's standard input. When sudo is enabled, the
// command is prepended with "sudo".
//
// Pass secrets this way rather than in arguments: stdin is not interpreted by
// a shell, does not appear in the process list, and is never logged.
func (e *Exec) RunPrivilegedCmdWithStdin(
	name string,
	args []string,
	stdin string,
) (string, error) {
	if e.sudo {
		args = append([]string{name}, args...)
		name = "sudo"
	}

	return e.executor.ExecuteWithStdin(name, args, "", stdin)
}
