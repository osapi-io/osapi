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

// Package keyrotation rotates the controller's key with no new admission of
// the agents. The controller makes the next key, and writes a record to the
// state bucket that its current key signs. Each agent keeps the next key at
// its next heartbeat, and its registration reports it. When each admitted
// agent that reported within the grace period has it, or the grace period
// ends, the controller switches to the next key, and writes a record that
// the new key signs. The agents then take the new key, and keep the old one
// as the previous one, which still verifies the jobs that it signed.
package keyrotation

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

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

// Seen tells which admitted agents still report: the agents in the registry
// now, with the controller keys that each reports, and the last report of
// each machine. The job client is one.
type Seen interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
	AgentsLastSeen(ctx context.Context) (map[string]time.Time, error)
}

// Config is what a Rotation uses.
type Config struct {
	// KV and Bucket are the state bucket, which holds the record.
	KV     KV
	Bucket string
	// Keys are the controller's keys, which Start and the switch read and
	// replace.
	Keys *pki.Manager
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
	// Current is the fingerprint of the key that the controller signs with.
	Current string
	// Next is the fingerprint of the key that it switches to, during a
	// rotation.
	Next string
	// StartedAt is when the rotation started, and SwitchBy when it switches
	// at the latest.
	StartedAt time.Time
	SwitchBy  time.Time
	// Agents are the admitted agents that the switch waits for, and whether
	// each has the next key.
	Agents []Agent
}

// Agent is one admitted agent of a rotation.
type Agent struct {
	MachineID string
	Hostname  string
	HasNext   bool
}
