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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	_ "github.com/osapi-io/osapi/internal/behaviors/all/gen"
	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
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

// TestUIOperationsAreCurrent checks that the UI's permission of each
// operation, which go generate writes from the definitions and the
// hand-written domains' specs, is what the joined spec says, and that the
// file is current.
func (s *AllPublicTestSuite) TestUIOperationsAreCurrent() {
	root := filepath.Join("..", "..", "..")
	got, err := authtoken.Operations(root)
	s.Require().NoError(err)
	spec, err := os.ReadFile(
		filepath.Join(root, "internal", "controller", "api", "gen", "api.yaml"),
	)
	s.Require().NoError(err)
	joined, err := authtoken.OperationPermissions(spec)
	s.Require().NoError(err)
	s.Equal(joined, got, "the definitions and the hand-written specs agree with the joined spec")
	file, err := os.ReadFile(filepath.Join(root, authtoken.OperationsFile))
	s.Require().NoError(err)
	s.Equal(
		authtoken.OperationsTypeScript(got),
		string(file),
		"%s is stale: run go generate ./internal/authtoken/gen/operations",
		authtoken.OperationsFile,
	)
}

// TestPermissionsAreDeclared checks the permissions of every registered
// endpoint group and behavior: a method that names a permission that nothing
// declares answers 403 to every caller, and a permission that is admin only
// on one method must be on each.
func (s *AllPublicTestSuite) TestPermissionsAreDeclared() {
	s.Empty(authtoken.Check())
}

// TestSDKPermissionsAreCurrent checks that the SDK's permission constants,
// which go generate writes from internal/authtoken, are each permission a
// token can have now, a behavior's included.
func (s *AllPublicTestSuite) TestSDKPermissionsAreCurrent() {
	got, err := os.ReadFile(filepath.Join("..", "..", "..", authtoken.GoFile))
	s.Require().NoError(err)
	s.Equal(authtoken.Go(), string(got),
		"%s is stale: run go generate ./internal/behaviors/all", authtoken.GoFile)
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

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}
