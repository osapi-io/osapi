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

package job_test

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
)

type ConfigPublicTestSuite struct {
	suite.Suite
}

func (suite *ConfigPublicTestSuite) SetupTest() {}

func (suite *ConfigPublicTestSuite) SetupSubTest() {}

func (suite *ConfigPublicTestSuite) TearDownTest() {}

func (suite *ConfigPublicTestSuite) TestGetJobsStreamConfig() {
	jobConfig := &config.Job{
		StreamName:     "JOBS",
		StreamSubjects: "job.>",
		Stream: config.JobStream{
			MaxAge:   "24h",
			MaxMsgs:  10000,
			Storage:  "file",
			Replicas: 1,
			Discard:  "old",
		},
	}
	streamConfig := job.GetJobsStreamConfig(jobConfig)

	// Verify stream configuration
	suite.NotNil(streamConfig)
	suite.Equal("JOBS", streamConfig.Name)
	suite.Equal("Stream for job request and processing", streamConfig.Description)
	suite.Equal([]string{"job.>"}, streamConfig.Subjects)
	suite.Equal(nats.FileStorage, streamConfig.Storage)
	suite.Equal(1, streamConfig.Replicas)
	suite.Equal(24*time.Hour, streamConfig.MaxAge)
	suite.Equal(int64(10000), streamConfig.MaxMsgs)
	suite.Equal(nats.DiscardOld, streamConfig.Discard)
}

func (suite *ConfigPublicTestSuite) TestGetJobsConsumerConfig() {
	jobConfig := &config.Job{
		StreamSubjects: "job.>",
		ConsumerName:   "jobs-worker",
		Consumer: config.JobConsumer{
			MaxDeliver:    5,
			AckWait:       "30s",
			MaxAckPending: 100,
			ReplayPolicy:  "instant",
		},
	}
	consumerConfig := job.GetJobsConsumerConfig(jobConfig)

	// Verify consumer configuration
	suite.Equal("jobs-worker", consumerConfig.Name)
	suite.Equal("Consumer for processing job requests", consumerConfig.Description)
	suite.Equal("jobs-worker", consumerConfig.Durable)
	suite.Equal(jetstream.AckExplicitPolicy, consumerConfig.AckPolicy)
	suite.Equal(5, consumerConfig.MaxDeliver)
	suite.Equal(30*time.Second, consumerConfig.AckWait)
	suite.Equal(100, consumerConfig.MaxAckPending)
	suite.Equal("job.>", consumerConfig.FilterSubject)
	suite.Equal(jetstream.ReplayInstantPolicy, consumerConfig.ReplayPolicy)
}

func (suite *ConfigPublicTestSuite) TestGetKVBucketConfig() {
	jobConfig := &config.Job{
		KVResponseBucket: "job-responses",
		KV: config.JobKV{
			TTL:      "1h",
			MaxBytes: 104857600, // 100MB
			Storage:  "file",
			Replicas: 1,
		},
	}
	kvConfig := job.GetKVBucketConfig(jobConfig)

	// Verify KV bucket configuration
	suite.NotNil(kvConfig)
	suite.Equal("job-responses", kvConfig.Bucket)
	suite.Equal("Storage for job responses indexed by request ID", kvConfig.Description)
	suite.Equal(1*time.Hour, kvConfig.TTL)
	suite.Equal(int64(100*1024*1024), kvConfig.MaxBytes)
	suite.Equal(nats.FileStorage, kvConfig.Storage)
	suite.Equal(1, kvConfig.Replicas)
}

func TestConfigPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigPublicTestSuite))
}
