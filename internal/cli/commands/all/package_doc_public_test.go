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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PackageDocPublicTestSuite struct {
	suite.Suite
}

// packageDoc is a comment line that starts the doc of a package.
var packageDoc = regexp.MustCompile(`^// Package (\w+) `)

// TestEachPackageDocNamesItsPackage fails when a file of a local group has a
// "// Package x" comment for a package that is not its own. A copy of the
// token group's comment once headed each file of config, skill and version,
// and go doc joined it to each package's doc (osapi-io/osapi#1251).
func (s *PackageDocPublicTestSuite) TestEachPackageDocNamesItsPackage() {
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
		s.Require().NoError(err, f)
		data, err := os.ReadFile(f)
		s.Require().NoError(err, f)
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "package ") {
				break
			}
			if m := packageDoc.FindStringSubmatch(line); m != nil {
				pkg := strings.TrimSuffix(parsed.Name.Name, "_test")
				s.Equal(pkg, m[1], "%s: the doc names package %s", f, m[1])
			}
		}
	}
}

func TestPackageDocPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PackageDocPublicTestSuite))
}
