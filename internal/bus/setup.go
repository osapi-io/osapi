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
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// Prepare connects to NATS with opts, creates each stream and bucket with
// Setup, then closes the connection.
func Prepare(
	ctx context.Context,
	log *slog.Logger,
	opts *natsclient.Options,
	namespace string,
	cfg config.NATS,
) error {
	nc, err := Dial(ctx, log, opts)
	if err != nil {
		return err
	}
	defer Close(nc)

	return Setup(ctx, nc, namespace, cfg)
}

// Dial connects to NATS with opts. The deadline of ctx bounds the dial.
func Dial(
	ctx context.Context,
	log *slog.Logger,
	opts *natsclient.Options,
) (*natsclient.Client, error) {
	nc := natsclient.New(log, opts)
	if err := nc.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	return nc, nil
}

// Telemetry turns on the spans and the metrics of nats-client in opts. The
// spans use the global tracer provider, which tracing sets. The metrics go to
// mp, the meter provider of the component; nil sends them to the global
// meter provider, which records nothing unless a program sets one.
func Telemetry(
	opts *natsclient.Options,
	mp *sdkmetric.MeterProvider,
) *natsclient.Options {
	opts.Telemetry = true
	if mp != nil {
		opts.MeterProvider = mp
	}

	return opts
}

// JetStream is the access to NATS that Setup uses. A *natsclient.Client is
// one.
type JetStream interface {
	StreamCreateOrUpdate(ctx context.Context, config jetstream.StreamConfig) error
	KVCreateOrUpdate(ctx context.Context, config jetstream.KeyValueConfig) error
	ObjectStoreCreateOrUpdate(ctx context.Context, config jetstream.ObjectStoreConfig) error
}

// checkDurations returns an error for the first duration of cfg that does
// not parse. The builders of the streams and buckets read a duration that
// does not parse as 0, which is no limit. Setup is exported, and a caller
// such as osapi nats setup can give it a config that was not validated. An
// empty duration is no limit, and passes.
func checkDurations(
	cfg config.NATS,
) error {
	for _, d := range []struct{ name, value string }{
		{"nats.stream.max_age", cfg.Stream.MaxAge},
		{"nats.kv.ttl", cfg.KV.TTL},
		{"nats.registry.ttl", cfg.Registry.TTL},
		{"nats.facts.ttl", cfg.Facts.TTL},
		{"nats.audit.max_age", cfg.Audit.MaxAge},
		{"nats.dlq.max_age", cfg.DLQ.MaxAge},
	} {
		if d.value == "" {
			continue
		}
		if _, err := time.ParseDuration(d.value); err != nil {
			return fmt.Errorf("%s: %w", d.name, err)
		}
	}

	return nil
}

// Setup creates or updates each stream and bucket that osapi uses, with the
// names in namespace. A second run with the same config changes nothing. A
// bucket with no name in cfg is not made.
func Setup(
	ctx context.Context,
	nc JetStream,
	namespace string,
	cfg config.NATS,
) error {
	if err := checkDurations(cfg); err != nil {
		return err
	}
	streamName := job.ApplyNamespaceToInfraName(namespace, cfg.Stream.Name)
	streamMaxAge, _ := time.ParseDuration(cfg.Stream.MaxAge)

	discard := jetstream.DiscardOld
	if cfg.Stream.Discard == "new" {
		discard = jetstream.DiscardNew
	}

	if err := nc.StreamCreateOrUpdate(ctx, jetstream.StreamConfig{
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
	if cfg.Secrets.Bucket != "" {
		kvs = append(kvs, SecretsKV(namespace, cfg.Secrets))
	}
	for _, kv := range kvs {
		if err := nc.KVCreateOrUpdate(ctx, kv); err != nil {
			return fmt.Errorf("create KV bucket %s: %w", kv.Bucket, err)
		}
	}

	if cfg.Audit.Stream != "" {
		audit := AuditStream(namespace, cfg.Audit)
		if err := nc.StreamCreateOrUpdate(ctx, audit); err != nil {
			return fmt.Errorf("create audit stream %s: %w", audit.Name, err)
		}
	}

	if cfg.Objects.Bucket != "" {
		objects := ObjectStore(namespace, cfg.Objects)
		if err := nc.ObjectStoreCreateOrUpdate(ctx, objects); err != nil {
			return fmt.Errorf("create Object Store bucket %s: %w", objects.Bucket, err)
		}
	}

	dlqMaxAge, _ := time.ParseDuration(cfg.DLQ.MaxAge)
	if err := nc.StreamCreateOrUpdate(ctx, jetstream.StreamConfig{
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
