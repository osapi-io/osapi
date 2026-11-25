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

package sdktypes_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/guide/sdktypes"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type SDKTypesPublicTestSuite struct {
	suite.Suite
}

// TestAppend adds a section of each closed set of the SDK after the spec's
// own, and leaves the spec it was given as it was.
func (s *SDKTypesPublicTestSuite) TestAppend() {
	in := guide.Spec{
		Kinds: []guide.Kind{{Kind: "option", Title: "Options"}},
		Terms: []guide.Term{{Kind: "option", Name: "Doc(text)"}},
	}

	out := sdktypes.Append(in)

	s.Len(in.Kinds, 1, "the spec it was given is unchanged")
	s.Equal(sdktypes.Kind, out.Kinds[1].Kind)
	s.Len(out.Of(sdktypes.Kind), len(types.All()))
}

// TestTerms names each type as a caller writes it, with its values, and
// says to use it with the enum rule.
func (s *SDKTypesPublicTestSuite) TestTerms() {
	terms := sdktypes.Terms([]types.TypeSpec{
		{Name: "Shade", Doc: "Shade is a tone.", Values: []string{"light", "dark"}},
	})

	s.Require().Len(terms, 1)
	s.Equal("types.Shade", terms[0].Name)
	s.Equal("Shade is a tone. Values: light, dark.", terms[0].Does)
	s.Contains(terms[0].When, "validate:\"enum\"")
	s.Contains(terms[0].When, "validate refuses oneof")
}

func TestSDKTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SDKTypesPublicTestSuite))
}
