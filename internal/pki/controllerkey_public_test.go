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

package pki_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type ControllerKeyPublicTestSuite struct {
	suite.Suite
}

func (s *ControllerKeyPublicTestSuite) TestRecord() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	nextPub, _, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	started := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	signed := func() pki.ControllerKeyRecord {
		r := pki.ControllerKeyRecord{Serial: 3, Current: pub, Next: nextPub, StartedAt: started}
		r.Sign(priv)

		return r
	}

	tests := []struct {
		name        string
		record      func() pki.ControllerKeyRecord
		key         ed25519.PublicKey
		wantSigned  bool
		wantRotates bool
	}{
		{
			name:        "the key that signed it verifies it, and it rotates to its next key",
			record:      signed,
			key:         pub,
			wantSigned:  true,
			wantRotates: true,
		},
		{
			name:        "another key does not verify it",
			record:      signed,
			key:         otherPub,
			wantRotates: true,
		},
		{
			name:        "a key of the wrong size does not verify it",
			record:      signed,
			key:         ed25519.PublicKey{1, 2, 3},
			wantRotates: true,
		},
		{
			name: "a field changed after the signature does not verify",
			record: func() pki.ControllerKeyRecord {
				r := signed()
				r.Serial++

				return r
			},
			key:         pub,
			wantRotates: true,
		},
		{
			name: "a time changed after the signature does not verify",
			record: func() pki.ControllerKeyRecord {
				r := signed()
				r.SwitchedAt = started

				return r
			},
			key:         pub,
			wantRotates: true,
		},
		{
			name: "a record with retired keys verifies with them",
			record: func() pki.ControllerKeyRecord {
				r := pki.ControllerKeyRecord{Serial: 4, Current: pub, Retired: [][]byte{otherPub}}
				r.Sign(priv)

				return r
			},
			key:        pub,
			wantSigned: true,
		},
		{
			name: "a retired key added after the signature does not verify",
			record: func() pki.ControllerKeyRecord {
				r := signed()
				r.Retired = [][]byte{otherPub}

				return r
			},
			key:         pub,
			wantRotates: true,
		},
		{
			name: "no signature does not verify",
			record: func() pki.ControllerKeyRecord {
				r := signed()
				r.Signature = nil

				return r
			},
			key:         pub,
			wantRotates: true,
		},
		{
			name: "a record with no next key does not rotate",
			record: func() pki.ControllerKeyRecord {
				r := pki.ControllerKeyRecord{Serial: 1, Current: pub}
				r.Sign(priv)

				return r
			},
			key:        pub,
			wantSigned: true,
		},
		{
			name: "a record whose next key is its current one does not rotate",
			record: func() pki.ControllerKeyRecord {
				r := pki.ControllerKeyRecord{Serial: 1, Current: pub, Next: pub}
				r.Sign(priv)

				return r
			},
			key:        pub,
			wantSigned: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := tc.record()
			s.Equal(tc.wantSigned, r.SignedBy(tc.key))
			s.Equal(tc.wantRotates, r.Rotating())
		})
	}
}

func (s *ControllerKeyPublicTestSuite) TestNextAndPreviousKeys() {
	keyOf := func() (ed25519.PublicKey, ed25519.PrivateKey) {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(s.T(), err)

		return pub, priv
	}
	data := []byte("job")

	tests := []struct {
		name         string
		setup        func(m *pki.Manager) []byte
		wantVerified bool
		validateFunc func(m *pki.Manager)
	}{
		{
			name: "the next key verifies before the switch",
			setup: func(m *pki.Manager) []byte {
				current, _ := keyOf()
				next, nextPriv := keyOf()
				m.SetControllerPublicKey(current)
				m.SetNextControllerPublicKey(next)

				return ed25519.Sign(nextPriv, data)
			},
			wantVerified: true,
			validateFunc: func(m *pki.Manager) {
				s.NotNil(m.NextControllerPublicKey())
			},
		},
		{
			name: "a rotation clears the next key and keeps the old one as the previous",
			setup: func(m *pki.Manager) []byte {
				current, currentPriv := keyOf()
				next, _ := keyOf()
				m.SetControllerPublicKey(current)
				m.SetNextControllerPublicKey(next)
				m.RotateControllerKey(next)

				return ed25519.Sign(currentPriv, data)
			},
			wantVerified: true,
			validateFunc: func(m *pki.Manager) {
				s.Nil(m.NextControllerPublicKey())
				s.NotNil(m.PreviousControllerPublicKey())
			},
		},
		{
			name: "a previous key set at the start verifies",
			setup: func(m *pki.Manager) []byte {
				current, _ := keyOf()
				previous, previousPriv := keyOf()
				m.SetControllerPublicKey(current)
				m.SetPreviousControllerPublicKey(previous)

				return ed25519.Sign(previousPriv, data)
			},
			wantVerified: true,
			validateFunc: func(m *pki.Manager) {
				s.NotNil(m.PreviousControllerPublicKey())
			},
		},
		{
			name: "a retired key verifies no message, only a stored record",
			setup: func(m *pki.Manager) []byte {
				current, _ := keyOf()
				retired, retiredPriv := keyOf()
				m.SetControllerPublicKey(current)
				m.SetRetiredControllerPublicKeys([]ed25519.PublicKey{retired})

				return ed25519.Sign(retiredPriv, data)
			},
			validateFunc: func(m *pki.Manager) {
				s.Len(m.StoredRecordKeys(), 2)
			},
		},
		{
			name: "a cleared next key does not verify",
			setup: func(m *pki.Manager) []byte {
				current, _ := keyOf()
				next, nextPriv := keyOf()
				m.SetControllerPublicKey(current)
				m.SetNextControllerPublicKey(next)
				m.SetNextControllerPublicKey(nil)

				return ed25519.Sign(nextPriv, data)
			},
			validateFunc: func(m *pki.Manager) {
				s.Nil(m.NextControllerPublicKey())
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := pki.New(memfs.New(), "/keys", "agent")
			sig := tc.setup(m)
			s.Equal(tc.wantVerified, m.VerifyWithGrace(data, sig))
			tc.validateFunc(m)
		})
	}
}

// TestStoredRecordKeys reads the keys that verify a record the controller
// stored: the current, the next, the previous and each retired key, in that
// order, and no key that is not set.
func (s *ControllerKeyPublicTestSuite) TestStoredRecordKeys() {
	keyOf := func() ed25519.PublicKey {
		pub, _, err := ed25519.GenerateKey(rand.Reader)
		s.Require().NoError(err)

		return pub
	}
	current, next, previous, retired := keyOf(), keyOf(), keyOf(), keyOf()

	tests := []struct {
		name     string
		setup    func(m *pki.Manager)
		wantKeys []ed25519.PublicKey
	}{
		{
			name: "each key that is set, in order",
			setup: func(m *pki.Manager) {
				m.SetControllerPublicKey(current)
				m.SetNextControllerPublicKey(next)
				m.SetPreviousControllerPublicKey(previous)
				m.SetRetiredControllerPublicKeys([]ed25519.PublicKey{retired})
			},
			wantKeys: []ed25519.PublicKey{current, next, previous, retired},
		},
		{
			name: "a key that is not set, or of the wrong size, is left out",
			setup: func(m *pki.Manager) {
				m.SetControllerPublicKey(current)
				m.SetRetiredControllerPublicKeys([]ed25519.PublicKey{{1, 2, 3}, retired})
			},
			wantKeys: []ed25519.PublicKey{current, retired},
		},
		{
			name:     "no key",
			setup:    func(*pki.Manager) {},
			wantKeys: []ed25519.PublicKey{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := pki.New(memfs.New(), "/keys", "controller")
			tc.setup(m)
			s.Equal(tc.wantKeys, m.StoredRecordKeys())
		})
	}
}

func TestControllerKeyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerKeyPublicTestSuite))
}
