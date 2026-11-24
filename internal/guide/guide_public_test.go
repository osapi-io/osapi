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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/guide"
)

type GuidePublicTestSuite struct {
	suite.Suite
}

func (s *GuidePublicTestSuite) TestSpecOf() {
	spec := guide.Spec{Terms: []guide.Term{
		{Kind: "option", Name: "Node"},
		{Kind: "tag", Name: `json:"name"`},
		{Kind: "option", Name: "API"},
	}}

	tests := []struct {
		name string
		kind string
		want []string
	}{
		{
			name: "when the kind has terms returns them in order",
			kind: "option",
			want: []string{"Node", "API"},
		},
		{name: "when the kind has none returns none", kind: "method", want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got []string
			for _, t := range spec.Of(tc.kind) {
				got = append(got, t.Name)
			}
			s.Equal(tc.want, got)
		})
	}
}

func (s *GuidePublicTestSuite) TestSections() {
	spec := guide.Spec{
		Kinds: []guide.Kind{{Kind: "option", Title: "Options"}, {Kind: "tag", Title: "Tags"}},
		Terms: []guide.Term{{Kind: "option", Name: "Node"}, {Kind: "option", Name: "API"}},
	}

	got := spec.Sections()
	s.Require().Len(got, 2)
	s.Equal("Options", got[0].Title)
	s.Len(got[0].Terms, 2)
	s.Equal("Tags", got[1].Title)
	s.Empty(got[1].Terms)
}

func (s *GuidePublicTestSuite) TestChecks() {
	checks := guide.Checks{
		{Rule: "name", Requires: "A name.", Fix: "Rename it."},
		{Rule: "doc", Requires: "A description.", Fix: "Describe it."},
	}

	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "when the rule has a check Fix returns its fix",
			validateFunc: func() {
				s.Equal("Describe it.", checks.Fix("doc"))
			},
		},
		{
			name: "when the rule has no check Fix returns nothing",
			validateFunc: func() {
				s.Empty(checks.Fix("nope"))
			},
		},
		{
			name: "when the rule has a check Requires returns what it requires",
			validateFunc: func() {
				s.Equal("A description.", checks.Requires("doc"))
				s.Empty(checks.Requires("nope"))
			},
		},
		{
			name: "when listed Help has one item for each rule",
			validateFunc: func() {
				s.Equal(
					"- `name`: A name. Fix: Rename it.\n- `doc`: A description. Fix: Describe it.\n",
					checks.Help(),
				)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, tc.validateFunc)
	}
}

func (s *GuidePublicTestSuite) TestMDX() {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "when the text has a tag or an expression",
			in:   "osapi-<key>.conf and {k}",
			want: "osapi-&lt;key&gt;.conf and &#123;k&#125;",
		},
		{
			name: "when the text has a code span",
			in:   "a `<b>` c <d>",
			want: "a `<b>` c &lt;d&gt;",
		},
		{
			name: "when the text has a bracket",
			in:   "Uses[Deps]()",
			want: `Uses\[Deps]()`,
		},
		{
			name: "when the text has a link, it stays a link",
			in:   "Refer to [Configuration](../../configuration.md) and [a <b>](x.md).",
			want: "Refer to [Configuration](../../configuration.md) and [a &lt;b&gt;](x.md).",
		},
		{
			name: "when the text has two asterisks",
			in:   "/etc/a/osapi-*.crt, /etc/b/osapi-*.crt",
			want: `/etc/a/osapi-\*.crt, /etc/b/osapi-\*.crt`,
		},
		{
			name: "when the text has an asterisk in a code span",
			in:   "`*slog.Logger` and *",
			want: "`*slog.Logger` and \\*",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, guide.MDX(tc.in))
		})
	}
}

func (s *GuidePublicTestSuite) TestMarkdown() {
	spec := guide.Spec{
		Kinds: []guide.Kind{{Kind: "a", Title: "Parts"}, {Kind: "b", Title: "Empty"}},
		Terms: []guide.Term{
			{Kind: "a", Name: `Key("field")`, Does: "Puts <it> in the path.", When: "On a get."},
			{Kind: "a", Name: "json:\"`x`\"", Does: "Names it.", When: "Always."},
		},
	}
	page := spec.Markdown("Behavior DSL", "The parts.", 2)
	s.Contains(page, "---\ntitle: Behavior DSL\nsidebar_position: 2\n---\n\n"+guide.Generated+
		"\n\n# Behavior DSL\n\nThe parts.\n")
	s.Contains(page, "\n## Parts\n")
	s.NotContains(page, "## Empty", "a kind with no term has no section")
	s.Contains(page, "\n### `Key(\"field\")`\n\nPuts &lt;it&gt; in the path.\n\nUse: On a get.\n")
	s.Contains(page, "### `` json:\"`x`\" ``", "a name with a backtick gets a longer fence")
}

func TestGuidePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GuidePublicTestSuite))
}
