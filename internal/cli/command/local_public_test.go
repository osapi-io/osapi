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
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// localDeps is what the methods of a test group use.
type localDeps struct {
	Prefix string
}

// greetIn is the input of greet: each field is a flag.
type greetIn struct {
	Name  string   `json:"name"  validate:"required,max=5"           doc:"Who to greet"`
	Tone  string   `json:"tone"  validate:"omitempty,oneof=dry warm" doc:"How to greet"`
	Tags  []string `json:"tags"  validate:"omitempty,dive,zzlocal"   doc:"Tags of the greeting"`
	Inner inner    `json:"inner"`
}

type inner struct {
	Code string `json:"code" validate:"omitempty,len=2"`
}

type greetOut struct {
	Text string `json:"text"`
}

// badOut is an output that JSON cannot encode.
type badOut struct {
	C chan int `json:"c"`
}

func greet(
	_ context.Context,
	d localDeps,
	in greetIn,
) (greetOut, error) {
	if in.Name == "fail" {
		return greetOut{}, errors.New("greet failed")
	}
	if in.Name == "bad" {
		return greetOut{Text: d.Prefix + in.Name}, command.Fail()
	}

	return greetOut{Text: d.Prefix + in.Name}, nil
}

func greetBad(
	_ context.Context,
	_ localDeps,
	_ greetIn,
) (badOut, error) {
	return badOut{}, nil
}

func greetEmpty(
	_ context.Context,
	_ struct{},
	in greetIn,
) (greetOut, error) {
	return greetOut{Text: in.Name}, nil
}

func greetInt(
	_ context.Context,
	n int,
	in greetIn,
) (greetOut, error) {
	return greetOut{Text: strings.Repeat(in.Name, n+1)}, nil
}

type LocalPublicTestSuite struct {
	suite.Suite
}

func (s *LocalPublicTestSuite) SetupSuite() {
	rule.Declare(rule.Custom{
		Tag:     "zzlocal",
		Check:   func(v string) bool { return v != "bad" },
		Schema:  rule.None,
		Message: "is not a good tag",
	})
}

func (s *LocalPublicTestSuite) TestDefine() {
	m := command.Method[greetIn, greetOut]("get").
		Short("Greet someone").
		Long("It greets.").
		Example("--name ann", "Greet Ann").
		Sub("hello").
		Fails("no name", greetIn{}).
		Handle(greet)
	g := command.Define("zzdefine", "Greetings",
		command.Uses[localDeps](),
		command.Under("zzparent", "zzchild"),
		m,
	)

	s.Equal("zzdefine", g.GroupName())
	s.Equal("Greetings", g.GroupDoc())
	s.Empty(g.GroupNotes())
	s.Empty(g.GroupLong())
	s.Empty(g.SubDoc("roles"))
	documented := command.Define("zzdoc", "Doc",
		command.Long("More."),
		command.SubDoc("roles", "Read the roles"),
		command.SubDoc("keys", "Read the keys"),
	)
	s.Equal("Read the roles", documented.SubDoc("roles"))
	s.Equal("More.", documented.GroupLong())
	s.Equal("Read the keys", documented.SubDoc("keys"))
	s.Equal([]string{"zzparent", "zzchild", "zzdefine"}, g.CommandPath())
	s.Equal([]string{"zzparent", "zzchild"}, g.Parents, "CommandPath does not change Parents")
	s.Equal(reflect.TypeFor[localDeps](), g.Deps)
	s.Equal([]command.Command{m}, g.Commands())
	s.Empty(g.CommandPermission(m))
	s.Same(m, g.Method("hello get"), "a method is its Sub and its name")
	s.Nil(g.Method("get"))
	s.Nil(g.Method("nosuch"))

	s.Equal("hello get", m.MethodName())
	s.Equal("get", m.CommandName())
	s.Equal(reflect.TypeFor[greetIn](), m.Input())
	s.Equal("hello", m.SubPath())
	s.Equal("Greet someone", m.CommandShort())
	s.Equal("It greets.", m.CommandLong())
	s.Equal([]command.Example{{Flags: "--name ann", Doc: "Greet Ann"}}, m.Examples())
	s.Empty(m.PageNotes())
	s.Empty(m.PageResult())
	s.Equal([]command.Refusal{{Doc: "no name", Input: greetIn{}}}, m.Refusals())
	s.NotNil(m.Handler())
}

func (s *LocalPublicTestSuite) TestRegistry() {
	a := command.Register(command.Define("zzregb", "B"))
	b := command.Register(command.Define("zzrega", "A", command.Under("zzup")))

	s.Same(a, command.Lookup("zzregb"))
	s.Same(b, command.Lookup("zzup zzrega"))
	s.Nil(command.Lookup("zzrega"))
	all := command.All()
	ia, ib := -1, -1
	for i, g := range all {
		switch g {
		case a:
			ia = i
		case b:
			ib = i
		}
	}
	s.Less(ia, ib, "All is by command path")
	// Two flat groups under one parent have one command path: their path
	// orders them.
	y := command.Register(command.Define("zzflaty", "Y", command.Flat(), command.Under("zzflat")))
	x := command.Register(command.Define("zzflatx", "X", command.Flat(), command.Under("zzflat")))
	ix, iy := -1, -1
	for i, g := range command.All() {
		switch g {
		case x:
			ix = i
		case y:
			iy = i
		}
	}
	s.Less(ix, iy, "a tie of command path is by path")
	s.PanicsWithValue(`command: group "zzregb" registered twice`, func() {
		command.Register(command.Define("zzregb", "B again"))
	})
}

func (s *LocalPublicTestSuite) TestProvide() {
	g := command.Define("zzprovide", "Provided", command.Uses[localDeps]())
	s.PanicsWithValue(
		"command: zzprovide uses command_test.localDeps, and Provide gives int",
		func() { command.Provide(g, func() int { return 1 }) },
	)
	s.NotPanics(func() { command.Provide(g, func() localDeps { return localDeps{} }) })
}

func (s *LocalPublicTestSuite) TestRun() {
	var printed string
	s.Require().NoError(command.AttachLocal(&cobra.Command{Use: "osapi"}, command.Deps{
		JSON:      func() bool { return false },
		PrintJSON: func(body []byte) { printed = string(body) },
	}))

	provided := command.Define("zzrun", "Run",
		command.Uses[localDeps](),
		command.Method[greetIn, greetOut]("get").Handle(greet),
		command.Method[greetIn, badOut]("bad").Handle(greetBad),
	)
	command.Provide(provided, func() localDeps { return localDeps{Prefix: "hi "} })
	notGiven := command.Define("zzrun", "Run",
		command.Uses[localDeps](),
		command.Method[greetIn, greetOut]("get").Handle(greet),
	)
	empty := command.Define("zzrun", "Run",
		command.Uses[struct{}](),
		command.Method[greetIn, greetOut]("get").Handle(greetEmpty),
	)
	notStruct := command.Define("zzrun", "Run",
		command.Uses[int](),
		command.Method[greetIn, greetOut]("get").Handle(greetInt),
	)
	noUses := command.Define("zzrun", "Run",
		command.Method[greetIn, greetOut]("get").Handle(greetEmpty),
	)
	secretGroup := command.Define("zzrun", "Run",
		command.Uses[struct{}](),
		command.Method[secretIn, greetOut]("get").
			Handle(func(context.Context, struct{}, secretIn) (greetOut, error) {
				return greetOut{}, nil
			}),
	)

	tests := []struct {
		name    string
		group   *command.LocalGroup
		method  string
		in      any
		ctx     context.Context
		want    string
		wantErr string
		usage   bool
	}{
		{
			name:   "when the input is valid, the output prints",
			group:  provided,
			method: "get",
			in:     greetIn{Name: "ann"},
			want:   `{"text":"hi ann"}`,
		},
		{
			name:   "when the command has a context, the handler gets it",
			group:  provided,
			method: "get",
			in:     greetIn{Name: "bob"},
			ctx:    context.Background(),
			want:   `{"text":"hi bob"}`,
		},
		{
			name:   "when the Deps has no field, its zero value is given",
			group:  empty,
			method: "get",
			in:     greetIn{Name: "cy"},
			want:   `{"text":"cy"}`,
		},
		{
			name:   "when the Deps is not a struct, its zero value is given",
			group:  notStruct,
			method: "get",
			in:     greetIn{Name: "cy"},
			want:   `{"text":"cy"}`,
		},
		{
			name:    "when a secret is a reference, no vault resolves it",
			group:   secretGroup,
			method:  "get",
			in:      secretIn{Token: "secret://default/token"},
			wantErr: `secret://default/token: the vault "default" does not exist`,
		},
		{
			name:    "when the method is not in the group",
			group:   provided,
			method:  "nosuch",
			in:      greetIn{Name: "ann"},
			wantErr: "the zzrun group has no method nosuch",
		},
		{
			name:    "when a required flag is empty, it is a usage error",
			group:   provided,
			method:  "get",
			in:      greetIn{},
			wantErr: "--name is required",
			usage:   true,
		},
		{
			name:    "when a value is not one of a list, it says the list",
			group:   provided,
			method:  "get",
			in:      greetIn{Name: "ann", Tone: "loud"},
			wantErr: `--tone: "loud" is not one of dry, warm`,
			usage:   true,
		},
		{
			name:    "when a value breaks a custom rule, it has the rule's message",
			group:   provided,
			method:  "get",
			in:      greetIn{Name: "ann", Tags: []string{"ok", "bad"}},
			wantErr: `--tags: "bad" is not a good tag`,
			usage:   true,
		},
		{
			name:    "when values break two rules, it names both",
			group:   provided,
			method:  "get",
			in:      greetIn{Name: "annabel", Inner: inner{Code: "x"}},
			wantErr: `--name: "annabel" breaks max=5; --Inner.Code: "x" breaks len=2`,
			usage:   true,
		},
		{
			name:    "when the input is not a struct, it is an error",
			group:   provided,
			method:  "get",
			in:      "ann",
			wantErr: "validator: (nil string)",
		},
		{
			name:    "when the Deps has fields and nothing gives it",
			group:   notGiven,
			method:  "get",
			in:      greetIn{Name: "ann"},
			wantErr: "the zzrun group's command_test.localDeps is not given",
		},
		{
			name:    "when the group Uses no type",
			group:   noUses,
			method:  "get",
			in:      greetIn{Name: "ann"},
			wantErr: "the zzrun group Uses no type",
		},
		{
			name:    "when the handler fails, its error is the command's",
			group:   provided,
			method:  "get",
			in:      greetIn{Name: "fail"},
			wantErr: "greet failed",
		},
		{
			name:    "when the handler fails its output, the output prints and the command fails",
			group:   provided,
			method:  "get",
			in:      greetIn{Name: "bad"},
			want:    `{"text":"hi bad"}`,
			wantErr: "the command failed: its output says why",
		},
		{
			name:    "when the output does not encode",
			group:   provided,
			method:  "bad",
			in:      greetIn{Name: "ann"},
			wantErr: "encode zzrun bad",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			printed = ""
			cmd := &cobra.Command{}
			if tc.ctx != nil {
				cmd.SetContext(tc.ctx)
			}

			err := command.Run(cmd, tc.group, tc.method, tc.in)

			if tc.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.wantErr)
				var ue *command.UsageError
				s.Equal(tc.usage, errors.As(err, &ue))
				s.Equal(tc.want, printed)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, printed)
		})
	}
}

func (s *LocalPublicTestSuite) TestCheck() {
	s.NoError(command.Check(context.Background(), greetIn{Name: "ann"}))
	s.EqualError(command.Check(context.Background(), greetIn{}), "--name is required")
}

func (s *LocalPublicTestSuite) TestFlagName() {
	t := reflect.TypeFor[struct {
		SigningKey string `json:"signing_key,omitempty"`
		Plain      string
	}]()

	s.Equal("signing-key", command.FlagName(t.Field(0)))
	s.Equal("Plain", command.FlagName(t.Field(1)))
}

func TestLocalPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LocalPublicTestSuite))
}

// secretIn is an input with a secret, which a command resolves itself.
type secretIn struct {
	Token secret.Value `json:"token" validate:"omitempty" doc:"The token"`
}
