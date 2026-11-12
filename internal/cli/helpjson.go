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
	"encoding/json"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootContract is the part of the CLI an agent needs before running anything:
// how errors are reported and what the exit codes mean. It is printed with
// the root command's JSON help.
const rootContract = `With --json, a command that fails writes its error to stderr as one JSON
document:

    {"error": "<message>", "code": "<error_code>"}

Match on "code"; the message can change between releases.

Exit codes, shared by every command:

    0    success
    1    error
    2    usage error: unknown command or flag, or a missing required flag`

// HelpCommand is one command in the JSON help. Subcommands are listed by name
// and description only, so an agent asks for the one it needs next rather
// than reading the whole tree at once.
type HelpCommand struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	Aliases     []string         `json:"aliases"`
	Description string           `json:"description"`
	Long        string           `json:"long,omitempty"`
	Usage       string           `json:"usage"`
	Example     string           `json:"example,omitempty"`
	Runnable    bool             `json:"runnable"`
	Flags       []HelpFlag       `json:"flags"`
	Subcommands []HelpSubcommand `json:"subcommands"`
}

// HelpFlag is one flag of a command, local or inherited.
type HelpFlag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
	Required  bool   `json:"required"`
	Inherited bool   `json:"inherited"`
}

// HelpSubcommand names a subcommand without describing it in full.
type HelpSubcommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HelpDocument is what `osapi help <command> --json` prints.
type HelpDocument struct {
	Version string      `json:"version"`
	Command HelpCommand `json:"command"`
}

// HelpJSON describes c as JSON: its flags, which are required, and its
// subcommands. For the root command the description also carries the error
// and exit-code contract every command shares.
func HelpJSON(
	c *cobra.Command,
	version string,
) ([]byte, error) {
	doc := HelpDocument{
		Version: version,
		Command: HelpCommand{
			Name:        c.Name(),
			Path:        c.CommandPath(),
			Aliases:     append([]string{}, c.Aliases...),
			Description: c.Short,
			Long:        c.Long,
			Usage:       c.UseLine(),
			Example:     c.Example,
			Runnable:    c.Runnable(),
			Flags:       []HelpFlag{},
			Subcommands: []HelpSubcommand{},
		},
	}

	if !c.HasParent() {
		doc.Command.Long = rootContract
	}

	required := map[string]bool{}
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok &&
			len(ann) > 0 && ann[0] == "true" {
			required[f.Name] = true
		}
	})

	add := func(inherited bool) func(*pflag.Flag) {
		return func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			doc.Command.Flags = append(doc.Command.Flags, HelpFlag{
				Name:      f.Name,
				Shorthand: f.Shorthand,
				Type:      f.Value.Type(),
				Default:   f.DefValue,
				Usage:     f.Usage,
				Required:  required[f.Name],
				Inherited: inherited,
			})
		}
	}
	c.LocalFlags().VisitAll(add(false))
	c.InheritedFlags().VisitAll(add(true))

	for _, sub := range c.Commands() {
		if !sub.IsAvailableCommand() {
			continue
		}
		doc.Command.Subcommands = append(doc.Command.Subcommands, HelpSubcommand{
			Name:        sub.Name(),
			Description: sub.Short,
		})
	}
	sort.Slice(doc.Command.Subcommands, func(i, j int) bool {
		return doc.Command.Subcommands[i].Name < doc.Command.Subcommands[j].Name
	})

	return json.MarshalIndent(doc, "", "  ")
}

// wantsJSON reports whether --json was given. It reads the root's persistent
// flag rather than a variable, because help runs before cobra's initializers.
func wantsJSON(
	c *cobra.Command,
) bool {
	f := c.Root().PersistentFlags().Lookup("json")

	return f != nil && f.Value.String() == "true"
}
