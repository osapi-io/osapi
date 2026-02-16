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

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/config"
)

// GetJobsStreamConfig returns the stream configuration for job processing.
// This creates a stream that accepts all job-related subjects (job.>).
func GetJobsStreamConfig(
	jobConfig *config.Job,
) *nats.StreamConfig {
	// Parse duration string to time.Duration
	maxAge, _ := time.ParseDuration(jobConfig.Stream.MaxAge)

	// Parse storage type
	var storage nats.StorageType
	if jobConfig.Stream.Storage == "memory" {
		storage = nats.MemoryStorage
	} else {
		storage = nats.FileStorage
	}

	// Parse discard policy
	var discard nats.DiscardPolicy
	if jobConfig.Stream.Discard == "new" {
		discard = nats.DiscardNew
	} else {
		discard = nats.DiscardOld
	}

	return &nats.StreamConfig{
		Name:        jobConfig.StreamName,
		Description: "Stream for job request and processing",
		Subjects:    []string{jobConfig.StreamSubjects},
		Storage:     storage,
		Replicas:    jobConfig.Stream.Replicas,
		MaxAge:      maxAge,
		MaxMsgs:     jobConfig.Stream.MaxMsgs,
		Discard:     discard,
	}
}

// GetJobsConsumerConfig returns the consumer configuration for processing job requests.
func GetJobsConsumerConfig(
	jobConfig *config.Job,
) jetstream.ConsumerConfig {
	// Parse duration string to time.Duration
	ackWait, _ := time.ParseDuration(jobConfig.Consumer.AckWait)

	// Parse replay policy
	var replayPolicy jetstream.ReplayPolicy
	if jobConfig.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	return jetstream.ConsumerConfig{
		Name:          jobConfig.ConsumerName,
		Description:   "Consumer for processing job requests",
		Durable:       jobConfig.ConsumerName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    jobConfig.Consumer.MaxDeliver,
		AckWait:       ackWait,
		MaxAckPending: jobConfig.Consumer.MaxAckPending,
		FilterSubject: jobConfig.StreamSubjects,
		ReplayPolicy:  replayPolicy,
	}
}

// GetKVBucketConfig returns the KeyValue bucket configuration for storing job responses.
func GetKVBucketConfig(
	jobConfig *config.Job,
) *nats.KeyValueConfig {
	// Parse duration string to time.Duration
	ttl, _ := time.ParseDuration(jobConfig.KV.TTL)

	// Parse storage type
	var storage nats.StorageType
	if jobConfig.KV.Storage == "memory" {
		storage = nats.MemoryStorage
	} else {
		storage = nats.FileStorage
	}

	return &nats.KeyValueConfig{
		Bucket:      jobConfig.KVResponseBucket,
		Description: "Storage for job responses indexed by request ID",
		TTL:         ttl,
		MaxBytes:    jobConfig.KV.MaxBytes,
		Storage:     storage,
		Replicas:    jobConfig.KV.Replicas,
	}
}
