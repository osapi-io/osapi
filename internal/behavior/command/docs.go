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

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

// exampleHost is the target of each example command, and exampleJobID and
// exampleTraceID fill the output that a docs page shows for it.
const (
	exampleHost    = "web-01"
	exampleJobID   = "550e8400-e29b-41d4-a716-446655440000"
	exampleTraceID = "26e448c9dbe1c4a8d216a98c675079c9"
)

// Example is a method's examples as the command lines of its cobra Example,
// each one under a comment that tells what it does.
func Example(
	b *behavior.Behavior,
	m *behavior.Method,
) string {
	command := b.Command(m)
	lines := make([]string, 0, 3*len(m.Examples))
	for i, ex := range m.Examples {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(
			lines,
			"  # "+ex.Doc,
			"  "+command+" --target "+exampleHost+exampleArgs(ex.Input),
		)
	}

	return strings.Join(lines, "\n")
}

// Annotations are the docs annotations of a method's command: the
// permission that it needs, and its result, which is what each example does
// on the host and the calls that it refuses.
func Annotations(
	b *behavior.Behavior,
	m *behavior.Method,
) map[string]string {
	out := map[string]string{}
	if b.API {
		out[reference.Permission] = b.Permission(m.Verb)
	}

	var parts []string
	if out := Output(m); out != "" {
		parts = append(parts, out)
	}
	var effects []string
	for _, ex := range m.Examples {
		if len(ex.Effects) == 0 {
			continue
		}
		lines := make([]string, 0, len(ex.Effects))
		for _, e := range ex.Effects {
			lines = append(lines, effectLine(e))
		}
		effects = append(
			effects,
			fmt.Sprintf("| %s | %s |", cell(ex.Doc), cell(strings.Join(lines, "; "))),
		)
	}
	if len(effects) > 0 {
		parts = append(parts, "With the default settings, each example does this on the host:\n\n"+
			"| Example | On the host |\n| --- | --- |\n"+strings.Join(effects, "\n"))
	}

	if len(m.Refusals) > 0 {
		rows := make([]string, 0, len(m.Refusals))
		for _, r := range m.Refusals {
			flags := commandLine(Args(r.Input))
			if flags != "" {
				flags = "`" + flags + "`"
			}
			rows = append(
				rows,
				fmt.Sprintf("| %s | %s | %s |", cell(r.Doc), cell(flags), cell(r.Error)),
			)
		}
		parts = append(parts, "The command fails in these cases:\n\n"+
			"| Call | Flags | Fails with |\n| --- | --- | --- |\n"+strings.Join(rows, "\n"))
	}
	if len(parts) > 0 {
		out[reference.Result] = strings.Join(parts, "\n\n")
	}

	return out
}

// Output is what each example of m prints, as osapi client prints it: the
// CLI's own printer, with its colors, in a fenced block of language ansi,
// which the docs site shows in those colors. Each example is one host,
// web-01, that answers the example's output. An example with no output, such
// as one of a write that answers only changed, still shows its row.
func Output(
	m *behavior.Method,
) string {
	blocks := make([]string, 0, len(m.Examples))
	for _, ex := range m.Examples {
		result, ok := exampleResult(m, ex.Output)
		if !ok {
			continue
		}
		screen := cli.Capture(func() {
			// As Print, with m itself, which need not be registered.
			cli.PrintJobID(result.JobID)
			cli.PrintTraceID(exampleTraceID)
			cli.PrintCompactTable([]cli.Section{Table(m, result)})
			// A table with no summary line leaves the trace pending; it is
			// not this example's to print later.
			cli.PrintTraceID("")
		})
		blocks = append(blocks, fmt.Sprintf("%s:\n\n```ansi\n%s\n```",
			ex.Doc, strings.Trim(screen, "\n")))
	}
	if len(blocks) == 0 {
		return ""
	}

	return "Each example prints this:\n\n" + strings.Join(blocks, "\n\n")
}

// exampleResult is the answer that example output gives from one host,
// web-01, shaped as the server shapes it: a row for each element of a list,
// with the hostname and the status. A write's row has changed. ok is false
// for an output that is not an object or a list of objects.
func exampleResult(
	m *behavior.Method,
	output any,
) (Result, bool) {
	// An example's output is the method's output type, which encodes.
	raw, _ := json.Marshal(output)
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		var one map[string]any
		if err := json.Unmarshal(raw, &one); err != nil {
			return Result{}, false
		}
		rows = []map[string]any{one}
	}
	for _, r := range rows {
		r["hostname"] = exampleHost
		r["status"] = "ok"
		if _, ok := r["changed"]; !ok && m.Verb == behavior.VerbWrite {
			r["changed"] = true
		}
	}

	return Result{JobID: exampleJobID, Results: rows}, true
}

// cellEscaper keeps text in one cell of a markdown table: a newline would
// end the row, and a pipe would end the cell.
var cellEscaper = strings.NewReplacer("\n", `\n`, "|", `\|`)

// cell makes s safe in one cell of a markdown table.
func cell(
	s string,
) string {
	return cellEscaper.Replace(s)
}

// GroupAnnotations are the docs annotations of a behavior's own command: a
// note that links to the page of the behavior.
func GroupAnnotations(
	b *behavior.Behavior,
) map[string]string {
	// The command's page is in the directory of its path under the CLI
	// reference, usage/cli, which is two levels under the sidebar; the
	// behavior's page is under reference/behaviors.
	words := strings.Fields(strings.TrimPrefix(b.Command(b.Methods[0]), "osapi "))
	up := strings.Repeat("../", 2+len(words)-1)

	return map[string]string{
		reference.Notes: fmt.Sprintf(
			"The [%s behavior](%sreference/behaviors/%s/%s.md) page tells where the behavior runs, "+
				"what it is built on and its settings.",
			b.Name,
			up,
			b.Name,
			b.Name,
		),
	}
}

// exampleArgs renders an example input as flags, in the input type's field
// order.
func exampleArgs(
	input any,
) string {
	line := commandLine(Args(input))
	if line == "" {
		return ""
	}

	return " " + line
}

// safeArg is an argument a shell reads as it is written.
var safeArg = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// commandLine joins args as a command line a reader can paste: an argument
// with a space or a control character in it is quoted, so a newline shows as
// \n rather than breaking the line, or a table row, it is printed in.
func commandLine(
	args []string,
) string {
	quoted := make([]string, 0, len(args))
	for _, a := range args {
		if !safeArg.MatchString(a) {
			a = strconv.Quote(a)
		}
		quoted = append(quoted, a)
	}

	return strings.Join(quoted, " ")
}

// initialisms are the behavior names that a title keeps in capitals.
var initialisms = map[string]string{"dns": "DNS", "os": "OS", "ntp": "NTP"}

// Title is the name of b as the title of its page and its sidebar entry:
// capitalized, with an initialism such as DNS in capitals.
func Title(
	b *behavior.Behavior,
) string {
	if t, ok := initialisms[b.Name]; ok {
		return t
	}

	return strings.ToUpper(b.Name[:1]) + b.Name[1:]
}

// effectLine says what one effect does, for a docs page.
func effectLine(
	e behavior.Effect,
) string {
	switch e.Kind {
	case behavior.EffectWrote:
		return fmt.Sprintf("writes `%s`: `%s`", e.Path, strings.TrimRight(e.Content, "\n"))
	case behavior.EffectRemoved:
		return fmt.Sprintf("removes `%s`", e.Path)
	default:
		return fmt.Sprintf("runs `%s`", e.Command)
	}
}

// ExampleFlags is the flags of an example's command line: --target and the
// flags of input, as the help and the pages write them.
func ExampleFlags(
	input any,
) string {
	return "--target " + exampleHost + exampleArgs(input)
}
