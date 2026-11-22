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

package commandgen

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
)

// ExampleTraceID is the trace that each example output ends with, and
// ExampleNow is the time that an age in one is from, so a page is the same
// on each run. A sample output gives its times relative to ExampleNow.
const ExampleTraceID = "26e448c9dbe1c4a8d216a98c675079c9"

// ExampleNow is the time that an age in an example output is from.
var ExampleNow = time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)

// docsPrinter prints as the CLI does, with each time in UTC: a page does
// not change with the time zone of the machine that writes it.
type docsPrinter struct {
	cli.Printer
}

// Time is t in UTC, in the one time format.
func (docsPrinter) Time(
	t time.Time,
) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

// Outputs are the outputs of the examples of m that Prints gave one: for
// each, the command line and what the CLI prints for it, in a fenced ansi
// block, then the API's response as --json prints it. line is the command
// with its group, such as osapi client agent list.
func Outputs(
	g command.Group,
	m command.Command,
	line string,
) string {
	var sections []string
	for _, ex := range m.Examples() {
		if ex.Output == nil {
			continue
		}
		// An output is a value of a Go type, which always encodes.
		answer, _ := json.Marshal(ex.Output)
		body, _ := json.MarshalIndent(ex.Output, "", "  ")
		printed := screen(g, m, answer)
		doc := ex.Doc
		if doc == "" {
			doc = "The command prints"
		}
		cmd := strings.TrimRight(line+" "+ex.Flags, " ")
		sections = append(sections, doc+":\n\n```ansi\n$ "+cmd+"\n"+printed+"\n```\n\n"+
			"With `--json`, the API's response:\n\n```json\n"+string(body)+"\n```")
	}

	return strings.Join(sections, "\n\n")
}

// screen is what the CLI prints for answer, the JSON of m's output.
func screen(
	g command.Group,
	m command.Command,
	answer []byte,
) string {
	command.SetNow(ExampleNow)
	defer command.SetNow(time.Time{})
	printed := cli.Capture(func() {
		cli.PrintTraceID(ExampleTraceID)
		if err := command.Derive(docsPrinter{}, g, m, nil, answer); err != nil {
			panic("example output of " + g.GroupName() + " " + m.MethodName() + ": " + err.Error())
		}
		cli.FlushTraceID()
	})

	return strings.Trim(printed, "\n")
}
