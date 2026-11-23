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

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
)

// NATSClient is the access to NATS that the job client uses. A
// *natsclient.Client is one.
type NATSClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
	StreamInfo(ctx context.Context, name string) (*jetstream.StreamInfo, error)
	ConsumerCreateOrUpdate(
		ctx context.Context,
		stream string,
		config jetstream.ConsumerConfig,
	) error
	Consume(
		ctx context.Context,
		stream string,
		consumer string,
		handler natsclient.JetStreamMessageHandler,
		opts *natsclient.ConsumeOptions,
	) error
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVDelete(ctx context.Context, bucket string, key string) error
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
	KVWatch(
		ctx context.Context,
		bucket string,
		pattern string,
		opts ...jetstream.WatchOpt,
	) (<-chan jetstream.KeyValueEntry, error)
}

// ConsumerStore looks up and removes a JetStream consumer. The JetStream
// API of a *natsclient.Client (its ExtJS) is one.
type ConsumerStore interface {
	Consumer(ctx context.Context, stream string, consumer string) (jetstream.Consumer, error)
	DeleteConsumer(ctx context.Context, stream string, consumer string) error
}

// Ensure natsclient.Client implements NATSClient interface.
var _ NATSClient = (*natsclient.Client)(nil)
