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

package types

import "slices"

// Signal is a signal that osapi sends to a process, by its name without SIG.
type Signal string

// The values of Signal: the signals.
const (
	// SignalTerm asks the process to stop.
	SignalTerm Signal = "TERM"
	// SignalKill stops the process at once.
	SignalKill Signal = "KILL"
	// SignalHup tells the process that its terminal closed; many daemons read their config again.
	SignalHup Signal = "HUP"
	// SignalInt interrupts the process, as Ctrl-C does.
	SignalInt Signal = "INT"
	// SignalUsr1 has the meaning that the program gives it.
	SignalUsr1 Signal = "USR1"
	// SignalUsr2 has the meaning that the program gives it.
	SignalUsr2 Signal = "USR2"
)

// Values are the signals.
func (Signal) Values() []Signal {
	return []Signal{SignalTerm, SignalKill, SignalHup, SignalInt, SignalUsr1, SignalUsr2}
}

// Valid reports whether s is one of the values.
func (s Signal) Valid() bool {
	return slices.Contains(s.Values(), s)
}
