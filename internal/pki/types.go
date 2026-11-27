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
	"crypto/ecdh"
	"crypto/ed25519"
	"sync"

	"github.com/avfs/avfs"
)

// Manager handles Ed25519 keypair lifecycle.
type Manager struct {
	// mu guards the two keys, which Reload replaces while the agent runs.
	// It is a pointer, so a copy of a Manager shares it.
	mu *sync.RWMutex

	fs         avfs.VFS
	keyDir     string
	keyPrefix  string // "agent" or "controller"
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	// controllerPubKey, previousControllerPubKey and nextControllerPubKey are
	// the keys of the controller that the manager accepts, which a rotation
	// of the controller's key changes while the process runs: mu guards them.
	controllerPubKey         ed25519.PublicKey
	previousControllerPubKey ed25519.PublicKey
	nextControllerPubKey     ed25519.PublicKey
	// retiredControllerPubKeys is each key that a switch before the last
	// one replaced: it verifies only the records that the controller
	// stored, such as its jobs, never a message.
	retiredControllerPubKeys []ed25519.PublicKey

	// withEncryption makes LoadOrGenerate also load or generate the P-256
	// ECDH key that opens values the controller seals for this agent.
	withEncryption      bool
	encryptionKey       *ecdh.PrivateKey
	encryptionPublicKey string
	// previousEncryptionKey is the encryption key that the last Reload
	// replaced: a value sealed to it just before a rotation still opens.
	previousEncryptionKey *ecdh.PrivateKey
}

// Option configures a Manager.
type Option func(*Manager)

// WithEncryptionKey gives the manager a P-256 ECDH encryption key next to
// its Ed25519 signing key, in {prefix}.ekey. The agent has one; the controller
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

	// JoinRequestSuffix is the NATS subject suffix of a join over NATS: a
	// request that the controller replies to with a JoinReply.
	JoinRequestSuffix = "join.request"
)

// EnrollmentRequest is sent by the agent to the controller. It carries both
// public keys: the Ed25519 signing key and the P-256 encryption key, in
// EncryptionKeyEncoding.
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
