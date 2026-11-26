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
	"encoding/base64"
	"strconv"
)

// RotateRequestSuffix is the NATS subject suffix of the rotation of an
// agent's keys: a request that the controller replies to with a
// RotateReply.
const RotateRequestSuffix = "join.rotate"

// rotateDomain starts the bytes that a rotation signs, so that its
// signatures are not signatures of anything else.
const rotateDomain = "osapi-rotate-v1"

// RotateRequest is what an admitted agent sends to replace its two keys: who
// it is, its new public keys, and when. Its current signing key signs it,
// which the controller checks against the admission, and its new signing key
// signs it too, which shows that the agent holds the new key.
type RotateRequest struct {
	MachineID string `json:"machine_id"`
	// PublicKey is the new Ed25519 signing key.
	PublicKey ed25519.PublicKey `json:"public_key"`
	// EncryptionKey is the new P-256 key, in EncryptionKeyEncoding.
	EncryptionKey string `json:"encryption_key"`
	At            int64  `json:"at"`
	// Signature is the signature of the current key.
	Signature []byte `json:"signature"`
	// NewSignature is the signature of the new key.
	NewSignature []byte `json:"new_signature"`
}

// RotateReply is the controller's answer to a RotateRequest: the
// fingerprint that the admission holds now, or why the controller refused.
type RotateReply struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	Error       string `json:"error,omitempty"`
}

// signed is the bytes that both signatures cover.
func (r RotateRequest) signed() []byte {
	var b bytes.Buffer
	for _, f := range []string{
		rotateDomain,
		r.MachineID,
		base64.RawURLEncoding.EncodeToString(r.PublicKey),
		r.EncryptionKey,
		strconv.FormatInt(r.At, 10),
	} {
		b.WriteString(f)
		b.WriteByte(0)
	}

	return b.Bytes()
}

// Sign signs r with current, the agent's signing key, and with next, the
// new key that r carries.
func (r *RotateRequest) Sign(
	current ed25519.PrivateKey,
	next ed25519.PrivateKey,
) {
	r.Signature = ed25519.Sign(current, r.signed())
	r.NewSignature = ed25519.Sign(next, r.signed())
}

// Verify reports whether pub, the key that the admission holds, signed r,
// and whether the new key that r carries signed it too.
func (r RotateRequest) Verify(
	pub ed25519.PublicKey,
) bool {
	return len(pub) == ed25519.PublicKeySize &&
		ed25519.Verify(pub, r.signed(), r.Signature) &&
		r.VerifyNew()
}

// VerifyNew reports whether the new key that r carries signed it.
func (r RotateRequest) VerifyNew() bool {
	return len(r.PublicKey) == ed25519.PublicKeySize &&
		ed25519.Verify(r.PublicKey, r.signed(), r.NewSignature)
}
