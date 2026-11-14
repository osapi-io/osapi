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
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
)

// BehaviorSmokeSuite runs every registered behavior's generated commands
// against the live server and agent, from its definition, so a new behavior
// is exercised on a real host without a test written for it.
//
// Every read method without a key runs always. Writes change the host, so
// they run only with OSAPI_INTEGRATION_WRITES=1 or
// OSAPI_INTEGRATION_WRITE_BEHAVIORS=1: each write example runs after the calls
// its example names with behavior.After, must report the example's changed,
// and run again must report changed: false.
type BehaviorSmokeSuite struct {
	suite.Suite
}

// command is the generated command line for one method, with input's flags:
// the docs pages' command without the binary's name.
func command(
	b *behavior.Behavior,
	m *behavior.Method,
	input any,
) []string {
	args := append(strings.Fields(b.CLICommand(m))[1:], "--target", "_any", "--json")

	return append(args, behaviorcmd.Args(input)...)
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

// TestPrograms: the image has every program a behavior declares, so a
// behavior's examples fail here for what it does, not for what is missing.
func (s *BehaviorSmokeSuite) TestPrograms() {
	for _, b := range behavior.All() {
		for _, p := range append(slices.Clone(b.Sudo), b.Runs...) {
			_, err := exec.LookPath(p)
			s.NoError(err,
				"%s runs %s, which this image lacks: add its package to the apt-get "+
					"install line in the justfile's behavior-int-linux recipe", b.Name, p)
		}
	}
}

func (s *BehaviorSmokeSuite) TestReads() {
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			if m.Verb != behavior.VerbRead || m.KeyField != "" {
				continue
			}
			s.Run(b.Name+" "+m.Name, func() {
				s.results(command(b, m, nil))
			})
		}
	}
}

func (s *BehaviorSmokeSuite) TestWrites() {
	if os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") != "1" {
		skipWrite(s.T())
	}

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
					for _, g := range ex.Given {
						if g.Kind != behavior.GivenCall {
							continue
						}
						for _, before := range b.Methods {
							if before.In == reflect.TypeOf(g.Input) {
								s.results(command(b, before, g.Input))
							}
						}
					}

					want, _ := behavior.Changed(ex.Output)
					for _, r := range s.results(command(b, m, ex.Input)) {
						if r["status"] == "ok" {
							s.Equal(want, r["changed"], "the example's changed")
						}
					}
					for _, r := range s.results(command(b, m, ex.Input)) {
						if r["status"] == "ok" {
							s.Equal(false, r["changed"], "the same call again changes nothing")
						}
					}
				})
			}
		}
	}
}

func TestBehaviorSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorSmokeSuite))
}
