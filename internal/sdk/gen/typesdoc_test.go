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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TypesDocTestSuite struct {
	suite.Suite

	dir string
}

func (s *TypesDocTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
}

func (s *TypesDocTestSuite) put(
	name string,
	text string,
) {
	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, name), []byte(text), 0o600))
}

// TestTypesDocFile writes Doc for each type with Values, from its comment,
// and All of them; a type with its own Doc, or with no Values, is left out.
func (s *TypesDocTestSuite) TestTypesDocFile() {
	tests := []struct {
		name         string
		setup        func()
		dir          func() string
		validateFunc func(string, bool, error)
	}{
		{
			name: "when a type has Values, it gets Doc and a row of All",
			setup: func() {
				s.put(
					"shade.go",
					"package types\n\n// Shade is a tone\n// of a colour.\ntype Shade string\n\n"+
						"func (Shade) Values() []Shade { return nil }\n\n"+
						"// Tint has its own.\ntype Tint string\n\nfunc (Tint) Values() []Tint { return nil }\n\n"+
						"func (Tint) Doc() string { return \"\" }\n\n// Name is free.\ntype Name string\n",
				)
				s.put("shade_test.go", "package types\n")
				s.put(typesDocName, "this is not Go")
			},
			validateFunc: func(src string, ok bool, err error) {
				s.Require().NoError(err)
				s.True(ok)
				s.Contains(src, "func (Shade) Doc() string")
				s.Contains(src, "\"Shade is a tone of a colour.\"")
				s.NotContains(src, "func (Tint) Doc()")
				s.NotContains(src, "func (Name)")
				s.Contains(src, "func All() []TypeSpec")
			},
		},
		{
			name:  "when no type has Values, there is no file",
			setup: func() { s.put("name.go", "package types\n\ntype Name string\n") },
			validateFunc: func(_ string, ok bool, err error) {
				s.Require().NoError(err)
				s.False(ok)
			},
		},
		{
			name:  "when there is no package, there is no file",
			dir:   func() string { return filepath.Join(s.dir, "gone") },
			setup: func() {},
			validateFunc: func(_ string, ok bool, err error) {
				s.Require().NoError(err)
				s.False(ok)
			},
		},
		{
			name:  "when the path is a file, it is an error",
			dir:   func() string { return filepath.Join(s.dir, "file.go") },
			setup: func() { s.put("file.go", "package types\n") },
			validateFunc: func(_ string, _ bool, err error) {
				s.Error(err)
			},
		},
		{
			name:  "when a file does not parse, it is an error",
			setup: func() { s.put("bad.go", "package types\n\ntype\n") },
			validateFunc: func(_ string, _ bool, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.dir = s.T().TempDir()
			tc.setup()
			dir := s.dir
			if tc.dir != nil {
				dir = tc.dir()
			}
			f, err := typesDocFile(dir)
			var buf bytes.Buffer
			if f != nil {
				s.Require().NoError(f.Render(&buf))
			}
			tc.validateFunc(buf.String(), f != nil, err)
		})
	}
}

func TestTypesDocTestSuite(t *testing.T) {
	suite.Run(t, new(TypesDocTestSuite))
}
