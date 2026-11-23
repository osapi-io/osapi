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
		out[reference.Permission] = string(b.Permission(m.PermissionVerb()))
	}

	var parts []string
	if out := Output(b, m); out != "" {
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

// Output is what each example of m prints, as the CLI prints it, colors
// included: the command and its table for one host, the same command for
// the fleet, where one host skips, and the API's response with --json. Each
// screen is a fenced ansi block, which the docs site shows as a terminal.
// An example whose output is not an object, or a list of them, has none.
func Output(
	b *behavior.Behavior,
	m *behavior.Method,
) string {
	command := b.Command(m)
	sections := make([]string, 0, len(m.Examples))
	for _, ex := range m.Examples {
		one, ok := exampleResult(m, ex.Output)
		if !ok {
			continue
		}
		args := exampleArgs(ex.Input)
		// one is decoded JSON, which always encodes again.
		body, _ := json.MarshalIndent(one, "", "  ")
		sections = append(sections, fmt.Sprintf(
			"%s, on one host:\n\n%s\n\nOn each host, where one cannot run it:\n\n%s\n\n"+
				"With `--json`, the API's response:\n\n```json\n%s\n```",
			ex.Doc,
			screen(m, command+" --target "+exampleHost+args, one),
			screen(m, command+" --target _all"+args, fleetResult(one)),
			body,
		))
	}
	if len(sections) == 0 {
		return ""
	}

	return strings.Join(sections, "\n\n")
}

// screen is line, as a person types it, and what the CLI prints for result,
// in a fenced ansi block.
func screen(
	m *behavior.Method,
	line string,
	result Result,
) string {
	printed := cli.Capture(func() {
		// As Print, with m itself, which need not be registered.
		cli.PrintJobID(result.JobID)
		cli.PrintTraceID(exampleTraceID)
		cli.PrintCompactTable([]cli.Section{Table(m, result)})
		// A table with no summary line leaves the trace pending; it is not
		// this example's to print later.
		cli.PrintTraceID("")
	})

	return "```ansi\n$ " + line + "\n" + strings.Trim(printed, "\n") + "\n```"
}

// fleetHosts are the hosts of the fleet example: two that return the
// example's output, and a Mac that skips, as it does a Linux-only behavior.
var fleetHosts = []string{"web-01", "web-02", "mac-01"}

// fleetResult is one as the fleet answers it: its rows from each of the
// first two hosts, and the last host skipped.
func fleetResult(
	one Result,
) Result {
	out := Result{JobID: one.JobID}
	for _, host := range fleetHosts[:2] {
		for _, row := range one.Results {
			out.Results = append(out.Results, rowFor(row, host))
		}
	}
	out.Results = append(out.Results, map[string]any{
		"hostname": fleetHosts[2],
		"status":   "skipped",
		"error":    behavior.ErrUnsupported.Error(),
	})

	return out
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

// InternalFolder is the folder of the reference pages of the internal
// behaviors, which have no area, after the folders of the areas.
const InternalFolder = "internal"

// InternalLabel is the label of InternalFolder in the sidebar and the navbar.
const InternalLabel = "Internal"

// ReferenceFolder is the folder of b's reference page under
// reference/behaviors: the folder of its area, such as hardware, or
// InternalFolder for an internal behavior, which has no area.
func ReferenceFolder(
	b *behavior.Behavior,
) string {
	if b.Area == "" {
		return InternalFolder
	}

	return string(b.Area)
}

// ReferencePage is b's reference page from reference/behaviors, such as
// hardware/disk/disk.md: each behavior's page is in a folder of its own, in
// the folder of its area. A behavior that Of nests has its folder in the
// folder of its parent, as its command is under the parent's, such as
// security/user/sshkey/sshkey.md.
func ReferencePage(
	b *behavior.Behavior,
) string {
	return ReferenceFolder(b) + "/" + pageFolder(b) + "/" + b.Name + ".md"
}

// pageFolder is the folder of b's page in the folder of its area: its name,
// in the folder of its parent's page when Of nests it.
func pageFolder(
	b *behavior.Behavior,
) string {
	if b.Parent != nil {
		return pageFolder(b.Parent) + "/" + b.Name
	}

	return b.Name
}

// ReferenceUp is the path from the folder of b's reference page up to
// reference/behaviors, such as ../../ for hardware/disk/disk.md.
func ReferenceUp(
	b *behavior.Behavior,
) string {
	return strings.Repeat("../", strings.Count(ReferencePage(b), "/"))
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
			"The [%s behavior](%sreference/behaviors/%s) page tells where the behavior runs, "+
				"what it is built on and its settings.",
			b.Name,
			up,
			ReferencePage(b),
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

// exampleFailure is the error that a host gives in a screen of Screen where
// the operation failed on it.
const exampleFailure = "permission denied"

// Screen is the screen of the first example of a behavior's method, as the
// CLI prints it, for a page that a person wrote: the osapi:output marker
// gives its arguments. They are the behavior, the method, then each host,
// as host, host:skipped or host:failed. One host is --target host; more are
// --target _all, and each answers in the order given. With --json last, it
// is the API's response, which the CLI prints with no table and no color.
func Screen(
	args []string,
) string {
	if len(args) < 3 {
		return "osapi:output needs a behavior, a method and a host"
	}
	b := behavior.Lookup(args[0])
	if b == nil || b.Method(args[1]) == nil || len(b.Method(args[1]).Examples) == 0 {
		return fmt.Sprintf(
			"osapi:output: no behavior %s with a method %s and an example",
			args[0],
			args[1],
		)
	}
	m := b.Method(args[1])
	hosts := args[2:]
	asJSON := hosts[len(hosts)-1] == "--json"
	if asJSON {
		hosts = hosts[:len(hosts)-1]
	}
	ex := m.Examples[0]
	one, _ := exampleResult(m, ex.Output)

	out := Result{JobID: one.JobID}
	for _, h := range hosts {
		name, status, _ := strings.Cut(h, ":")
		switch status {
		case "skipped":
			out.Results = append(out.Results, map[string]any{
				"hostname": name, "status": "skipped", "error": behavior.ErrUnsupported.Error(),
			})
		case "failed":
			out.Results = append(out.Results, map[string]any{
				"hostname": name, "status": "failed", "error": exampleFailure,
			})
		default:
			for _, row := range one.Results {
				out.Results = append(out.Results, rowFor(row, name))
			}
		}
	}

	target := "_all"
	if len(hosts) == 1 {
		target, _, _ = strings.Cut(hosts[0], ":")
	}
	line := b.Command(m) + " --target " + target + exampleArgs(ex.Input)
	if asJSON {
		// The body as the API sends it: compact, and with no color.
		body, _ := json.Marshal(out)

		return "```ansi\n$ " + line + " --json\n" + string(body) + "\n```"
	}

	return screen(m, line, out)
}

// rowFor is row as host answers it: each value that is the example's host,
// such as the name that hostname get reads, is host's own.
func rowFor(
	row map[string]any,
	host string,
) map[string]any {
	out := make(map[string]any, len(row))
	for k, v := range row {
		if v == exampleHost {
			v = host
		}
		out[k] = v
	}
	out["hostname"] = host

	return out
}
