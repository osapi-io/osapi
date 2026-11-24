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

// testConfig is the config block of the type that the tests register.
type testConfig struct {
	Address string `mapstructure:"address" validate:"required" doc:"The address of the store."`
	Port    int    `mapstructure:"port"                        doc:"The port of the store."`
}

type VaultPublicTestSuite struct {
	suite.Suite
}

// registerTypes registers two types for a test, and removes them when
// the test ends: test, whose config block is a testConfig that it puts in
// built, and plain, with no config and no check.
func registerTypes(
	t *testing.T,
	ctrl *gomock.Controller,
	built *testConfig,
) {
	vault.Register(vault.Define(
		"test", "A store for the tests",
		func(spec vault.Spec) error {
			if spec.KeyFile == "bad" {
				return errors.New("bad key")
			}

			return nil
		},
		func(_ context.Context, _ vault.Env, spec vault.Spec, cfg testConfig) (vault.Vault, error) {
			*built = cfg
			if cfg.Address == "fail" {
				return nil, errors.New("no answer")
			}
			v := mocks.NewMockVault(ctrl)
			v.EXPECT().Name().Return(spec.Name).AnyTimes()

			return v, nil
		},
	))
	vault.Register(vault.Define[struct{}]("plain", "A store with no check", nil,
		func(context.Context, vault.Env, vault.Spec, struct{}) (vault.Vault, error) {
			return mocks.NewMockVault(ctrl), nil
		},
	))
	t.Cleanup(func() {
		vault.Unregister("test")
		vault.Unregister("plain")
	})
}

func (s *VaultPublicTestSuite) SetupTest() {
	registerTypes(s.T(), gomock.NewController(s.T()), &testConfig{})
}

func (s *VaultPublicTestSuite) TestRegister() {
	s.Panics(func() {
		vault.Register(vault.Define[struct{}]("test", "", nil, nil))
	}, "two types with one name panic")
}

func (s *VaultPublicTestSuite) TestLookup() {
	t, ok := vault.Lookup("test")
	s.True(ok)
	s.Equal("A store for the tests", t.Doc)
	s.Equal("testConfig", t.Config.Name())

	_, ok = vault.Lookup("nope")
	s.False(ok)
}

func (s *VaultPublicTestSuite) TestTypes() {
	types := vault.Types()
	names := make([]string, 0, len(types))
	for _, t := range types {
		names = append(names, t.Name)
	}

	s.Equal([]string{"plain", "test"}, names, "sorted by name")
}

func (s *VaultPublicTestSuite) TestTypeOf() {
	t, err := vault.TypeOf(vault.Spec{Name: "default", Type: "test"})
	s.Require().NoError(err)
	s.Equal("test", t.Name)

	_, err = vault.TypeOf(vault.Spec{Name: "default", Type: "nope"})
	s.EqualError(
		err,
		`vault "default": type "nope" is not a type of vault: the types are plain, test`,
	)
}

func (s *VaultPublicTestSuite) TestDecode() {
	t, _ := vault.Lookup("test")
	tests := []struct {
		name    string
		config  map[string]any
		want    testConfig
		wantErr string
	}{
		{
			name:   "when the block has the keys of the struct decodes it",
			config: map[string]any{"address": "https://vault:8200", "port": "8200"},
			want:   testConfig{Address: "https://vault:8200", Port: 8200},
		},
		{
			name: "when there is no block is the zero struct",
		},
		{
			name:    "when the block has a key that the struct does not have refuses it",
			config:  map[string]any{"adress": "x", "zone": "y"},
			wantErr: `vault "default": config: the type test has no setting adress, zone`,
		},
		{
			name:    "when a value does not decode to its field refuses it",
			config:  map[string]any{"port": "high"},
			wantErr: `vault "default": config: `,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := t.Decode(vault.Spec{Name: "default", Type: "test", Config: tc.config})
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(&tc.want, got)
		})
	}
}

func (s *VaultPublicTestSuite) TestCheck() {
	t, _ := vault.Lookup("test")
	s.NoError(t.Check(vault.Spec{Name: "default"}))
	s.EqualError(t.Check(vault.Spec{Name: "default", KeyFile: "bad"}), `vault "default": bad key`)

	plain, _ := vault.Lookup("plain")
	s.NoError(plain.Check(vault.Spec{Name: "default", KeyFile: "bad"}), "a type with no check")
}

func TestVaultPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(VaultPublicTestSuite))
}
