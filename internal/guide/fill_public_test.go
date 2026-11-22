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
package guide_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/guide"
)

type FillPublicTestSuite struct {
	suite.Suite
}

func TestFillPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FillPublicTestSuite))
}

func (s *FillPublicTestSuite) TestFill() {
	render := func(args []string) string { return "T(" + strings.Join(args, ",") + ")" }
	page := "A\n<!-- osapi:t x y -->\nold\n<!-- /osapi:t -->\nB\n<!-- osapi:t -->\n<!-- /osapi:t -->\nC\n"

	got := guide.Fill(page, "t", render)

	s.Equal("A\n<!-- osapi:t x y -->\n\nT(x,y)\n\n<!-- /osapi:t -->\nB\n"+
		"<!-- osapi:t -->\n\nT()\n\n<!-- /osapi:t -->\nC\n", got,
		"each text between markers is rendered again, with the words after the name")
	s.Equal(got, guide.Fill(got, "t", render), "a page filled again does not change")
	s.Equal(page, guide.Fill(page, "u", render), "markers of another name stay")

	for _, broken := range []string{
		"<!-- osapi:t x",
		"<!-- osapi:t x -->\nno end",
		"<!-- /osapi:t -->\n<!-- osapi:t x -->",
	} {
		s.Equal(broken, guide.Fill(broken, "t", render), "a start with no end stays")
	}
}
