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

package bus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// JetStream is the part of a NATS client that Setup calls.
type JetStream interface {
	CreateOrUpdateStreamWithConfig(
		ctx context.Context,
		streamConfig jetstream.StreamConfig,
	) error
	CreateOrUpdateKVBucketWithConfig(
		ctx context.Context,
		config jetstream.KeyValueConfig,
	) (jetstream.KeyValue, error)
	CreateOrUpdateObjectStore(
		ctx context.Context,
		cfg jetstream.ObjectStoreConfig,
	) (jetstream.ObjectStore, error)
}

// Prepare connects to NATS with opts, creates each stream and bucket with
// Setup, then closes the connection.
func Prepare(
	ctx context.Context,
	log *slog.Logger,
	opts *natsclient.Options,
	namespace string,
	cfg config.NATS,
) error {
	nc := natsclient.New(log, opts)
	if err := nc.Connect(); err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}
	defer Close(nc)

	return Setup(ctx, nc, namespace, cfg)
}

// Setup creates or updates each stream and bucket that osapi uses, with the
// names in namespace. A second run with the same config changes nothing. A
// bucket with no name in cfg is not made.
func Setup(
	ctx context.Context,
	js JetStream,
	namespace string,
	cfg config.NATS,
) error {
	streamName := job.ApplyNamespaceToInfraName(namespace, cfg.Stream.Name)
	streamMaxAge, _ := time.ParseDuration(cfg.Stream.MaxAge)

	discard := jetstream.DiscardOld
	if cfg.Stream.Discard == "new" {
		discard = jetstream.DiscardNew
	}

	if err := js.CreateOrUpdateStreamWithConfig(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{job.ApplyNamespaceToSubjects(namespace, cfg.Stream.Subjects)},
		MaxAge:   streamMaxAge,
		MaxMsgs:  cfg.Stream.MaxMsgs,
		Storage:  StorageType(cfg.Stream.Storage),
		Replicas: cfg.Stream.Replicas,
		Discard:  discard,
	}); err != nil {
		return fmt.Errorf("create JOBS stream: %w", err)
	}

	kvs := []jetstream.KeyValueConfig{
		JobKV(namespace, cfg.KV),
		ResponseKV(namespace, cfg.KV),
	}
	if cfg.Registry.Bucket != "" {
		kvs = append(kvs, RegistryKV(namespace, cfg.Registry))
	}
	if cfg.Facts.Bucket != "" {
		kvs = append(kvs, FactsKV(namespace, cfg.Facts))
	}
	if cfg.State.Bucket != "" {
		kvs = append(kvs, StateKV(namespace, cfg.State))
	}
	if cfg.FileState.Bucket != "" {
		kvs = append(kvs, FileStateKV(namespace, cfg.FileState))
	}
	for _, kv := range kvs {
		if _, err := js.CreateOrUpdateKVBucketWithConfig(ctx, kv); err != nil {
			return fmt.Errorf("create KV bucket %s: %w", kv.Bucket, err)
		}
	}

	if cfg.Audit.Stream != "" {
		audit := AuditStream(namespace, cfg.Audit)
		if err := js.CreateOrUpdateStreamWithConfig(ctx, audit); err != nil {
			return fmt.Errorf("create audit stream %s: %w", audit.Name, err)
		}
	}

	if cfg.Objects.Bucket != "" {
		objects := ObjectStore(namespace, cfg.Objects)
		if _, err := js.CreateOrUpdateObjectStore(ctx, objects); err != nil {
			return fmt.Errorf("create Object Store bucket %s: %w", objects.Bucket, err)
		}
	}

	dlqMaxAge, _ := time.ParseDuration(cfg.DLQ.MaxAge)
	if err := js.CreateOrUpdateStreamWithConfig(ctx, jetstream.StreamConfig{
		Name: streamName + "-DLQ",
		Subjects: []string{
			"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES." + streamName + ".*",
		},
		Storage:  StorageType(cfg.DLQ.Storage),
		MaxAge:   dlqMaxAge,
		MaxMsgs:  cfg.DLQ.MaxMsgs,
		Replicas: cfg.DLQ.Replicas,
		Metadata: map[string]string{
			"dead_letter_queue": "true",
		},
	}); err != nil {
		return fmt.Errorf("create DLQ stream: %w", err)
	}

	return nil
}
