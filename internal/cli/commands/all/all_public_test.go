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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
)

type AllPublicTestSuite struct {
	suite.Suite
}

// TestValidate holds each registered local group to the rules of a
// definition.
func (s *AllPublicTestSuite) TestValidate() {
	s.NotEmpty(command.All())
	for _, g := range command.All() {
		s.Run(strings.Join(g.CommandPath(), " "), func() {
			s.Empty(command.Validate(g))
		})
	}
}

// TestRefusals runs each input that a method's Fails declares, as the
// generated command does, and expects a usage error: the command exits 2,
// and the handler does not run. Validate makes each rule of an input have
// one, so this covers each rule of each local command.
func (s *AllPublicTestSuite) TestRefusals() {
	count := 0
	for _, g := range command.All() {
		for _, m := range g.Methods {
			for _, r := range m.Refusals() {
				count++
				s.Run(g.Name+" "+m.MethodName()+": "+r.Doc, func() {
					err := command.Run(&cobra.Command{}, g, m.MethodName(), r.Input)

					var ue *command.UsageError
					s.Require().True(errors.As(err, &ue), "a usage error, not %v", err)
					s.NotEmpty(ue.Msg)
					s.True(strings.HasPrefix(ue.Msg, "--"), "it names the flag: %s", ue.Msg)
				})
			}
		}
	}
	s.NotZero(count)
}

// TestEachPackageRegistered holds that each package of commands is in the
// binary: all.go imports it, and gen.go imports its generated commands. A
// package that neither imported, such as agentservice at its merge, had
// tests that passed and no command in the binary.
func (s *AllPublicTestSuite) TestEachPackageRegistered() {
	all, err := os.ReadFile("all.go")
	s.Require().NoError(err)
	gen, err := os.ReadFile(filepath.Join("gen", "gen.go"))
	s.Require().NoError(err)
	dirs, err := os.ReadDir("..")
	s.Require().NoError(err)
	checked := 0
	for _, d := range dirs {
		if !d.IsDir() || d.Name() == "all" {
			continue
		}
		files, _ := filepath.Glob(filepath.Join("..", d.Name(), "*.go"))
		registers := false
		for _, f := range files {
			data, _ := os.ReadFile(f)
			if !strings.HasSuffix(f, "_test.go") &&
				strings.Contains(string(data), "command.Register(") {
				registers = true
			}
		}
		if !registers {
			continue
		}
		checked++
		pkg := "github.com/osapi-io/osapi/internal/cli/commands/" + d.Name()
		s.Contains(string(all), `"`+pkg+`"`, "all.go does not import %s", d.Name())
		s.Contains(
			string(gen),
			`"`+pkg+`/gen"`,
			"gen/gen.go does not import %s/gen: run go generate ./internal/cli/commands/all",
			d.Name(),
		)
	}
	s.NotZero(checked)
}

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}
