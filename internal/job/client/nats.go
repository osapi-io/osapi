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

// NATSClient defines the NATS operations needed by the job client.
type NATSClient interface {
	Publish(
		ctx context.Context,
		subject string,
		data []byte,
	) error
	GetStreamInfo(
		ctx context.Context,
		streamName string,
	) (*jetstream.StreamInfo, error)
	KVPut(
		bucket string,
		key string,
		value []byte,
	) error
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
}

// Ensure natsclient.Client implements NATSClient interface.
var _ NATSClient = (*natsclient.Client)(nil)
