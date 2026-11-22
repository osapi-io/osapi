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
	"os"
	"path/filepath"
	"strings"
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
	s.Equal(map[string]string{"getNodeStatus": "status:read"}, got,
		"a public operation needs no permission, and parameters is not an operation")

	ts := authtoken.OperationsTypeScript(got)
	s.Contains(ts, "  getNodeStatus: \"status:read\",\n")
	s.Contains(ts, "export type OperationId")
}

func (s *OperationsPublicTestSuite) TestOperationPermissionsRefusesBadYAML() {
	_, err := authtoken.OperationPermissions([]byte("paths: ["))
	s.Error(err)
	_, err = authtoken.OperationPermissions([]byte("paths:\n  /a:\n    get: [1]\n"))
	s.Error(err)
}

func (s *OperationsPublicTestSuite) TestRegistryOperations() {
	got := authtoken.RegistryOperations()
	motd := behavior.Lookup("motdperm")
	for _, m := range motd.Methods {
		id := motd.OperationID(m)
		s.Equal(string(motd.Permission(m.PermissionVerb())), got[strings.ToLower(id[:1])+id[1:]],
			"a served behavior's method has its permission")
	}
	hidden := behavior.Lookup("hiddenperm")
	s.NotContains(got, "get"+hidden.OperationID(hidden.Methods[0])[3:],
		"an internal behavior has no route")
	s.Equal("helth:read", got["getZztypoStatus"], "a method that names a permission has it")
	s.Equal("zzsplit:write", got["postZzsplitOne"])
	s.NotContains(got, "getZzsplitOpen", "a public method needs no permission")
}

func (s *OperationsPublicTestSuite) TestOperations() {
	spec := `
paths:
  /node/{hostname}/clock:
    get:
      operationId: GetNodeClock
      security:
        - BearerAuth:
            - clock:read
`
	write := func(root, domain, text string) {
		dir := filepath.Join(root, "internal", "controller", "api", "node", domain, "gen")
		s.Require().NoError(os.MkdirAll(dir, 0o755))
		s.Require().NoError(os.WriteFile(filepath.Join(dir, "api.yaml"), []byte(text), 0o600))
	}

	tests := []struct {
		name         string
		setup        func(root string)
		root         func(root string) string
		validateFunc func(map[string]string, error)
	}{
		{
			name:  "when a hand-written domain has a spec, its operations join the definitions'",
			setup: func(root string) { write(root, "clock", spec) },
			validateFunc: func(got map[string]string, err error) {
				s.Require().NoError(err)
				s.Equal("clock:read", got["getNodeClock"])
				s.Equal("helth:read", got["getZztypoStatus"])
			},
		},
		{
			name: "when a spec gives an operation another permission than a definition",
			setup: func(root string) {
				write(root, "typo", `
paths:
  /zztypo/status:
    get:
      operationId: GetZztypoStatus
      security:
        - BearerAuth:
            - health:read
`)
			},
			validateFunc: func(_ map[string]string, err error) {
				s.ErrorContains(err, "operation getZztypoStatus needs health:read")
			},
		},
		{
			name:  "when a spec is not YAML",
			setup: func(root string) { write(root, "bad", "paths: [") },
			validateFunc: func(_ map[string]string, err error) {
				s.ErrorContains(err, "read the spec")
			},
		},
		{
			name: "when a spec cannot be read",
			setup: func(root string) {
				s.Require().NoError(os.MkdirAll(
					filepath.Join(
						root,
						"internal",
						"controller",
						"api",
						"node",
						"dir",
						"gen",
						"api.yaml",
					),
					0o755,
				))
			},
			validateFunc: func(_ map[string]string, err error) {
				s.ErrorContains(err, "read ")
			},
		},
		{
			name: "when the root makes a bad pattern",
			root: func(root string) string { return filepath.Join(root, "[") },
			validateFunc: func(_ map[string]string, err error) {
				s.ErrorContains(err, "find the hand-written specs")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			root := s.T().TempDir()
			if tc.setup != nil {
				tc.setup(root)
			}
			if tc.root != nil {
				root = tc.root(root)
			}
			tc.validateFunc(authtoken.Operations(root))
		})
	}
}

func TestOperationsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OperationsPublicTestSuite))
}
