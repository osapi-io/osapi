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

package natsvault_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault/mocks"
)

type NATSVaultPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	ctx  context.Context
}

func (s *NATSVaultPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.ctx = vault.WithActor(context.Background(), "ops@example.com")
}

func (s *NATSVaultPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
}

// mockVault is a vault default on a mock KV, with a key in a file.
func (s *NATSVaultPublicTestSuite) mockVault() (*mocks.MockKV, *natsvault.Vault, avfs.VFS) {
	kv := mocks.NewMockKV(s.ctrl)
	kv.EXPECT().KVStatus(gomock.Any(), testBucket).Return(nil, nil)
	fs, keys := keyFile(s.T())
	v, err := natsvault.Open(s.ctx, kv, testBucket, "default", keys)
	s.Require().NoError(err)

	return kv, v, fs
}

// realVault is a vault default on a NATS server, with a key in a file.
func (s *NATSVaultPublicTestSuite) realVault() (*natsvault.Vault, avfs.VFS) {
	nc := startNATS(s.T())
	fs, keys := keyFile(s.T())
	v, err := natsvault.Open(s.ctx, nc, testBucket, "default", keys)
	s.Require().NoError(err)

	return v, fs
}

// entry is a KV entry with value at revision.
func (s *NATSVaultPublicTestSuite) entry(
	value []byte,
	revision uint64,
) *mocks.MockKeyValueEntry {
	e := mocks.NewMockKeyValueEntry(s.ctrl)
	e.EXPECT().Value().Return(value).AnyTimes()
	e.EXPECT().Revision().Return(revision).AnyTimes()

	return e
}

// sealed is the record that v writes for value under key.
func (s *NATSVaultPublicTestSuite) sealed(
	kv *mocks.MockKV,
	v *natsvault.Vault,
	key string,
	value string,
) []byte {
	var data []byte
	kv.EXPECT().
		KVGet(gomock.Any(), testBucket, "default."+key).
		Return(nil, jetstream.ErrKeyNotFound)
	kv.EXPECT().KVCreate(gomock.Any(), testBucket, "default."+key, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, b []byte) (uint64, error) {
			data = b

			return 1, nil
		})
	_, err := v.Put(s.ctx, key, []byte(value))
	s.Require().NoError(err)

	return data
}

func (s *NATSVaultPublicTestSuite) TestType() {
	tests := []struct {
		name    string
		spec    vault.Spec
		wantErr string
	}{
		{
			name: "when the vault names its key file opens it",
			spec: vault.Spec{Name: "default", Type: "nats", KeyFile: "/etc/osapi/secrets.key"},
		},
		{
			name:    "when the vault names no key refuses it",
			spec:    vault.Spec{Name: "default", Type: "nats"},
			wantErr: `vault "default": a nats vault needs its key: set key_file or key_env`,
		},
		{
			name: "when the vault names a file and a variable refuses it",
			spec: vault.Spec{
				Name: "default", Type: "nats", KeyFile: "/k", KeyEnv: "OSAPI_SECRETS_KEY",
			},
			wantErr: `vault "default": set key_file or key_env, not both`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs, _ := keyFile(s.T())
			set, err := vault.Open(s.ctx, vault.Env{
				NATS:   startNATS(s.T()),
				Bucket: testBucket,
				FS:     fs,
			}, []vault.Spec{tc.spec})
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			v, ok := set.Get("default")
			s.True(ok)
			s.Equal("default", v.Name())
			s.Equal(natsvault.TypeName, natsvault.Type.Name)
			s.Equal("Config", natsvault.Type.Config.Name())
		})
	}
}

func (s *NATSVaultPublicTestSuite) TestOpen() {
	tests := []struct {
		name    string
		bucket  string
		status  error
		keys    func() natsvault.Keys
		wantErr string
	}{
		{
			name:   "when the bucket and the key are there opens the vault",
			bucket: testBucket,
		},
		{
			name:    "when no bucket is named refuses",
			wantErr: "nats.secrets.bucket is empty: a nats vault needs its bucket",
		},
		{
			name:    "when the bucket is not there tells to set it up",
			bucket:  testBucket,
			status:  jetstream.ErrBucketNotFound,
			wantErr: "open the bucket test-secrets: nats: bucket not found: run osapi nats setup",
		},
		{
			name:   "when the key is not there tells to make it",
			bucket: testBucket,
			keys: func() natsvault.Keys {
				return natsvault.Keys{FS: memfs.New(), File: "/etc/osapi/secrets.key"}
			},
			wantErr: "make it with osapi pki key create --name vault:<name>",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			kv := mocks.NewMockKV(s.ctrl)
			if tc.bucket != "" {
				kv.EXPECT().KVStatus(gomock.Any(), tc.bucket).Return(nil, tc.status)
			}
			_, keys := keyFile(s.T())
			if tc.keys != nil {
				keys = tc.keys()
			}

			v, err := natsvault.Open(s.ctx, kv, tc.bucket, "default", keys)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(v)

				return
			}
			s.Require().NoError(err)
			s.Equal("default", v.Name())
		})
	}
}

func (s *NATSVaultPublicTestSuite) TestPut() {
	s.Run("when the key is new and then changes, each Put is a new version", func() {
		v, _ := s.realVault()

		first, err := v.Put(s.ctx, "registry-token", []byte("one"))
		s.Require().NoError(err)
		s.Equal("registry-token", first.Key)
		s.Equal(uint64(1), first.Version)
		s.Equal("ops@example.com", first.By)
		s.False(first.Changed.IsZero())

		second, err := v.Put(context.Background(), "registry-token", []byte("two"))
		s.Require().NoError(err)
		s.Equal(uint64(2), second.Version)
		s.Empty(second.By, "a Put with no actor records nobody")

		value, meta, err := v.Get(s.ctx, "registry-token")
		s.Require().NoError(err)
		s.Equal("two", string(value))
		s.Equal(second, meta)
	})

	s.Run("when the same value is put two times, the stored values differ", func() {
		nc := startNATS(s.T())
		_, keys := keyFile(s.T())
		v, err := natsvault.Open(s.ctx, nc, testBucket, "default", keys)
		s.Require().NoError(err)

		_, err = v.Put(s.ctx, "a", []byte("same"))
		s.Require().NoError(err)
		_, err = v.Put(s.ctx, "b", []byte("same"))
		s.Require().NoError(err)
		a, err := nc.KVGet(s.ctx, testBucket, "default.a")
		s.Require().NoError(err)
		b, err := nc.KVGet(s.ctx, testBucket, "default.b")
		s.Require().NoError(err)
		s.NotEqual(a.Value(), b.Value(), "each value has its own nonce")
		s.NotContains(string(a.Value()), "same", "the bucket has no plain value")
	})

	s.Run("when the key is not a valid name refuses it", func() {
		_, v, _ := s.mockVault()

		_, err := v.Put(s.ctx, "a.b", []byte("x"))

		s.ErrorContains(err, `key "a.b": a key is 1 to 128 letters, digits, - or _`)
	})

	s.Run("when the key file is gone returns the error", func() {
		_, v, fs := s.mockVault()
		s.Require().NoError(fs.Remove("/etc/osapi/secrets.key"))

		_, err := v.Put(s.ctx, "token", []byte("x"))

		s.ErrorContains(err, "read the vault key: ")
	})

	s.Run("when the bucket does not answer returns the error", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(nil, errors.New("timeout"))

		_, err := v.Put(s.ctx, "token", []byte("x"))

		s.EqualError(err, "get default/token: timeout")
	})

	s.Run("when the stored value is not a record returns the error", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry([]byte("{"), 3), nil)

		_, err := v.Put(s.ctx, "token", []byte("x"))

		s.ErrorContains(err, "get default/token: ")
	})

	s.Run("when a different Put changed the key at the same time fails", func() {
		kv, v, _ := s.mockVault()
		data := s.sealed(kv, v, "token", "old")
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry(data, 4), nil)
		kv.EXPECT().KVUpdate(gomock.Any(), testBucket, "default.token", gomock.Any(), uint64(4)).
			Return(uint64(0), jetstream.ErrKeyExists)

		_, err := v.Put(s.ctx, "token", []byte("new"))

		s.ErrorIs(err, jetstream.ErrKeyExists)
	})
}

func (s *NATSVaultPublicTestSuite) TestGet() {
	s.Run("when the key is there is its value and metadata", func() {
		v, _ := s.realVault()
		_, err := v.Put(s.ctx, "token", []byte("s3cret"))
		s.Require().NoError(err)

		value, meta, err := v.Get(s.ctx, "token")

		s.Require().NoError(err)
		s.Equal("s3cret", string(value))
		s.Equal(uint64(1), meta.Version)
	})

	s.Run("when the key is not there is ErrNotFound", func() {
		v, _ := s.realVault()

		_, _, err := v.Get(s.ctx, "token")

		s.ErrorIs(err, vault.ErrNotFound)
		s.EqualError(err, "default/token: secret not found")
	})

	s.Run("when the key file is gone returns the error", func() {
		kv, v, fs := s.mockVault()
		data := s.sealed(kv, v, "token", "x")
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry(data, 1), nil)
		s.Require().NoError(fs.Remove("/etc/osapi/secrets.key"))

		_, _, err := v.Get(s.ctx, "token")

		s.ErrorContains(err, "read the vault key: ")
	})

	s.Run("when the value is under a key that the file does not have fails", func() {
		kv, v, fs := s.mockVault()
		data := s.sealed(kv, v, "token", "x")
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry(data, 1), nil)
		s.Require().NoError(fs.Remove("/etc/osapi/secrets.key"))
		_, err := natsvault.CreateKey(s.ctx, fs, "/etc/osapi/secrets.key")
		s.Require().NoError(err)

		_, _, err = v.Get(s.ctx, "token")

		s.ErrorContains(err, "which the vault key does not have")
	})

	s.Run("when the value was copied from a different key does not decrypt", func() {
		kv, v, _ := s.mockVault()
		data := s.sealed(kv, v, "token", "x")
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.other").
			Return(s.entry(data, 1), nil)

		_, _, err := v.Get(s.ctx, "other")

		s.ErrorContains(err, "get default/other: the value does not decrypt: ")
	})

	s.Run("when the nonce has the wrong size does not decrypt", func() {
		kv, v, _ := s.mockVault()
		data := s.sealed(kv, v, "token", "x")
		var rec map[string]any
		s.Require().NoError(json.Unmarshal(data, &rec))
		rec["nonce"] = "AAAA"
		bad, _ := json.Marshal(rec)
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry(bad, 1), nil)

		_, _, err := v.Get(s.ctx, "token")

		s.ErrorContains(err, "its nonce has the wrong size")
	})
}

func (s *NATSVaultPublicTestSuite) TestDelete() {
	s.Run("when the key is there deletes it, and a new Put starts again", func() {
		v, _ := s.realVault()
		_, err := v.Put(s.ctx, "token", []byte("x"))
		s.Require().NoError(err)

		s.Require().NoError(v.Delete(s.ctx, "token"))

		_, _, err = v.Get(s.ctx, "token")
		s.ErrorIs(err, vault.ErrNotFound)
		s.ErrorIs(v.Delete(s.ctx, "token"), vault.ErrNotFound)
		meta, err := v.Put(s.ctx, "token", []byte("y"))
		s.Require().NoError(err)
		s.Equal(uint64(1), meta.Version)
	})

	s.Run("when the bucket does not delete returns the error", func() {
		kv, v, _ := s.mockVault()
		data := s.sealed(kv, v, "token", "x")
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(s.entry(data, 1), nil)
		kv.EXPECT().KVDelete(gomock.Any(), testBucket, "default.token").
			Return(errors.New("timeout"))

		s.EqualError(v.Delete(s.ctx, "token"), "delete default/token: timeout")
	})
}

func (s *NATSVaultPublicTestSuite) TestList() {
	s.Run("when the vault has keys is each one, sorted, and no key of a different vault", func() {
		nc := startNATS(s.T())
		fs, keys := keyFile(s.T())
		v, err := natsvault.Open(s.ctx, nc, testBucket, "default", keys)
		s.Require().NoError(err)
		_, err = natsvault.CreateKey(s.ctx, fs, "/etc/osapi/prod.key")
		s.Require().NoError(err)
		other, err := natsvault.Open(s.ctx, nc, testBucket, "prod", natsvault.Keys{
			FS: fs, File: "/etc/osapi/prod.key",
		})
		s.Require().NoError(err)

		empty, err := v.List(s.ctx)
		s.Require().NoError(err)
		s.Empty(empty)

		for _, k := range []string{"zeta", "alpha"} {
			_, err = v.Put(s.ctx, k, []byte("x"))
			s.Require().NoError(err)
		}
		_, err = other.Put(s.ctx, "beta", []byte("x"))
		s.Require().NoError(err)

		got, err := v.List(s.ctx)

		s.Require().NoError(err)
		s.Require().Len(got, 2)
		s.Equal("alpha", got[0].Key)
		s.Equal("zeta", got[1].Key)
	})

	s.Run("when the bucket does not list returns the error", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
			Return(nil, errors.New("timeout"))

		_, err := v.List(s.ctx)

		s.EqualError(err, "list default: timeout")
	})

	s.Run("when a key goes after the list leaves it out", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
			Return([]string{"default.token"}, nil)
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(nil, jetstream.ErrKeyNotFound)

		got, err := v.List(s.ctx)

		s.Require().NoError(err)
		s.Empty(got)
	})

	s.Run("when a key does not read returns the error", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
			Return([]string{"default.token"}, nil)
		kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
			Return(nil, errors.New("timeout"))

		_, err := v.List(s.ctx)

		s.EqualError(err, "get default/token: timeout")
	})
}

func (s *NATSVaultPublicTestSuite) TestRotate() {
	s.Run("when values are under the old key encrypts them again, one time", func() {
		v, fs := s.realVault()
		for _, k := range []string{"a", "b"} {
			_, err := v.Put(s.ctx, k, []byte("value-"+k))
			s.Require().NoError(err)
		}
		before, _, err := v.Get(s.ctx, "a")
		s.Require().NoError(err)
		data, err := fs.ReadFile("/etc/osapi/secrets.key")
		s.Require().NoError(err)
		old, err := natsvault.ParseKeyring(data)
		s.Require().NoError(err)
		ring := natsvault.NewKeyring(old)
		s.Require().NoError(fs.WriteFile("/etc/osapi/secrets.key", ring.Marshal(), 0o600))

		n, err := v.Rotate(s.ctx, ring)
		s.Require().NoError(err)
		s.Equal(2, n)
		again, err := v.Rotate(s.ctx, ring)
		s.Require().NoError(err)
		s.Equal(0, again)

		current := ring.Current()
		s.Require().NoError(fs.WriteFile("/etc/osapi/secrets.key", current.Marshal(), 0o600))
		after, meta, err := v.Get(s.ctx, "a")
		s.Require().NoError(err)
		s.Equal(before, after)
		s.Equal(uint64(1), meta.Version, "a rotation keeps the version")
	})

	s.Run("when the list fails returns the error", func() {
		kv, v, _ := s.mockVault()
		kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
			Return(nil, errors.New("timeout"))

		_, err := v.Rotate(s.ctx, natsvault.NewKeyring(natsvault.Keyring{}))

		s.EqualError(err, "list default: timeout")
	})

	tests := []struct {
		name    string
		setup   func(kv *mocks.MockKV, data []byte)
		ring    func(fs avfs.VFS) natsvault.Keyring
		want    int
		wantErr string
	}{
		{
			name: "when a Put changes the key during the rotation reads it again",
			setup: func(kv *mocks.MockKV, data []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(s.entry(data, 1), nil).Times(2)
				gomock.InOrder(
					kv.EXPECT().
						KVUpdate(gomock.Any(), testBucket, "default.token", gomock.Any(), uint64(1)).
						Return(uint64(0), jetstream.ErrKeyExists),
					kv.EXPECT().
						KVUpdate(gomock.Any(), testBucket, "default.token", gomock.Any(), uint64(1)).
						Return(uint64(2), nil),
				)
			},
			want: 1,
		},
		{
			name: "when the key changes three times stops",
			setup: func(kv *mocks.MockKV, data []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(s.entry(data, 1), nil).Times(3)
				kv.EXPECT().
					KVUpdate(gomock.Any(), testBucket, "default.token", gomock.Any(), uint64(1)).
					Return(uint64(0), jetstream.ErrKeyExists).
					Times(3)
			},
			wantErr: "rotate default/token: the key changed three times during the rotation",
		},
		{
			name: "when the write fails returns the error",
			setup: func(kv *mocks.MockKV, data []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(s.entry(data, 1), nil)
				kv.EXPECT().
					KVUpdate(gomock.Any(), testBucket, "default.token", gomock.Any(), uint64(1)).
					Return(uint64(0), errors.New("timeout"))
			},
			wantErr: "put default/token: timeout",
		},
		{
			name: "when the key goes during the rotation leaves it",
			setup: func(kv *mocks.MockKV, _ []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(nil, jetstream.ErrKeyNotFound)
			},
		},
		{
			name: "when the key does not read returns the error",
			setup: func(kv *mocks.MockKV, _ []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(nil, errors.New("timeout"))
			},
			wantErr: "get default/token: timeout",
		},
		{
			name: "when the value is under a key that the ring does not have fails",
			setup: func(kv *mocks.MockKV, data []byte) {
				kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
					Return(s.entry(data, 1), nil)
			},
			ring: func(avfs.VFS) natsvault.Keyring {
				return natsvault.NewKeyring(natsvault.Keyring{})
			},
			wantErr: "rotate default/token: the value is under key ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			kv, v, fs := s.mockVault()
			data := s.sealed(kv, v, "token", "x")
			// The list, then each read of the key.
			kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
				Return([]string{"default.token"}, nil)
			kv.EXPECT().KVGet(gomock.Any(), testBucket, "default.token").
				Return(s.entry(data, 1), nil)
			tc.setup(kv, data)
			ring := func() natsvault.Keyring {
				raw, err := fs.ReadFile("/etc/osapi/secrets.key")
				s.Require().NoError(err)
				old, err := natsvault.ParseKeyring(raw)
				s.Require().NoError(err)

				return natsvault.NewKeyring(old)
			}()
			if tc.ring != nil {
				ring = tc.ring(fs)
			}

			n, err := v.Rotate(s.ctx, ring)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, n)
		})
	}
}

func TestNATSVaultPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NATSVaultPublicTestSuite))
}
