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

package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// newFn makes the job client of Connect (injectable for testing).
var newFn = New

// Connection is a NATS connection with a job client and the buckets that the
// job client uses. The controller and the agent each make one.
type Connection struct {
	// NATS is the connection.
	NATS *natsclient.Client
	// Jobs is the job client.
	Jobs *Client
	// JobsKV is the bucket of the jobs.
	JobsKV jetstream.KeyValue
	// RegistryKV is the bucket where each component registers.
	RegistryKV jetstream.KeyValue
	// FactsKV is the bucket of the agent facts, or nil if cfg names none.
	FactsKV jetstream.KeyValue
	// StateKV is the bucket of the agent state, or nil if cfg names none.
	StateKV jetstream.KeyValue
}

// Connect connects to NATS as conn tells, opens the buckets that cfg names,
// and makes a job client on them. A job waits for its answers for at most
// timeout. resolve gives the value that a target has in a subject.
func Connect(
	ctx context.Context,
	log *slog.Logger,
	conn config.NATSConnection,
	cfg config.NATS,
	timeout time.Duration,
	resolve func(string) string,
) (*Connection, error) {
	namespace := conn.Namespace

	nc, err := bus.Dial(log, &natsclient.Options{
		Host: conn.Host,
		Port: conn.Port,
		Auth: bus.AuthOptions(conn.Auth),
		Name: conn.ClientName,
	})
	if err != nil {
		return nil, err
	}

	type bucket struct {
		kv   *jetstream.KeyValue
		name string
		cfg  jetstream.KeyValueConfig
	}

	c := &Connection{NATS: nc}
	buckets := []bucket{
		{&c.JobsKV, "jobs", bus.JobKV(namespace, cfg.KV)},
		{&c.RegistryKV, "registry", bus.RegistryKV(namespace, cfg.Registry)},
	}
	if cfg.Facts.Bucket != "" {
		buckets = append(buckets, bucket{&c.FactsKV, "facts", bus.FactsKV(namespace, cfg.Facts)})
	}
	if cfg.State.Bucket != "" {
		buckets = append(buckets, bucket{&c.StateKV, "state", bus.StateKV(namespace, cfg.State)})
	}
	for _, b := range buckets {
		*b.kv, err = nc.CreateOrUpdateKVBucketWithConfig(ctx, b.cfg)
		if err != nil {
			bus.Close(nc)

			return nil, fmt.Errorf("create %s KV bucket: %w", b.name, err)
		}
	}

	c.Jobs, err = newFn(log, nc, &Options{
		Timeout:        timeout,
		KVBucket:       c.JobsKV,
		RegistryKV:     c.RegistryKV,
		FactsKV:        c.FactsKV,
		StateKV:        c.StateKV,
		StreamName:     job.ApplyNamespaceToInfraName(namespace, cfg.Stream.Name),
		TargetResolver: resolve,
	})
	if err != nil {
		bus.Close(nc)

		return nil, fmt.Errorf("create job client: %w", err)
	}

	return c, nil
}
