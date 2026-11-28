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
package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type TypesGenPublicTestSuite struct {
	suite.Suite
}

func TestTypesGenPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TypesGenPublicTestSuite))
}

// TestDoc: a closed set says what it is.
func (s *TypesGenPublicTestSuite) TestDoc() {
	tests := []struct {
		name         string
		doc          func() string
		validateFunc func(doc string)
	}{
		{
			name: "when the set is ContentType",
			doc:  types.ContentRaw.Doc,
			validateFunc: func(doc string) {
				s.Contains(doc, "how a deploy treats a file")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(tc.doc())
		})
	}
}

// TestAll: each closed set of the SDK is listed once, says what it is, and
// lists values.
func (s *TypesGenPublicTestSuite) TestAll() {
	tests := []struct {
		name         string
		validateFunc func(all []types.TypeSpec)
	}{
		{
			name: "when each set is listed once with a doc and values",
			validateFunc: func(all []types.TypeSpec) {
				s.NotEmpty(all)
				seen := map[string]bool{}
				for _, t := range all {
					s.False(seen[t.Name])
					seen[t.Name] = true
					s.NotEmpty(t.Doc)
					s.NotEmpty(t.Values)
				}
				s.True(seen["ContentType"])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(types.All())
		})
	}
}
