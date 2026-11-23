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

	"github.com/osapi-io/osapi/internal/atomicfile"
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
	if _, err := n.h.FileState.KVPut(
		ctx,
		n.h.FileStateBucket,
		notePath,
		[]byte(in.Text),
	); err != nil {
		return noteOut{}, err
	}
	if in.Mode != "nowrite" {
		_ = n.h.FS.MkdirAll("/etc", 0o755)
		_ = atomicfile.Write(ctx, n.h.FS, notePath, []byte(in.Text), 0o644)
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

// hiddenIn has a rule on a field that JSON leaves out, so its refusal breaks
// the rule as a Go value and not once it is sent.
type hiddenIn struct {
	Text   string `json:"text" validate:"omitempty,max=3"`
	Secret string `json:"-"    validate:"omitempty,max=1"`
}

func readHidden(context.Context, note, hiddenIn) (noteOut, error) {
	return noteOut{}, errors.New("refused")
}

func getNote(
	ctx context.Context,
	n note,
	in getIn,
) (noteOut, error) {
	if in.Strict {
		entry, err := n.h.FileState.KVGet(ctx, n.h.FileStateBucket, notePath)
		if err != nil {
			return noteOut{}, errors.New("not managed")
		}
		if string(entry.Value()) == "{}" {
			panic("trusted a damaged record")
		}
	}
	// A read runs its commands through gohai's runner, as a collector does.
	out, err := n.h.Sources().Exec.Execute(ctx, "cat", notePath)

	return noteOut{Text: string(out)}, err
}

// rawRead reads the note with h.Exec, outside gohai.
func rawRead(
	ctx context.Context,
	n note,
	_ getIn,
) (noteOut, error) {
	out, err := n.h.Exec.RunCmd(ctx, "cat", []string{notePath})

	return noteOut{Text: out}, err
}

// shelfIn is the input of a parent's method, which a child's After names.
type shelfIn struct {
	Name string `json:"name"`
}

// addShelf is the parent's method: it fails for the name "fail".
func addShelf(
	_ context.Context,
	_ note,
	in shelfIn,
) (noteOut, error) {
	if in.Name == "fail" {
		return noteOut{}, errors.New("no shelf")
	}

	return noteOut{Text: in.Name, Changed: true}, nil
}

// shelf is a parent behavior, whose provider fails to build with err.
func shelf(
	err error,
) *behavior.Behavior {
	return behavior.Define("shelf", "A shelf", behavior.Node, onDebian(err),
		behavior.Write("add", addShelf))
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
	if _, err := p.h.Exec.RunPrivilegedCmd(
		ctx,
		"apt-get",
		[]string{"install", "--", in.Name},
	); err != nil {
		return pkgOut{}, err
	}

	return pkgOut{Changed: true}, nil
}

// installByFile is installPkg with the state in a file that the install
// writes, as /etc/group is after groupadd.
func installByFile(
	ctx context.Context,
	p pkg,
	in pkgIn,
) (pkgOut, error) {
	if data, _ := p.h.FS.ReadFile("/var/lib/pkg/" + in.Name); string(data) == "installed" {
		return pkgOut{}, nil
	}
	if _, err := p.h.Exec.RunPrivilegedCmd(
		ctx,
		"apt-get",
		[]string{"install", "--", in.Name},
	); err != nil {
		return pkgOut{}, err
	}

	return pkgOut{Changed: true}, nil
}

type CheckPublicTestSuite struct {
	suite.Suite
}

// TestExistsAfter: a file that a command writes, declared with ExistsAfter,
// holds its later content once the command runs, so a second run of the
// example reports changed: false.
func (s *CheckPublicTestSuite) TestExistsAfter() {
	tests := []struct {
		name  string
		given []behavior.HostOption
		want  []string
	}{
		{
			name: "when ExistsAfter says the install writes the file",
			given: []behavior.HostOption{
				behavior.Exists("/var/lib/pkg/tree", "not-installed"),
				behavior.ExistsAfter("apt-get install -- tree", "/var/lib/pkg/tree", "installed"),
			},
		},
		{
			name: "when the file is the same after the install",
			given: []behavior.HostOption{
				behavior.Exists("/var/lib/pkg/tree", "not-installed"),
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
				"pkgfile",
				"Packages",
				behavior.Node,
				behavior.Platform(
					"debian",
					func(h host.Host) (pkg, error) { return pkg{h: h}, nil },
				),
				behavior.Sudo("apt-get"),
				behavior.Write("install", installByFile).Describe("Install one").
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
			name:     "when a method runs only on a platform its provider does not serve",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).Describe("Get it").On(behavior.PlatformDarwin).
					Example("A wrong output it does not check", getIn{}, noteOut{Text: "nope"},
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when an example gives a symbolic link",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).Describe("Get it").
					Example("The note through a link", getIn{}, noteOut{Text: "hello"},
						behavior.Links("/etc/note.link", notePath),
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when an example runs only on its platform",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).Describe("Get it").
					Example("The note on debian", getIn{}, noteOut{Text: "hello"},
						behavior.OnPlatform("debian"),
						behavior.Prints("cat /etc/note", "hello")).
					Example("A wrong output for a Mac, not run here", getIn{}, noteOut{Text: "nope"},
						behavior.OnPlatform(behavior.PlatformDarwin),
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when no example runs on a platform the method serves",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).Describe("Get it").
					Example("Only for a Mac", getIn{}, noteOut{Text: "hello"},
						behavior.OnPlatform(behavior.PlatformDarwin),
						behavior.Prints("cat /etc/note", "hello")),
			},
			want: []string{"note get on debian: no example runs on this platform"},
		},
		{
			name:     "when the programs are declared by a behavior it requires",
			platform: onDebian(nil),
			programs: []behavior.Option{
				behavior.Writes("/etc/note"),
				behavior.Requires(behavior.Define(
					"tools", "Tools", behavior.Node, behavior.Internal, onDebian(nil),
					behavior.Sudo("apply"), behavior.Runs("cat"), behavior.KeepsState,
				)),
			},
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
				"keeps state in the file-state bucket but does not declare behavior.KeepsState",
				"writes /etc/note but does not declare it with behavior.Writes",
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
			name:     "when After names a method of the parent",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Of(shelf(nil)),
				behavior.Read("get", getNote).
					Example("The note", getIn{}, noteOut{Text: "hello"},
						behavior.After(shelfIn{Name: "top"}),
						behavior.Prints("cat /etc/note", "hello")),
			},
		},
		{
			name:     "when the parent's method that After names fails",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Of(shelf(nil)),
				behavior.Read("get", getNote).
					Example("The note", getIn{}, noteOut{}, behavior.After(shelfIn{Name: "fail"})),
			},
			want: []string{"After add failed: no shelf"},
		},
		{
			name:     "when the parent's provider cannot be built",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Of(shelf(errors.New("no socket"))),
				behavior.Read("get", getNote).
					Example("The note", getIn{}, noteOut{}, behavior.After(shelfIn{Name: "top"})),
			},
			want: []string{"no provider of shelf on the fake host"},
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
				// The method's validate rules refuse it before the method
				// runs, as the agent does: the provider need not check it.
				behavior.Read("hidden", readHidden).
					Fails("A text over 3 characters", hiddenIn{Text: "abcd"}, "'max' tag").
					Fails("A text the provider refuses as well", hiddenIn{Text: "abcd"}, "refused"),
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
			name:     "when a refusal breaks a rule that the agent's check does not see",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("hidden", readHidden).
					Fails("A secret over 1 character", hiddenIn{Secret: "ab"}, "refused"),
			},
			want: []string{
				"refusal \"A secret over 1 character\": the input breaks Secret:max, and Method.Check passed it",
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
				programs = []behavior.Option{
					behavior.Sudo("apply"), behavior.Runs("cat"),
					behavior.Writes("/etc/note"), behavior.KeepsState,
				}
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

// TestCheckGohai: a read that runs a command outside gohai fails, unless it
// says why with NotAFact.
func (s *CheckPublicTestSuite) TestCheckGohai() {
	tests := []struct {
		name string
		m    *behavior.Method
		want []string
	}{
		{
			name: "when a read runs a command outside gohai",
			m:    behavior.Read("get", rawRead),
			want: []string{"a read runs cat outside gohai"},
		},
		{
			name: "when the read says what it reads is not a fact",
			m:    behavior.Read("get", rawRead).NotAFact("a note is not a fact"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("note", "Notes", behavior.Node, onDebian(nil),
				behavior.Runs("cat"),
				tc.m.Describe("Get it").
					Example("The note", getIn{}, noteOut{Text: "hi"},
						behavior.Prints("cat "+notePath, "hi")),
			)

			got := behaviortest.Check(b)

			s.Require().Len(got, len(tc.want), strings.Join(got, "\n"))
			for i, want := range tc.want {
				s.Contains(got[i], want)
			}
		})
	}
}

// TestCheckObjects: a provider reads the objects that the example stores,
// and an object the example does not store is not found.
func (s *CheckPublicTestSuite) TestCheckObjects() {
	read := func(ctx context.Context, p pkg, in pkgIn) (pkgOut, error) {
		content, contentType, err := p.h.Objects.Object(ctx, in.Name)
		if err != nil {
			return pkgOut{}, err
		}

		return pkgOut{Changed: string(content) == "x" && contentType == "raw"}, nil
	}
	b := behavior.Define(
		"obj",
		"Objects",
		behavior.Node,
		behavior.Platform("debian", func(h host.Host) (pkg, error) { return pkg{h: h}, nil }),
		behavior.Read("get", read).Describe("Get one").
			Example("A stored object", pkgIn{Name: "a"}, pkgOut{Changed: true},
				behavior.Stored("a", "x", "raw")).
			Fails("An object the store does not have", pkgIn{Name: "b"}, "not found"),
	)

	s.Empty(behaviortest.Check(b))
}

// TestCheckProgramState: a write whose state a program keeps passes the
// second-run check only when the example says what the program prints once
// the write has run.
func (s *CheckPublicTestSuite) TestCheckProgramState() {
	tests := []struct {
		name   string
		given  []behavior.HostOption
		always bool
		want   []string
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
		{
			name: "when an action changes the host each time it runs",
			given: []behavior.HostOption{
				behavior.Prints("dpkg-query -W -- tree", "not-installed"),
			},
			always: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := behavior.Write("install", installPkg)
			if tc.always {
				m = behavior.Action("install", installPkg).Always()
			}
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
				m.Describe("Install one").
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
