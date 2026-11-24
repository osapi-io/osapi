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

package vaultkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/cli/commands/vaultkey"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

const keyFile = "/etc/osapi/secrets.key"

// deps are the deps of a config with a nats vault default, a vault prod of
// a different type and a nats vault env whose key is in a variable, with
// kv as the NATS that Dial connects to.
func deps(
	fs avfs.VFS,
	kv natsvault.KV,
	dialErr error,
) vaultkey.Deps {
	var cfg config.Config
	cfg.Controller.NATS.Namespace = "osapi"
	cfg.NATS.Secrets.Bucket = "secrets"
	cfg.Controller.Secrets.Vaults = []vault.Spec{
		{Name: "default", Type: "nats", KeyFile: keyFile},
		{Name: "prod", Type: "hashicorp"},
		{Name: "env", Type: "nats", KeyEnv: "OSAPI_TEST_VAULT_KEY_UNSET"},
	}

	return vaultkey.Deps{
		FS:     fs,
		Config: cfg,
		Dial: func(context.Context) (natsvault.KV, func(), error) {
			if dialErr != nil {
				return nil, nil, dialErr
			}

			return kv, func() {}, nil
		},
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	tests := []struct {
		name    string
		in      vaultkey.Input
		setup   func(fs avfs.VFS)
		wantErr string
	}{
		{
			name: "when the vault has no key writes one",
			in:   vaultkey.Input{Name: "default"},
		},
		{
			name: "when the key is there refuses",
			in:   vaultkey.Input{Name: "default"},
			setup: func(fs avfs.VFS) {
				_, err := natsvault.CreateKey(context.Background(), fs, keyFile)
				s.Require().NoError(err)
			},
			wantErr: `vault "default": /etc/osapi/secrets.key exists: `,
		},
		{
			name:    "when the config has no such vault refuses",
			in:      vaultkey.Input{Name: "nope"},
			wantErr: `no vault "nope" in controller.secrets.vaults of the config`,
		},
		{
			name:    "when the vault is not of type nats refuses",
			in:      vaultkey.Input{Name: "prod"},
			wantErr: `vault "prod" is of type hashicorp: only a nats vault has a key of osapi`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			if tc.setup != nil {
				tc.setup(fs)
			}

			got, err := vaultkey.Create(context.Background(), deps(fs, nil, nil), tc.in)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal("default", got.Name)
			s.Equal(keyFile, got.KeyFile)
			ring, err := natsvault.Keys{FS: fs, File: keyFile}.Load()
			s.Require().NoError(err)
			s.Equal(ring.ID(), got.KeyID)
		})
	}
}

func (s *MethodsPublicTestSuite) TestRotate() {
	tests := []struct {
		name    string
		in      vaultkey.Input
		setup   func(kv *mocks.MockKV)
		dialErr error
		noKey   bool
		bucket  string
		wantErr string
	}{
		{
			name: "when the vault has a key rotates it",
			in:   vaultkey.Input{Name: "default"},
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVStatus(gomock.Any(), "osapi-secrets").Return(nil, nil)
				kv.EXPECT().KVKeys(gomock.Any(), "osapi-secrets", "default.>").Return(nil, nil)
			},
		},
		{
			name:    "when the config has no such vault refuses",
			in:      vaultkey.Input{Name: "nope"},
			wantErr: `no vault "nope"`,
		},
		{
			name:    "when NATS does not answer returns the error",
			in:      vaultkey.Input{Name: "default"},
			dialErr: errors.New("connect to NATS: no route"),
			wantErr: "connect to NATS: no route",
		},
		{
			name: "when the bucket is not there tells to set it up",
			in:   vaultkey.Input{Name: "default"},
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVStatus(gomock.Any(), "osapi-secrets").
					Return(nil, jetstream.ErrBucketNotFound)
			},
			wantErr: `vault "default": open the bucket osapi-secrets: `,
		},
		{
			name: "when the key of the vault is in a variable that is not set returns the error",
			in:   vaultkey.Input{Name: "env"},
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVStatus(gomock.Any(), "osapi-secrets").Return(nil, nil)
			},
			wantErr: `vault "env": the vault key: OSAPI_TEST_VAULT_KEY_UNSET is not set`,
		},
		{
			name:    "when no bucket is named refuses",
			in:      vaultkey.Input{Name: "default"},
			bucket:  "-",
			wantErr: `vault "default": nats.secrets.bucket is empty`,
		},
		{
			name: "when the rotation fails returns the error",
			in:   vaultkey.Input{Name: "default"},
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVStatus(gomock.Any(), "osapi-secrets").Return(nil, nil)
				kv.EXPECT().KVKeys(gomock.Any(), "osapi-secrets", "default.>").
					Return(nil, errors.New("timeout"))
			},
			wantErr: `vault "default": list default: timeout`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			kv := mocks.NewMockKV(gomock.NewController(s.T()))
			if tc.setup != nil {
				tc.setup(kv)
			}
			fs := memfs.New()
			old, err := natsvault.CreateKey(context.Background(), fs, keyFile)
			s.Require().NoError(err)
			d := deps(fs, kv, tc.dialErr)
			if tc.bucket == "-" {
				d.Config.NATS.Secrets.Bucket = ""
			}

			got, err := vaultkey.Rotate(context.Background(), d, tc.in)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal("default", got.Name)
			s.NotEqual(old.ID(), got.KeyID)
			s.Equal(0, got.Values)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
