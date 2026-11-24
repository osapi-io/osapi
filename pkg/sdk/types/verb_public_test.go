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

	"github.com/osapi-io/osapi/pkg/sdk/types"
	"github.com/stretchr/testify/suite"
)

type VerbPublicTestSuite struct {
	suite.Suite
}

func (s *VerbPublicTestSuite) TestNewPermission() {
	tests := []struct {
		name     string
		resource string
		verb     types.Verb
		want     types.Permission
	}{
		{
			name:     "when a read",
			resource: "sysctl",
			verb:     types.VerbRead,
			want:     types.PermSysctlRead,
		},
		{
			name:     "when an admin verb",
			resource: "docker",
			verb:     types.VerbExec,
			want:     types.PermDockerExec,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, types.NewPermission(tc.resource, tc.verb))
		})
	}
}

func TestVerbPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(VerbPublicTestSuite))
}
