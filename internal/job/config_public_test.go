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

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

type ConfigPublicTestSuite struct {
	suite.Suite
}

func (suite *ConfigPublicTestSuite) SetupTest() {}

func (suite *ConfigPublicTestSuite) SetupSubTest() {}

func (suite *ConfigPublicTestSuite) TearDownTest() {}

func (suite *ConfigPublicTestSuite) TestGetJobsStreamConfig() {
	tests := []struct {
		name         string
		streamConfig *config.NATSStream
		validateFunc func(config *jetstream.StreamConfig)
	}{
		{
			name: "when using file storage and old discard policy",
			streamConfig: &config.NATSStream{
				Name:     "JOBS",
				Subjects: "job.>",
				MaxAge:   "24h",
				MaxMsgs:  10000,
				Storage:  "file",
				Replicas: 1,
				Discard:  "old",
			},
			validateFunc: func(config *jetstream.StreamConfig) {
				suite.NotNil(config)
				suite.Equal("JOBS", config.Name)
				suite.Equal("Stream for job request and processing", config.Description)
				suite.Equal([]string{"job.>"}, config.Subjects)
				suite.Equal(jetstream.FileStorage, config.Storage)
				suite.Equal(1, config.Replicas)
				suite.Equal(24*time.Hour, config.MaxAge)
				suite.Equal(int64(10000), config.MaxMsgs)
				suite.Equal(jetstream.DiscardOld, config.Discard)
			},
		},
		{
			name: "when using memory storage and new discard policy",
			streamConfig: &config.NATSStream{
				Name:     "JOBS_MEMORY",
				Subjects: "job.memory.>",
				MaxAge:   "1h",
				MaxMsgs:  5000,
				Storage:  "memory",
				Replicas: 3,
				Discard:  "new",
			},
			validateFunc: func(config *jetstream.StreamConfig) {
				suite.NotNil(config)
				suite.Equal("JOBS_MEMORY", config.Name)
				suite.Equal("Stream for job request and processing", config.Description)
				suite.Equal([]string{"job.memory.>"}, config.Subjects)
				suite.Equal(jetstream.MemoryStorage, config.Storage)
				suite.Equal(3, config.Replicas)
				suite.Equal(1*time.Hour, config.MaxAge)
				suite.Equal(int64(5000), config.MaxMsgs)
				suite.Equal(jetstream.DiscardNew, config.Discard)
			},
		},
		{
			name: "when using unknown storage defaults to file",
			streamConfig: &config.NATSStream{
				Name:     "JOBS_UNKNOWN",
				Subjects: "job.unknown.>",
				MaxAge:   "12h",
				MaxMsgs:  1000,
				Storage:  "unknown",
				Replicas: 1,
				Discard:  "old",
			},
			validateFunc: func(config *jetstream.StreamConfig) {
				suite.NotNil(config)
				suite.Equal(jetstream.FileStorage, config.Storage)
			},
		},
		{
			name: "when using unknown discard defaults to old",
			streamConfig: &config.NATSStream{
				Name:     "JOBS_DISCARD",
				Subjects: "job.discard.>",
				MaxAge:   "6h",
				MaxMsgs:  2000,
				Storage:  "file",
				Replicas: 1,
				Discard:  "unknown",
			},
			validateFunc: func(config *jetstream.StreamConfig) {
				suite.NotNil(config)
				suite.Equal(jetstream.DiscardOld, config.Discard)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.validateFunc(job.GetJobsStreamConfig(tt.streamConfig))
		})
	}
}

func TestConfigPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConfigPublicTestSuite))
}
