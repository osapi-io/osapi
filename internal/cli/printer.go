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
	"os"
	"time"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// Printer is the command layer's printer: the output of a method that has
// no renderer prints through the helpers here, in the one theme.
type Printer struct{}

var _ command.Printer = Printer{}

// Heading prints a heading, as PrintHeading does.
func (Printer) Heading(
	title string,
) {
	PrintHeading(title)
}

// KV adds pairs to the key-value block, as PrintKV does.
func (Printer) KV(
	pairs ...string,
) {
	PrintKV(pairs...)
}

// Table prints t under its heading, as PrintCompactTable does: an empty
// one says None., then its hint.
func (Printer) Table(
	t command.Table,
) {
	empty := "None."
	if t.Empty != "" {
		empty += " " + t.Empty
	}
	PrintCompactTable([]Section{{
		Title:     t.Title,
		Headers:   t.Headers,
		Rows:      t.Rows,
		Empty:     empty,
		Reference: true,
	}})
}

// Verdict prints text as a one-line verdict at the left margin.
func (Printer) Verdict(
	text string,
) {
	FlushKV()
	fmt.Println(Success(os.Stdout, text))
}

// Status is word in its status color, as RenderStatus renders it.
func (Printer) Status(
	word string,
) string {
	return RenderStatus(word)
}

// Bytes is n as a size, as FormatBytes renders it.
func (Printer) Bytes(
	n int64,
) string {
	return FormatBytes(int(n))
}

// Age is d as an age, as FormatAge renders it.
func (Printer) Age(
	d time.Duration,
) string {
	return FormatAge(d)
}

// Time is t in the one time format, as LocalTime renders it.
func (Printer) Time(
	t time.Time,
) string {
	return LocalTime(t)
}

// Duration is d rounded, as FormatDuration renders it.
func (Printer) Duration(
	d time.Duration,
) string {
	return FormatDuration(d.String())
}
