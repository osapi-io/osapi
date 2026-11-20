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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type GuidePublicTestSuite struct {
	suite.Suite
}

func (s *GuidePublicTestSuite) TestPrintBlock() {
	got := captureStdout(func() {
		cli.PrintHeading("Options")
		cli.PrintBlock("Flat()", "", "Adds no word.", "Use", "Alone.")
		cli.PrintBlock("Long()", "", "Help.")
	})
	s.Equal(
		"\nOptions:\n  Flat()\n    Adds no word.\n    Use: Alone.\n\n  Long()\n    Help.\n",
		got,
	)
}

// TestPrintBlockWraps breaks a long paragraph at spaces, under the title.
func (s *GuidePublicTestSuite) TestPrintBlockWraps() {
	got := captureStdout(func() {
		cli.PrintHeading("Keys")
		cli.PrintBlock("Key", "", strings.TrimSpace(strings.Repeat("word ", 16)))
	})
	s.Equal(
		"\nKeys:\n  Key\n    "+strings.TrimSpace(strings.Repeat("word ", 14))+"\n    word word\n",
		got,
	)
}

func TestGuidePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GuidePublicTestSuite))
}
