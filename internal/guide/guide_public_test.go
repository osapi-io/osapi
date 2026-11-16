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

func TestGuidePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GuidePublicTestSuite))
}
