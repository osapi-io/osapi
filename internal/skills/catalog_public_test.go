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

package skills_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/skills"
)

type CatalogPublicTestSuite struct {
	suite.Suite

	installed fstest.MapFS
	dev       fstest.MapFS
}

func (s *CatalogPublicTestSuite) SetupTest() {
	s.installed = fstest.MapFS{
		"osapi/SKILL.md": {
			Data: []byte(
				"---\nname: osapi\ndescription: >\n  Operate hosts,\n  through osapi.\n---\nbody\n",
			),
		},
		"osapi/references/targets.md": {Data: []byte("targets\n")},
		"stray.txt":                   {Data: []byte("not a skill\n")},
	}
	s.dev = fstest.MapFS{
		"build-a-behavior/SKILL.md": {
			Data: []byte(
				"---\nname: build-a-behavior\ndescription: Build a behavior.\n---\nbuild\n",
			),
		},
		"build-a-behavior/references/services.md": {Data: []byte("services\n")},
	}
}

func (s *CatalogPublicTestSuite) TestCatalog() {
	tests := []struct {
		name      string
		installed fstest.MapFS
		dev       fstest.MapFS
		want      []skills.Skill
		wantErr   string
	}{
		{
			name: "when both sets have skills lists them sorted, with init set",
			want: []skills.Skill{
				{
					Name:        "build-a-behavior",
					Description: "Build a behavior.",
					Init:        false,
					References:  []string{"references/services.md"},
				},
				{
					Name:        "osapi",
					Description: "Operate hosts, through osapi.",
					Init:        true,
					References:  []string{"references/targets.md"},
				},
			},
		},
		{
			name:      "when a directory has no SKILL.md returns an error",
			installed: fstest.MapFS{"broken/README.md": {Data: []byte("x")}},
			wantErr:   `skill "broken"`,
		},
		{
			name: "when the front matter is not YAML returns an error",
			installed: fstest.MapFS{"bad/SKILL.md": {
				Data: []byte("---\nname: bad\ndescription: [unclosed\n---\n"),
			}},
			wantErr: `skill "bad": front matter`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			installed, dev := s.installed, s.dev
			if tt.installed != nil {
				installed = tt.installed
			}
			if tt.dev != nil {
				dev = tt.dev
			}

			got, err := skills.Catalog(installed, dev)

			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.wantErr)

				return
			}
			s.Require().NoError(err)
			if tt.want != nil {
				s.Equal(tt.want, got)
			}
		})
	}
}

func (s *CatalogPublicTestSuite) TestCatalogUnreadable() {
	_, err := skills.Catalog(s.installed, unreadable{})

	s.Require().Error(err)
	s.Contains(err.Error(), "list skills")
}

func (s *CatalogPublicTestSuite) TestRead() {
	tests := []struct {
		name      string
		skill     string
		reference string
		want      string
		wantErr   string
	}{
		{
			name:  "when the skill is installed returns its SKILL.md",
			skill: "osapi",
			want:  "---\nname: osapi\ndescription: >\n  Operate hosts,\n  through osapi.\n---\nbody\n",
		},
		{
			name:  "when the skill is a dev skill returns its SKILL.md",
			skill: "build-a-behavior",
			want:  "---\nname: build-a-behavior\ndescription: Build a behavior.\n---\nbuild\n",
		},
		{
			name:      "when a reference is given returns that file",
			skill:     "build-a-behavior",
			reference: "references/services.md",
			want:      "services\n",
		},
		{
			name:    "when the skill does not exist returns not found",
			skill:   "nope",
			wantErr: `skill "nope": not found`,
		},
		{
			name:      "when the reference does not exist returns not found",
			skill:     "osapi",
			reference: "references/nope.md",
			wantErr:   "not found",
		},
		{
			name:      "when the reference climbs out of the skill returns not found",
			skill:     "osapi",
			reference: "../build-a-behavior/SKILL.md",
			wantErr:   "not found",
		},
		{
			name:    "when the name has a slash returns not found",
			skill:   "osapi/references",
			wantErr: "not found",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := skills.Read(s.installed, s.dev, tt.skill, tt.reference)

			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.wantErr)
				s.ErrorIs(err, skills.ErrNotFound)

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.want, string(got))
		})
	}
}

// unreadable is a file system whose root cannot be read.
type unreadable struct{}

func (unreadable) Open(string) (fs.File, error) { return nil, errors.New("denied") }

func TestCatalogPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CatalogPublicTestSuite))
}
