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

type RenewPublicTestSuite struct {
	suite.Suite
}

func TestRenewPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RenewPublicTestSuite))
}

func (s *RenewPublicTestSuite) TestSignVerify() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	other, _, _ := ed25519.GenerateKey(rand.Reader)
	r := pki.RenewRequest{MachineID: "m-1", NATSUserKey: "UA1", At: 1700000000}
	s.False(r.Verify(pub), "unsigned")
	r.Sign(priv)
	s.True(r.Verify(pub))
	s.False(r.Verify(other), "another key")
	s.False(r.Verify(pub[:10]), "a key of the wrong size")
	changed := r
	changed.NATSUserKey = "UA2"
	s.False(changed.Verify(pub), "a change breaks the signature")
	later := r
	later.At++
	s.False(later.Verify(pub), "the time is signed")
}
