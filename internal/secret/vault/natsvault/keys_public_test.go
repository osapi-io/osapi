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
	"strings"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

type KeysPublicTestSuite struct {
	suite.Suite
}

func (s *KeysPublicTestSuite) TestLoad() {
	key := strings.Repeat("ab", 32)
	env := func(vars map[string]string) func(string) (string, bool) {
		return func(k string) (string, bool) {
			v, ok := vars[k]

			return v, ok
		}
	}
	tests := []struct {
		name    string
		keys    func() natsvault.Keys
		wantErr string
	}{
		{
			name: "when the file has a key reads it",
			keys: func() natsvault.Keys {
				fs := memfs.New()
				s.Require().NoError(fs.WriteFile("/k", []byte(key), 0o600))

				return natsvault.Keys{FS: fs, File: "/k"}
			},
		},
		{
			name: "when the file is not there tells to make it",
			keys: func() natsvault.Keys {
				return natsvault.Keys{FS: memfs.New(), File: "/k"}
			},
			wantErr: "read the vault key: open /k: ",
		},
		{
			name: "when the file is not a key names it",
			keys: func() natsvault.Keys {
				fs := memfs.New()
				s.Require().NoError(fs.WriteFile("/k", []byte("x"), 0o600))

				return natsvault.Keys{FS: fs, File: "/k"}
			},
			wantErr: "the vault key /k: line 1: a vault key is 64 hex digits",
		},
		{
			name: "when the variable has a key reads it",
			keys: func() natsvault.Keys {
				return natsvault.Keys{
					Env:       "OSAPI_SECRETS_KEY",
					LookupEnv: env(map[string]string{"OSAPI_SECRETS_KEY": key}),
				}
			},
		},
		{
			name: "when the variable is not set names it",
			keys: func() natsvault.Keys {
				return natsvault.Keys{Env: "OSAPI_SECRETS_KEY", LookupEnv: env(nil)}
			},
			wantErr: "the vault key: OSAPI_SECRETS_KEY is not set",
		},
		{
			name: "when the variable is not a key names it",
			keys: func() natsvault.Keys {
				return natsvault.Keys{
					Env:       "OSAPI_SECRETS_KEY",
					LookupEnv: env(map[string]string{"OSAPI_SECRETS_KEY": "x"}),
				}
			},
			wantErr: "the vault key in OSAPI_SECRETS_KEY: line 1: a vault key is 64 hex digits",
		},
		{
			name:    "when there is no file and no variable refuses",
			keys:    func() natsvault.Keys { return natsvault.Keys{} },
			wantErr: "the vault has no key: set key_file or key_env",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ring, err := tc.keys().Load()
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(1, ring.Len())
		})
	}
}

func TestKeysPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeysPublicTestSuite))
}
