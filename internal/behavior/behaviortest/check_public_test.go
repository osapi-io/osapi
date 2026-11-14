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

package behaviortest_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

const notePath = "/etc/note"

// note keeps a note in a file, applies it with a command and records it in
// the store. Its input's Mode makes it misbehave in one way each.
type note struct{ h host.Host }

type noteIn struct {
	Text string `json:"text"`
	Mode string `json:"mode"`
}

type noteOut struct {
	Text    string `json:"text"`
	Changed bool   `json:"changed"`
}

type getIn struct {
	Strict bool `json:"strict"`
}

func (n note) set(
	ctx context.Context,
	in noteIn,
) (noteOut, error) {
	if in.Mode == "fail" {
		return noteOut{}, errors.New("refused")
	}
	if old, err := n.h.FS.ReadFile(notePath); err == nil && string(old) == in.Text &&
		in.Mode != "always" {
		return noteOut{Text: in.Text}, nil
	}
	if _, err := n.h.Exec.RunPrivilegedCmd(ctx, "apply", []string{notePath}); err != nil {
		if in.Mode == "panic" {
			panic("unchecked failure")
		}

		return noteOut{}, err
	}
	if _, err := n.h.FileState.Put(ctx, notePath, []byte(in.Text)); err != nil {
		return noteOut{}, err
	}
	if in.Mode != "nowrite" {
		_ = n.h.FS.MkdirAll("/etc", 0o755)
		_ = n.h.FS.WriteFile(notePath, []byte(in.Text), 0o644)
	}

	return noteOut{Text: in.Text, Changed: true}, nil
}

func setNote(
	ctx context.Context,
	n note,
	in noteIn,
) (noteOut, error) {
	return n.set(ctx, in)
}

func getNote(
	ctx context.Context,
	n note,
	in getIn,
) (noteOut, error) {
	if in.Strict {
		entry, err := n.h.FileState.Get(ctx, notePath)
		if err != nil {
			return noteOut{}, errors.New("not managed")
		}
		if string(entry.Value()) == "{}" {
			panic("trusted a damaged record")
		}
	}
	out, err := n.h.Exec.RunCmd(ctx, "cat", []string{notePath})

	return noteOut{Text: out}, err
}

func onDebian(
	err error,
) behavior.Option {
	return behavior.Platform("debian", func(h host.Host) (note, error) {
		return note{h: h}, err
	})
}

// pkg installs a package, which only the package manager's query shows:
// state a program keeps rather than a file.
type pkg struct{ h host.Host }

type pkgIn struct {
	Name string `json:"name"`
}

type pkgOut struct {
	Changed bool `json:"changed"`
}

func installPkg(
	ctx context.Context,
	p pkg,
	in pkgIn,
) (pkgOut, error) {
	out, err := p.h.Exec.RunCmd(ctx, "dpkg-query", []string{"-W", "--", in.Name})
	if err != nil {
		return pkgOut{}, err
	}
	if out == "installed" {
		return pkgOut{}, nil
	}
	if _, err := p.h.Exec.RunPrivilegedCmd(ctx, "apt-get", []string{"install", "--", in.Name}); err != nil {
		return pkgOut{}, err
	}

	return pkgOut{Changed: true}, nil
}

type CheckPublicTestSuite struct {
	suite.Suite
}

func (s *CheckPublicTestSuite) TestCheck() {
	wrote := behavior.Wrote(notePath, "hello")
	tests := []struct {
		name     string
		platform behavior.Option
		// programs declares what the provider runs; nil is Sudo("apply")
		// and Runs("cat").
		programs []behavior.Option
		methods  []behavior.Option
		want     []string
	}{
		{
			name:     "when every example holds",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).Describe("Set it").
					Example("A note", noteIn{Text: "hello"}, noteOut{Text: "hello", Changed: true},
						behavior.Exists("/etc/other", "x"),
						wrote, behavior.Ran("apply", notePath)),
				behavior.Read("get", getNote).Describe("Get it").
					Example("The note", getIn{}, noteOut{Text: "hello"},
						behavior.After(noteIn{Text: "hello"}),
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when the programs are declared by a behavior it requires",
			platform: onDebian(nil),
			programs: []behavior.Option{behavior.Requires(behavior.Define(
				"tools", "Tools", behavior.Node, behavior.Internal, onDebian(nil),
				behavior.Sudo("apply"), behavior.Runs("cat"),
			))},
			methods: []behavior.Option{
				behavior.Write("set", setNote).Describe("Set it").
					Example("A note", noteIn{Text: "hello"}, noteOut{Text: "hello", Changed: true},
						wrote, behavior.Ran("apply", notePath)),
				behavior.Read("get", getNote).Describe("Get it").
					Example("The note", getIn{}, noteOut{Text: "hello"},
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when it runs programs it does not declare",
			platform: onDebian(nil),
			programs: []behavior.Option{},
			methods: []behavior.Option{
				behavior.Write("set", setNote).Describe("Set it").
					Example("A note", noteIn{Text: "hello"}, noteOut{Text: "hello", Changed: true},
						wrote, behavior.Ran("apply", notePath)),
				behavior.Read("get", getNote).Describe("Get it").
					Example("The note", getIn{}, noteOut{Text: "hello"},
						behavior.Prints("cat /etc/note", "hello")),
			},
			want: []string{
				"runs apply with sudo but does not declare it with behavior.Sudo",
				"runs cat but does not declare it with behavior.Runs",
			},
		},
		{
			name:     "when the provider cannot be built",
			platform: onDebian(errors.New("no socket")),
			methods: []behavior.Option{
				behavior.Read("get", getNote).Example("The note", getIn{}, noteOut{}),
			},
			want: []string{"no provider on the fake host"},
		},
		{
			name:     "when After names an input no method takes",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).
					Example("The note", getIn{}, noteOut{}, behavior.After(42)),
			},
			want: []string{"no method takes int"},
		},
		{
			name:     "when After fails",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).Example("Unused", noteIn{Text: "hello"},
					noteOut{Text: "hello", Changed: true}, wrote),
				behavior.Read("get", getNote).
					Example("The note", getIn{}, noteOut{}, behavior.After(noteIn{Mode: "fail"})),
			},
			want: []string{"After set failed: refused"},
		},
		{
			name:     "when the method fails",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).
					Example("A note", noteIn{Mode: "fail"}, noteOut{}, wrote),
			},
			want: []string{"failed: refused"},
		},
		{
			name:     "when the output, the effects and a second run are wrong",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).
					Example("A note", noteIn{Text: "hello", Mode: "always"}, noteOut{Text: "bye"},
						behavior.Wrote(notePath, "bye"),
						behavior.Wrote("/etc/missing", "x"),
						behavior.Removed(notePath),
						behavior.Ran("reload")),
			},
			want: []string{
				"output",
				"wrote /etc/note as",
				"did not write /etc/missing",
				"did not remove /etc/note",
				"did not run \"reload\"",
				"running it again reports changed: true",
			},
		},
		{
			name:     "when a write declares no effects",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).
					Example("A note", noteIn{Text: "hello"},
						noteOut{Text: "hello", Changed: true}),
			},
			want: []string{"a write declares no effects"},
		},
		{
			name:     "when a write changes nothing, it needs no effects",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).
					Example("The same note", noteIn{Text: "hello"}, noteOut{Text: "hello"},
						behavior.Exists(notePath, "hello")),
			},
		},
		{
			name:     "when the method refuses as declared",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).
					Fails("An unmanaged note", getIn{Strict: true}, "not managed"),
				behavior.Write("set", setNote).
					Fails("A refused note", noteIn{Mode: "fail"}, "refused"),
			},
		},
		{
			name:     "when a refusal succeeds, fails differently, or cannot be set up",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).
					Fails("Should fail", getIn{}, "not managed",
						behavior.Prints("cat /etc/note", "hello")).
					Fails("Wrong error", getIn{Strict: true}, "denied").
					Fails("Bad setup", getIn{}, "x", behavior.After(42)),
			},
			want: []string{
				"refusal \"Should fail\": succeeded with",
				"refusal \"Wrong error\": failed with",
				"refusal \"Bad setup\": no method takes int",
			},
		},
		{
			name:     "when a damaged record makes it panic",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).Example("Unused", noteIn{Text: "hello"},
					noteOut{Text: "hello", Changed: true}, wrote),
				behavior.Read("get", getNote).
					Example("The note", getIn{Strict: true}, noteOut{Text: "hello"},
						behavior.After(noteIn{Text: "hello"}),
						behavior.Prints("cat /etc/note", "hello")),
			},
			want: []string{
				"with store read 1 returning \"{}\": panicked: trusted a damaged record",
			},
		},
		{
			name:     "when a failing call panics",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Write("set", setNote).
					Example("A note", noteIn{Text: "hello", Mode: "panic"},
						noteOut{Text: "hello", Changed: true}, wrote),
			},
			want: []string{"failing: panicked: unchecked failure"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			programs := tc.programs
			if programs == nil {
				programs = []behavior.Option{behavior.Sudo("apply"), behavior.Runs("cat")}
			}
			opts := append([]behavior.Option{behavior.Node, tc.platform}, programs...)
			b := behavior.Define("note", "A note", append(opts, tc.methods...)...)

			problems := behaviortest.Check(b)

			s.Len(problems, len(tc.want), strings.Join(problems, "\n"))
			for i, want := range tc.want {
				if i < len(problems) {
					s.Contains(problems[i], want)
				}
			}
		})
	}
}

// recorder is a testing.TB that keeps the errors reported to it.
type recorder struct {
	testing.TB
	errs []string
}

func (r *recorder) Helper() {}

func (r *recorder) Error(args ...any) {
	r.errs = append(r.errs, args[0].(string))
}

func (s *CheckPublicTestSuite) TestRun() {
	b := behavior.Define("note", "A note", behavior.Node, onDebian(nil), behavior.Runs("cat"),
		behavior.Read("get", getNote).Example("The note", getIn{}, noteOut{Text: "x"},
			behavior.Prints("cat /etc/note", "hello")),
	)
	r := &recorder{}

	behaviortest.Run(r, b)

	s.Len(r.errs, 1)
	s.Contains(r.errs[0], "output")
}

// TestCheckProgramState: a write whose state a program keeps passes the
// second-run check only when the example says what the program prints once
// the write has run.
func (s *CheckPublicTestSuite) TestCheckProgramState() {
	tests := []struct {
		name  string
		given []behavior.HostOption
		want  []string
	}{
		{
			name: "when PrintsAfter says the query answers installed after the install",
			given: []behavior.HostOption{
				behavior.Prints("dpkg-query -W -- tree", "not-installed"),
				behavior.PrintsAfter(
					"apt-get install -- tree",
					"dpkg-query -W -- tree",
					"installed",
				),
			},
		},
		{
			name: "when the query answers the same every time",
			given: []behavior.HostOption{
				behavior.Prints("dpkg-query -W -- tree", "not-installed"),
			},
			want: []string{"running it again reports changed: true"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			opts := append(
				[]behavior.HostOption{behavior.Ran("apt-get", "install", "--", "tree")},
				tc.given...)
			b := behavior.Define(
				"pkg",
				"Packages",
				behavior.Node,
				behavior.Platform(
					"debian",
					func(h host.Host) (pkg, error) { return pkg{h: h}, nil },
				),
				behavior.Runs("dpkg-query"),
				behavior.Sudo("apt-get"),
				behavior.Write("install", installPkg).Describe("Install one").
					Example("Install tree", pkgIn{Name: "tree"}, pkgOut{Changed: true}, opts...),
			)

			got := behaviortest.Check(b)

			s.Require().Len(got, len(tc.want), strings.Join(got, "\n"))
			for i, want := range tc.want {
				s.Contains(got[i], want)
			}
		})
	}
}

func TestCheckPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CheckPublicTestSuite))
}
