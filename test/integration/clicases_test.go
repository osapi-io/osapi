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

// This file has no build tag: TestEachCommandHasACase runs with the unit
// tests, as the gate of osapi-io/osapi#1449, and CommandSmokeSuite, in
// cli_test.go, runs the cases against a live controller.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
)

// cliCase is one CLI command from its definition: its words, its examples
// and its refusals, and whether it is a process or calls the API.
type cliCase struct {
	// path is the command without the binary, such as client job get.
	path  string
	group string
	name  string
	api   bool
	// verb is the HTTP method of an API command, such as GET.
	verb     string
	process  bool
	examples []string
	refusals []any
}

// cliCases is each command that an endpoint group or a local group defines,
// in the order of their registration.
func cliCases() []cliCase {
	var out []cliCase
	add := func(g clicommand.Group, api bool) {
		prefix := g.CommandPath()
		if api {
			prefix = append([]string{"client"}, prefix...)
		}
		for _, c := range g.Commands() {
			words := append([]string{}, prefix...)
			// A sub-path that is the command's name is one word, as the
			// command generator makes it.
			if sub := c.SubPath(); sub != "" && sub != c.CommandName() {
				words = append(words, strings.Fields(strings.ReplaceAll(sub, "/", " "))...)
			}
			words = append(words, c.CommandName())
			cc := cliCase{
				path: strings.Join(words, " "), group: strings.Join(prefix, " "),
				name: c.CommandName(), api: api,
			}
			for _, e := range c.Examples() {
				cc.examples = append(cc.examples, e.Flags)
			}
			switch m := c.(type) {
			case *endpoint.Method:
				cc.verb = m.HTTPMethod()
				for _, r := range m.Refusals() {
					cc.refusals = append(cc.refusals, r.Input)
				}
			case *clicommand.LocalMethod:
				cc.process = m.IsProcess()
				for _, r := range m.Refusals() {
					cc.refusals = append(cc.refusals, r.Input)
				}
			}
			out = append(out, cc)
		}
	}
	for _, g := range endpoint.All() {
		add(g, true)
	}
	for _, g := range clicommand.All() {
		add(g, false)
	}

	return out
}

// cliPolicy is how CommandSmokeSuite runs the examples of a command that is
// not run as its kind would be. With none, an API read runs always, an API
// write runs with OSAPI_INTEGRATION_WRITES=1, and a local command runs
// always. The refusals of each command run always: they change nothing.
type cliPolicy struct {
	// skip is why the suite does not run the examples, which the gate
	// requires of a process and of each command that the suite cannot run.
	skip string
	// host is a command that reads or writes the files of a component on
	// its host: it runs against the suite's own server, not a deployment
	// such as the lab, whose files are not where the suite runs.
	host bool
	// write is a local command that changes the server's state, which runs
	// with OSAPI_INTEGRATION_WRITES=1.
	write bool
}

const (
	skipProcess   = "a process that runs until it stops: TestMain starts the suite's server with osapi start"
	skipAdmission = "changes the admission of a machine, and the suite has no machine to spend: " +
		"each other test needs its agent; the admission endpoints' unit tests and the lab's join run it"
	skipScaffold = "writes the source of a definition into a checkout: the scaffold checks of the " +
		"justfile (behavior-, endpoint- and command-scaffold-check) run it in a copy"
	skipService = "installs or removes a system service on the host: osapi-lab's join of a VM runs " +
		"it (osapi-io/osapi-lab#73)"
)

// cliPolicies are the commands whose examples do not run as their kind.
var cliPolicies = map[string]cliPolicy{
	"start":                   {skip: skipProcess},
	"agent start":             {skip: skipProcess},
	"controller start":        {skip: skipProcess},
	"nats start":              {skip: skipProcess},
	"client admission admit":  {skip: skipAdmission},
	"client admission reject": {skip: skipAdmission},
	"client admission revoke": {skip: skipAdmission},
	"client admission delete": {skip: skipAdmission},
	"agent service install":   {skip: skipService},
	"agent service uninstall": {skip: skipService},
	"dsl behavior create":     {skip: skipScaffold},
	"dsl endpoint create":     {skip: skipScaffold},
	"dsl command create":      {skip: skipScaffold},
	"dsl docs generate": {
		skip: "writes the docs of a checkout: docs-gen-check and cli-docs-gen-check run it",
	},
	"pki key create": {
		skip: "makes a key on the host, and refuses one that is there: the suite's server made " +
			"each key that it uses, and sets no JetStream key_file; its unit tests make each kind",
	},
	"pki cert rotate": {
		skip: "the running server keeps serving the old certificate, so each later call of the " +
			"suite, which trusts the new file, would fail TLS; its unit tests rotate one",
	},
	"controller join token create": {host: true, write: true},
	"controller join token list":   {host: true},
	"controller join token revoke": {host: true, write: true},
	"controller token generate":    {host: true},
	"controller token validate":    {host: true},
	"nats setup":                   {host: true, write: true},
	"pki cert list":                {host: true},
	"pki key list":                 {host: true},
	"pki key get":                  {host: true},
	"pki key rotate":               {host: true, write: true},
	"pki list":                     {host: true},
	"pki cert get":                 {host: true},
}

// undo is a command that undoes an example on the agent that the other
// tests use, and the error that says it has nothing to undo.
type undo struct {
	args    []string
	nothing string
}

// undoes are the undo of each example that changes the agent that the
// other tests use, by the command of the example. A config update with no
// labels removes each label that the controller gave; undrain has nothing to
// undo once undrain's own example ran.
var undoes = map[string]undo{
	"client agent drain": {
		args:    []string{"client", "agent", "undrain"},
		nothing: "not in draining",
	},
	"client agent config update": {args: []string{"client", "agent", "config", "update"}},
}

// CLICasesTestSuite is the gate: each command has a case that the
// integration suite runs, or a reason that it does not.
type CLICasesTestSuite struct {
	suite.Suite
}

func (s *CLICasesTestSuite) TestEachCommandHasACase() {
	seen := map[string]bool{}
	for _, c := range cliCases() {
		seen[c.path] = true
		p := cliPolicies[c.path]
		if c.process {
			s.NotEmpty(p.skip, "%s is a process: give it a skip in cliPolicies", c.path)
		}
		if p.skip == "" {
			s.True(len(c.examples) > 0 || len(c.refusals) > 0,
				"%s has no example and no refusal for the integration suite to run: add an "+
					"Example or a Fails to its definition, or a skip with the reason to cliPolicies",
				c.path)
		}
	}
	for path := range undoes {
		s.True(seen[path], "undoes names %s, which no definition has", path)
	}
	for path, p := range cliPolicies {
		s.True(seen[path], "cliPolicies names %s, which no definition has", path)
		if p.skip != "" {
			s.GreaterOrEqual(len(p.skip), 40, "%s: say why the suite cannot run it", path)
		}
	}
}

func (s *CLICasesTestSuite) TestRefusalsAreInputs() {
	for _, c := range cliCases() {
		for _, r := range c.refusals {
			s.Equal(reflect.Struct, reflect.TypeOf(r).Kind(), "%s: a refusal is an input", c.path)
		}
	}
}

func TestCLICasesTestSuite(t *testing.T) {
	suite.Run(t, new(CLICasesTestSuite))
}
