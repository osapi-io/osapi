// Copyright (c) 2024 John Dewey

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

package cli_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

// TreePublicTestSuite holds the rows of a tree in a table: the tree lines
// keep their spacing, and they are dim in any column.
type TreePublicTestSuite struct {
	suite.Suite
}

func (s *TreePublicTestSuite) TestTreeCell() {
	s.Equal("job.publish", ansi(cli.ExportRenderCell("SPAN", "job.publish")))
	s.Equal("│  └─ job.process", ansi(cli.ExportRenderCell("SPAN", "│  └─ job.process")))
	s.Equal("└─ api", ansi(cli.ExportRenderCell("TYPE", "└─ api")), "a tree in any column")
	s.Equal("   x", ansi(cli.ExportRenderCell("TYPE", "   x")), "spaces alone are no tree")
}

func (s *TreePublicTestSuite) TestTreeSpacingIsKept() {
	out := cli.Capture(func() {
		cli.PrintCompactTable([]cli.Section{{
			Title:     "Spans",
			Headers:   []string{"TYPE", "DETAIL"},
			Rows:      [][]string{{"│     └─ file.write", "a  b"}, {"two\nlines", "x"}},
			Reference: true,
		}})
	})

	plain := ansi(out)
	s.Contains(plain, "│     └─ file.write")
	s.Contains(plain, "a b", "other cells are still flattened")
	s.Contains(plain, "two lines")
}

// ansi drops the color codes of s.
func ansi(
	s string,
) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			skip = true
		case skip && r == 'm':
			skip = false
		case !skip:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func TestTreePublicTestSuite(t *testing.T) {
	suite.Run(t, new(TreePublicTestSuite))
}
