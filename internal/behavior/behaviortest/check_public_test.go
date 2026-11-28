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
	"os"
	"path/filepath"
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

			s.Require().Len(got, len(tc.want))
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
			name:     "when a refusal runs only on its platform",
			platform: onDebian(nil),
			methods: []behavior.Option{
				behavior.Read("get", getNote).
					Fails("An unmanaged note", getIn{Strict: true}, "not managed",
						behavior.OnPlatform("debian")).
					Fails("A refusal for a Mac, not run here", getIn{}, "not managed",
						behavior.OnPlatform(behavior.PlatformDarwin),
						behavior.Prints("cat /etc/note", "hello")),
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

			s.Len(problems, len(tc.want))
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

			s.Require().Len(got, len(tc.want))
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

			s.Require().Len(got, len(tc.want))
			for i, want := range tc.want {
				s.Contains(got[i], want)
			}
		})
	}
}

func TestCheckPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CheckPublicTestSuite))
}

// fragile is a provider that panics on one input, so Check must report the
// panic and keep the rest of its problems.
type fragile struct {
	// runs counts the calls, for a panic on the second.
	runs *int
}

type fragileIn struct {
	Key string `json:"key"`
}

type fragileOut struct {
	Changed bool `json:"changed"`
}

func setFragile(
	_ context.Context,
	f fragile,
	in fragileIn,
) (fragileOut, error) {
	*f.runs++
	switch {
	case strings.Contains(in.Key, ".."):
		var keys []string
		_ = keys[1]
	case in.Key == "again" && *f.runs > 1:
		var keys []string
		_ = keys[1]
	}

	return fragileOut{}, nil
}

func (s *CheckPublicTestSuite) TestCheckReportsAPanic() {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "when a probe makes the provider panic, it is a problem",
			key:  "a",
			want: "panicked",
		},
		{
			name: "when the second run of a write panics, it is a problem",
			key:  "again",
			want: "running it again panicked",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("fragile", "Panics on one input", behavior.Node,
				behavior.Platform("debian", func(host.Host) (fragile, error) {
					return fragile{runs: new(int)}, nil
				}),
				behavior.Write("set", setFragile).Describe("Set it").
					Example("A set", fragileIn{Key: tc.key}, fragileOut{}),
			)

			var problems []string
			s.NotPanics(func() { problems = behaviortest.Check(b) })
			s.Contains(strings.Join(problems, "\n"), tc.want)
		})
	}
}

// files is a provider that changes /etc/w/x one way per op.
type files struct{ h host.Host }

type filesIn struct {
	Op string `json:"op"`
}

type filesOut struct {
	Changed bool `json:"changed"`
}

const (
	filesPath = "/etc/w/x"
	// linkPath is where a link goes: a new path, so the link is made.
	linkPath = "/etc/w/y"
)

func changeFile(
	ctx context.Context,
	p files,
	in filesIn,
) (filesOut, error) {
	fs := p.h.FS
	_ = fs.MkdirAll("/etc/w", 0o755)
	var err error
	switch in.Op {
	case "open":
		var f interface{ Close() error }
		f, err = fs.OpenFile(filesPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
		}
	case "rename":
		// /tmp-x is the example's own file, so only the rename is a write.
		err = fs.Rename("/tmp-x", filesPath)
	case "remove":
		err = fs.Remove(filesPath)
	case "chmod":
		err = fs.Chmod(filesPath, 0o600)
	case "truncate":
		err = fs.Truncate(filesPath, 0)
	case "symlink":
		err = fs.Symlink(filesPath, linkPath)
	case "link":
		err = fs.Link(filesPath, linkPath)
	case "chown":
		err = fs.Chown(filesPath, 0, 0)
	case "lchown":
		err = fs.Lchown(filesPath, 0, 0)
	case "chtimes":
		err = fs.Chtimes(filesPath, behaviortest.Now, behaviortest.Now)
	case "read":
		_, err = fs.ReadFile(filesPath)
	case "atomic":
		err = atomicfile.Write(ctx, fs, filesPath, []byte("x"), 0o644)
	default:
		err = fs.WriteFile(filesPath, []byte("x"), 0o644)
	}

	return filesOut{}, err
}

func (s *CheckPublicTestSuite) TestWrites() {
	tests := []struct {
		name   string
		op     string
		writes []string
		want   []string
	}{
		{
			name:   "a write under a declared pattern is fine",
			op:     "write",
			writes: []string{"/etc/w/*"},
		},
		{name: "any path is fine with *", op: "chmod", writes: []string{"*"}},
		{name: "a read is not a write", op: "read"},
		{name: "a write", op: "write", want: []string{"writes /etc/w/x"}},
		{name: "an open for writing", op: "open", want: []string{"writes /etc/w/x"}},
		{name: "a rename onto the path", op: "rename", want: []string{"writes /etc/w/x"}},
		{name: "a removal", op: "remove", want: []string{"writes /etc/w/x"}},
		{name: "a mode change", op: "chmod", want: []string{"writes /etc/w/x"}},
		{name: "a truncation", op: "truncate", want: []string{"writes /etc/w/x"}},
		{name: "a symlink at a path", op: "symlink", want: []string{"writes /etc/w/y"}},
		{name: "a hard link at a path", op: "link", want: []string{"writes /etc/w/y"}},
		{name: "an owner change", op: "chown", want: []string{"writes /etc/w/x"}},
		{name: "an owner change of a link", op: "lchown", want: []string{"writes /etc/w/x"}},
		{name: "a time change", op: "chtimes", want: []string{"writes /etc/w/x"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("files", "Files", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (files, error) {
					return files{h: h}, nil
				}),
				behavior.Writes(tc.writes...),
				behavior.Read("get", changeFile).Describe("Change it").
					Example("A change", filesIn{Op: tc.op}, filesOut{},
						// The example's own file is set up, not written.
						behavior.Exists(filesPath, "x"),
						behavior.Exists("/tmp-x", "x")),
			)

			var undeclared []string
			for _, p := range behaviortest.Check(b) {
				if strings.Contains(p, "does not declare it with behavior.Writes") {
					undeclared = append(undeclared, p)
				}
			}

			if tc.want == nil {
				s.Empty(undeclared)
				return
			}
			s.Require().NotEmpty(undeclared)
			s.Contains(undeclared[0], tc.want[0])
		})
	}
}

func (s *CheckPublicTestSuite) TestFileSpans() {
	tests := []struct {
		name string
		op   string
		want bool
	}{
		{name: "when a write goes through atomicfile it has its span", op: "atomic"},
		{name: "when a write goes through h.FS it has none", op: "write", want: true},
		{name: "when a removal goes through h.FS it has none", op: "remove", want: true},
		{name: "when an owner change goes through h.FS it has none", op: "chown", want: true},
		{name: "when a time change goes through h.FS it has none", op: "chtimes", want: true},
		{name: "when a read goes through h.FS it needs none", op: "read"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("files", "Files", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (files, error) {
					return files{h: h}, nil
				}),
				behavior.Writes("*"),
				behavior.Read("get", changeFile).Describe("Change it").
					Example("A change", filesIn{Op: tc.op}, filesOut{},
						behavior.Exists(filesPath, "x")),
			)

			var spanless []string
			for _, p := range behaviortest.Check(b) {
				if strings.Contains(p, "so the trace has no file span") {
					spanless = append(spanless, p)
				}
			}

			if !tc.want {
				s.Empty(spanless)
				return
			}
			s.Require().Len(spanless, 1)
			s.Contains(spanless[0], "changes /etc/w/x through h.FS")
			s.Contains(spanless[0], "Fix: For a file that osapi manages, require managedfile")
		})
	}
}

func (s *CheckPublicTestSuite) TestDirectCalls() {
	tools := behavior.Define("tools", "Tools", behavior.Node, behavior.Internal,
		behavior.Write("set", setNote))
	b := behavior.Define("user", "A user of tools", behavior.Node, behavior.Requires(tools))
	method := tools.Methods[0].Func()
	pkg := method[:strings.LastIndex(method, ".")]
	fn := method[strings.LastIndex(method, ".")+1:]

	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "a call through behavior.Call is fine",
			src: "package user\n\nimport (\n\t\"context\"\n\n\tt \"" + pkg + "\"\n" +
				"\t\"github.com/osapi-io/osapi/internal/behavior\"\n)\n\n" +
				"func f() { _, _ = behavior.Call(context.Background(), nil, t." + fn + ", nil, nil) }\n" +
				// A call of a local function, and of a field, are not of a package.
				"func h() {}\n" +
				"func g() { h(); var x struct{ y struct{ z func() } }; x.y.z() }\n",
		},
		{
			name: "a direct call is reported with its place and the fix",
			src: "package user\n\nimport t \"" + pkg + "\"\n\n" +
				"func f() { _, _ = t." + fn + "(nil, nil, nil) }\n",
			want: []string{"user.go:5:", "calls t." + fn + " directly", "behavior.Call"},
		},
		{
			name: "a file that does not parse is reported",
			src:  "package user\n\nfunc {",
			want: []string{"read "},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			dir := s.T().TempDir()
			s.Require().NoError(os.WriteFile(filepath.Join(dir, "user.go"), []byte(tc.src), 0o644))
			s.Require().NoError(os.WriteFile(filepath.Join(dir, "user_test.go"),
				[]byte("package user\n\nfunc {"), 0o644))
			defer behaviortest.SetSourceDir(dir)()

			r := &recorder{}
			behaviortest.Run(r, b)
			// user has no methods, so no example calls tools: that is the
			// used rule, which chain_public_test.go covers.
			var errs []string
			for _, e := range r.errs {
				if !strings.Contains(e, "no example calls it") {
					errs = append(errs, e)
				}
			}

			if tc.want == nil {
				s.Empty(errs)
				return
			}
			s.Require().Len(errs, 1)
			for _, w := range tc.want {
				s.Contains(errs[0], w)
			}
		})
	}

	s.Run("a behavior that requires nothing is not read", func() {
		defer behaviortest.SetSourceDir("/nonexistent")()

		r := &recorder{}
		behaviortest.Run(r, behavior.Define("alone", "Alone", behavior.Node))
		s.Empty(r.errs)
	})
}
