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

// Package pki provides Ed25519 keypair management for agent identity,
// including generation, loading, saving, signing, and verification.
package pki

import (
	"crypto/ed25519"

	"github.com/avfs/avfs"
	"github.com/nats-io/nkeys"
)

// Manager handles Ed25519 keypair lifecycle.
type Manager struct {
	fs                       avfs.VFS
	keyDir                   string
	keyPrefix                string // "agent" or "controller"
	publicKey                ed25519.PublicKey
	privateKey               ed25519.PrivateKey
	controllerPubKey         ed25519.PublicKey
	previousControllerPubKey ed25519.PublicKey

	// withEncryption makes LoadOrGenerate also load or generate the X25519
	// key that opens values the controller seals for this agent.
	withEncryption      bool
	encryptionKey       nkeys.KeyPair
	encryptionPublicKey string
}

// Option configures a Manager.
type Option func(*Manager)

// WithEncryptionKey gives the manager an X25519 encryption key next to its
// Ed25519 signing key, in {prefix}.xkey. The agent has one; the controller
// does not.
func WithEncryptionKey() Option {
	return func(m *Manager) {
		m.withEncryption = true
	}
}

const (
	// EnrollRequestSuffix is the NATS subject suffix for enrollment requests.
	EnrollRequestSuffix = "enroll.request"

	// EnrollResponsePrefix is the NATS subject prefix for enrollment responses.
	// Append ".<machineID>" to form the full subject.
	EnrollResponsePrefix = "enroll.response"

	// PKIRotateSuffix is the NATS subject suffix for key rotation notifications.
	PKIRotateSuffix = "pki.rotate"

	// JoinRequestSuffix is the NATS subject suffix of a join over NATS: a
	// request that the controller replies to with a JoinReply.
	JoinRequestSuffix = "join.request"
)

// EnrollmentRequest is sent by the agent to the controller. It carries both
// public keys: the Ed25519 signing key and the X25519 encryption key, in the
// nkeys form that starts with "X".
type EnrollmentRequest struct {
	MachineID     string `json:"machine_id"`
	Hostname      string `json:"hostname"`
	PublicKey     []byte `json:"public_key"`
	EncryptionKey string `json:"encryption_key"`
	Fingerprint   string `json:"fingerprint"`
}

// EnrollmentResponse is sent by the controller to the agent.
type EnrollmentResponse struct {
	Accepted            bool   `json:"accepted"`
	ControllerPublicKey []byte `json:"controller_public_key,omitempty"`
	Reason              string `json:"reason,omitempty"`
}
