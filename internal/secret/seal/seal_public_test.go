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

package seal_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

type SealPublicTestSuite struct {
	suite.Suite

	agent *ecdh.PrivateKey
	other *ecdh.PrivateKey
}

func (s *SealPublicTestSuite) SetupTest() {
	var err error
	s.agent, err = ecdh.P256().GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.other, err = ecdh.P256().GenerateKey(rand.Reader)
	s.Require().NoError(err)
}

func (s *SealPublicTestSuite) TearDownTest() {
	seal.ResetCreateKey()
}

// sealed is a value sealed for the job jobID, the job key sealed to
// s.agent, and the key that sealed it.
func (s *SealPublicTestSuite) sealed(
	jobID string,
	value secret.Value,
) (string, string, string) {
	sealer, err := seal.New(jobID)
	s.Require().NoError(err)
	key, err := sealer.KeyFor(pki.EncodeEncryptionKey(s.agent.PublicKey()))
	s.Require().NoError(err)

	return sealer.Value(value), key, sealer.From()
}

// gcmOf is AES-256-GCM with key, as the seal uses it: the 12-byte nonce in
// front of the ciphertext.
func gcmOf(
	key []byte,
) cipher.AEAD {
	block, _ := aes.NewCipher(key)
	aead, _ := cipher.NewGCMWithRandomNonce(block)

	return aead
}

// kek is the key-encryption key of the package documentation, for a key
// that from sealed to the agent.
func (s *SealPublicTestSuite) kek(
	from string,
) []byte {
	fromPub, err := pki.ParseEncryptionKey(from)
	s.Require().NoError(err)
	shared, err := s.agent.ECDH(fromPub)
	s.Require().NoError(err)
	k, err := hkdf.Key(sha256.New, shared, nil,
		"osapi-seal-v1"+string(fromPub.Bytes())+string(s.agent.PublicKey().Bytes()), 32)
	s.Require().NoError(err)

	return k
}

func (s *SealPublicTestSuite) TestRoundTrip() {
	value, key, from := s.sealed("job-1", "hunter2")

	s.True(strings.HasPrefix(value, seal.Prefix))
	s.NotContains(value, "hunter2")
	o, err := seal.Open("job-1", key, from, s.agent)
	s.Require().NoError(err)
	got, err := o.Value(secret.Value(value))
	s.Require().NoError(err)
	s.Equal(secret.Value("hunter2"), got)

	again, _, _ := s.sealed("job-1", "hunter2")
	s.NotEqual(value, again, "each seal has its own nonce")
}

// TestWireFormat opens a sealed key and a sealed value with the steps that
// the package documentation gives, and no code of the package.
func (s *SealPublicTestSuite) TestWireFormat() {
	value, key, from := s.sealed("job-1", "hunter2")

	raw, err := base64.RawURLEncoding.DecodeString(key)
	s.Require().NoError(err)
	s.Len(raw, 12+32+16, "the nonce, the job key and the tag")
	jobKey, err := gcmOf(s.kek(from)).Open(nil, nil, raw, []byte("job-1"))
	s.Require().NoError(err)
	s.Len(jobKey, 32)

	sealedValue, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, seal.Prefix))
	s.Require().NoError(err)
	plain, err := gcmOf(jobKey).Open(nil, nil, sealedValue, []byte("job-1"))
	s.Require().NoError(err)
	s.Equal("hunter2", string(plain))
}

func (s *SealPublicTestSuite) TestOpenRefuses() {
	value, key, from := s.sealed("job-1", "hunter2")
	otherValue, _, _ := s.sealed("job-2", "hunter2")
	tampered := []byte(key)
	tampered[len(tampered)-1] ^= 1
	tests := []struct {
		name    string
		open    func() (*seal.Opener, error)
		value   string
		wantErr string
	}{
		{
			name:    "when the job key is not base64, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", "!", from, s.agent) },
			wantErr: "open the job key",
		},
		{
			name:    "when the key that sealed it is not a P-256 key, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, "junk!", s.agent) },
			wantErr: "open the job key: encryption key",
		},
		{
			name:    "when the key of a different agent opens it, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.other) },
			wantErr: "it is not sealed to this agent for this job",
		},
		{
			name:    "when the job key was sealed for a different job, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-2", key, from, s.agent) },
			wantErr: "it is not sealed to this agent for this job",
		},
		{
			name:    "when the sealed job key was changed, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", string(tampered), from, s.agent) },
			wantErr: "open the job key",
		},
		{
			name:    "when the sealed job key is shorter than a nonce, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", "AAAA", from, s.agent) },
			wantErr: "it is not sealed to this agent for this job",
		},
		{
			name: "when the job key has the wrong size, it does not open",
			open: func() (*seal.Opener, error) {
				short := gcmOf(s.kek(from)).Seal(nil, nil, []byte("short"), []byte("job-1"))

				return seal.Open(
					"job-1",
					base64.RawURLEncoding.EncodeToString(short),
					from,
					s.agent,
				)
			},
			wantErr: "it is not sealed to this agent for this job",
		},
		{
			name:    "when the value is of a different job, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   otherValue,
			wantErr: "a sealed secret of the job does not open",
		},
		{
			name:    "when the value is not sealed, it is refused",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   "hunter2",
			wantErr: "a secret of the job is not sealed",
		},
		{
			name:    "when the sealed value is not base64, it is refused",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   seal.Prefix + "!",
			wantErr: "a sealed secret of the job is not valid",
		},
		{
			name:    "when the sealed value is shorter than a nonce, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   seal.Prefix + "AAAA",
			wantErr: "a sealed secret of the job does not open",
		},
		{
			name:    "when the sealed value was changed, it does not open",
			open:    func() (*seal.Opener, error) { return seal.Open("job-1", key, from, s.agent) },
			value:   value[:len(value)-2] + "AA",
			wantErr: "a sealed secret of the job does not open",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			o, err := tc.open()
			if tc.value == "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			got, err := o.Value(secret.Value(tc.value))
			s.EqualError(err, tc.wantErr)
			s.Empty(got)
		})
	}
}

func (s *SealPublicTestSuite) TestNewAndKeyForFail() {
	sealer, err := seal.New("job-1")
	s.Require().NoError(err)
	_, err = sealer.KeyFor("not-a-key")
	s.ErrorContains(err, "seal the job key")

	seal.SetCreateKey(func() (*ecdh.PrivateKey, error) { return nil, errors.New("no entropy") })
	_, err = seal.New("job-1")
	s.ErrorContains(err, "no entropy")
}

func (s *SealPublicTestSuite) TestOpenerContext() {
	s.Nil(seal.OpenerFrom(context.Background()))
	_, key, from := s.sealed("job-1", "x")
	o, err := seal.Open("job-1", key, from, s.agent)
	s.Require().NoError(err)
	s.Same(o, seal.OpenerFrom(seal.WithOpener(context.Background(), o)))
}

func TestSealPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SealPublicTestSuite))
}
