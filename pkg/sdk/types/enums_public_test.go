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

type EnumsPublicTestSuite struct {
	suite.Suite
}

// TestValid: a value of a closed set is valid when its type lists it.
func (s *EnumsPublicTestSuite) TestValid() {
	tests := []struct {
		name  string
		valid func() bool
		want  bool
	}{
		{name: "when a content type is listed", valid: types.ContentTemplate.Valid, want: true},
		{name: "when a content type is not listed", valid: types.ContentType("zip").Valid},
		{name: "when a reason is listed", valid: types.ReasonNotManaged.Valid, want: true},
		{name: "when a reason is not listed", valid: types.Reason("not_manged").Valid},
		{name: "when a permission is listed", valid: types.PermFileRead.Valid, want: true},
		{name: "when a permission is not listed", valid: types.Permission("helth:read").Valid},
		{name: "when a job status is listed", valid: types.JobStatusSkipped.Valid, want: true},
		{name: "when a job status is not listed", valid: types.JobStatus("done").Valid},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.valid())
		})
	}
}

func TestEnumsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EnumsPublicTestSuite))
}
