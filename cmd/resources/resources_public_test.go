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

package resources_test

import (
	"io/fs"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/cmd/resources"
)

type ResourcesPublicTestSuite struct {
	suite.Suite
}

// frontmatter is what Claude Code and Codex need to load a skill: a YAML block
// first in the file with its name and description. A Markdown formatter once
// rewrote it into a heading, and the installed skill silently never loaded.
var frontmatter = regexp.MustCompile(`(?s)\A---\nname: ([a-z0-9-]+)\ndescription: .+?\n---\n`)

func (s *ResourcesPublicTestSuite) TestSkillsHaveFrontmatter() {
	dirs, err := fs.ReadDir(resources.Skills, "skills")
	s.Require().NoError(err)
	s.Require().NotEmpty(dirs)

	for _, d := range dirs {
		s.Run(d.Name(), func() {
			data, err := fs.ReadFile(resources.Skills, path.Join("skills", d.Name(), "SKILL.md"))
			s.Require().NoError(err)

			m := frontmatter.FindSubmatch(data)
			s.Require().NotNil(m, "SKILL.md must open with name and description frontmatter")
			s.Equal(d.Name(), string(m[1]), "the skill's name is its directory's")
		})
	}
}

func TestResourcesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ResourcesPublicTestSuite))
}
