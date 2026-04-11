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
)

// Manager handles Ed25519 keypair lifecycle.
type Manager struct {
	fs               avfs.VFS
	keyDir           string
	publicKey        ed25519.PublicKey
	privateKey       ed25519.PrivateKey
	controllerPubKey         ed25519.PublicKey
	previousControllerPubKey ed25519.PublicKey
}

// EnrollmentState represents the agent's PKI enrollment state.
type EnrollmentState string

const (
	// StateUnregistered indicates the agent has not yet enrolled.
	StateUnregistered EnrollmentState = "unregistered"

	// StatePending indicates the agent has submitted an enrollment request.
	StatePending EnrollmentState = "pending"

	// StateAccepted indicates the controller has accepted the agent.
	StateAccepted EnrollmentState = "accepted"

	// StateRejected indicates the controller has rejected the agent.
	StateRejected EnrollmentState = "rejected"
)

// EnrollmentRequest is sent by the agent to the controller.
type EnrollmentRequest struct {
	MachineID   string `json:"machine_id"`
	Hostname    string `json:"hostname"`
	PublicKey   []byte `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
}

// EnrollmentResponse is sent by the controller to the agent.
type EnrollmentResponse struct {
	Accepted            bool   `json:"accepted"`
	ControllerPublicKey []byte `json:"controller_public_key,omitempty"`
	Reason              string `json:"reason,omitempty"`
}
