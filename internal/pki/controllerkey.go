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
	"strings"
	"time"
)

// ControllerKeyStateKey is the key of the ControllerKeyRecord in the state
// bucket, which each agent reads at its heartbeat.
const ControllerKeyStateKey = "pki.controller"

// controllerKeyDomain separates the signed bytes of a ControllerKeyRecord
// from each other message that the controller's key signs.
const controllerKeyDomain = "osapi-controller-key-v1"

// ControllerKeyRecord is the state of the rotation of the controller's key.
// The key that is current when the controller writes it signs it: an agent
// takes only a record that a controller key it trusts signed. A record with
// a Next is a rotation in progress, which the current key signs. After the
// switch, Current is the new key, which signs the record, and Previous is
// the old one.
type ControllerKeyRecord struct {
	// Serial grows by one with each record that the controller writes.
	Serial uint64 `json:"serial"`
	// Current is the public key that the controller signs with.
	Current []byte `json:"current"`
	// Next is the public key that the controller switches to, during a
	// rotation.
	Next []byte `json:"next,omitempty"`
	// Previous is the public key that the last switch replaced.
	Previous []byte `json:"previous,omitempty"`
	// Retired is each key that a switch before the last one replaced, oldest
	// first. The controller still verifies the records that they signed,
	// such as the jobs in the queue.
	Retired [][]byte `json:"retired,omitempty"`
	// StartedAt is when the rotation in progress started.
	StartedAt time.Time `json:"started_at,omitzero"`
	// SwitchedAt is when the last switch was.
	SwitchedAt time.Time `json:"switched_at,omitzero"`
	// Signature is the Ed25519 signature of the record by its signer.
	Signature []byte `json:"signature"`
}

// signed is the bytes that Signature signs: each field but the signature,
// in a fixed order, after a domain string.
func (r ControllerKeyRecord) signed() []byte {
	var b strings.Builder
	b.WriteString(controllerKeyDomain)
	for _, f := range [][2]string{
		{"serial", strconv.FormatUint(r.Serial, 10)},
		{"current", base64.StdEncoding.EncodeToString(r.Current)},
		{"next", base64.StdEncoding.EncodeToString(r.Next)},
		{"previous", base64.StdEncoding.EncodeToString(r.Previous)},
		{"started_at", stamp(r.StartedAt)},
		{"switched_at", stamp(r.SwitchedAt)},
	} {
		b.WriteString("\n" + f[0] + "=" + f[1])
	}
	// A record with no retired key signs the same bytes as before the field.
	for _, k := range r.Retired {
		b.WriteString("\nretired=" + base64.StdEncoding.EncodeToString(k))
	}

	return []byte(b.String())
}

// stamp is t in the signed bytes: empty for the zero time.
func stamp(
	t time.Time,
) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format(time.RFC3339Nano)
}

// Sign signs the record with priv, the private key of its signer.
func (r *ControllerKeyRecord) Sign(
	priv ed25519.PrivateKey,
) {
	r.Signature = ed25519.Sign(priv, r.signed())
}

// SignedBy reports whether pub signed the record.
func (r ControllerKeyRecord) SignedBy(
	pub ed25519.PublicKey,
) bool {
	return len(pub) == ed25519.PublicKeySize &&
		len(r.Signature) == ed25519.SignatureSize &&
		ed25519.Verify(pub, r.signed(), r.Signature)
}

// Rotating reports whether a rotation is in progress: the record has a next
// key that is not its current one.
func (r ControllerKeyRecord) Rotating() bool {
	return len(r.Next) > 0 && !bytes.Equal(r.Next, r.Current)
}
