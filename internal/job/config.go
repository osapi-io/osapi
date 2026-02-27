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

package job

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/config"
)

// GetJobsStreamConfig returns the stream configuration for job processing.
// This creates a stream that accepts all job-related subjects (job.>).
func GetJobsStreamConfig(
	streamConfig *config.NATSStream,
) *jetstream.StreamConfig {
	// Parse duration string to time.Duration
	maxAge, _ := time.ParseDuration(streamConfig.MaxAge)

	// Parse storage type
	var storage jetstream.StorageType
	if streamConfig.Storage == "memory" {
		storage = jetstream.MemoryStorage
	} else {
		storage = jetstream.FileStorage
	}

	// Parse discard policy
	var discard jetstream.DiscardPolicy
	if streamConfig.Discard == "new" {
		discard = jetstream.DiscardNew
	} else {
		discard = jetstream.DiscardOld
	}

	return &jetstream.StreamConfig{
		Name:        streamConfig.Name,
		Description: "Stream for job request and processing",
		Subjects:    []string{streamConfig.Subjects},
		Storage:     storage,
		Replicas:    streamConfig.Replicas,
		MaxAge:      maxAge,
		MaxMsgs:     streamConfig.MaxMsgs,
		Discard:     discard,
	}
}

// GetJobsConsumerConfig returns the consumer configuration for processing job requests.
func GetJobsConsumerConfig(
	consumerConfig *config.NodeAgentConsumer,
	streamSubjects string,
) jetstream.ConsumerConfig {
	// Parse duration string to time.Duration
	ackWait, _ := time.ParseDuration(consumerConfig.AckWait)

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if consumerConfig.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	return jetstream.ConsumerConfig{
		Name:          consumerConfig.Name,
		Description:   "Consumer for processing job requests",
		Durable:       consumerConfig.Name,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    consumerConfig.MaxDeliver,
		AckWait:       ackWait,
		MaxAckPending: consumerConfig.MaxAckPending,
		FilterSubject: streamSubjects,
		ReplayPolicy:  replayPolicy,
	}
}

// GetKVBucketConfig returns the KeyValue bucket configuration for storing job responses.
func GetKVBucketConfig(
	kvConfig *config.NATSKV,
) jetstream.KeyValueConfig {
	// Parse duration string to time.Duration
	ttl, _ := time.ParseDuration(kvConfig.TTL)

	// Parse storage type
	var storage jetstream.StorageType
	if kvConfig.Storage == "memory" {
		storage = jetstream.MemoryStorage
	} else {
		storage = jetstream.FileStorage
	}

	return jetstream.KeyValueConfig{
		Bucket:      kvConfig.ResponseBucket,
		Description: "Storage for job responses indexed by request ID",
		TTL:         ttl,
		MaxBytes:    kvConfig.MaxBytes,
		Storage:     storage,
		Replicas:    kvConfig.Replicas,
	}
}
