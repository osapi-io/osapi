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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/dsl"
	"github.com/osapi-io/osapi/internal/skills"
)

type DSLPublicTestSuite struct {
	suite.Suite
}

// TestEveryDSLHasTheSharedParts holds each DSL to the contract that they
// share: each shared command under the command of its group, unless it says
// why the command does not apply, its skill in the binary, and its docs page.
// A change that reaches one DSL and not the others fails here.
func (s *DSLPublicTestSuite) TestEveryDSLHasTheSharedParts() {
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
			for _, c := range dsl.Commands() {
				path := strings.Join(d.Command, " ") + " " + c
				why, without := d.Without[c]
				if without {
					s.NotEmpty(why, "%s says %s does not apply, with no reason", d.Name, c)
					s.False(
						paths[path],
						"osapi %s exists, but %s says it does not apply",
						path,
						d.Name,
					)

					continue
				}
				s.True(
					paths[path],
					"osapi %s does not exist: add it, or say in Without why it does not apply",
					path,
				)
			}
			for c := range d.Without {
				s.Contains(
					dsl.Commands(),
					c,
					"%s names %q in Without, which is not a shared command",
					d.Name,
					c,
				)
			}
			_, openapi := d.Without[dsl.CommandOpenAPI]
			s.Equal(d.ServesAPI(), !openapi,
				"a DSL that serves the API has openapi get, and one that does not says so")
			s.Equal(d.ServesAPI(), d.Operations != nil,
				"a DSL that serves the API gives its permissions and its operations")

			_, err := fs.Stat(skills.Dev(), d.Skill+"/SKILL.md")
			s.NoError(err, "the skill %s is not in the binary", d.Skill)
			_, err = os.Stat(filepath.Join(
				"..", "..", "..", "..", "docs", "docs", "sidebar", "development", "dsl", d.Page))
			s.NoError(err, "the docs page %s does not exist", d.Page)
		})
	}
	s.Equal([]string{"behavior", "chain", "command", "endpoint"}, names)
}

func TestDSLPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DSLPublicTestSuite))
}
