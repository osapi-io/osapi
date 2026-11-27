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
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// rawResults runs one command straight to the controller, not through
// the spec proxy, and returns its exit code, its standard error, and its
// result rows, which may have failed. A refusal can break the spec on
// purpose, such as a name that the path's pattern refuses: the proxy would
// count it against the run.
func (s *BehaviorSmokeSuite) rawResults(
	c call,
) (int, string, []map[string]any) {
	stdout, stderr, code := runDirect(s.T().TempDir(), c.stdin, c.args...)
	var body struct {
		Results []map[string]any `json:"results"`
	}
	_ = parseJSON(stdout, &body)

	return code, stderr, body.Results
}

// TestRefusals runs each refusal of each behavior method that names only
// an input, not a host that the fake host sets up, against the live target
// (osapi-io/osapi#1538): the controller or the agent must refuse it. A
// refusal of a write runs only with the writes, and only when the input's
// validation refuses it, since a refusal that the code lets through changes
// the host.
func (s *BehaviorSmokeSuite) TestRefusals() {
	writes := runWrites || os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") == "1"
	count := 0
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			if m.Verb == types.VerbWrite && (!writes || m.Disruptive) {
				continue
			}
			for _, r := range m.Refusals {
				if len(r.Given) > 0 || r.NotLive != "" || !runsOn(r.Example, s.platform()) {
					continue
				}
				// A write refusal runs only when the input's validation
				// refuses it, as its error, a rule's name such as min, says:
				// one whose error is the host's state, such as an interface
				// that osapi does not configure, can change a host whose
				// state differs from the fake one.
				if m.Verb == types.VerbWrite && strings.ContainsAny(r.Error, " :") {
					continue
				}
				count++
				s.Run(b.Name+" "+m.Name+" "+r.Doc, func() {
					c := commandLine(b, m, r.Input)
					code, stderr, rows := s.rawResults(c)
					s.NotRegexp(
						connectionError,
						stderr,
						"%s: it failed for its connection, not its input",
						strings.Join(c.args, " "),
					)
					if code != 0 {
						return
					}
					for _, row := range rows {
						if row["status"] == "ok" {
							s.Failf("a refusal succeeded",
								"%s: %v", strings.Join(c.args, " "), row)
						}
					}
				})
			}
		}
	}
	s.NotZero(count, "no refusal of a behavior runs live")
}

// TestKeyedReads runs each example of each read that names an item, such
// as a service's get, on the live target (osapi-io/osapi#1538). An example
// that makes its item with a call of the behavior runs that call first, so
// it runs with the writes.
func (s *BehaviorSmokeSuite) TestKeyedReads() {
	writes := runWrites || os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") == "1"
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			if m.Verb != types.VerbRead || (len(m.PathKeys()) == 0 && !needsInput(m)) {
				continue
			}
			for _, ex := range m.Examples {
				s.Run(b.Name+" "+m.Name+" "+ex.Doc, func() {
					switch {
					case ex.NotLive != "":
						s.T().Skip("runs on the fake host only: " + ex.NotLive)
					case !runsOn(ex, s.platform()):
						s.T().
							Skipf("an example of %v, and the target is %s", ex.Platforms, s.platform())
					}
					var calls []behavior.Given
					for _, g := range ex.Given {
						if g.Kind == behavior.GivenCall {
							calls = append(calls, g)
						}
					}
					if len(calls) > 0 {
						if !writes {
							s.T().
								Skip("makes its item with a write: set OSAPI_INTEGRATION_WRITES=1")
						}
						defer s.restore(b)()
						made := madeBy(b, ex)
						defer s.remove(made)
						s.store(ex.Given)
						for _, g := range calls {
							on, before := methodFor(b, g.Input)
							s.Require().NotNil(before, "no method takes %T", g.Input)
							s.results(commandLine(on, before, g.Input))
						}
					}
					s.results(commandLine(b, m, s.liveInput(ex)))
				})
			}
		}
	}
}

// liveInput is the input of ex with each field that ex marks Live set to
// the target's fact (osapi-io/osapi#1541).
func (s *BehaviorSmokeSuite) liveInput(
	ex behavior.Example,
) any {
	if len(ex.Live) == 0 {
		return ex.Input
	}
	v := reflect.New(reflect.TypeOf(ex.Input)).Elem()
	v.Set(reflect.ValueOf(ex.Input))
	for _, l := range ex.Live {
		f, ok := liveField(v, l.Field)
		s.Require().True(ok, "%T has no field %s", ex.Input, l.Field)
		value := s.fact(l.Fact)
		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
		case reflect.Int, reflect.Int32, reflect.Int64:
			n, err := strconv.Atoi(value)
			s.Require().NoError(err)
			f.SetInt(int64(n))
		default:
			s.FailNowf("a live field is a string or an int", "%s of %T", l.Field, ex.Input)
		}
	}

	return v.Interface()
}

// liveField is the field of the struct v whose JSON name is name.
func liveField(
	v reflect.Value,
	name string,
) (reflect.Value, bool) {
	for i := range v.NumField() {
		if strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0] == name {
			return v.Field(i), true
		}
	}

	return reflect.Value{}, false
}

// fact is the target's value of fact.
func (s *BehaviorSmokeSuite) fact(
	fact behavior.Fact,
) string {
	switch fact {
	case behavior.FactInterface:
		return s.primaryInterface()
	case behavior.FactAdminGroup:
		if s.platform() == behavior.PlatformDarwin {
			return "wheel"
		}

		return "root"
	case behavior.FactUser:
		return "root"
	case behavior.FactPID:
		return "1"
	case behavior.FactSSHService:
		if s.platform() == behavior.PlatformDebian {
			return "ssh"
		}

		return "sshd"
	case behavior.FactSSHUnit:
		return s.fact(behavior.FactSSHService) + ".service"
	}
	s.FailNow("no value for the fact " + string(fact))

	return ""
}

// primaryInterface is the primary interface of the target, as agent get
// reports it.
func (s *BehaviorSmokeSuite) primaryInterface() string {
	out, stderr, code := runCLI("client", "agent", "list", "--json")
	s.Require().Zero(code, stderr)
	var resp struct {
		Items []struct {
			Hostname         string `json:"hostname"`
			PrimaryInterface string `json:"primary_interface"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(out, &resp))
	for _, a := range resp.Items {
		if (target == "_any" || a.Hostname == target) && a.PrimaryInterface != "" {
			return a.PrimaryInterface
		}
	}
	s.FailNow("no agent of the target " + target + " reports its primary interface")

	return ""
}
