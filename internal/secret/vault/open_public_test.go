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

package vault_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/secret/vault"
)

type OpenPublicTestSuite struct {
	suite.Suite

	// built is the config that the last build of the type test got.
	built testConfig
}

func (s *OpenPublicTestSuite) SetupTest() {
	registerTypes(s.T(), gomock.NewController(s.T()), &s.built)
}

func (s *OpenPublicTestSuite) TestOpen() {
	tests := []struct {
		name    string
		specs   []vault.Spec
		wantErr string
	}{
		{
			name: "when each vault opens is each one, in the order of the config",
			specs: []vault.Spec{
				{
					Name: "prod", Type: "test", Targets: []string{"group:prod"},
					Config: map[string]any{"address": "a"},
				},
				{Name: "default", Type: "plain"},
			},
		},
		{
			name:    "when two vaults have one name refuses",
			specs:   []vault.Spec{{Name: "a", Type: "plain"}, {Name: "a", Type: "plain"}},
			wantErr: `vault "a": two vaults have the name`,
		},
		{
			name:    "when a type is not registered refuses",
			specs:   []vault.Spec{{Name: "a", Type: "nope"}},
			wantErr: `vault "a": type "nope" is not a type of vault`,
		},
		{
			name:    "when a config block does not decode refuses",
			specs:   []vault.Spec{{Name: "a", Type: "test", Config: map[string]any{"x": 1}}},
			wantErr: `vault "a": config: `,
		},
		{
			name:    "when the check of the type fails refuses",
			specs:   []vault.Spec{{Name: "a", Type: "test", KeyFile: "bad"}},
			wantErr: `vault "a": bad key`,
		},
		{
			name: "when a vault does not open refuses",
			specs: []vault.Spec{
				{Name: "a", Type: "test", Config: map[string]any{"address": "fail"}},
			},
			wantErr: `vault "a": no answer`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			set, err := vault.Open(context.Background(), vault.Env{}, tc.specs)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(set)

				return
			}
			s.Require().NoError(err)
			s.Equal([]string{"prod", "default"}, set.Names())
			v, ok := set.Get("prod")
			s.Require().True(ok)
			s.Equal("prod", v.Name())
			s.Equal("a", s.built.Address, "build gets the decoded block")
			spec, ok := set.Spec("prod")
			s.True(ok)
			s.Equal([]string{"group:prod"}, spec.Targets)
			_, ok = set.Get("nope")
			s.False(ok)
			_, ok = set.Spec("nope")
			s.False(ok)
		})
	}
}

func (s *OpenPublicTestSuite) TestCheckKey() {
	for _, key := range []string{"registry-token", "JWT_key", "a"} {
		s.NoError(vault.CheckKey(key), key)
	}
	for _, key := range []string{"", "a.b", "a/b", "a b", string(make([]byte, 129))} {
		s.Error(vault.CheckKey(key))
	}
}

func TestOpenPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OpenPublicTestSuite))
}
