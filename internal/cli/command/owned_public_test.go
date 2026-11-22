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

package command_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// OwnedPublicTestSuite covers the checks the CLI owns and the behavior DSL
// asks of it (osapi #973, #981): which types a flag holds, and which
// output tags the output knows.
type OwnedPublicTestSuite struct {
	suite.Suite
}

func (s *OwnedPublicTestSuite) TestFlaggable() {
	for t, want := range map[reflect.Type]bool{
		reflect.TypeFor[string]():            true,
		reflect.TypeFor[bool]():              true,
		reflect.TypeFor[int]():               true,
		reflect.TypeFor[[]string]():          true,
		reflect.TypeFor[map[string]string](): true,
		reflect.TypeFor[[]byte]():            true,
		reflect.TypeFor[int64]():             false,
		reflect.TypeFor[float64]():           false,
		reflect.TypeFor[[]int]():             false,
	} {
		s.Equal(want, command.Flaggable(t), t.String())
	}
}

func (s *OwnedPublicTestSuite) TestTagProblems() {
	type out struct {
		Size int `json:"size" render:"bytes"`
		Odd  int `json:"odd"  render:"furlongs"`
	}
	got := command.TagProblems(reflect.TypeFor[out]())
	s.Len(got, 1)
	s.Contains(got[0], `render:"furlongs"`)
}

func (s *OwnedPublicTestSuite) TestTagKeys() {
	tests := []struct {
		name string
		tag  reflect.StructTag
		want []string
	}{
		{
			name: "when the tag has keys, each in order",
			tag:  `json:"a" doc:"The \"a\""`,
			want: []string{"json", "doc"},
		},
		{
			name: "when the tag ends in a space, it stops there",
			tag:  `json:"a" `,
			want: []string{"json"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, command.TagKeys(tc.tag))
		})
	}
}

func TestOwnedPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OwnedPublicTestSuite))
}
