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
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type CertRecordPublicTestSuite struct {
	suite.Suite
}

func certPEM(
	der string,
) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte(der)})
}

func (s *CertRecordPublicTestSuite) TestSignedWith() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	other, _, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	verify := func(key ed25519.PublicKey) func([]byte, []byte) bool {
		return func(data, sig []byte) bool { return ed25519.Verify(key, data, sig) }
	}

	rec := pki.CertRecord{
		Serial: 1, Name: "nats", Current: certPEM("a"), Next: certPEM("b"),
		StartedAt: "2026-11-26T00:00:00Z",
	}
	s.False(rec.SignedWith(verify(pub)), "an unsigned record is not signed")
	rec.Sign(priv)

	tests := []struct {
		name   string
		mutate func(*pki.CertRecord)
		key    ed25519.PublicKey
		want   bool
	}{
		{name: "when its signer checks it, it is signed", key: pub, want: true},
		{name: "when another key checks it, it is not", key: other},
		{
			name:   "when its next certificate changed, it is not",
			mutate: func(r *pki.CertRecord) { r.Next = certPEM("c") },
			key:    pub,
		},
		{
			name:   "when its serial changed, it is not",
			mutate: func(r *pki.CertRecord) { r.Serial = 2 },
			key:    pub,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := rec
			if tc.mutate != nil {
				tc.mutate(&r)
			}
			s.Equal(tc.want, r.SignedWith(verify(tc.key)))
		})
	}
}

func (s *CertRecordPublicTestSuite) TestRotating() {
	s.False(pki.CertRecord{Current: certPEM("a")}.Rotating(), "no next")
	s.False(pki.CertRecord{Current: certPEM("a"), Next: certPEM("a")}.Rotating(), "next is current")
	s.True(pki.CertRecord{Current: certPEM("a"), Next: certPEM("b")}.Rotating())
}

func (s *CertRecordPublicTestSuite) TestCertFingerprint() {
	s.Equal(
		"SHA256:ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
		pki.CertFingerprint(certPEM("a")),
	)
	s.Empty(pki.CertFingerprint(nil), "no PEM")
	s.Empty(pki.CertFingerprint(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY"})), "not a cert")
}

func TestCertRecordPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CertRecordPublicTestSuite))
}
