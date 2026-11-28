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
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
	"github.com/osapi-io/osapi/internal/dsl"
	"github.com/osapi-io/osapi/internal/skills"
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
					s.Require().ErrorAs(err, &ue)
					s.Regexp(`^--`, ue.Msg)
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
	var missing []string
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
		if !strings.Contains(string(all), `"`+pkg+`"`) {
			missing = append(missing, "all.go: "+pkg)
		}
		if !strings.Contains(string(gen), `"`+pkg+`/gen"`) {
			missing = append(missing, "gen/gen.go: "+pkg+"/gen")
		}
	}
	s.NotZero(checked)
	s.Empty(missing)
}

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}

// TestEveryDSLHasTheSharedParts holds each DSL to the contract that they
// share: each shared command under the command of its group, unless it says
// why the command does not apply, its skill in the binary, and its docs page.
// A change that reaches one DSL and not the others fails here.
func (s *AllPublicTestSuite) TestEveryDSLHasTheSharedParts() {
	paths := map[string]bool{}
	for _, g := range command.All() {
		for _, m := range g.Methods {
			words := slices.Clone(g.CommandPath())
			if sub := m.SubPath(); sub != "" && sub != m.CommandName() {
				words = append(words, sub)
			}
			paths[strings.Join(append(words, m.CommandName()), " ")] = true
		}
	}

	all := dsl.All()
	names := make([]string, 0, len(all))
	for _, d := range all {
		names = append(names, d.Name)
		s.Run(d.Name, func() {
			var wrong []string
			for _, c := range dsl.Commands() {
				path := strings.Join(d.Command, " ") + " " + c
				why, without := d.Without[c]
				switch {
				case without && why == "":
					wrong = append(wrong, "no reason in Without: "+c)
				case without && paths[path]:
					wrong = append(wrong, "in Without but exists: osapi "+path)
				case !without && !paths[path]:
					wrong = append(wrong, "missing: osapi "+path)
				}
			}
			s.Empty(wrong)
			for c := range d.Without {
				s.Contains(dsl.Commands(), c)
			}
			_, openapi := d.Without[dsl.CommandOpenAPI]
			s.Equal(d.ServesAPI(), !openapi)
			s.Equal(d.ServesAPI(), d.Operations != nil)

			_, err := fs.Stat(skills.Dev(), d.Skill+"/SKILL.md")
			s.NoError(err)
			_, err = os.Stat(filepath.Join(
				"..", "..", "..", "..", "docs", "docs", "sidebar", "development", "dsl", d.Page))
			s.NoError(err)
		})
	}
	s.Equal([]string{"behavior", "chain", "command", "endpoint"}, names)
}

// packageDoc is a comment line that starts the doc of a package.
var packageDoc = regexp.MustCompile(`^// Package (\w+) `)

// TestEachPackageDocNamesItsPackage fails when a file of a local group has a
// "// Package x" comment for a package that is not its own. A copy of the
// token group's comment once headed each file of config, skill and version,
// and go doc joined it to each package's doc (osapi-io/osapi#1251).
func (s *AllPublicTestSuite) TestEachPackageDocNamesItsPackage() {
	files, err := filepath.Glob(filepath.Join("..", "*", "*.go"))
	s.Require().NoError(err)
	s.Require().NotEmpty(files)

	for _, f := range files {
		parsed, err := parser.ParseFile(
			token.NewFileSet(),
			f,
			nil,
			parser.ParseComments|parser.PackageClauseOnly,
		)
		s.Require().NoError(err)
		data, err := os.ReadFile(f)
		s.Require().NoError(err)
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "package ") {
				break
			}
			if m := packageDoc.FindStringSubmatch(line); m != nil {
				pkg := strings.TrimSuffix(parsed.Name.Name, "_test")
				s.Equal(f+": "+pkg, f+": "+m[1])
			}
		}
	}
}
