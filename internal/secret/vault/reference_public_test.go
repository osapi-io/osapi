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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/mocks"
)

type ReferencePublicTestSuite struct {
	suite.Suite
}

func (s *ReferencePublicTestSuite) TestParseReference() {
	tests := []struct {
		name    string
		in      string
		want    vault.Reference
		wantErr string
	}{
		{
			name: "when the reference names a vault and a key parses it",
			in:   "secret://default/jwt-signing-key",
			want: vault.Reference{Vault: "default", Key: "jwt-signing-key"},
		},
		{
			name:    "when the value is not a reference refuses it",
			in:      "a-literal",
			wantErr: `"a-literal" is not a reference: a reference is secret://<vault>/<key>`,
		},
		{
			name:    "when the reference has no key refuses it",
			in:      "secret://default",
			wantErr: "secret://default: a reference is secret://<vault>/<key>, with a vault name of 1 to 64 letters, digits, - or _",
		},
		{
			name:    "when the vault name is not valid refuses it",
			in:      "secret://de.fault/key",
			wantErr: "secret://de.fault/key: a reference is secret://<vault>/<key>, with a vault name of 1 to 64 letters, digits, - or _",
		},
		{
			name:    "when the key is not valid refuses it",
			in:      "secret://default/a/b",
			wantErr: `secret://default/a/b: key "a/b": a key is 1 to 128 letters, digits, - or _, such as registry-token`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := vault.ParseReference(tc.in)
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
			s.Equal(tc.in, got.String())
			s.True(vault.IsReference(tc.in))
		})
	}
}

func (s *ReferencePublicTestSuite) TestResolve() {
	tests := []struct {
		name    string
		ref     vault.Reference
		setup   func(v *mocks.MockVault)
		want    string
		wantErr string
	}{
		{
			name: "when the key exists gives its value",
			ref:  vault.Reference{Vault: "refs", Key: "jwt"},
			setup: func(v *mocks.MockVault) {
				v.EXPECT().Get(gomock.Any(), "jwt").Return([]byte("the-value"), vault.Meta{}, nil)
			},
			want: "the-value",
		},
		{
			name:    "when the vault does not exist refuses it",
			ref:     vault.Reference{Vault: "other", Key: "jwt"},
			setup:   func(*mocks.MockVault) {},
			wantErr: `secret://other/jwt: the vault "other" does not exist`,
		},
		{
			name: "when the key does not exist refuses it",
			ref:  vault.Reference{Vault: "refs", Key: "jwt"},
			setup: func(v *mocks.MockVault) {
				v.EXPECT().Get(gomock.Any(), "jwt").Return(nil, vault.Meta{}, vault.ErrNotFound)
			},
			wantErr: `secret://refs/jwt: the vault "refs" has no key "jwt"`,
		},
		{
			name: "when the vault fails gives the error with the reference",
			ref:  vault.Reference{Vault: "refs", Key: "jwt"},
			setup: func(v *mocks.MockVault) {
				v.EXPECT().
					Get(gomock.Any(), "jwt").
					Return(nil, vault.Meta{}, errors.New("no answer"))
			},
			wantErr: "secret://refs/jwt: no answer",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			v := mocks.NewMockVault(gomock.NewController(s.T()))
			tc.setup(v)
			vault.Register(vault.Define[struct{}]("refs", "A store of the tests of Resolve", nil,
				func(context.Context, vault.Env, vault.Spec, struct{}) (vault.Vault, error) {
					return v, nil
				},
			))
			defer vault.Unregister("refs")
			set, err := vault.Open(context.Background(), vault.Env{},
				[]vault.Spec{{Name: "refs", Type: "refs"}})
			s.Require().NoError(err)

			got, err := set.Resolve(context.Background(), tc.ref)
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, string(got))
		})
	}
}

func TestReferencePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ReferencePublicTestSuite))
}
