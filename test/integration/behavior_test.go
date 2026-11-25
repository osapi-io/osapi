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
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/stretchr/testify/suite"

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

// call is a command line and what it reads from standard input.
type call struct {
	args  []string
	stdin string
}

// commandLine is the generated command line for one method, with input's flags:
// the docs pages' command without the binary's name. A secret of input with a
// value is the flag -, and the value is the call's standard input.
func commandLine(
	b *behavior.Behavior,
	m *behavior.Method,
	input any,
) call {
	args := append(strings.Fields(b.Command(m))[1:], "--target", target, "--json")

	return call{args: append(args, command.Args(input)...), stdin: command.Stdin(input)}
}

// results runs one command and returns its result rows.
func (s *BehaviorSmokeSuite) results(
	c call,
) []map[string]any {
	args := c.args
	stdout, stderr, code := runCLIStdin(c.stdin, args...)
	s.Require().Equal(0, code, "%s: %s", strings.Join(args, " "), stderr)

	var body struct {
		Results []map[string]any `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &body), stdout)
	skipIfCommandOff(s.T(), stdout)
	for _, r := range body.Results {
		s.Contains([]any{"ok", "skipped"}, r["status"], "%s: %v", strings.Join(args, " "), r)
	}

	return body.Results
}

// TestPrograms: each host the suite targets has every program a behavior
// declares for its platform, so a behavior's examples fail for what it does,
// not for what is missing. The agent finds its missing programs at start and
// reports them, so this checks the host the behaviors run on: the CI runner,
// or each lab VM.
func (s *BehaviorSmokeSuite) TestPrograms() {
	// The agent finds its missing programs again on each heartbeat, so a
	// program that TestMain's prepare installed, such as chrony, leaves the
	// list at the next one. Wait for it, for up to two heartbeats.
	var missing map[string][]string
	deadline := time.Now().Add(25 * time.Second)
	for {
		missing = s.missingPrograms()
		if len(missing) == 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Second)
	}
	for host, programs := range missing {
		s.Empty(programs,
			"%s lacks programs that its behaviors run: install their packages in the "+
				"suite's prepare in test/integration/prepare_test.go", host)
	}
}

// missingPrograms is the missing programs of each agent that the suite
// targets that lacks one.
func (s *BehaviorSmokeSuite) missingPrograms() map[string][]string {
	out, stderr, code := runCLI("client", "agent", "list", "--json")
	s.Require().Zero(code, stderr)
	var resp struct {
		Items []struct {
			Hostname        string   `json:"hostname"`
			MissingPrograms []string `json:"missing_programs"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(out, &resp))

	checked := 0
	missing := map[string][]string{}
	for _, a := range resp.Items {
		if target != "_any" && a.Hostname != target {
			continue
		}
		checked++
		if len(a.MissingPrograms) > 0 {
			missing[a.Hostname] = a.MissingPrograms
		}
	}
	s.Require().NotZero(checked, "no agent matches the target %s", target)

	return missing
}

func (s *BehaviorSmokeSuite) TestReads() {
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			// A read that names an item, or the item of a parent such as an
			// interface's routes, has no input that every host has.
			if m.Verb != types.VerbRead || len(m.PathKeys()) > 0 || needsInput(m) {
				continue
			}
			s.Run(b.Name+" "+m.Name, func() {
				s.results(commandLine(b, m, nil))
			})
		}
	}
}

// needsInput reports whether a read has a required field, such as the path
// of a file's status: no input names it on every host, and its examples run
// with the writes.
func needsInput(
	m *behavior.Method,
) bool {
	if m.In == nil {
		return false
	}
	for i := range m.In.NumField() {
		if strings.Contains(m.In.Field(i).Tag.Get("validate"), "required") {
			return true
		}
	}

	return false
}

func (s *BehaviorSmokeSuite) TestWrites() {
	if os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") != "1" {
		skipWrite(s.T())
	}

	s.refreshIndex()
	var examples []made
	defer func() { s.leftNothing(examples) }()
	// Before leftNothing, which requires that the interface examples left
	// no config of dummy. TestMain's prepare made it.
	defer func() { s.NoError(dropDummy()) }()

	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			// A method that takes the host down, such as a reboot, never runs
			// on the host the test runs on.
			if m.Verb != types.VerbWrite || m.Disruptive {
				continue
			}
			for _, ex := range m.Examples {
				s.Run(b.Name+" "+m.Name+" "+ex.Doc, func() {
					if ex.NotLive != "" {
						s.T().Skip("runs on the fake host only: " + ex.NotLive)
					}
					// An interface example deletes dummy, which the
					// examples after it need. This runs last, after the
					// example's own remove.
					if b.Name == "interface" {
						defer func() { s.NoError(makeDummy()) }()
					}
					// A rename makes the host answer to its new name. The
					// suite follows it there, so the restore reaches the host
					// by the name that it has, then goes back to the target's
					// own name. Registered first, so it runs last.
					renames := b.Name == "hostname" && target != "_any"
					var machineID string
					if renames {
						named := target
						machineID = s.agentHostname("", named)
						defer func() { target = named }()
					}
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
						on, before := methodFor(b, g.Input)
						s.Require().NotNil(before, "no method takes %T", g.Input)
						s.results(commandLine(on, before, g.Input))
					}

					want, _ := behavior.Changed(ex.Output)
					for _, r := range s.results(commandLine(b, m, ex.Input)) {
						if r["status"] == "ok" {
							s.Equal(want, r["changed"], "the example's changed")
							if name, ok := r["name"].(string); renames && ok && name != "" {
								target = s.follow(machineID, name)
							}
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
	b := behavior.Lookup("updates")
	if b == nil || b.Method("update") == nil {
		return
	}
	m := b.Method("update")
	s.results(commandLine(b, m, reflect.New(m.In).Elem().Interface()))
}

// methodFor is the method whose input type is input's, and its behavior:
// one of b, or of b's parent, as After can name.
func methodFor(
	b *behavior.Behavior,
	input any,
) (*behavior.Behavior, *behavior.Method) {
	for _, on := range []*behavior.Behavior{b, b.Parent} {
		if on == nil {
			continue
		}
		for _, m := range on.Methods {
			if m.In == reflect.TypeOf(input) {
				return on, m
			}
		}
	}

	return nil, nil
}

// item is an input that names an item, and the behavior whose delete
// removes it.
type item struct {
	b     *behavior.Behavior
	input any
}

// made is what a write example can leave on the target: the items that its
// input and its Given calls name, which their behavior's delete removes, and
// the objects that it uploaded. The example's own item is first, so a key
// goes before the user that it is under.
type made struct {
	items   []item
	objects []string
}

// madeBy is what ex can leave.
func madeBy(
	b *behavior.Behavior,
	ex behavior.Example,
) made {
	m := made{items: []item{{b, ex.Input}}}
	for _, g := range ex.Given {
		switch g.Kind {
		case behavior.GivenCall:
			on, _ := methodFor(b, g.Input)
			if on == nil {
				on = b
			}
			m.items = append(m.items, item{on, g.Input})
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
	for _, it := range m.items {
		s.reset(it.b, it.input)
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
		for _, it := range m.items {
			for _, r := range s.deleteResults(it.b, it.input) {
				if r["status"] == "ok" {
					s.Equal(false, r["changed"],
						"%s %v was left on the target", it.b.Name, it.input)
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
			"--file", path, "--content-type", string(g.ContentType))
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
	// named is the target's name before the example: a rename moves the
	// target to the new name, the update puts the old one back, and the
	// read after it is by the old name.
	named := target

	return func() {
		for _, in := range before {
			s.results(commandLine(b, update, in))
		}
		target = named
		s.Equal(before, read(), "%s is back as the example found it", b.Name)
	}
}

// agentHostname is, with a machine ID, the hostname that the agent reports;
// with "" and a hostname, the machine ID of that agent. Each is "" when no
// agent matches.
func (s *BehaviorSmokeSuite) agentHostname(
	machineID string,
	hostname string,
) string {
	out, stderr, code := runCLI("client", "agent", "list", "--json")
	s.Require().Zero(code, stderr)
	var resp struct {
		Items []struct {
			Hostname  string `json:"hostname"`
			MachineID string `json:"machine_id"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(out, &resp), out)
	for _, a := range resp.Items {
		switch {
		case machineID != "" && a.MachineID == machineID:
			return a.Hostname
		case machineID == "" && a.Hostname == hostname:
			return a.MachineID
		}
	}

	return ""
}

// follow is the name that the renamed agent answers to: name, once the agent
// registers with it, or the name that the agent keeps, when its config pins
// agent.hostname and the rename of the host does not change it.
func (s *BehaviorSmokeSuite) follow(
	machineID string,
	name string,
) string {
	var got string
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if got = s.agentHostname(machineID, ""); got == name {
			return name
		}
		time.Sleep(250 * time.Millisecond)
	}
	if got == "" {
		return name
	}

	return got
}

func TestBehaviorSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorSmokeSuite))
}
