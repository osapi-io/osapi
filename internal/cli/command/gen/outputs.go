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
	"context"
	"encoding/json"
	"log/slog"
	"slices"
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

// Outputs are the outputs of the examples of m: each example that Prints
// gave one, or, for a method that answers with no body, each example
// without --json. For each, the command line and what the CLI prints for
// it, in a fenced ansi block, then what --json prints. line is the command
// with its group, such as osapi client agent list.
func Outputs(
	g command.Group,
	m command.Command,
	line string,
) string {
	_, local := g.(*command.LocalGroup)
	jsonDoc := "With `--json`, the API's response:"
	if local {
		jsonDoc = "With `--json`, the command prints:"
	}
	sections := make([]string, 0, len(m.Examples()))
	for _, ex := range m.Examples() {
		if len(ex.Logs) > 0 {
			sections = append(sections, logSection(ex, strings.TrimRight(line+" "+ex.Flags, " ")))

			continue
		}
		var in any
		var answer, body []byte
		switch {
		case ex.Output != nil:
			// An output is a value of a Go type, which always encodes.
			answer, _ = json.Marshal(ex.Output)
			body, _ = json.MarshalIndent(ex.Output, "", "  ")
		case !slices.Contains(strings.Fields(ex.Flags), "--json"):
			i, done, ok := command.Done(g, m, ex.Flags)
			if !ok {
				continue
			}
			in, body = i, done
		default:
			continue
		}
		printed := screen(g, m, in, answer, !local)
		said := jsonDoc
		if in != nil {
			// A method with no body answers nothing: the CLI writes the JSON.
			said = "With `--json`, the command prints:"
		}
		doc := ex.Doc
		if doc == "" {
			doc = "The command prints"
		}
		cmd := strings.TrimRight(line+" "+ex.Flags, " ")
		sections = append(sections, doc+":\n\n```ansi\n$ "+cmd+"\n"+printed+"\n```\n\n"+
			said+"\n\n```json\n"+string(body)+"\n```")
	}

	return strings.Join(sections, "\n\n")
}

// logSection is the section of a process's example: the command line and
// the lines it writes when it starts, as its log handler writes them in a
// terminal, then as JSON for --json.
func logSection(
	ex command.Example,
	cmd string,
) string {
	doc := ex.Doc
	if doc == "" {
		doc = "The command prints"
	}

	return doc + ":\n\n```ansi\n$ " + cmd + "\n" + logged(ex.Logs, false) + "\n```\n\n" +
		"With `--json`, the command prints:\n\n```json\n" + logged(ex.Logs, true) + "\n```"
}

// logged is lines as the CLI's log handler writes them, each at ExampleNow:
// in color, or as JSON.
func logged(
	lines []command.LogLine,
	asJSON bool,
) string {
	var buf strings.Builder
	h := cli.LogHandler(&buf, slog.LevelInfo, asJSON, true)
	for _, l := range lines {
		r := slog.NewRecord(ExampleNow, l.Level, l.Msg, 0)
		r.AddAttrs(l.Attrs...)
		// A handler that writes to a strings.Builder does not fail.
		_ = h.Handle(context.Background(), r)
	}

	return strings.TrimRight(buf.String(), "\n")
}

// screen is what the CLI prints for answer, the JSON of m's output, to the
// input in. traced is true for a command that calls the API, which ends
// with the trace of its call.
func screen(
	g command.Group,
	m command.Command,
	in any,
	answer []byte,
	traced bool,
) string {
	command.SetNow(ExampleNow)
	defer command.SetNow(time.Time{})
	printed := cli.Capture(func() {
		if traced {
			cli.PrintTraceID(ExampleTraceID)
		}
		if err := command.Derive(docsPrinter{}, g, m, in, answer); err != nil {
			panic("example output of " + g.GroupName() + " " + m.MethodName() + ": " + err.Error())
		}
		if traced {
			cli.FlushTraceID()
		}
	})

	return strings.Trim(printed, "\n")
}
