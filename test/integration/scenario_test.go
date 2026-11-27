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

// This file has no build tag: a scenario is data, and TestEachStepIsACommand
// checks each step against the command tree with the unit tests. The runner
// is in scenario_run_test.go, which runs against a live controller.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/cmd"
	"github.com/osapi-io/osapi/internal/pki"
)

// scenario is a flow of osapi commands, each step with the checks of its
// result, that the runner runs as subtests: a join, a rotation, a secret
// that a job opens.
type scenario struct {
	name string
	// writeOp names the OSAPI_INTEGRATION_SKIP_* switch of a scenario that
	// writes, such as JOIN. A scenario that only reads has none.
	writeOp string
	// fixture runs the steps against a fixture server of the scenario's
	// own, with controller.pki.auto_admit as fixtureAutoAdmit.
	fixture          bool
	fixtureAutoAdmit bool
	// fixtureConsumer replaces keys of the fixture's agent.consumer, such
	// as a short ack_wait; a key whose value is nil is removed.
	fixtureConsumer map[string]any
	// fixtureAgent replaces keys of the fixture's agent config, such as
	// command: {enabled: false}.
	fixtureAgent map[string]any
	// ownServer is a scenario that changes the host that runs the suite,
	// such as the keys of its agent: it skips against a deployment.
	ownServer bool
	// rootLinux is a scenario that needs root on Linux with systemd.
	rootLinux bool
	// labOnly is a scenario that runs only against a deployment, such as a
	// reboot of a lab VM: the CI runner and the suite's own host are not
	// for it to lose.
	labOnly bool
	steps   []step
	// cleanup runs after the steps, also when one fails. Its failures are
	// not the scenario's.
	cleanup []step
}

// on is where a step's command runs.
type on int

const (
	// onSuite is a command against the suite's controller, or the lab's.
	onSuite on = iota
	// onHost is a command on the agent's host, with the server's config,
	// such as osapi pki key rotate --name agent.
	onHost
	// onFixture is a command against the scenario's fixture server.
	onFixture
	// onFixtureHost is a command on the fixture's host with no config
	// file, such as osapi agent service install.
	onFixtureHost
)

// step is one command, or the start or the stop of a fixture's agent,
// and its checks. A step with wait runs again until its exit and its
// checks pass, or wait ends.
type step struct {
	name string
	// cli is the command and its flags, without osapi. A word in braces
	// is a placeholder: see expand.
	cli   []string
	on    on
	stdin string
	// keyDir, for onHost and onFixture, is the agent's key directory, as
	// OSAPI_AGENT_PKI_KEY_DIR: "{dir:name}" is a directory of the scenario.
	keyDir string
	// start starts the fixture's agent with that name, with keyDir and
	// join. stop stops it.
	start string
	join  string
	stop  string
	// kill stops the fixture's agent with that name with SIGKILL, so it
	// acks nothing that it runs.
	kill string
	// background, with cli on the fixture, starts the command and goes on
	// at once: a job whose agent the next steps kill. The scenario's end
	// stops it.
	background bool
	// exits, with start, is that the agent must stop with an error within
	// wait: a join that the controller refuses.
	exits bool
	// fails is that the command must exit with an error.
	fails bool
	// exec is a program of the host, not osapi, such as systemctl, whose
	// output the checks read.
	exec   []string
	checks []check
	// capture keeps the value of a JSON path of the output under a name,
	// for the placeholders of the later steps.
	capture map[string]string
	// wait is how long the step may try again for its exit and its checks,
	// pause apart. A step with no wait sleeps pause after it runs.
	wait  time.Duration
	pause time.Duration
	// skipIf skips the scenario when it matches the output: the host
	// cannot run it, such as one with no docker.
	skipIf *check
	// ownServer is a step that reads what only the suite's own server
	// makes, such as a join string of its own NATS: it is skipped against
	// a deployment.
	ownServer bool
}

// check is a relation of the values at a JSON path of a step's output.
//
// A path is fields by dots: results[0].status, items[*].fingerprint, and
// items[state=Ready].hostname, the items whose state is Ready. "" is the
// whole output, "!stderr" the error output, and "!all" both.
//
// ==, !=, prefix, nonempty, empty and p256 hold for each value at the path,
// and the path must have one. has and lacks are that the values include
// want, or do not. contains and absent are that each value has want in it,
// or does not. file and nofile are that want names a file, or does not: the
// path is not read.
type check struct {
	path string
	op   string
	want string
}

var checkOps = []string{
	"==", "!=", "prefix", "nonempty", "empty", "p256",
	"has", "lacks", "contains", "absent", "file", "nofile",
}

// values is each value at path in doc, a decoded JSON document.
func values(
	doc any,
	path string,
) []any {
	cur := []any{doc}
	for _, seg := range splitStepPath(path) {
		name, sel, hasSel := strings.Cut(seg, "[")
		var next []any
		for _, v := range cur {
			if name != "" {
				m, ok := v.(map[string]any)
				if !ok {
					continue
				}
				v, ok = m[name]
				if !ok {
					continue
				}
			}
			if !hasSel {
				next = append(next, v)

				continue
			}
			list, ok := v.([]any)
			if !ok {
				continue
			}
			sel = strings.TrimSuffix(sel, "]")
			switch k, want, isFilter := strings.Cut(sel, "="); {
			case sel == "*":
				next = append(next, list...)
			case isFilter:
				for _, item := range list {
					if m, ok := item.(map[string]any); ok && text(m[k]) == want {
						next = append(next, item)
					}
				}
			default:
				var i int
				if _, err := fmt.Sscan(sel, &i); err == nil && i >= 0 && i < len(list) {
					next = append(next, list[i])
				}
			}
		}
		cur = next
	}

	return cur
}

// splitStepPath splits a path at each dot that is not in brackets, so a filter
// value such as a hostname can have dots.
func splitStepPath(
	path string,
) []string {
	if path == "" {
		return nil
	}
	var out []string
	depth, start := 0, 0
	for i, r := range path {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case '.':
			if depth == 0 {
				out = append(out, path[start:i])
				start = i + 1
			}
		}
	}

	return append(out, path[start:])
}

// text is a JSON value as the checks compare it.
func text(
	v any,
) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%v", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, _ := json.Marshal(t)

		return string(b)
	}
}

// eval is the reason that c does not hold for vals, or "".
func (c check) eval(
	vals []string,
	exists func(string) bool,
) string {
	each := func(ok func(string) bool) string {
		if len(vals) == 0 {
			return fmt.Sprintf("%q has no value", c.path)
		}
		for _, v := range vals {
			if !ok(v) {
				return fmt.Sprintf("%q is %q: not %s %q", c.path, v, c.op, c.want)
			}
		}

		return ""
	}
	switch c.op {
	case "==":
		return each(func(v string) bool { return v == c.want })
	case "!=":
		return each(func(v string) bool { return v != c.want })
	case "prefix":
		return each(func(v string) bool { return strings.HasPrefix(v, c.want) })
	case "nonempty":
		return each(func(v string) bool { return v != "" })
	case "empty":
		if len(vals) == 0 {
			return ""
		}

		return each(func(v string) bool { return v == "" })
	case "p256":
		return each(pki.ValidEncryptionKey)
	case "contains":
		return each(func(v string) bool { return strings.Contains(v, c.want) })
	case "absent":
		for _, v := range vals {
			if strings.Contains(v, c.want) {
				return fmt.Sprintf("%q has %q", c.path, c.want)
			}
		}

		return ""
	case "has":
		if slices.Contains(vals, c.want) {
			return ""
		}

		return fmt.Sprintf("%q is %q: it has no %q", c.path, vals, c.want)
	case "lacks":
		if !slices.Contains(vals, c.want) {
			return ""
		}

		return fmt.Sprintf("%q has %q", c.path, c.want)
	case "file":
		if exists(c.want) {
			return ""
		}

		return c.want + " is not there"
	case "nofile":
		if !exists(c.want) {
			return ""
		}

		return c.want + " is there"
	}

	return "unknown check " + c.op
}

// placeholder is a word in braces: a name that a step captured, or one of
// target, stamp, dir:<name>, wrongfp:<name> and urlq:<name>, the value of
// name as a URL query escapes it.
var placeholder = regexp.MustCompile(`\{([a-z][a-zA-Z0-9]*)(?::([a-zA-Z0-9-]+))?\}`)

// builtins are the placeholders that the runner gives each scenario.
var builtins = []string{"target", "stamp", "dir", "wrongfp", "urlq"}

// expand fills each placeholder of s from vars, and the others with fn.
func expand(
	s string,
	vars map[string]string,
	fn func(name string, arg string) (string, bool),
) (string, error) {
	var err error
	out := placeholder.ReplaceAllStringFunc(s, func(m string) string {
		p := placeholder.FindStringSubmatch(m)
		if v, ok := fn(p[1], p[2]); ok {
			return v
		}
		if v, ok := vars[p[1]]; ok {
			return v
		}
		err = fmt.Errorf("%s: no step captured %s", s, p[1])

		return m
	})

	return out, err
}

// commandOf is the command of a step's words in root, and its flags.
func commandOf(
	root *cobra.Command,
	words []string,
) (*cobra.Command, []string, error) {
	var path, flags []string
	for _, w := range words {
		switch {
		case strings.HasPrefix(w, "--"):
			name, _, _ := strings.Cut(strings.TrimPrefix(w, "--"), "=")
			flags = append(flags, name)
		case len(flags) == 0:
			path = append(path, w)
		}
	}
	c, rest, err := root.Find(path)
	if err != nil {
		return nil, nil, err
	}
	if len(rest) > 0 || !c.Runnable() {
		return nil, nil, fmt.Errorf("%q is not a command", strings.Join(path, " "))
	}

	return c, flags, nil
}

// ScenarioTestSuite checks each scenario's steps with no controller.
type ScenarioTestSuite struct {
	suite.Suite
}

// TestEachStepIsACommand: each step names a command that exists, with
// flags of that command, its checks are known checks, and each placeholder
// is a builtin or a name that an earlier step captured.
func (s *ScenarioTestSuite) TestEachStepIsACommand() {
	root := cmd.Root()
	s.NotEmpty(scenarios())
	for _, sc := range scenarios() {
		s.Run(sc.name, func() {
			s.NotEmpty(sc.steps, "a scenario has steps")
			captured := map[string]bool{}
			for _, st := range append(append([]step{}, sc.steps...), sc.cleanup...) {
				s.checkStep(root, sc, st, captured)
			}
		})
	}
}

func (s *ScenarioTestSuite) checkStep(
	root *cobra.Command,
	sc scenario,
	st step,
	captured map[string]bool,
) {
	s.NotEmpty(st.name, "%s: each step has a name", sc.name)
	kinds := 0
	for _, set := range []bool{len(st.cli) > 0, st.start != "", st.stop != "", st.kill != "", len(st.exec) > 0} {
		if set {
			kinds++
		}
	}
	s.Equal(1, kinds, "%s: %s: a step is one of cli, start, stop, kill and exec", sc.name, st.name)
	if len(st.cli) > 0 {
		c, flags, err := commandOf(root, st.cli)
		if s.NoError(err, "%s: %s", sc.name, st.name) {
			for _, f := range flags {
				s.NotNil(
					c.Flag(f),
					"%s: %s: %s has no flag --%s",
					sc.name,
					st.name,
					c.CommandPath(),
					f,
				)
			}
		}
	}
	if st.background {
		s.Equal(onFixture, st.on, "%s: %s: a background step runs on the fixture", sc.name, st.name)
	}
	if st.on == onFixture || st.on == onFixtureHost || st.start != "" || st.stop != "" ||
		st.kill != "" {
		s.True(sc.fixture, "%s: %s: a fixture step in a scenario with no fixture", sc.name, st.name)
	}
	if len(st.exec) > 0 {
		s.Contains(
			[]string{"systemctl"},
			st.exec[0],
			"%s: %s: exec runs only known programs",
			sc.name,
			st.name,
		)
	}
	all := slices.Clone(st.checks)
	if st.skipIf != nil {
		all = append(all, *st.skipIf)
	}
	known := func(words ...string) {
		for _, w := range words {
			for _, m := range placeholder.FindAllStringSubmatch(w, -1) {
				s.True(
					slices.Contains(builtins, m[1]) || captured[m[1]],
					"%s: %s: %s is not a builtin, and no step captured it before",
					sc.name,
					st.name,
					m[0],
				)
			}
		}
	}
	known(
		append(append(append([]string{}, st.cli...), st.exec...), st.keyDir, st.join, st.stdin)...)
	// A step's captures come before its checks, so a check can name a file
	// that the same step's output gives.
	for name := range st.capture {
		captured[name] = true
	}
	for _, c := range all {
		s.Contains(checkOps, c.op, "%s: %s: an unknown check", sc.name, st.name)
		known(c.path, c.want)
	}
}

// TestChecks: the checks of a step read the output as their doc says.
func (s *ScenarioTestSuite) TestChecks() {
	var doc any
	s.Require().NoError(json.Unmarshal([]byte(`{"items":[
		{"hostname":"a.local","state":"Ready","n":2,"ok":true},
		{"hostname":"b","state":"Pending","n":3,"ok":false}]}`), &doc))
	none := func(string) bool { return false }
	tests := []struct {
		name string
		path string
		c    check
		fail bool
	}{
		{
			name: "a filter, a value with a dot",
			path: "items[state=Ready].hostname",
			c:    check{op: "==", want: "a.local"},
		},
		{name: "an index", path: "items[1].n", c: check{op: "==", want: "3"}},
		{name: "each value", path: "items[*].state", c: check{op: "==", want: "Ready"}, fail: true},
		{name: "has", path: "items[*].hostname", c: check{op: "has", want: "b"}},
		{name: "lacks", path: "items[*].hostname", c: check{op: "lacks", want: "c"}},
		{
			name: "lacks a value that is there",
			path: "items[*].hostname",
			c:    check{op: "lacks", want: "b"},
			fail: true,
		},
		{name: "a bool", path: "items[0].ok", c: check{op: "==", want: "true"}},
		{name: "a path with no value", path: "items[9].n", c: check{op: "nonempty"}, fail: true},
		{name: "empty, with no value", path: "nothing", c: check{op: "empty"}},
		{name: "a prefix", path: "items[0].hostname", c: check{op: "prefix", want: "a."}},
		{name: "contains", path: "items[*].state", c: check{op: "contains", want: "e"}},
		{name: "absent", path: "items[*].state", c: check{op: "absent", want: "Ready"}, fail: true},
		{name: "nofile", c: check{op: "nofile", want: "/nothing/here"}},
		{name: "file", c: check{op: "file", want: "/nothing/here"}, fail: true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var vals []string
			for _, v := range values(doc, tt.path) {
				vals = append(vals, text(v))
			}
			tt.c.path = tt.path
			got := tt.c.eval(vals, none)
			if tt.fail {
				s.NotEmpty(got)
			} else {
				s.Empty(got)
			}
		})
	}
}

// TestExpand: a placeholder is a builtin, or a captured value.
func (s *ScenarioTestSuite) TestExpand() {
	fn := func(name, arg string) (string, bool) {
		if name == "dir" {
			return "/tmp/" + arg, true
		}

		return "", false
	}
	got, err := expand("--key-dir={dir:token} --id {id}", map[string]string{"id": "7"}, fn)
	s.NoError(err)
	s.Equal("--key-dir=/tmp/token --id 7", got)
	_, err = expand("{missing}", nil, fn)
	s.Error(err)
}

func TestScenarioTestSuite(t *testing.T) {
	suite.Run(t, new(ScenarioTestSuite))
}
