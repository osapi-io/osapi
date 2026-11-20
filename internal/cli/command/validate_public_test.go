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

package command_test

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// roleIn has a rule after dive, which breaks on an item.
type roleIn struct {
	Roles []string `json:"roles" validate:"required,dive,oneof=a b" doc:"The roles"`
}

type roleOut struct{}

func handleRoles(
	_ context.Context,
	_ localDeps,
	_ roleIn,
) (roleOut, error) {
	return roleOut{}, nil
}

// badFlags has a field for each way a field is not a flag.
type badFlags struct {
	NoJSON  string `validate:"omitempty" doc:"No json name"`
	Renamed string `validate:"omitempty" doc:"Renamed"      json:"renamed"   flag:"other"`
	Same    string `validate:"omitempty" doc:"Same"         json:"same_name" flag:"same-name"`
	Routed  string `validate:"omitempty" doc:"Routed"       json:"routed"                     in:"path"`
	Wide    int64  `validate:"omitempty" doc:"Wide"         json:"wide"`
	Bare    string `                                        json:"bare"`
	Skipped string `                                        json:"-"`
	hidden  string //nolint:unused
}

// badTags breaks each rule of the output's tags.
type badTags struct {
	_    struct{} `title:"Bad"`
	Size int      `             json:"size" render:"nope"`
	Name string   `title:"Name" json:"name" render:"short=0" empty:"None."`
}

type ValidatePublicTestSuite struct {
	suite.Suite
}

func (s *ValidatePublicTestSuite) TestValidate() {
	tests := []struct {
		name  string
		group *command.LocalGroup
		want  []string
	}{
		{
			name: "when the group is valid",
			group: command.Define("zzvalid", "Valid",
				command.Uses[localDeps](),
				command.Method[roleIn, roleOut]("list").
					Short("List roles").
					Example("--roles a").
					Fails("no roles", roleIn{}).
					Fails("an unknown role", roleIn{Roles: []string{"c"}}).
					Handle(handleRoles),
			),
		},
		{
			name:  "when the group has a bad word, no methods and no Uses",
			group: command.Define("Bad", "Bad", command.Under("ok", "Up")),
			want: []string{
				`the word "Up" is not lowercase letters and digits.`,
				`the word "Bad" is not lowercase letters and digits.`,
				"the group has no methods.",
				"the group Uses no type.",
			},
		},
		{
			name: "when the methods break the rules of names, verbs and help",
			group: command.Define("zznames", "Names",
				command.Uses[localDeps](),
				command.Method[struct{}, greetOut]("List").Sub("Bad").Short("X").
					Handle(func(context.Context, localDeps, struct{}) (greetOut, error) {
						return greetOut{}, nil
					}),
				command.Method[roleIn, roleOut]("frob").
					Fails("no roles", roleIn{}).
					Fails("an unknown role", roleIn{Roles: []string{"c"}}).
					Handle(handleRoles),
				command.Method[roleIn, roleOut]("frob").Short("X").Example("").
					Fails("no roles", roleIn{}).
					Fails("an unknown role", roleIn{Roles: []string{"c"}}).
					Handle(handleRoles),
			),
			want: []string{
				`"List" is not lowercase letters and digits.`,
				`"Bad" is not lowercase letters and digits.`,
				`"List" is not a verb or an action.`,
				`the Sub segment "Bad" has no SubDoc.`,
				`"frob" is not a verb or an action.`,
				"the method has no Short.",
				"the method has input and no Example.",
				"two methods have the name frob.",
				`"frob" is not a verb or an action.`,
			},
		},
		{
			name: "when the input breaks the rules of flags, docs and tags",
			group: command.Define("zzflags", "Flags",
				command.Uses[localDeps](),
				command.Method[string, greetOut]("get").Short("X").Handle(greet),
				command.Method[badFlags, greetOut]("list").Short("X").Example(""),
			),
			want: []string{
				"the input is string, not a struct.",
				"the handler is func(context.Context, command_test.localDeps, command_test.greetIn) (command_test.greetOut, error); " +
					"it must be func(context.Context, command_test.localDeps, string) (command_test.greetOut, error).",
				"input field NoJSON has no json name.",
				"input field Renamed has the flag tag other: its flag is its json name.",
				"input field Routed has the in tag path: a local command has no route.",
				"input field Wide is int64, which no flag holds.",
				"input field bare has no doc.",
				"input field bare has no validate tag.",
				"the method has no handler.",
			},
		},
		{
			name: "when a rule has no Fails, or a Fails is not the input",
			group: command.Define("zzfails", "Fails",
				command.Uses[localDeps](),
				command.Method[greetIn, greetOut]("get").Short("X").Example("--name a").
					Fails("not the input", "x").
					Handle(greet),
			),
			want: []string{
				"input field Inner is command_test.inner, which no flag holds.",
				"input field inner has no doc.",
				"input field inner has no validate tag.",
				`the .Fails input "not the input" is string, not command_test.greetIn.`,
				"no .Fails input breaks required on Name.",
				"no .Fails input breaks max=5 on Name.",
				"no .Fails input breaks oneof=dry warm on Tone.",
				"no .Fails input breaks zzlocal on Tags.",
			},
		},
		{
			name: "when a method has a handler and the group Uses no type",
			group: command.Define("zzhandler", "Handler",
				command.Method[struct{}, greetOut]("get").Short("X").
					Handle(func(context.Context, localDeps, struct{}) (greetOut, error) {
						return greetOut{}, nil
					}),
			),
			want: []string{"the group Uses no type."},
		},
	}

	seen := map[string]bool{}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got []string
			for _, p := range command.Validate(tc.group) {
				got = append(got, p.Message)
				seen[p.Rule] = true
				s.Equal(command.Checks.Fix(p.Rule), p.Fix, "every rule has a check")
				s.NotEmpty(p.Fix)
				s.Equal(p.Message+" Fix: "+p.Fix, p.String())
				s.Equal(strings.Join(tc.group.CommandPath(), " "), p.Group)
			}

			s.Equal(tc.want, got)
		})
	}

	// The rules of the output apply to the registered groups: a renderer
	// with no Because, a tag the output does not know, and a method that
	// answers nothing with no past tense.
	out := command.Define("zzshape", "Output",
		command.Uses[localDeps](),
		command.Method[struct{}, greetOut]("list").Short("X").
			Handle(func(context.Context, localDeps, struct{}) (greetOut, error) {
				return greetOut{}, nil
			}),
		command.Method[struct{}, badTags]("get").Short("X").
			Handle(func(context.Context, localDeps, struct{}) (badTags, error) {
				return badTags{}, nil
			}),
		command.Method[struct{}, struct{}]("frob").Short("X").
			Handle(func(context.Context, localDeps, struct{}) (struct{}, error) {
				return struct{}{}, nil
			}),
	)
	command.AddCommands(out)
	command.Render(out, "list", func(*cobra.Command, struct{}, greetOut) {})
	var got []string
	for _, p := range command.ValidateOutput() {
		if p.Group != "zzshape" {
			continue
		}
		got = append(got, p.Message)
		seen[p.Rule] = true
		s.Equal(command.Checks.Fix(p.Rule), p.Fix, "every rule has a check")
	}
	s.Equal([]string{
		"the renderer has no Because.",
		`badTags.Size has render:"nope", which the output does not know.`,
		`badTags.Name has render:"short=0", which the output does not know.`,
		"badTags.Name has an empty tag and is not a list.",
		"badTags.Name has a title tag: it goes on the field _ struct{}.",
		`"frob" has no past tense.`,
	}, got)

	// Each check is a rule that Validate applies, and each rule has a
	// check: the cases above trigger every one.
	checked := make([]string, 0, len(command.Checks))
	for _, c := range command.Checks {
		checked = append(checked, c.Rule)
		s.NotEmpty(c.Requires, c.Rule)
	}
	triggered := make([]string, 0, len(seen))
	for r := range seen {
		triggered = append(triggered, r)
	}
	s.ElementsMatch(
		checked,
		triggered,
		"each check is triggered by a case, and each rule has a check",
	)
}

func TestValidatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ValidatePublicTestSuite))
}
