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

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

type KeyringPublicTestSuite struct {
	suite.Suite
}

func (s *KeyringPublicTestSuite) TestParseKeyring() {
	key := strings.Repeat("ab", 32)
	other := strings.Repeat("cd", 32)
	tests := []struct {
		name    string
		data    string
		wantLen int
		wantErr string
	}{
		{
			name:    "when the file has one key reads it",
			data:    key + "\n",
			wantLen: 1,
		},
		{
			name:    "when the file has comments and blank lines reads the keys",
			data:    "# the vault key\n\n" + key + "\n  " + other + "  \n",
			wantLen: 2,
		},
		{
			name:    "when a line is not hex refuses it",
			data:    key + "\nnot-hex\n",
			wantErr: "line 2: a vault key is 64 hex digits",
		},
		{
			name:    "when a line is a reference refuses it",
			data:    "secret://default/vault-key\n",
			wantErr: "line 1 is a reference: the vault key cannot come from a vault, because the controller needs the key to read the vault: put the key itself here",
		},
		{
			name:    "when a key is short refuses it",
			data:    "abcd",
			wantErr: "line 1: a vault key is 64 hex digits",
		},
		{
			name:    "when the file has no key refuses it",
			data:    "# nothing\n",
			wantErr: "has no key",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ring, err := natsvault.ParseKeyring([]byte(tc.data))
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantLen, ring.Len())
			again, err := natsvault.ParseKeyring(ring.Marshal())
			s.Require().NoError(err)
			s.Equal(ring.ID(), again.ID())
		})
	}
}

func (s *KeyringPublicTestSuite) TestNewKeyring() {
	old := natsvault.NewKeyring(natsvault.Keyring{})
	s.Equal(1, old.Len())
	s.Len(old.ID(), 16)

	ring := natsvault.NewKeyring(old)

	s.Equal(2, ring.Len(), "the old key stays for reads")
	s.NotEqual(old.ID(), ring.ID(), "the new key is the current key")
	s.Equal(1, ring.Current().Len())
	s.Equal(ring.ID(), ring.Current().ID())
}

func TestKeyringPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeyringPublicTestSuite))
}
