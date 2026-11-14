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
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	_ "github.com/osapi-io/osapi/internal/behaviors/all/gen"
	behaviorapi "github.com/osapi-io/osapi/internal/controller/api/behavior"
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

func (s *AllPublicTestSuite) TestEveryRouteIsMounted() {
	e := echo.New()
	for _, register := range behaviorapi.Handler(slog.Default(), nil, "", nil) {
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
			route := m.HTTPMethod() + " " + behaviorapi.Path(b, m)
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
	root := &cobra.Command{Use: "next"}
	behaviorcmd.Attach(root, behaviorcmd.Deps{})

	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			line := b.CLICommand(m)
			s.Run(line, func() {
				args := strings.Fields(strings.TrimPrefix(line, behavior.CommandPrefix+" "))[1:]
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
