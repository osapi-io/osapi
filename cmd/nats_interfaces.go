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

package cmd

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
)

// NATSClient defines the NATS operations needed by cmd setup, runtime, and
// metrics functions. Each function that receives a NATSClient uses only a
// subset of these methods; the full interface is defined here so the
// natsBundle can store a single value that satisfies all consumers.
type NATSClient interface {
	// Connection management
	Connect() error
	Close()

	// Stream operations
	CreateOrUpdateStreamWithConfig(
		ctx context.Context,
		streamConfig jetstream.StreamConfig,
	) error
	GetStreamInfo(
		ctx context.Context,
		streamName string,
	) (*jetstream.StreamInfo, error)

	// KV operations
	CreateOrUpdateKVBucket(
		ctx context.Context,
		bucketName string,
	) (jetstream.KeyValue, error)
	CreateOrUpdateKVBucketWithConfig(
		ctx context.Context,
		config jetstream.KeyValueConfig,
	) (jetstream.KeyValue, error)

	// Object Store operations
	CreateOrUpdateObjectStore(
		ctx context.Context,
		cfg jetstream.ObjectStoreConfig,
	) (jetstream.ObjectStore, error)
	ObjectStore(
		ctx context.Context,
		name string,
	) (jetstream.ObjectStore, error)

	// Message publishing
	Publish(
		ctx context.Context,
		subject string,
		data []byte,
	) error

	// Message consumption
	ConsumeMessages(
		ctx context.Context,
		streamName string,
		consumerName string,
		handler natsclient.JetStreamMessageHandler,
		opts *natsclient.ConsumeOptions,
	) error
	CreateOrUpdateConsumerWithConfig(
		ctx context.Context,
		streamName string,
		consumerConfig jetstream.ConsumerConfig,
	) error

	// KV convenience operations
	KVPut(
		bucket string,
		key string,
		value []byte,
	) error

	// Connection inspection (replaces type assertions against *natsclient.Client)
	ConnectedURL() string
	ConnectedServerVersion() string

	// JetStream handle access (replaces ExtJS field access)
	KeyValue(
		ctx context.Context,
		bucket string,
	) (jetstream.KeyValue, error)
	Stream(
		ctx context.Context,
		name string,
	) (jetstream.Stream, error)
}

// Ensure natsclient.Client implements NATSClient interface.
var _ NATSClient = (*natsclient.Client)(nil)
