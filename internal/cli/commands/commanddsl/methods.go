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

package commanddsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// The layers that make a command.
const (
	LayerCommand     = "command"
	LayerEndpoint    = "endpoint"
	LayerBehavior    = "behavior"
	LayerHandWritten = "hand-written"
)

// List is each command of the binary that runs, and the layer that made it.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Summary{}}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			return
		}
		if c.Runnable() && len(c.Commands()) == 0 {
			layer, group := layerOf(c, d.IsBehavior)
			out.Items = append(out.Items, Summary{
				Command: c.CommandPath(), Layer: layer, Group: group, Doc: c.Short,
			})
			switch layer {
			case LayerCommand:
				out.Command++
			case LayerEndpoint:
				out.Endpoint++
			case LayerBehavior:
				out.Behavior++
			default:
				out.HandWritten++
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(d.Root)

	return out, nil
}

// layerOf is the layer that made c, and its group's command path.
func layerOf(
	c *cobra.Command,
	isBehavior func(string) bool,
) (string, string) {
	g, ok := command.Origin(c)
	if !ok {
		return LayerHandWritten, ""
	}
	if lg, ok := g.(*command.LocalGroup); ok {
		return LayerCommand, lg.Path()
	}
	group := strings.Join(g.CommandPath(), " ")
	if isBehavior(g.GroupName()) {
		return LayerBehavior, group
	}

	return LayerEndpoint, group
}

// Get is the group with the command path, and each of its commands.
func Get(
	_ context.Context,
	_ Deps,
	in GetInput,
) (GetOutput, error) {
	g := find(in.Name)
	if g == nil {
		return GetOutput{}, fmt.Errorf(
			"no group named %q: osapi cli command list shows them",
			in.Name,
		)
	}
	name := strings.Join(g.CommandPath(), " ")
	if lg, ok := g.(*command.LocalGroup); ok {
		name = lg.Path()
	}
	out := GetOutput{Name: name, Doc: g.GroupDoc(), Commands: []Command{}}
	for _, c := range g.Commands() {
		out.Commands = append(out.Commands, describe(g, c))
	}

	return out, nil
}

// find is the group with the command path: a local group by its path, or
// any group by its command path.
func find(
	path string,
) command.Group {
	if g := command.Lookup(path); g != nil {
		return g
	}
	for _, g := range command.Groups() {
		if strings.Join(g.CommandPath(), " ") == path {
			return g
		}
	}

	return nil
}

// describe is c of g: its whole command, and the flags and refusals of a
// local command.
func describe(
	g command.Group,
	c command.Command,
) Command {
	root := "osapi client"
	_, local := g.(*command.LocalGroup)
	if local {
		root = "osapi"
	}
	full := strings.Join(
		[]string{root, strings.Join(g.CommandPath(), " "), c.SubPath(), c.CommandName()}, " ",
	)
	out := Command{
		Command: strings.Join(strings.Fields(full), " "),
		Method:  c.MethodName(),
		Short:   c.CommandShort(),
	}
	// Only a local command has flags from its input: an API command's come
	// from its request.
	if !local {
		return out
	}
	m, _ := c.(*command.LocalMethod)
	for f := range m.In.Fields() {
		if !f.IsExported() {
			continue
		}
		out.Flags = append(out.Flags, Flag{
			Flag: "--" + command.FlagName(f), Short: f.Tag.Get("short"),
			Type: f.Type.String(), Validate: f.Tag.Get("validate"), Doc: f.Tag.Get("doc"),
		})
	}
	for _, r := range m.Refusals() {
		out.Fails = append(out.Fails, r.Doc)
	}

	return out
}

// Create writes a new local group in the checkout, and is its files and the
// steps after it.
func Create(
	_ context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	files, err := command.Scaffold(d.Dir, in.Name, in.Description)
	if err != nil {
		return CreateOutput{}, err
	}
	out := CreateOutput{Name: in.Name}
	for _, f := range files {
		out.Files = append(out.Files, File{Path: f})
	}
	for _, step := range []string{
		"go generate ./internal/cli/commands/all",
		"go test ./internal/cli/commands/" + in.Name + "/...",
		"go run . cli command validate",
		"mise exec -- just cli-docs-generate",
	} {
		out.Next = append(out.Next, Step{Command: step})
	}

	return out, nil
}

// Spec is each part of the command DSL, by kind.
func Spec(
	_ context.Context,
	_ Deps,
	_ SpecInput,
) (SpecOutput, error) {
	return SpecOutput{Kinds: command.Spec.Sections()}, nil
}

// Validate checks each local group, and the output of each group. When a
// rule fails, it is the report with command.Fail: the report prints, then
// the command exits 1.
func Validate(
	_ context.Context,
	_ Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	out := ValidateOutput{Local: len(command.All()), Groups: len(command.Groups())}
	for _, g := range command.All() {
		out.Problems = append(out.Problems, command.Validate(g)...)
	}
	out.Problems = append(out.Problems, command.ValidateOutput()...)
	out.Valid = len(out.Problems) == 0
	if out.Valid {
		return out, nil
	}
	out.Fixes = map[string]any{}
	for _, p := range out.Problems {
		if p.Fix != "" {
			out.Fixes[p.Rule] = p.Fix
		}
	}

	return out, command.Fail()
}
