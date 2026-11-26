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

package pkilist_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/pkikey"
	"github.com/osapi-io/osapi/internal/cli/commands/pkilist"
	"github.com/osapi-io/osapi/internal/pki"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) TestList() {
	expires := time.Date(2031, 11, 25, 12, 0, 0, 0, time.UTC)
	d := pkilist.Deps{
		Keys: func() []pkikey.Key {
			return []pkikey.Key{
				{Name: "controller", State: "ok", Algorithm: "Ed25519", Fingerprint: "SHA256:aa"},
				{Name: "jetstream", State: "ok", Algorithm: "ChaCha20-Poly1305", KeyID: "7f3a"},
				{Name: "agent-encryption", State: "ok", Algorithm: "ECDH P-256", PublicKey: "BA"},
				{Name: "vault:default", State: "not made"},
			}
		},
		Certs: func() []pki.Cert {
			return []pki.Cert{
				{
					Name:        "api",
					State:       "ok",
					Key:         "ECDSA P-256",
					Fingerprint: "SHA256:bb",
					Expires:     expires,
				},
				{Name: "nats-metrics", State: "not made"},
			}
		},
	}

	got, err := pkilist.List(context.Background(), d, pkilist.ListInput{})

	s.Require().NoError(err)
	s.Require().Len(got.Items, 6)
	s.Equal(pkilist.KindKey, got.Items[0].Kind)
	s.Equal("SHA256:aa", got.Items[0].Fingerprint)
	s.Equal("7f3a", got.Items[1].Fingerprint, "a secret key shows its ID")
	s.Equal("BA", got.Items[2].Fingerprint, "an encryption key shows its public key")
	s.Empty(got.Items[3].Fingerprint)
	s.Equal(pkilist.KindCert, got.Items[4].Kind)
	s.Equal("ECDSA P-256", got.Items[4].Algorithm)
	s.Equal(expires, *got.Items[4].Expires)
	s.Nil(got.Items[5].Expires, "a certificate that is not made has no expiry")
}
