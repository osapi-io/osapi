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
	"github.com/osapi-io/osapi/internal/guide"
)

type GuidePublicTestSuite struct {
	suite.Suite
}

func (s *GuidePublicTestSuite) TestPrintTerms() {
	tests := []struct {
		name  string
		terms []guide.Term
		want  string
	}{
		{
			name: "when a term has a use prints the name, what it does and the use",
			terms: []guide.Term{
				{Name: "Node", Does: "Runs on a host.", When: "Always."},
				{Name: "API", Does: "Serves it."},
			},
			want: "  Node\n    Runs on a host.\n    Use: Always.\n\n  API\n    Serves it.\n",
		},
		{
			name: "when the text is long wraps it under the name",
			terms: []guide.Term{
				{Name: "Key", Does: strings.TrimSpace(strings.Repeat("word ", 16))},
			},
			want: "  Key\n    " + strings.TrimSpace(strings.Repeat("word ", 14)) +
				"\n    word word\n",
		},
		{
			name:  "when a word is longer than a line keeps it whole",
			terms: []guide.Term{{Name: "x", Does: strings.Repeat("y", 80) + " z"}},
			want:  "  x\n    " + strings.Repeat("y", 80) + "\n    z\n",
		},
		{name: "when there are no terms prints nothing", want: ""},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, captureStdout(func() { cli.PrintTerms(tc.terms) }))
		})
	}
}

func TestGuidePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GuidePublicTestSuite))
}
