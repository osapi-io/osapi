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
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type ThemePublicTestSuite struct {
	suite.Suite
}

func (suite *ThemePublicTestSuite) TearDownTest() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// withColor makes the default renderer emit truecolor, as a terminal would.
func withColor() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func (suite *ThemePublicTestSuite) TestRoleHelpers() {
	tests := []struct {
		name      string
		color     bool
		render    func(*bytes.Buffer) string
		want      string
		wantColor bool
	}{
		{
			name:   "when plain Mute returns the text",
			render: func(b *bytes.Buffer) string { return cli.Mute(b, "x") },
			want:   "x",
		},
		{
			name:   "when plain OK returns the text",
			render: func(b *bytes.Buffer) string { return cli.OK(b, "x") },
			want:   "x",
		},
		{
			name:   "when plain Err returns the text",
			render: func(b *bytes.Buffer) string { return cli.Err(b, "x") },
			want:   "x",
		},
		{
			name:   "when plain Success marks with ok",
			render: func(b *bytes.Buffer) string { return cli.Success(b, "done") },
			want:   "[ok] done",
		},
		{
			name:   "when plain Failure marks with err",
			render: func(b *bytes.Buffer) string { return cli.Failure(b, "broke") },
			want:   "[err] broke",
		},
		{
			name:      "when colored Success uses a check",
			color:     true,
			render:    func(b *bytes.Buffer) string { return cli.Success(b, "done") },
			want:      "✓",
			wantColor: true,
		},
		{
			name:      "when colored Failure uses a cross",
			color:     true,
			render:    func(b *bytes.Buffer) string { return cli.Failure(b, "broke") },
			want:      "✗",
			wantColor: true,
		},
		{
			name:      "when colored Sub is bold",
			color:     true,
			render:    func(b *bytes.Buffer) string { return cli.Sub(b, "agent") },
			want:      "agent",
			wantColor: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			lipgloss.SetColorProfile(termenv.Ascii)
			if tc.color {
				withColor()
			}

			got := tc.render(&bytes.Buffer{})

			assert.Contains(suite.T(), got, tc.want)
			assert.Equal(suite.T(), tc.wantColor, strings.ContainsRune(got, 0x1b))
		})
	}
}

func (suite *ThemePublicTestSuite) TestFileWriter() {
	// A file that is not a terminal renders plain.
	f, err := os.CreateTemp(suite.T().TempDir(), "out")
	suite.Require().NoError(err)
	defer func() { _ = f.Close() }()

	assert.Equal(suite.T(), "x", cli.Mute(f, "x"))
}

func (suite *ThemePublicTestSuite) TestHeading() {
	tests := []struct {
		name  string
		order []lipgloss.Color
		count int
		want  []string
	}{
		{
			name:  "when headings are rendered each takes the next color",
			order: []lipgloss.Color{cli.LogoCyan, cli.LogoMagenta},
			count: 2,
			want:  []string{"0;213;255", "255;0;255"},
		},
		{
			name:  "when more headings than colors the order wraps",
			order: []lipgloss.Color{cli.LogoCyan, cli.LogoMagenta},
			count: 3,
			want:  []string{"0;213;255", "255;0;255", "0;213;255"},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			withColor()
			cli.SetHeadingOrder(tc.order)

			for i := 0; i < tc.count; i++ {
				got := cli.Heading(&bytes.Buffer{}, "Usage:")
				assert.Contains(suite.T(), got, tc.want[i])
				assert.Contains(suite.T(), got, "Usage:")
			}
		})
	}
}

func (suite *ThemePublicTestSuite) TestFlags() {
	usages := "  -j, --json                Enable JSON output\n" +
		"  -f, --osapi-file string   Path to config file"

	tests := []struct {
		name         string
		color        bool
		validateFunc func(string)
	}{
		{
			name: "when plain returns the usages unchanged",
			validateFunc: func(got string) {
				assert.Equal(suite.T(), usages, got)
			},
		},
		{
			name:  "when colored styles names and keeps descriptions",
			color: true,
			validateFunc: func(got string) {
				assert.Contains(suite.T(), got, "--json")
				assert.Contains(suite.T(), got, "--osapi-file")
				assert.Contains(suite.T(), got, "Path to config file")
				assert.Contains(suite.T(), got, "\x1b[")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			lipgloss.SetColorProfile(termenv.Ascii)
			if tc.color {
				withColor()
			}

			tc.validateFunc(cli.Flags(&bytes.Buffer{}, usages))
		})
	}
}

func (suite *ThemePublicTestSuite) TestBanner() {
	tests := []struct {
		name         string
		color        bool
		validateFunc func(string)
	}{
		{
			name: "when plain draws nothing",
			validateFunc: func(got string) {
				assert.Empty(suite.T(), got)
			},
		},
		{
			name:  "when colored draws the four rows of the logo",
			color: true,
			validateFunc: func(got string) {
				lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
				assert.Len(suite.T(), lines, 4)
				assert.Contains(suite.T(), got, "██")
				// Every color of the logo appears.
				for _, rgb := range []string{
					"138;138;138", "0;213;255", "69;174;23",
					"255;209;0", "255;0;255", "35;0;253", "18;12;76",
				} {
					assert.Contains(suite.T(), got, rgb)
				}
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			lipgloss.SetColorProfile(termenv.Ascii)
			if tc.color {
				withColor()
			}

			tc.validateFunc(cli.Banner(&bytes.Buffer{}))
		})
	}
}

func (suite *ThemePublicTestSuite) TestPrintJSONAndNote() {
	tests := []struct {
		name  string
		print func()
		want  string
	}{
		{
			name:  "when PrintJSON prints the body on one line",
			print: func() { cli.PrintJSON([]byte(`{"a":1}`)) },
			want:  "{\"a\":1}\n",
		},
		{
			name:  "when PrintNote prints an indented note",
			print: func() { cli.PrintNote("No files found.") },
			want:  "  No files found.\n",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			assert.Equal(suite.T(), tc.want, captureStdout(tc.print))
		})
	}
}

func TestThemePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ThemePublicTestSuite))
}
