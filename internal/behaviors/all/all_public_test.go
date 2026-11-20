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

package all_test

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	_ "github.com/osapi-io/osapi/internal/behaviors/all/gen"
	"github.com/osapi-io/osapi/internal/behaviors/sysctl"
	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
)

// AllPublicTestSuite checks that every behavior all.go registers is valid
// and that its generated code is imported, so a behavior added without
// running the generator, or whose gen package is missing from
// internal/behaviors/all/gen, fails here.
type AllPublicTestSuite struct {
	suite.Suite
}

func (s *AllPublicTestSuite) TestEveryBehaviorIsValid() {
	for _, b := range behavior.All() {
		s.Run(b.Name, func() {
			s.Empty(behavior.ValidateBehavior(b))
		})
	}
}

// TestUIPermissionsAreCurrent checks that the UI's permissions, which go
// generate writes from internal/authtoken, are what a token can have now.
func (s *AllPublicTestSuite) TestUIPermissionsAreCurrent() {
	got, err := os.ReadFile(filepath.Join("..", "..", "..", authtoken.TypeScriptFile))
	s.Require().NoError(err)
	s.Equal(authtoken.TypeScript(), string(got),
		"%s is stale: run go generate ./internal/behaviors/all", authtoken.TypeScriptFile)
}

// TestUIStatusesAreCurrent checks that the UI's status tones, which go
// generate writes from internal/cli, are the tones the CLI prints now.
func (s *AllPublicTestSuite) TestUIStatusesAreCurrent() {
	got, err := os.ReadFile(filepath.Join("..", "..", "..", cli.StatusTypeScriptFile))
	s.Require().NoError(err)
	s.Equal(cli.StatusTypeScript(), string(got),
		"%s is stale: run go generate ./internal/behaviors/all", cli.StatusTypeScriptFile)
}

func (s *AllPublicTestSuite) TestEveryRouteIsMounted() {
	e := echo.New()
	for _, register := range server.Handler(slog.Default(), "", nil, dispatch.Deps{}) {
		register(e)
	}
	mounted := map[string]bool{}
	for _, r := range e.Router().Routes() {
		mounted[r.Method+" "+r.Path] = true
	}

	for _, b := range behavior.All() {
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			g := behavior.GroupOf(b)
			s.Require().
				NotNil(g, "%s has no endpoint group; run go generate ./internal/behaviors/all", b.Name)
			route := m.HTTPMethod() + " " + server.Path(g, g.Method(m.Name))
			s.Run(route, func() {
				s.True(
					mounted[route],
					"%s is not mounted; run go generate ./internal/behaviors/all",
					route,
				)
			})
		}
	}
}

func (s *AllPublicTestSuite) TestEveryCommandIsAttached() {
	// osapi client, with node for the served behaviors and next node for
	// those still beside a hand-written domain, as cmd builds it.
	root := &cobra.Command{Use: "client"}
	node := &cobra.Command{Use: "node"}
	next := &cobra.Command{Use: "next"}
	nextNode := &cobra.Command{Use: "node"}
	root.AddCommand(node, next)
	next.AddCommand(nextNode)
	s.Require().NoError(clicommand.Attach(root, clicommand.Deps{}))

	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			line := b.Command(m)
			s.Run(line, func() {
				args := strings.Fields(strings.TrimPrefix(line, "osapi client "))
				cmd, _, err := root.Find(args)
				s.Require().NoError(err)
				s.Equal(
					m.Name,
					cmd.Name(),
					"%s is not attached; run go generate ./internal/behaviors/all",
					line,
				)
			})
		}
	}
}

// handWritten is every operation the hand-written specs declare, as "VERB
// path", with the directory of the spec that declares it.
func handWritten(
	t *testing.T,
) map[string]string {
	root, err := filepath.Abs("../../..")
	require.NoError(t, err)
	var specs []string
	for _, pattern := range []string{
		"internal/controller/api/*/gen/api.yaml",
		"internal/controller/api/node/*/gen/api.yaml",
	} {
		found, err := filepath.Glob(filepath.Join(root, pattern))
		require.NoError(t, err)
		specs = append(specs, found...)
	}

	ops := map[string]string{}
	for _, spec := range specs {
		dir := filepath.Dir(filepath.Dir(spec))
		rel, _ := filepath.Rel(root, dir)
		// The joined spec is not hand-written.
		if rel == "internal/controller/api" {
			continue
		}
		data, err := os.ReadFile(spec)
		require.NoError(t, err)
		var doc struct {
			Paths map[string]map[string]any `yaml:"paths"`
		}
		require.NoError(t, yaml.Unmarshal(data, &doc))
		for path, item := range doc.Paths {
			for verb := range item {
				ops[strings.ToUpper(verb)+" "+path] = rel
			}
		}
	}

	return ops
}

// collisions is each of b's operations a hand-written spec in ops also
// declares, with what to do about it.
func collisions(
	b *behavior.Behavior,
	ops map[string]string,
) []string {
	var out []string
	for _, m := range b.Methods {
		op := m.HTTPMethod() + " " + b.Path(m)
		if dir, ok := ops[op]; ok {
			out = append(out, fmt.Sprintf(
				"%s is served by the %s behavior and by %s: delete the hand-written domain "+
					"(its cutover)",
				op, b.Name, dir))
		}
	}

	return out
}

// TestNoRouteIsServedTwice fails while a behavior shares a route with a
// hand-written domain, naming the domain's directory: a domain's cutover
// deletes what this names in the change that adds its behavior.
func (s *AllPublicTestSuite) TestNoRouteIsServedTwice() {
	ops := handWritten(s.T())
	s.Require().NotEmpty(ops)

	for _, b := range behavior.All() {
		if !b.API {
			continue
		}
		s.Run(b.Name, func() {
			s.Empty(collisions(b, ops))
		})
	}

	s.Run("when a behavior shares a hand-written route", func() {
		served := behavior.Define("ntp", "NTP", behavior.Node, behavior.API,
			behavior.Read("list", sysctl.List).Describe("List"))
		got := collisions(served, ops)
		s.Require().Len(got, 1)
		s.Contains(got[0], "GET /api/node/{hostname}/ntp is served by the ntp behavior "+
			"and by internal/controller/api/node/ntp")
	})
}

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}
