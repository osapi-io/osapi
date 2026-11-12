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
	"errors"
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type HelpPublicTestSuite struct {
	suite.Suite
}

func (suite *HelpPublicTestSuite) TearDownTest() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// newRoot builds a small command tree with the theme applied.
func newRoot() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "osapi", Short: "test root"}
	child := &cobra.Command{
		Use:   "get",
		Short: "get a thing",
		RunE:  func(*cobra.Command, []string) error { return nil },
	}
	child.Flags().String("key", "", "the key")
	_ = child.MarkFlagRequired("key")
	root.AddCommand(child)
	root.PersistentFlags().BoolP("json", "j", false, "Enable JSON output")

	cli.ApplyTheme(root)

	return root, child
}

func (suite *HelpPublicTestSuite) TestHelp() {
	tests := []struct {
		name         string
		color        bool
		args         []string
		validateFunc func(string)
	}{
		{
			name: "when plain root help keeps cobra's text and no logo",
			args: []string{"--help"},
			validateFunc: func(out string) {
				assert.Contains(suite.T(), out, "Usage:")
				assert.Contains(suite.T(), out, "Available Commands:")
				assert.Contains(suite.T(), out, "Flags:\n")
				assert.NotContains(suite.T(), out, "██")
				assert.NotContains(suite.T(), out, "\x1b[")
			},
		},
		{
			name:  "when colored root help draws the logo above the help",
			color: true,
			args:  []string{"--help"},
			validateFunc: func(out string) {
				assert.Contains(suite.T(), out, "██")
				assert.Contains(suite.T(), out, "Usage:")
			},
		},
		{
			name:  "when colored child help has no logo",
			color: true,
			args:  []string{"get", "--help"},
			validateFunc: func(out string) {
				assert.NotContains(suite.T(), out, "██")
				assert.Contains(suite.T(), out, "--key")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			lipgloss.SetColorProfile(termenv.Ascii)
			if tc.color {
				lipgloss.SetColorProfile(termenv.TrueColor)
			}

			root, _ := newRoot()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetArgs(tc.args)

			suite.Require().NoError(root.Execute())
			tc.validateFunc(out.String())
		})
	}
}

func (suite *HelpPublicTestSuite) TestIsUsageError() {
	tests := []struct {
		name string
		args []string
		err  error
		want bool
	}{
		{
			name: "when the command is unknown",
			args: []string{"gett"},
			want: true,
		},
		{
			name: "when a flag is unknown",
			args: []string{"get", "--nope"},
			want: true,
		},
		{
			name: "when a required flag is missing",
			args: []string{"get"},
			want: true,
		},
		{
			name: "when the error came from running the command",
			err:  errors.New("connection refused"),
			want: false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := tc.err
			if err == nil {
				root, _ := newRoot()
				root.SetArgs(tc.args)
				_, err = root.ExecuteC()
				suite.Require().Error(err)
			}

			assert.Equal(suite.T(), tc.want, cli.IsUsageError(err))
		})
	}
}

func (suite *HelpPublicTestSuite) TestPrintError() {
	tests := []struct {
		name         string
		args         []string
		err          error
		json         bool
		validateFunc func(string)
	}{
		{
			name: "when a usage error points at the help",
			args: []string{"get", "--nope"},
			validateFunc: func(out string) {
				assert.Contains(suite.T(), out, "[err] unknown flag: --nope")
				assert.Contains(suite.T(), out, "Run 'osapi get --help' for usage.")
			},
		},
		{
			name: "when a runtime error has no help pointer",
			err:  errors.New("connection refused"),
			validateFunc: func(out string) {
				assert.Equal(suite.T(), "[err] connection refused\n", out)
			},
		},
		{
			name: "when json prints a plain error line",
			args: []string{"get", "--nope"},
			json: true,
			validateFunc: func(out string) {
				assert.Equal(suite.T(), "Error: unknown flag: --nope\n", out)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			root, child := newRoot()
			cmd, err := child, tc.err
			if err == nil {
				root.SetArgs(tc.args)
				cmd, err = root.ExecuteC()
			}

			var out bytes.Buffer
			cli.PrintError(&out, cmd, err, tc.json)

			tc.validateFunc(out.String())
		})
	}
}

func (suite *HelpPublicTestSuite) TestUsageErrorUnwrap() {
	root, _ := newRoot()
	root.SetArgs([]string{"get", "--nope"})
	_, err := root.ExecuteC()

	assert.Equal(suite.T(), "unknown flag: --nope", fmt.Sprint(err))
	assert.NotNil(suite.T(), errors.Unwrap(err))
}

func TestHelpPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HelpPublicTestSuite))
}
