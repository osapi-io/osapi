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

// Package enrollment provides the controller-side enrollment system for
// accepting or rejecting agent enrollment requests via NATS.
package enrollment

import (
	"crypto/ed25519"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// PendingAgent represents an agent awaiting enrollment acceptance.
type PendingAgent struct {
	MachineID   string    `json:"machine_id"`
	Hostname    string    `json:"hostname"`
	PublicKey   []byte    `json:"public_key"`
	Fingerprint string    `json:"fingerprint"`
	RequestedAt time.Time `json:"requested_at"`
}

// NATSSubscriber defines the NATS operations needed by the enrollment
// watcher for subscribing to enrollment requests and publishing responses.
// Satisfied by the nats-client's *Client type (Subscribe + PublishCore).
type NATSSubscriber interface {
	Subscribe(
		subj string,
		cb nats.MsgHandler,
	) (*nats.Subscription, error)
	PublishCore(
		subj string,
		data []byte,
	) error
}

// PKIProvider defines the PKI operations needed by the enrollment watcher
// to provide the controller's public key in acceptance responses.
type PKIProvider interface {
	PublicKey() ed25519.PublicKey
}

// KVStore wraps the jetstream.KeyValue interface for testability.
type KVStore = jetstream.KeyValue
