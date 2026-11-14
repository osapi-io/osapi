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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// usageTemplate is cobra's default usage template with the headings, command
// names and flag names styled. Without color it prints the same text as
// cobra's, so anything that reads help output keeps working.
const usageTemplate = `{{heading "Usage:"}}{{if runnable .}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{heading "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{heading "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{heading "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{sub (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{heading .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{sub (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{heading "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{sub (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{heading "Flags:"}}
{{flags (.LocalFlags.FlagUsages | trimTrailingWhitespaces)}}{{end}}{{if .HasAvailableInheritedFlags}}

{{heading "Global Flags:"}}
{{flags (.InheritedFlags.FlagUsages | trimTrailingWhitespaces)}}{{end}}{{if .HasHelpSubCommands}}

{{heading "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{sub (rpad .CommandPath .CommandPathPadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{mute (printf "Use \"%s [command] --help\" for more information about a command." .CommandPath)}}{{end}}
`

// usageError marks an error caused by how the command was invoked, so the
// error output adds a pointer to the help.
type usageError struct {
	err error
}

func (e *usageError) Error() string { return e.err.Error() }

func (e *usageError) Unwrap() error { return e.err }

// usagePrefixes are cobra's own messages for a bad invocation. Cobra returns
// them as plain errors, so the prefix is the only way to tell them apart.
var usagePrefixes = []string{
	"unknown command",
	"unknown flag",
	"unknown shorthand flag",
	"required flag",
	"accepts ",
	"requires ",
	"invalid argument",
	"flag needs an argument",
	"if any flags in the group",
}

// IsUsageError reports whether err came from how the command was invoked.
func IsUsageError(
	err error,
) bool {
	var ue *usageError
	if errors.As(err, &ue) {
		return true
	}

	for _, p := range usagePrefixes {
		if strings.HasPrefix(err.Error(), p) {
			return true
		}
	}

	return false
}

// ApplyTheme styles help and errors for root and every command under it.
// The banner is printed above the root's help only. With --json, help is the
// command described as JSON instead, for an agent to read.
func ApplyTheme(
	root *cobra.Command,
	version string,
) {
	out := os.Stdout

	cobra.AddTemplateFunc("heading", func(s string) string { return Heading(out, s) })
	cobra.AddTemplateFunc("sub", func(s string) string { return Sub(out, s) })
	cobra.AddTemplateFunc("mute", func(s string) string { return Mute(out, s) })
	cobra.AddTemplateFunc("flags", func(s string) string { return Flags(out, s) })
	cobra.AddTemplateFunc("runnable", runnable)

	root.SetUsageTemplate(usageTemplate)
	root.SilenceUsage = true
	root.SilenceErrors = true

	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{err: err}
	})

	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(c *cobra.Command, args []string) {
		if wantsJSON(c) {
			// The document is plain structs and strings, so it always marshals.
			out, _ := HelpJSON(c, version)
			_, _ = fmt.Fprintln(c.OutOrStdout(), string(out))

			return
		}

		if c == root {
			w := c.OutOrStdout()
			if banner := Banner(w); banner != "" {
				_, _ = fmt.Fprintln(w)
				_, _ = fmt.Fprint(w, banner)
				_, _ = fmt.Fprintln(w)
			}
		}
		defaultHelp(c, args)
	})
}

// groupAnnotation marks a command GuardGroups gave a run function only to
// reject unknown words. It is still a group: help and help --json say so.
const groupAnnotation = "osapi.group"

// runnable is whether c does something of its own, rather than only group
// others.
func runnable(
	c *cobra.Command,
) bool {
	return c.Runnable() && c.Annotations[groupAnnotation] == ""
}

// GuardGroups makes every command that only groups others reject a word it
// does not know. Cobra checks for an unknown subcommand only at the root;
// below it, a group given one prints its help and exits 0, so a typo such as
// `osapi config lst` looks like success. A group run bare still prints its
// help. Call it once every command is added.
func GuardGroups(
	root *cobra.Command,
) {
	for _, c := range root.Commands() {
		GuardGroups(c)
	}
	if root.Runnable() || !root.HasSubCommands() || !root.HasParent() {
		return
	}

	if root.Annotations == nil {
		root.Annotations = map[string]string{}
	}
	root.Annotations[groupAnnotation] = "true"
	root.Args = cobra.ArbitraryArgs
	root.RunE = func(c *cobra.Command, args []string) error {
		if len(args) > 0 {
			return &usageError{
				err: fmt.Errorf("unknown command %q for %q", args[0], c.CommandPath()),
			}
		}

		return c.Help()
	}
}

// PrintError writes err for a person: a red cross and the message, then a
// pointer to the command's help when the invocation itself was wrong. With
// jsonOutput it writes the JSON error document instead.
func PrintError(
	w io.Writer,
	cmd *cobra.Command,
	err error,
	jsonOutput bool,
) {
	FlushKV()
	if jsonOutput {
		code := CodeError
		if IsUsageError(err) {
			code = CodeUsage
		}
		WriteErrorJSON(w, ErrorDocument{Error: err.Error(), Code: code})

		return
	}

	_, _ = fmt.Fprintln(w, Failure(w, err.Error()))

	// A usage error is followed by the command's usage, as cobra prints it,
	// in the theme: what was wrong, then how to call it.
	if IsUsageError(err) && cmd != nil {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprint(w, cmd.UsageString())
	}
}
