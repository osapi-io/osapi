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

package cmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/pkg/behavior"
)

// methodDescription is one method in `behavior describe --json`.
type methodDescription struct {
	Name       string             `json:"name"`
	Doc        string             `json:"doc"`
	Verb       behavior.Verb      `json:"verb"`
	HTTPMethod string             `json:"http_method"`
	Path       string             `json:"path"`
	Permission string             `json:"permission"`
	Command    string             `json:"command"`
	Input      *behavior.Schema   `json:"input"`
	Output     *behavior.Schema   `json:"output"`
	Examples   []behavior.Example `json:"examples,omitempty"`
}

// behaviorDescription is what `behavior describe --json` prints.
type behaviorDescription struct {
	Name      string              `json:"name"`
	Doc       string              `json:"doc"`
	Placement behavior.Placement  `json:"placement"`
	Group     string              `json:"group,omitempty"`
	Methods   []methodDescription `json:"methods"`
}

var behaviorDescribeCmd = &cobra.Command{
	Use:   "describe <behavior> [method]",
	Short: "Show a behavior's methods, or one method's input and output",
	Example: `  osapi behavior describe sysctl
  osapi behavior describe sysctl get
  osapi behavior describe sysctl get --json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(_ *cobra.Command, args []string) error {
		b := behavior.Lookup(args[0])
		if b == nil {
			return fmt.Errorf("no behavior named %q; run 'osapi behavior list'", args[0])
		}

		methods := b.Methods
		if len(args) == 2 {
			m := b.Method(args[1])
			if m == nil {
				return fmt.Errorf("behavior %s has no method %q", b.Name, args[1])
			}
			methods = []*behavior.Method{m}
		}

		desc, err := describeBehavior(b, methods)
		if err != nil {
			return err
		}

		if jsonOutput {
			out, _ := json.Marshal(desc)
			cli.PrintJSON(out)

			return nil
		}

		if len(args) == 2 {
			printMethod(desc.Methods[0])

			return nil
		}
		printBehavior(desc)

		return nil
	},
}

func describeBehavior(
	b *behavior.Behavior,
	methods []*behavior.Method,
) (behaviorDescription, error) {
	desc := behaviorDescription{
		Name: b.Name, Doc: b.Doc, Placement: b.Placement, Group: b.Group,
	}

	for _, m := range methods {
		in, err := behavior.SchemaFor(m.In)
		if err != nil {
			return desc, fmt.Errorf("%s %s input: %w", b.Name, m.Name, err)
		}
		out, err := behavior.SchemaFor(m.Out)
		if err != nil {
			return desc, fmt.Errorf("%s %s output: %w", b.Name, m.Name, err)
		}

		desc.Methods = append(desc.Methods, methodDescription{
			Name:       m.Name,
			Doc:        m.Doc,
			Verb:       m.Verb,
			HTTPMethod: m.HTTPMethod(),
			Path:       b.Path(m),
			Permission: b.Permission(m.Verb),
			Command:    b.Command(m),
			Input:      in,
			Output:     out,
			Examples:   m.Examples,
		})
	}

	return desc, nil
}

func printBehavior(
	desc behaviorDescription,
) {
	rows := make([][]string, 0, len(desc.Methods))
	for _, m := range desc.Methods {
		rows = append(rows, []string{
			m.Name, string(m.Verb), m.HTTPMethod + " " + m.Path, m.Permission, m.Doc,
		})
	}

	fmt.Println()
	cli.PrintKV("Behavior", desc.Name, "Placement", string(desc.Placement))
	cli.PrintKV("Description", desc.Doc)
	cli.PrintCompactTable([]cli.Section{{
		Title:     "Methods",
		Headers:   []string{"METHOD", "VERB", "ROUTE", "PERMISSION", "DESCRIPTION"},
		Rows:      rows,
		Reference: true,
	}})
}

func printMethod(
	m methodDescription,
) {
	fmt.Println()
	cli.PrintKV("Method", m.Name, "Verb", string(m.Verb))
	cli.PrintKV("Description", m.Doc)
	cli.PrintKV("Route", m.HTTPMethod+" "+m.Path, "Permission", m.Permission)
	cli.PrintKV("Command", m.Command)

	cli.PrintCompactTable([]cli.Section{
		{
			Title:     "Input",
			Headers:   []string{"FIELD", "TYPE", "REQUIRED", "CONSTRAINTS", "DESCRIPTION"},
			Rows:      fieldRows(m.Input, true),
			Reference: true,
		},
		{
			Title:     "Output",
			Headers:   []string{"FIELD", "TYPE", "DESCRIPTION"},
			Rows:      fieldRows(m.Output, false),
			Reference: true,
		},
	})
}

// fieldRows lists a schema's fields in declaration order. A slice output is
// described by its element, since each element is one result row.
func fieldRows(
	s *behavior.Schema,
	withConstraints bool,
) [][]string {
	if s.Type == "array" && s.Items != nil {
		s = s.Items
	}

	rows := make([][]string, 0, len(s.Order))
	for _, name := range s.Order {
		f := s.Properties[name]
		if withConstraints {
			rows = append(rows, []string{
				name, f.Type, strconv.FormatBool(slices.Contains(s.Required, name)),
				constraints(f), f.Description,
			})

			continue
		}
		rows = append(rows, []string{name, f.Type, f.Description})
	}

	return rows
}

// constraints summarizes a field's validation in one cell.
func constraints(
	f *behavior.Schema,
) string {
	var parts []string
	if f.Pattern != "" {
		parts = append(parts, "pattern "+f.Pattern)
	}
	if f.Format != "" {
		parts = append(parts, "format "+f.Format)
	}
	if len(f.Enum) > 0 {
		parts = append(parts, "one of "+strings.Join(f.Enum, " "))
	}
	if f.MinLength != nil {
		parts = append(parts, fmt.Sprintf("min length %d", *f.MinLength))
	}
	if f.MaxLength != nil {
		parts = append(parts, fmt.Sprintf("max length %d", *f.MaxLength))
	}
	if f.Minimum != nil {
		parts = append(parts, fmt.Sprintf("min %g", *f.Minimum))
	}
	if f.Maximum != nil {
		parts = append(parts, fmt.Sprintf("max %g", *f.Maximum))
	}

	return strings.Join(parts, ", ")
}

func init() {
	behaviorCmd.AddCommand(behaviorDescribeCmd)
}
