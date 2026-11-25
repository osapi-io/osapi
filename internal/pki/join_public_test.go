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

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type JoinPublicTestSuite struct {
	suite.Suite
}

func (s *JoinPublicTestSuite) TestSignVerify() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	r := pki.JoinRequest{
		Token:         "osapi_jt_a.b",
		MachineID:     "m1",
		Hostname:      "web-01",
		PublicKey:     pub,
		EncryptionKey: "X1",
	}
	s.False(r.Verify(), "unsigned")

	r.Sign(priv)
	s.True(r.Verify())

	changed := r
	changed.Hostname = "web-02"
	s.False(changed.Verify(), "a change to a field breaks the signature")

	other, _, _ := ed25519.GenerateKey(rand.Reader)
	swapped := r
	swapped.PublicKey = other
	s.False(swapped.Verify(), "another key does not verify it")

	short := r
	short.PublicKey = pub[:10]
	s.False(short.Verify(), "a key of the wrong size")

	withUser := r
	withUser.NATSUserKey = "UA1"
	s.False(withUser.Verify(), "the user key is signed too")
	withUser.Sign(priv)
	s.True(withUser.Verify())
	withUser.NATSUserKey = "UA2"
	s.False(withUser.Verify(), "another user key breaks the signature")
}

func TestJoinPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JoinPublicTestSuite))
}
