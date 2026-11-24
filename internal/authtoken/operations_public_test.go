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

package authtoken_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
)

type OperationsPublicTestSuite struct {
	suite.Suite
}

func (s *OperationsPublicTestSuite) TestOperationPermissions() {
	spec := []byte(`
paths:
  /node/{hostname}/status:
    parameters:
      - name: hostname
    get:
      operationId: GetNodeStatus
      security:
        - BearerAuth:
            - status:read
  /health:
    get:
      operationId: GetHealth
`)
	got, err := authtoken.OperationPermissions(spec)
	s.Require().NoError(err)
	s.Equal(map[string]authtoken.Permission{"GetNodeStatus": "status:read"}, got,
		"a public operation needs no permission, and parameters is not an operation")
}

func (s *OperationsPublicTestSuite) TestOperationPermissionsRefusesBadYAML() {
	_, err := authtoken.OperationPermissions([]byte("paths: ["))
	s.Error(err)
	_, err = authtoken.OperationPermissions([]byte("paths:\n  /a:\n    get: [1]\n"))
	s.Error(err)
}

func (s *OperationsPublicTestSuite) TestOperations() {
	got := authtoken.Operations()
	motd := behavior.Lookup("motdperm")
	for _, m := range motd.Methods {
		id := motd.OperationID(m)
		s.Equal(motd.Permission(m.PermissionVerb()), got[id],
			"a served behavior's method has its permission")
	}
	hidden := behavior.Lookup("hiddenperm")
	s.NotContains(got, hidden.OperationID(hidden.Methods[0]),
		"an internal behavior has no route")
	s.Equal(authtoken.Permission("helth:read"), got["GetZztypoStatus"],
		"a method that names a permission has it")
	s.Equal(authtoken.Permission("zzsplit:write"), got["PostZzsplitOne"])
	s.NotContains(got, "GetZzsplitOpen", "a public method needs no permission")
}

func TestOperationsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OperationsPublicTestSuite))
}
