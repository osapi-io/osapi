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
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
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
	examples []clicommand.Example
	refusals []any
	// noCheck is why no read shows the result of the examples, as NoCheck
	// says it.
	noCheck string
}

// writes reports whether the command changes what a later read shows: an
// API call other than GET, or a local command whose policy is write.
func (c cliCase) writes() bool {
	return c.api && c.verb != "GET" || cliPolicies[c.path].write
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
			cc.examples = c.Examples()
			switch m := c.(type) {
			case *endpoint.Method:
				cc.verb = m.HTTPMethod()
				cc.noCheck = m.Unchecked()
				for _, r := range m.Refusals() {
					cc.refusals = append(cc.refusals, r.Input)
				}
			case *clicommand.LocalMethod:
				cc.process = m.IsProcess()
				cc.noCheck = m.Unchecked()
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
	// issue is the open issue that makes a skipped command run, such as
	// osapi-io/osapi#1536 for the commands that need a second agent, and
	// elsewhere is the check that runs it in place of the suite, such as a
	// recipe of the justfile. The gate requires one of them of each skip, so
	// a new skip is a choice with a record (osapi-io/osapi#1535).
	issue     string
	elsewhere string
	// host is a command that reads or writes the files of a component on
	// its host: it runs against the suite's own server, not a deployment
	// such as the lab, whose files are not where the suite runs.
	host bool
	// write is a local command that changes the server's state, which runs
	// with OSAPI_INTEGRATION_WRITES=1.
	write bool
	// madeByWrite is a read whose item only a write example of the same run
	// makes, such as the file that file upload stores. It runs with
	// OSAPI_INTEGRATION_WRITES=1, as that write does (osapi-io/osapi#1612).
	madeByWrite bool
	// env is what the command runs with besides the client's environment,
	// such as the NATS user that it connects as.
	env []string
}

const (
	skipProcess = "a process that runs until it stops: the join suite's fixture server runs " +
		"it, with a short life"
	skipAdmission = "changes the admission of a machine, and each other test needs the suite's " +
		"agent: the join suite runs it against a fixture server of its own"
	skipScaffold = "writes the source of a definition into a checkout: the scaffold checks of the " +
		"justfile run it in a copy"
	skipService = "installs or removes a system service on the host: the agent service scenario " +
		"installs one with a fixture's join string, as root on Linux"
	// joinSuite runs the admission commands and the processes against a
	// fixture server of its own.
	joinSuite = "the join scenario in scenarios_test.go, against the fixture of fixture_test.go"
	// pkiScenario makes an agent's keys in a new directory, and refuses a
	// second create.
	pkiScenario = "the pki scenario in scenarios_test.go, against a fixture"
	// serviceSuite installs and removes the agent's service.
	serviceSuite = "the agent service scenario in scenarios_test.go"
	// scaffoldChecks are the recipes that run a scaffold in a copy.
	scaffoldChecks = "just behavior-scaffold-check, endpoint-scaffold-check and command-scaffold-check"
)

// cliPolicies are the commands whose examples do not run as their kind.
var cliPolicies = map[string]cliPolicy{
	"start": {
		skip:      "a process that runs until it stops: TestMain starts the suite's server with it",
		elsewhere: "TestMain in integration_test.go",
	},
	"agent start":             {skip: skipProcess, elsewhere: joinSuite},
	"controller start":        {skip: skipProcess, elsewhere: joinSuite},
	"nats start":              {skip: skipProcess, elsewhere: joinSuite},
	"client admission admit":  {skip: skipAdmission, elsewhere: joinSuite},
	"client admission reject": {skip: skipAdmission, elsewhere: joinSuite},
	"client admission revoke": {skip: skipAdmission, elsewhere: joinSuite},
	"client admission delete": {skip: skipAdmission, elsewhere: joinSuite},
	"agent service install":   {skip: skipService, elsewhere: serviceSuite},
	"agent service uninstall": {skip: skipService, elsewhere: serviceSuite},
	"dsl behavior create":     {skip: skipScaffold, elsewhere: scaffoldChecks},
	"dsl endpoint create":     {skip: skipScaffold, elsewhere: scaffoldChecks},
	"dsl command create":      {skip: skipScaffold, elsewhere: scaffoldChecks},
	"dsl docs generate": {
		skip:      "writes the docs of a checkout: a copy of the checkout runs it",
		elsewhere: "just docs-gen-check and cli-docs-gen-check",
	},
	"pki key create": {
		skip: "makes a key on the host, and refuses one that is there: the suite's server made " +
			"each key that it uses, and sets no JetStream key_file; its unit tests make each kind",
		elsewhere: pkiScenario,
	},
	// A rotation of the API's certificate reaches the running server, and
	// each later call of the suite reads the new api.crt (#1569).
	"pki cert rotate": {host: true, write: true},
	// A rotation needs a NATS server that serves its self-signed nats.crt, as
	// the suite's own server does: the lab's serves a cert_file, which a
	// rotation refuses with 409.
	"client pki nats rotate":       {host: true, write: true},
	"controller join token create": {host: true, write: true},
	"controller join token list":   {host: true},
	"controller join token revoke": {host: true, write: true},
	"controller token generate":    {host: true},
	"controller token validate":    {host: true},
	// An operator runs nats setup as the setup user, which may create and
	// update the streams; the controller's user may not (#1415).
	"nats setup": {host: true, write: true, env: []string{
		"OSAPI_CONTROLLER_NATS_AUTH_USERNAME=setup",
		"OSAPI_CONTROLLER_NATS_AUTH_PASSWORD=integration-setup",
	}},
	"pki cert list":  {host: true},
	"pki key list":   {host: true},
	"pki key get":    {host: true},
	"pki key rotate": {host: true, write: true},
	"pki list":       {host: true},
	"pki cert get":   {host: true},
	// The file that file upload stores, and the secret that secret set
	// stores.
	"client file get":      {madeByWrite: true},
	"client secret get":    {madeByWrite: true},
	"client secret reveal": {madeByWrite: true},
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

// issueRef is an issue as a skip names it, such as osapi-io/osapi#1536.
var issueRef = regexp.MustCompile(`^osapi-io/[a-z0-9.-]+#[0-9]+$`)

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
	s.eachWriteIsChecked(seen)
	for path := range undoes {
		s.True(seen[path], "undoes names %s, which no definition has", path)
	}
	for path, p := range cliPolicies {
		s.True(seen[path], "cliPolicies names %s, which no definition has", path)
		if p.skip != "" {
			s.GreaterOrEqual(len(p.skip), 40, "%s: say why the suite cannot run it", path)
			s.True(issueRef.MatchString(p.issue) || p.elsewhere != "",
				"%s: a skip names the open issue that makes it run, as osapi-io/<repo>#<n>, "+
					"or the check that runs it elsewhere", path)
		}
	}
}

// eachWriteIsChecked is the gate of osapi-io/osapi#1540: each example of a
// write that the suite runs has a read that shows its result, or its
// method says why none can, and each read is a command that reads.
func (s *CLICasesTestSuite) eachWriteIsChecked(
	seen map[string]bool,
) {
	reads := map[string]bool{}
	for _, c := range cliCases() {
		reads[c.path] = !c.writes()
	}
	for _, c := range cliCases() {
		if cliPolicies[c.path].skip != "" {
			continue
		}
		for _, e := range c.examples {
			if c.writes() && c.noCheck == "" {
				s.NotEmpty(e.Checks, "%s %s: a write with no read of its result: add a Check "+
					"or a CheckFails after the Example, or a NoCheck with the reason", c.path, e.Flags)
			}
			for _, rb := range e.Checks {
				s.True(
					seen[rb.Command],
					"%s: its Check runs %s, which no definition has",
					c.path,
					rb.Command,
				)
				s.True(
					reads[rb.Command],
					"%s: its Check runs %s, which is not a read",
					c.path,
					rb.Command,
				)
				s.True(len(rb.Fields) > 0 || rb.Fails != "",
					"%s: its Check of %s asserts nothing", c.path, rb.Command)
			}
		}
		if c.noCheck != "" {
			s.GreaterOrEqual(len(c.noCheck), 40, "%s: say why no read shows its result", c.path)
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

// TestEachLiveFieldIsAField is the gate of osapi-io/osapi#1541: each field
// that an example marks Live is a field of its input, by its JSON name, of
// a kind that a fact fills, with a fact that the suite knows.
func (s *CLICasesTestSuite) TestEachLiveFieldIsAField() {
	known := []behavior.Fact{
		behavior.FactInterface, behavior.FactAdminGroup, behavior.FactUser, behavior.FactPID,
		behavior.FactSSHService, behavior.FactSSHUnit,
	}
	count := 0
	for _, b := range behavior.All() {
		for _, m := range b.Methods {
			examples := append([]behavior.Example{}, m.Examples...)
			for _, r := range m.Refusals {
				examples = append(examples, r.Example)
			}
			for _, ex := range examples {
				for _, l := range ex.Live {
					count++
					where := b.Name + " " + m.Name + " " + ex.Doc
					s.Contains(known, l.Fact, "%s: no value for the fact %s", where, l.Fact)
					t := reflect.TypeOf(ex.Input)
					found := false
					for i := range t.NumField() {
						f := t.Field(i)
						if strings.Split(f.Tag.Get("json"), ",")[0] != l.Field {
							continue
						}
						found = true
						s.Contains(
							[]reflect.Kind{
								reflect.String,
								reflect.Int,
								reflect.Int32,
								reflect.Int64,
							},
							f.Type.Kind(),
							"%s: %s is not a string or an int",
							where,
							l.Field,
						)
					}
					s.True(found, "%s: %T has no field %s", where, ex.Input, l.Field)
				}
			}
		}
	}
	s.NotZero(count, "no example marks a field Live")
}
