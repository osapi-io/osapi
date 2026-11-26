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

package pki

import (
	"bytes"
	"crypto/ed25519"
	"strconv"
)

// RenewRequestSuffix is the NATS subject suffix of the renewal of an
// agent's NATS user in operator mode: a request that the controller replies
// to with a JoinReply that has a new JWT.
const RenewRequestSuffix = "join.renew"

// renewDomain starts the bytes that a renewal signs, so that its signature
// is not a signature of anything else.
const renewDomain = "osapi-renew-v1"

// RenewRequest is what an admitted agent sends to renew the JWT of its NATS
// user: who it is, its user key, and when, signed with its signing key.
type RenewRequest struct {
	MachineID   string
	NATSUserKey string
	At          int64
	Signature   []byte
}

// signed is the bytes that the signature covers.
func (r RenewRequest) signed() []byte {
	var b bytes.Buffer
	for _, f := range []string{renewDomain, r.MachineID, r.NATSUserKey, strconv.FormatInt(r.At, 10)} {
		b.WriteString(f)
		b.WriteByte(0)
	}

	return b.Bytes()
}

// Sign signs r with priv, the agent's signing key.
func (r *RenewRequest) Sign(
	priv ed25519.PrivateKey,
) {
	r.Signature = ed25519.Sign(priv, r.signed())
}

// Verify reports whether pub, the key that admission recorded for the
// machine, signed r.
func (r RenewRequest) Verify(
	pub ed25519.PublicKey,
) bool {
	return len(pub) == ed25519.PublicKeySize && ed25519.Verify(pub, r.signed(), r.Signature)
}
