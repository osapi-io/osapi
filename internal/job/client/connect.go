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
	"errors"
	"fmt"
	"log/slog"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

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
	// JobsBucket is the KV bucket of the jobs.
	JobsBucket string
	// RegistryBucket is the KV bucket where each component registers.
	RegistryBucket string
	// FactsBucket is the KV bucket of the agent facts, or "" if cfg names
	// none.
	FactsBucket string
	// StateBucket is the KV bucket of the agent state, or "" if cfg names
	// none.
	StateBucket string
}

// Connect connects to NATS as conn tells, opens the buckets that cfg names,
// and makes a job client on them. A job waits for its answers for at most
// timeout. resolve gives the value that a target has in a subject. The
// metrics of each call to NATS go to mp, the meter provider of the
// component, which can be nil.
func Connect(
	ctx context.Context,
	log *slog.Logger,
	conn config.NATSConnection,
	cfg config.NATS,
	timeout time.Duration,
	resolve func(ctx context.Context, target string) string,
	mp *sdkmetric.MeterProvider,
) (*Connection, error) {
	namespace := conn.Namespace

	nc, err := bus.Dial(log, bus.Telemetry(&natsclient.Options{
		Host: conn.Host,
		Port: conn.Port,
		Auth: bus.AuthOptions(conn.Auth),
		Name: conn.ClientName,
	}, mp))
	if err != nil {
		return nil, err
	}

	c := &Connection{
		NATS:           nc,
		JobsBucket:     bus.JobKV(namespace, cfg.KV).Bucket,
		RegistryBucket: bus.RegistryKV(namespace, cfg.Registry).Bucket,
	}
	buckets := []jetstream.KeyValueConfig{
		bus.JobKV(namespace, cfg.KV),
		bus.RegistryKV(namespace, cfg.Registry),
	}
	if cfg.Facts.Bucket != "" {
		facts := bus.FactsKV(namespace, cfg.Facts)
		c.FactsBucket = facts.Bucket
		buckets = append(buckets, facts)
	}
	if cfg.State.Bucket != "" {
		state := bus.StateKV(namespace, cfg.State)
		c.StateBucket = state.Bucket
		buckets = append(buckets, state)
	}
	for _, b := range buckets {
		if err := ensureBucket(ctx, nc, b); err != nil {
			bus.Close(nc)

			return nil, err
		}
	}

	c.Jobs, err = newFn(log, nc, &Options{
		Timeout:        timeout,
		JobsBucket:     c.JobsBucket,
		RegistryBucket: c.RegistryBucket,
		FactsBucket:    c.FactsBucket,
		StateBucket:    c.StateBucket,
		StreamName:     job.ApplyNamespaceToInfraName(namespace, cfg.Stream.Name),
		TargetResolver: resolve,
	})
	if err != nil {
		bus.Close(nc)

		return nil, fmt.Errorf("create job client: %w", err)
	}

	return c, nil
}

// ensureBucket binds to the bucket b names, and creates it with b only when
// it is missing. It never changes a bucket that exists: the agent and the
// controller each connect with their own config, and an update from each
// flipped the job queue's age between theirs, which removed every job older
// than the shorter one at once. osapi nats setup creates and updates the
// buckets from the config of the server.
func ensureBucket(
	ctx context.Context,
	nc *natsclient.Client,
	b jetstream.KeyValueConfig,
) error {
	_, err := nc.KVStatus(ctx, b.Bucket)
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, jetstream.ErrBucketNotFound):
		return fmt.Errorf("open KV bucket %s: %w", b.Bucket, err)
	}
	if err := nc.KVCreateOrUpdate(ctx, b); err != nil {
		return fmt.Errorf("create KV bucket %s: %w", b.Bucket, err)
	}

	return nil
}
