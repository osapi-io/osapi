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
	"errors"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault/mocks"
)

type KeyPublicTestSuite struct {
	suite.Suite
}

// failing is fs with op failing from its nth call on, and each other
// operation as fs does it.
func failing(
	fs avfs.VFS,
	op avfs.FnVFS,
	nth int,
) avfs.VFS {
	vfs := failfs.New(fs)
	calls := 0
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, fn avfs.FnVFS, _ *failfs.FailParam) error {
		if fn != op {
			return nil
		}
		calls++
		if calls >= nth {
			return errors.New("injected")
		}

		return nil
	})

	return vfs
}

func (s *KeyPublicTestSuite) TestCreateKey() {
	tests := []struct {
		name    string
		fs      func() avfs.VFS
		file    string
		wantErr string
	}{
		{
			name: "when there is no key writes one that only its owner reads",
			fs:   func() avfs.VFS { return memfs.New() },
			file: "/etc/osapi/secrets.key",
		},
		{
			name:    "when the vault has no key_file refuses",
			fs:      func() avfs.VFS { return memfs.New() },
			wantErr: "the vault has no key_file: set it to the file to write",
		},
		{
			name: "when the key is there does not replace it",
			fs: func() avfs.VFS {
				fs := memfs.New()
				_, err := natsvault.CreateKey(context.Background(), fs, "/etc/osapi/secrets.key")
				s.Require().NoError(err)

				return fs
			},
			file:    "/etc/osapi/secrets.key",
			wantErr: "/etc/osapi/secrets.key exists: rotate the key with osapi controller secret key rotate",
		},
		{
			name:    "when the directory cannot be made returns the error",
			fs:      func() avfs.VFS { return failing(memfs.New(), avfs.FnMkdirAll, 1) },
			file:    "/etc/osapi/secrets.key",
			wantErr: "make the directory of /etc/osapi/secrets.key: ",
		},
		{
			name:    "when the file cannot be written returns the error",
			fs:      func() avfs.VFS { return failing(memfs.New(), avfs.FnRename, 1) },
			file:    "/etc/osapi/secrets.key",
			wantErr: "write /etc/osapi/secrets.key: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := tc.fs()

			ring, err := natsvault.CreateKey(context.Background(), fs, tc.file)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			info, err := fs.Stat(tc.file)
			s.Require().NoError(err)
			s.Equal("-rw-------", info.Mode().Perm().String())
			loaded, err := natsvault.Keys{FS: fs, File: tc.file}.Load()
			s.Require().NoError(err)
			s.Equal(ring.ID(), loaded.ID())
		})
	}
}

func (s *KeyPublicTestSuite) TestRotateKey() {
	ctx := context.Background()
	const file = "/etc/osapi/secrets.key"

	s.Run("when the vault has values rotates the key and each value", func() {
		nc := startNATS(s.T())
		fs := memfs.New()
		old, err := natsvault.CreateKey(ctx, fs, file)
		s.Require().NoError(err)
		v, err := natsvault.Open(ctx, nc, testBucket, "default", natsvault.Keys{FS: fs, File: file})
		s.Require().NoError(err)
		for _, k := range []string{"a", "b", "c"} {
			_, err = v.Put(ctx, k, []byte("value-"+k))
			s.Require().NoError(err)
		}

		ring, n, err := natsvault.RotateKey(ctx, v)

		s.Require().NoError(err)
		s.Equal(3, n)
		s.NotEqual(old.ID(), ring.ID())
		loaded, err := natsvault.Keys{FS: fs, File: file}.Load()
		s.Require().NoError(err)
		s.Equal(1, loaded.Len(), "the old key goes")
		s.Equal(ring.ID(), loaded.ID())
		for _, k := range []string{"a", "b", "c"} {
			value, _, err := v.Get(ctx, k)
			s.Require().NoError(err)
			s.Equal("value-"+k, string(value))
		}
	})

	tests := []struct {
		name    string
		keys    func(fs avfs.VFS) natsvault.Keys
		fs      func(fs avfs.VFS) avfs.VFS
		after   func(fs avfs.VFS)
		setup   func(kv *mocks.MockKV)
		wantErr string
		check   func(fs avfs.VFS)
	}{
		{
			name: "when the key is in a variable refuses",
			keys: func(avfs.VFS) natsvault.Keys {
				return natsvault.Keys{
					Env: "OSAPI_SECRETS_KEY",
					LookupEnv: func(string) (string, bool) {
						return strings.Repeat("ab", 32), true
					},
				}
			},
			wantErr: "the key of default is in OSAPI_SECRETS_KEY, not a file: rotate writes key_file",
		},
		{
			name:    "when the key file is gone returns the error",
			after:   func(fs avfs.VFS) { s.Require().NoError(fs.Remove(file)) },
			wantErr: "read the vault key: ",
		},
		{
			name:    "when the new key cannot be written returns the error",
			fs:      func(fs avfs.VFS) avfs.VFS { return failing(fs, avfs.FnRename, 1) },
			wantErr: "write /etc/osapi/secrets.key: ",
		},
		{
			name: "when a value does not rotate keeps the old key in the file",
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").
					Return(nil, errors.New("timeout"))
			},
			wantErr: "list default: timeout: /etc/osapi/secrets.key keeps the old keys, so each value decrypts: run the rotation again",
			check: func(fs avfs.VFS) {
				loaded, err := natsvault.Keys{FS: fs, File: file}.Load()
				s.Require().NoError(err)
				s.Equal(2, loaded.Len())
			},
		},
		{
			name: "when the old key cannot be dropped returns the error",
			fs:   func(fs avfs.VFS) avfs.VFS { return failing(fs, avfs.FnRename, 2) },
			setup: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), testBucket, "default.>").Return(nil, nil)
			},
			wantErr: "write /etc/osapi/secrets.key: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			kv := mocks.NewMockKV(gomock.NewController(s.T()))
			kv.EXPECT().KVStatus(gomock.Any(), testBucket).Return(nil, nil)
			if tc.setup != nil {
				tc.setup(kv)
			}
			base := memfs.New()
			_, err := natsvault.CreateKey(ctx, base, file)
			s.Require().NoError(err)
			fs := avfs.VFS(base)
			if tc.fs != nil {
				fs = tc.fs(base)
			}
			keys := natsvault.Keys{FS: fs, File: file}
			if tc.keys != nil {
				keys = tc.keys(fs)
			}
			v, err := natsvault.Open(ctx, kv, testBucket, "default", keys)
			s.Require().NoError(err)
			if tc.after != nil {
				tc.after(fs)
			}

			_, _, err = natsvault.RotateKey(ctx, v)

			s.ErrorContains(err, tc.wantErr)
			if tc.check != nil {
				tc.check(fs)
			}
		})
	}
}

func TestKeyPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeyPublicTestSuite))
}
