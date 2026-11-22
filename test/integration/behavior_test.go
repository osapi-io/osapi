//go:build integration

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

package integration_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/gohai/pkg/gohai/family"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/command"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
)

// BehaviorSmokeSuite runs every registered behavior's generated commands
// against the live server and agent, from its definition, so a new behavior
// is exercised on a real host without a test written for it.
//
// Every read method with no key in its path runs always. Writes change the host, so
// they run only with OSAPI_INTEGRATION_WRITES=1 or
// OSAPI_INTEGRATION_WRITE_BEHAVIORS=1: each write example runs after the calls
// its example names with behavior.After, must report the example's changed,
// and run again must report changed: false.
type BehaviorSmokeSuite struct {
	suite.Suite
}

// commandLine is the generated command line for one method, with input's flags:
// the docs pages' command without the binary's name.
func commandLine(
	b *behavior.Behavior,
	m *behavior.Method,
	input any,
) []string {
	args := append(strings.Fields(b.Command(m))[1:], "--target", target, "--json")

	return append(args, command.Args(input)...)
}

// results runs one command and returns its result rows.
func (s *BehaviorSmokeSuite) results(
	args []string,
) []map[string]any {
	stdout, stderr, code := runCLI(args...)
	s.Require().Equal(0, code, "%s: %s", strings.Join(args, " "), stderr)

	var body struct {
		Results []map[string]any `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &body), stdout)
	for _, r := range body.Results {
		s.Contains([]any{"ok", "skipped"}, r["status"], "%s: %v", strings.Join(args, " "), r)
	}

	return body.Results
}

// TestPrograms: the image has every program a behavior declares for this
// platform, so a behavior's examples fail here for what it does, not for
// what is missing.
func (s *BehaviorSmokeSuite) TestPrograms() {
	for _, b := range behavior.All() {
		sudo, runs := b.Programs(behavior.PlatformName(family.Detect()))
		for _, p := range append(sudo, runs...) {
			_, err := exec.LookPath(p)
			s.NoError(err,
				"%s runs %s, which this host lacks: install its package on the CI "+
					"runner in .github/workflows/go-int.yml", b.Name, p)
		}
	}
}

func (s *BehaviorSmokeSuite) TestReads() {
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			// A read that names an item, or the item of a parent such as an
			// interface's routes, has no input that every host has.
			if m.Verb != behavior.VerbRead || len(m.PathKeys()) > 0 {
				continue
			}
			s.Run(b.Name+" "+m.Name, func() {
				s.results(commandLine(b, m, nil))
			})
		}
	}
}

func (s *BehaviorSmokeSuite) TestWrites() {
	if os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") != "1" {
		skipWrite(s.T())
	}

	s.refreshIndex()
	var examples []made
	defer func() { s.leftNothing(examples) }()

	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			if m.Verb != behavior.VerbWrite {
				continue
			}
			for _, ex := range m.Examples {
				s.Run(b.Name+" "+m.Name+" "+ex.Doc, func() {
					defer s.restore(b)()
					s.reset(b, ex.Input)
					// The example's items go again after it, also when it
					// fails, so the target is as the suite found it.
					examples = append(examples, madeBy(b, ex))
					defer s.remove(examples[len(examples)-1])
					s.store(ex.Given)
					for _, g := range ex.Given {
						if g.Kind != behavior.GivenCall {
							continue
						}
						for _, before := range b.Methods {
							if before.In == reflect.TypeOf(g.Input) {
								s.results(commandLine(b, before, g.Input))
							}
						}
					}

					want, _ := behavior.Changed(ex.Output)
					for _, r := range s.results(commandLine(b, m, ex.Input)) {
						if r["status"] == "ok" {
							s.Equal(want, r["changed"], "the example's changed")
						}
					}
					if m.AlwaysChanges {
						return
					}
					for _, r := range s.results(commandLine(b, m, ex.Input)) {
						if r["status"] == "ok" {
							s.Equal(false, r["changed"], "the same call again changes nothing")
						}
					}
				})
			}
		}
	}
}

// refreshIndex refreshes the package index of the target, as a person does
// before an install. The runner's lists are recent, but a lab VM's can be
// weeks old, and an install of a version that the mirror no longer serves
// fails with exit status 100. The refresh's own example can run after the
// package examples, so the suite runs it first.
func (s *BehaviorSmokeSuite) refreshIndex() {
	b := behavior.Lookup("updatepackages")
	if b == nil || b.Method("update") == nil {
		return
	}
	m := b.Method("update")
	s.results(commandLine(b, m, reflect.New(m.In).Elem().Interface()))
}

// made is what a write example can leave on the target: the items that its
// input and its Given calls name, which the behavior's delete removes, and
// the objects that it uploaded.
type made struct {
	b       *behavior.Behavior
	inputs  []any
	objects []string
}

// madeBy is what ex can leave.
func madeBy(
	b *behavior.Behavior,
	ex behavior.Example,
) made {
	m := made{b: b, inputs: []any{ex.Input}}
	for _, g := range ex.Given {
		switch g.Kind {
		case behavior.GivenCall:
			m.inputs = append(m.inputs, g.Input)
		case behavior.GivenObject:
			m.objects = append(m.objects, g.Path)
		}
	}

	return m
}

// remove deletes what m names: each item, with the behavior's delete, and
// each object, with osapi client file delete. An object that is not there
// is not an error.
func (s *BehaviorSmokeSuite) remove(
	m made,
) {
	for _, in := range m.inputs {
		s.reset(m.b, in)
	}
	for _, name := range m.objects {
		_, stderr, code := runCLI("client", "file", "delete", "--name", name)
		s.True(code == 0 || strings.Contains(stderr, "not found"),
			"delete object %s: %s", name, stderr)
	}
}

// leftNothing runs each delete of the write examples again, at the end of
// the suite, and requires that none changes the target: each example's
// remove left nothing behind.
func (s *BehaviorSmokeSuite) leftNothing(
	all []made,
) {
	for _, m := range all {
		for _, in := range m.inputs {
			for _, r := range s.deleteResults(m.b, in) {
				if r["status"] == "ok" {
					s.Equal(false, r["changed"],
						"%s %v was left on the target", m.b.Name, in)
				}
			}
		}
	}
}

// store uploads each object that the example stores with behavior.Stored,
// with osapi client file upload, as a person does before a deploy. An object
// that has the content already is unchanged.
func (s *BehaviorSmokeSuite) store(
	given []behavior.Given,
) {
	for _, g := range given {
		if g.Kind != behavior.GivenObject {
			continue
		}
		path := filepath.Join(s.T().TempDir(), "object")
		s.Require().NoError(os.WriteFile(path, []byte(g.Content), 0o600))
		_, stderr, code := runCLI("client", "file", "upload", "--name", g.Path,
			"--file", path, "--content-type", g.ContentType)
		s.Require().Equal(0, code, "upload %s: %s", g.Path, stderr)
	}
}

// reset starts a write example from the state the fake host gives it, with
// the item the example names absent. Each example on the fake host gets a
// new host, but here every example shares one: an earlier example can leave
// the item managed, and this one then changes nothing. Thus reset runs the
// behavior's delete with the fields of its input that the example's input
// carries: its key, the key of a parent, or a route's destination.
func (s *BehaviorSmokeSuite) reset(
	b *behavior.Behavior,
	input any,
) {
	s.deleteResults(b, input)
}

// deleteResults runs the behavior's delete with the fields of input that it
// takes, and returns its rows; none when the behavior has no delete or input
// cannot name the item.
func (s *BehaviorSmokeSuite) deleteResults(
	b *behavior.Behavior,
	input any,
) []map[string]any {
	del := b.Method("delete")
	if del == nil {
		return nil
	}
	data, err := json.Marshal(input)
	s.Require().NoError(err)
	var fields map[string]any
	s.Require().NoError(json.Unmarshal(data, &fields))
	keep := map[string]any{}
	for i := range del.In.NumField() {
		name := behavior.JSONName(del.In.Field(i))
		v, ok := fields[name]
		if !ok || v == nil || v == "" {
			// A delete that the example's input cannot name has nothing
			// to reset.
			if strings.Contains(del.In.Field(i).Tag.Get("validate"), "required") {
				return nil
			}

			continue
		}
		keep[name] = v
	}
	if len(keep) == 0 {
		return nil
	}
	key, err := json.Marshal(keep)
	s.Require().NoError(err)
	in := reflect.New(del.In)
	s.Require().NoError(json.Unmarshal(key, in.Interface()))
	return s.results(commandLine(b, del, in.Elem().Interface()))
}

// restore reads the item of a behavior that has no delete, such as the
// host's name, and returns a func that writes it back with update. Thus a
// write example on a live host, which renames the runner or a lab VM, leaves
// it as it found it. A behavior with a delete is reset with it instead.
func (s *BehaviorSmokeSuite) restore(
	b *behavior.Behavior,
) func() {
	var get, update *behavior.Method
	for _, m := range b.Methods {
		switch {
		case m.Name == "delete":
			return func() {}
		case m.Name == "get" && m.KeyField == "":
			get = m
		case m.Name == "update" && m.KeyField == "":
			update = m
		}
	}
	if get == nil || update == nil {
		return func() {}
	}

	// read is each host's item, as the input that would write it back.
	read := func() []any {
		var inputs []any
		for _, r := range s.results(commandLine(b, get, nil)) {
			if r["status"] != "ok" {
				continue
			}
			data, err := json.Marshal(r)
			s.Require().NoError(err)
			in := reflect.New(update.In)
			s.Require().NoError(json.Unmarshal(data, in.Interface()))
			inputs = append(inputs, in.Elem().Interface())
		}

		return inputs
	}
	before := read()

	return func() {
		for _, in := range before {
			s.results(commandLine(b, update, in))
		}
		s.Equal(before, read(), "%s is back as the example found it", b.Name)
	}
}

func TestBehaviorSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorSmokeSuite))
}
