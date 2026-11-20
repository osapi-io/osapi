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

package behaviordsl

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type MethodsTestSuite struct {
	suite.Suite
}

func (s *MethodsTestSuite) TestConstraints() {
	one, two := 1, 2
	low, high := 0.5, 9.0
	got := constraints(&behavior.Schema{
		Pattern: "^a$", Format: "ipv4", Enum: []string{"a", "b"},
		MinLength: &one, MaxLength: &two, Minimum: &low, Maximum: &high,
	})
	s.Equal("pattern ^a$, format ipv4, one of a b, min length 1, max length 2, min 0.5, max 9", got)
}

func (s *MethodsTestSuite) TestShape() {
	s.Equal(
		"ListInput",
		shape("ListInput", &behavior.Schema{}, true),
		"a type with no field is its name",
	)
}

func (s *MethodsTestSuite) TestPairs() {
	s.Equal("key=a, value=1", pairs(map[string]any{"key": "a", "value": 1}))
	s.Equal("[key=a] [key=b]", pairs([]map[string]string{{"key": "a"}, {"key": "b"}}))
	s.Equal("(none)", pairs(struct{}{}))
	s.Equal(`"text"`, pairs("text"))
}

func TestMethodsTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsTestSuite))
}
