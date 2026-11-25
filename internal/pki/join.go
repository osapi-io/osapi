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
)

// JoinRequest is what an agent sends to POST /api/agent/join: the join
// token, who it is, and its two public keys, signed with its signing key. The
// signature shows that the agent holds the key that it asks the controller
// to admit.
type JoinRequest struct {
	Token         string
	MachineID     string
	Hostname      string
	PublicKey     ed25519.PublicKey
	EncryptionKey string
	// NATSUserKey is the public key (U...) of the agent's own NATS user, for
	// a join in operator mode: the controller mints a JWT for it. The seed
	// stays with the agent.
	NATSUserKey string `json:",omitempty"`
	Signature   []byte
}

// joinDomain starts the bytes that a join request signs, so that a
// signature over a join cannot be used as a signature over anything else.
const joinDomain = "osapi-join-v1"

// signed is the bytes that the signature covers: each field but the
// signature, each ended by a zero byte, so that no two requests give the
// same bytes.
func (r JoinRequest) signed() []byte {
	var b bytes.Buffer
	for _, f := range [][]byte{
		[]byte(joinDomain),
		[]byte(r.Token),
		[]byte(r.MachineID),
		[]byte(r.Hostname),
		r.PublicKey,
		[]byte(r.EncryptionKey),
	} {
		b.Write(f)
		b.WriteByte(0)
	}
	// An older agent sends no user key, and signs the same bytes as before.
	if r.NATSUserKey != "" {
		b.WriteString(r.NATSUserKey)
		b.WriteByte(0)
	}

	return b.Bytes()
}

// Sign signs r with priv, the private key of r.PublicKey.
func (r *JoinRequest) Sign(
	priv ed25519.PrivateKey,
) {
	r.Signature = ed25519.Sign(priv, r.signed())
}

// Verify reports whether r's signature is by the private key of
// r.PublicKey.
func (r JoinRequest) Verify() bool {
	return len(r.PublicKey) == ed25519.PublicKeySize &&
		ed25519.Verify(r.PublicKey, r.signed(), r.Signature)
}

// JoinReply is the controller's reply to a join over NATS: its public key,
// and the NATS user that the agent connects with after the join. Error is
// why the join did not admit, and the other fields are empty then.
type JoinReply struct {
	ControllerPublicKey ed25519.PublicKey `json:"controller_public_key,omitempty"`
	NATS                JoinNATS          `json:"nats"`
	// Labels are the labels of the join token, which the agent takes.
	Labels map[string]string `json:"labels,omitempty"`
	Error  string            `json:"error,omitempty"`
}

// JoinNATS is the NATS user of an agent, which the join gives it.
type JoinNATS struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// JWT is, in operator mode, the user JWT of the agent's NATS user key.
	JWT string `json:"jwt,omitempty"`
}
