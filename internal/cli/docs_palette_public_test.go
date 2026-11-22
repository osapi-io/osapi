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
package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type DocsPalettePublicTestSuite struct {
	suite.Suite
}

// TestDocsPaletteCSS: the light page has the dark tints, which read on
// white, and the dark page the logo colors, as the UI and the CLI use them.
func (s *DocsPalettePublicTestSuite) TestDocsPaletteCSS() {
	css := cli.DocsPaletteCSS()

	s.Contains(css, ":root {\n  --osapi-heading-1: #0a8fb0;")
	s.Contains(css, "[data-theme='dark'] {\n  --osapi-heading-1: #00d5ff;")
	s.NotContains(css, "--osapi-link", "a link is a heading color, not a color of its own")
	s.Contains(css, "  --osapi-accent: #2300fd;")
	s.Contains(css, "  --osapi-ok: #67ea94;")
	s.Contains(css, "  --osapi-warn: #ffd787;")
	s.Contains(css, "  --osapi-error: #ef4444;")
}

// TestDocsPaletteTS: the code blocks' palette has the same colors.
func (s *DocsPalettePublicTestSuite) TestDocsPaletteTS() {
	ts := cli.DocsPaletteTS()

	s.Contains(ts, "export const palette = {")
	s.Contains(ts, `headings: ["#0a8fb0", "#b030b0", "#b08f00", "#3a8f14", "#5a4cd8"],`)
	s.Contains(ts, `headings: ["#00d5ff", "#ff00ff", "#ffd100", "#45ae17", "#6b5cff"],`)
	s.Contains(ts, "} as const;")
}

// TestFilesAreCurrent: the docs site's files are what the palette makes.
func (s *DocsPalettePublicTestSuite) TestFilesAreCurrent() {
	for file, want := range map[string]string{
		cli.DocsPaletteCSSFile: cli.DocsPaletteCSS(),
		cli.DocsPaletteTSFile:  cli.DocsPaletteTS(),
	} {
		got, err := os.ReadFile(filepath.Join("..", "..", file))
		s.Require().NoError(err)
		s.Equal(want, string(got), "%s is stale: run go generate ./internal/behaviors/all", file)
	}
}

func TestDocsPalettePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DocsPalettePublicTestSuite))
}
