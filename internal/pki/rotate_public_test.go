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

type RotatePublicTestSuite struct {
	suite.Suite
}

func TestRotatePublicTestSuite(t *testing.T) {
	suite.Run(t, new(RotatePublicTestSuite))
}

func (s *RotatePublicTestSuite) TestVerify() {
	oldPub, oldPriv, _ := ed25519.GenerateKey(rand.Reader)
	newPub, newPriv, _ := ed25519.GenerateKey(rand.Reader)
	otherPub, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	signed := func(edit func(r *pki.RotateRequest)) pki.RotateRequest {
		r := pki.RotateRequest{
			MachineID: "m-1", PublicKey: newPub, EncryptionKey: "ekey", At: 1700000000,
		}
		r.Sign(oldPriv, newPriv)
		if edit != nil {
			edit(&r)
		}

		return r
	}

	tests := []struct {
		name    string
		req     pki.RotateRequest
		pub     ed25519.PublicKey
		want    bool
		wantNew bool
	}{
		{name: "both keys signed it", req: signed(nil), pub: oldPub, want: true, wantNew: true},
		{
			name: "a key that is not the admitted one signed it",
			req:  signed(nil), pub: otherPub, want: false, wantNew: true,
		},
		{name: "the admitted key is short", req: signed(nil), pub: oldPub[:5], wantNew: true},
		{
			name: "a field changed after the signatures",
			req:  signed(func(r *pki.RotateRequest) { r.MachineID = "m-2" }), pub: oldPub,
		},
		{
			name: "another key signed it in place of the new one",
			req: signed(func(r *pki.RotateRequest) {
				r.Sign(oldPriv, otherPriv)
			}),
			pub: oldPub,
		},
		{
			name: "the new key is short",
			req:  signed(func(r *pki.RotateRequest) { r.PublicKey = newPub[:5] }),
			pub:  oldPub,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, tt.req.Verify(tt.pub))
			s.Equal(tt.wantNew, tt.req.VerifyNew())
		})
	}
}
