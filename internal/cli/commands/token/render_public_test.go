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

package token_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/token"
)

type RenderPublicTestSuite struct {
	suite.Suite
}

// printed is what fn writes to stdout.
func (s *RenderPublicTestSuite) printed(
	fn func(),
) string {
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	stdout := os.Stdout
	os.Stdout = w
	fn()
	cli.FlushKV()
	os.Stdout = stdout
	s.Require().NoError(w.Close())
	out, err := io.ReadAll(r)
	s.Require().NoError(err)

	return string(out)
}

func (s *RenderPublicTestSuite) TestRender() {
	s.Require().NoError(command.AttachLocal(&cobra.Command{Use: "osapi"}, command.Deps{
		JSON: func() bool { return false },
	}))
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name   string
		method string
		in     any
		out    any
		want   []string
		absent []string
	}{
		{
			name:   "when a token is generated, it prints with what it carries",
			method: "generate",
			in:     token.GenerateInput{},
			out: token.GenerateOutput{
				Token: "signed", Subject: "ops", Roles: []string{"read", "write"},
				Permissions: []string{"node:read"},
			},
			want: []string{"signed", "ops", "read, write", "Permissions", "node:read"},
		},
		{
			name:   "when a token has no permissions, the line is left out",
			method: "generate",
			in:     token.GenerateInput{},
			out:    token.GenerateOutput{Token: "signed", Subject: "ops", Roles: []string{"read"}},
			want:   []string{"signed"},
			absent: []string{"Permissions"},
		},
		{
			name:   "when the roles print, each permission has a row and each role a column",
			method: "list",
			in:     token.RolesInput{},
			out: token.RolesOutput{Roles: []token.Role{
				{Name: "admin", Source: "built-in", Permissions: []string{"node:read"}},
				{Name: "ops", Source: "config"},
			}},
			want: []string{"PERMISSION", "ADMIN", "OPS", "node:read", "yes"},
		},
		{
			name:   "when claims print, they have the effective permissions and the times",
			method: "validate",
			in:     token.ValidateInput{},
			out: token.Claims{
				Subject: "ops", Roles: []string{"read"}, Permissions: []string{"node:read"},
				EffectivePermissions: []string{"node:read"}, Audience: []string{"osapi"},
				IssuedAt: at, ExpiresAt: at,
			},
			want: []string{
				"Token",
				"ops",
				"Permissions",
				"Effective permissions",
				"osapi",
				"Expires",
			},
		},
		{
			name:   "when claims have no permissions, the line is left out",
			method: "validate",
			in:     token.ValidateInput{},
			out:    token.Claims{Subject: "ops", Roles: []string{"read"}},
			want:   []string{"Effective permissions"},
			absent: []string{"  Permissions"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			body, err := json.Marshal(tc.out)
			s.Require().NoError(err)

			got := s.printed(func() {
				command.Output(&cobra.Command{}, token.Group.Name, tc.method, tc.in, body)
			})

			for _, w := range tc.want {
				s.Contains(got, w)
			}
			for _, a := range tc.absent {
				s.NotContains(got, a)
			}
		})
	}
}

func TestRenderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RenderPublicTestSuite))
}
