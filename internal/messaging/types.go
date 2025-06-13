// Copyright (c) 2025 John Dewey

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

package messaging

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
)

// NATSClient defines the interface for NATS messaging operations.
type NATSClient interface {
	// Connection management
	Connect() error

	// JetStream setup and configuration
	CreateOrUpdateStreamWithConfig(ctx context.Context, streamConfig *nats.StreamConfig) error
	CreateOrUpdateJetStreamWithConfig(
		ctx context.Context,
		streamConfig *nats.StreamConfig,
		consumerConfigs ...jetstream.ConsumerConfig,
	) error
	CreateOrUpdateConsumerWithConfig(
		ctx context.Context,
		streamName string,
		consumerConfig jetstream.ConsumerConfig,
	) error

	// Key-Value operations
	CreateKVBucket(bucketName string) (nats.KeyValue, error)
	KVPut(bucket, key string, value []byte) error
	KVGet(bucket, key string) ([]byte, error)
	KVDelete(bucket, key string) error
	KVKeys(bucket string) ([]string, error)

	// Request-reply operations
	PublishAndWaitKV(
		ctx context.Context,
		subject string,
		data []byte,
		kv nats.KeyValue,
		opts *natsclient.RequestReplyOptions,
	) ([]byte, error)

	// Message consumption
	ConsumeMessages(
		ctx context.Context,
		streamName, consumerName string,
		handler natsclient.JetStreamMessageHandler,
		opts *natsclient.ConsumeOptions,
	) error

	// Stream operations
	GetStreamInfo(ctx context.Context, streamName string) (*nats.StreamInfo, error)
}

// Ensure natsclient.Client implements NATSClient interface
var _ NATSClient = (*natsclient.Client)(nil)
