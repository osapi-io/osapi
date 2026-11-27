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

// exitHooks run before the process exits, the last added first.
var exitHooks []func()

// AtExit adds fn to the work that runs before the CLI exits. A command that
// fails through HandleError or LogFatal exits from inside cobra, so its
// PersistentPostRun never runs. Work that must happen also on a failure,
// such as the end of the command's span and the flush of its tracer, goes
// here.
func AtExit(
	fn func(),
) {
	exitHooks = append(exitHooks, fn)
}

// RunExitHooks runs each hook that AtExit added, the last added first, and
// removes them, so a second call runs nothing.
func RunExitHooks() {
	hooks := exitHooks
	exitHooks = nil
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i]()
	}
}

// exit runs the exit hooks, then exits with code.
func exit(
	code int,
) {
	RunExitHooks()
	osExit(code)
}
