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

type CatalogPublicTestSuite struct {
	suite.Suite
}

func TestCatalogPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CatalogPublicTestSuite))
}

// TestValid: a value of a closed set is valid when its type lists it.
func (s *CatalogPublicTestSuite) TestValid() {
	tests := []struct {
		name  string
		valid func() bool
		want  bool
	}{
		{name: "when a FieldIn is listed", valid: types.FieldInPath.Valid, want: true},
		{name: "when a FieldIn is not listed", valid: types.FieldIn("nope").Valid},
		{name: "when a FieldKind is listed", valid: types.FieldString.Valid, want: true},
		{name: "when a FieldKind is not listed", valid: types.FieldKind("nope").Valid},
		{name: "when a ColumnKind is listed", valid: types.ColumnValue.Valid, want: true},
		{name: "when a ColumnKind is not listed", valid: types.ColumnKind("nope").Valid},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.valid())
		})
	}
}
