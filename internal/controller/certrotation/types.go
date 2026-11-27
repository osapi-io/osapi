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

// Package certrotation rotates the NATS server's self-signed certificate
// with no restart and no new join of the agents. Each agent pins that
// certificate, so the controller makes the next one beside it, and writes
// a record to the state bucket that its key signs, with both certificates.
// Each agent adds the next one to the certificates that it trusts at its
// next heartbeat, and its registration reports it. When each admitted agent
// that reported within the grace period has it, or the grace period ends,
// the controller renames the next files into place. The NATS server serves
// the new certificate at its next handshake, and the controller writes the
// record of the switch.
package certrotation

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

// ErrOperatorCert is a rotation of a certificate that the operator's
// cert_file names: osapi rotates only the self-signed one that it made.
var ErrOperatorCert = errors.New("the certificate is the operator's cert_file: replace that file")

// KV is the part of a NATS client that the rotation uses: the record in the
// state bucket. A *natsclient.Client is one.
type KV interface {
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVCreate(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVUpdate(
		ctx context.Context,
		bucket string,
		key string,
		value []byte,
		revision uint64,
	) (uint64, error)
}

// Admitted lists the admission records: the agents that the switch waits
// for are the admitted ones.
type Admitted interface {
	List(ctx context.Context) ([]admission.Record, error)
}

// Seen tells which admitted agents still report, with the next NATS
// certificate that each reports, and the last report of each machine. The
// job client is one.
type Seen interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
	AgentsLastSeen(ctx context.Context) (map[string]time.Time, error)
}

// Config is what a Rotation uses.
type Config struct {
	// KV and Bucket are the state bucket, which holds the record.
	KV     KV
	Bucket string
	// Keys are the controller's keys, which sign the record and check it.
	Keys *pki.Manager
	// Source is the NATS server's certificate.
	Source pki.CertSource
	// Admitted and Seen are the admitted agents and when each last
	// reported. A Rotation with none of them, such as the one of the CLI,
	// can start a rotation and show it, and does not switch.
	Admitted Admitted
	Seen     Seen
	// Grace is how long a rotation waits for the agents before it switches.
	Grace time.Duration
	// Logger logs the start and the switch.
	Logger *slog.Logger
	// Now is the clock. Nil is time.Now.
	Now func() time.Time
}

// Status is a rotation as the API shows it.
type Status struct {
	// Current is the fingerprint of the certificate that the server serves.
	Current string
	// Next is the fingerprint of the certificate that it switches to,
	// during a rotation.
	Next string
	// StartedAt is when the rotation started, and SwitchBy when it switches
	// at the latest.
	StartedAt time.Time
	SwitchBy  time.Time
	// Agents are the admitted agents that the switch waits for, and whether
	// each trusts the next certificate.
	Agents []Agent
}

// Agent is one admitted agent of a rotation.
type Agent struct {
	MachineID string
	Hostname  string
	HasNext   bool
}
