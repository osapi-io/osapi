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
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"strconv"
	"strings"
)

// NATSCertStateKey is the key of the CertRecord of the NATS server's
// certificate in the state bucket, which each agent reads at its heartbeat.
const NATSCertStateKey = "pki.cert.nats"

// certRecordDomain separates the signed bytes of a CertRecord from each
// other message that the controller's key signs.
const certRecordDomain = "osapi-cert-v1"

// CertRecord is the state of the rotation of a server's certificate that the
// agents trust: the NATS server's self-signed one, which each agent pins.
// The controller's key signs it, so an agent takes only a record that a
// controller key it trusts signed. A record with a Next is a rotation in
// progress: each agent adds Next to the certificates that it trusts, and
// its registration reports it. After the switch, Current is the new
// certificate, which the server serves, and it has no Next.
type CertRecord struct {
	// Serial grows by one with each record that the controller writes.
	Serial uint64 `json:"serial"`
	// Name is the certificate's, as list shows it, such as nats.
	Name string `json:"name"`
	// Current is the PEM certificate that the server serves.
	Current []byte `json:"current"`
	// Next is the PEM certificate that the server switches to, during a
	// rotation.
	Next []byte `json:"next,omitempty"`
	// StartedAt is when the rotation in progress started.
	StartedAt string `json:"started_at,omitempty"`
	// Signature is the controller key's Ed25519 signature of the record.
	Signature []byte `json:"signature"`
}

// signed is the bytes that Signature signs: each field but the signature,
// in a fixed order, after a domain string.
func (r CertRecord) signed() []byte {
	var b strings.Builder
	b.WriteString(certRecordDomain)
	for _, f := range [][2]string{
		{"serial", strconv.FormatUint(r.Serial, 10)},
		{"name", r.Name},
		{"current", CertFingerprint(r.Current)},
		{"next", CertFingerprint(r.Next)},
		{"started_at", r.StartedAt},
	} {
		b.WriteString("\n" + f[0] + "=" + f[1])
	}

	return []byte(b.String())
}

// Sign signs the record with priv, the controller's private key.
func (r *CertRecord) Sign(
	priv ed25519.PrivateKey,
) {
	r.Signature = ed25519.Sign(priv, r.signed())
}

// SignedWith reports whether verify, such as the VerifyWithGrace of an
// agent's keys, accepts the signature.
func (r CertRecord) SignedWith(
	verify func(data []byte, signature []byte) bool,
) bool {
	return len(r.Signature) == ed25519.SignatureSize && verify(r.signed(), r.Signature)
}

// Rotating reports whether a rotation is in progress: the record has a next
// certificate that is not its current one.
func (r CertRecord) Rotating() bool {
	return len(r.Next) > 0 && !bytes.Equal(r.Next, r.Current)
}

// CertFingerprint is the SHA-256 of the first certificate in pemData, as
// SHA256:<hex>, or empty when it has none.
func CertFingerprint(
	pemData []byte,
) string {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		return ""
	}
	sum := sha256.Sum256(block.Bytes)

	return "SHA256:" + hex.EncodeToString(sum[:])
}
