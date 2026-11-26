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
package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type AgentConfigPublicTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	nc   *jobmocks.MockNATSClient
}

func TestAgentConfigPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentConfigPublicTestSuite))
}

func (s *AgentConfigPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.nc = jobmocks.NewMockNATSClient(s.ctrl)
}

func (s *AgentConfigPublicTestSuite) client(
	state string,
) *client.Client {
	c, err := newTrusted(
		s.nc,
		&client.Options{Timeout: time.Second, JobsBucket: "jobs", StateBucket: state},
	)
	s.Require().NoError(err)

	return c
}

// entry is a KV entry with value v.
func (s *AgentConfigPublicTestSuite) entry(
	v []byte,
) jetstream.KeyValueEntry {
	e := jobmocks.NewMockKeyValueEntry(s.ctrl)
	e.EXPECT().Value().Return(v).AnyTimes()

	return e
}

func (s *AgentConfigPublicTestSuite) TestGetAgentConfig() {
	ctx := context.Background()
	stored, _ := json.Marshal(
		job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}},
	)
	tests := []struct {
		name    string
		state   string
		setup   func()
		want    job.AgentConfig
		wantErr string
	}{
		{name: "no state bucket", state: "", wantErr: "agent state bucket not configured"},
		{name: "an agent with no config has version 0", state: "state", setup: func() {
			s.nc.EXPECT().
				KVGet(gomock.Any(), "state", "config.m-1").
				Return(nil, jetstream.ErrKeyNotFound)
		}},
		{name: "the stored config", state: "state", setup: func() {
			s.nc.EXPECT().KVGet(gomock.Any(), "state", "config.m-1").Return(s.entry(stored), nil)
		}, want: job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}}},
		{name: "a bucket that cannot be read", state: "state", setup: func() {
			s.nc.EXPECT().
				KVGet(gomock.Any(), "state", "config.m-1").
				Return(nil, errors.New("nats down"))
		}, wantErr: "read the config of m-1: nats down"},
		{name: "a config that is not JSON", state: "state", setup: func() {
			s.nc.EXPECT().
				KVGet(gomock.Any(), "state", "config.m-1").
				Return(s.entry([]byte("{")), nil)
		}, wantErr: "the config of m-1"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := s.client(tt.state).GetAgentConfig(ctx, "m-1")
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tt.want, got)
		})
	}
}

func (s *AgentConfigPublicTestSuite) TestSetAgentConfig() {
	ctx := context.Background()
	stored, _ := json.Marshal(job.AgentConfig{Version: 2, Labels: map[string]string{"group": "db"}})
	labels := map[string]string{"group": "web"}

	s.Run("the next version, with the new labels", func() {
		s.nc.EXPECT().KVGet(gomock.Any(), "state", "config.m-1").Return(s.entry(stored), nil)
		s.nc.EXPECT().KVPut(gomock.Any(), "state", "config.m-1", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
				var cfg job.AgentConfig
				s.Require().NoError(json.Unmarshal(data, &cfg))
				s.Equal(3, cfg.Version)
				s.Equal(labels, cfg.Labels)
				s.Equal(4, cfg.MaxJobs)
				s.Equal(types.LogDebug, cfg.LogLevel)
				s.Equal("5m", cfg.FactsInterval)
				s.False(cfg.UpdatedAt.IsZero())

				return 1, nil
			})
		got, err := s.client("state").SetAgentConfig(ctx, "m-1",
			job.AgentConfig{
				Labels:        labels,
				MaxJobs:       4,
				LogLevel:      types.LogDebug,
				FactsInterval: "5m",
				Version:       9,
			})
		s.Require().NoError(err)
		s.Equal(3, got.Version)
	})
	s.Run("a config that cannot be read", func() {
		s.nc.EXPECT().
			KVGet(gomock.Any(), "state", "config.m-1").
			Return(nil, errors.New("nats down"))
		_, err := s.client("state").SetAgentConfig(ctx, "m-1", job.AgentConfig{Labels: labels})
		s.ErrorContains(err, "nats down")
	})
	s.Run("a config that cannot be written", func() {
		s.nc.EXPECT().
			KVGet(gomock.Any(), "state", "config.m-1").
			Return(nil, jetstream.ErrKeyNotFound)
		s.nc.EXPECT().
			KVPut(gomock.Any(), "state", "config.m-1", gomock.Any()).
			Return(uint64(0), errors.New("full"))
		_, err := s.client("state").SetAgentConfig(ctx, "m-1", job.AgentConfig{Labels: labels})
		s.ErrorContains(err, "write the config of m-1: full")
	})
	s.Run("a config that cannot be marshaled", func() {
		s.nc.EXPECT().
			KVGet(gomock.Any(), "state", "config.m-1").
			Return(nil, jetstream.ErrKeyNotFound)
		c := s.client("state")
		c.JSONMarshalFn = func(any) ([]byte, error) { return nil, errors.New("no json") }
		_, err := c.SetAgentConfig(ctx, "m-1", job.AgentConfig{Labels: labels})
		s.ErrorContains(err, "marshal the config of m-1")
	})
}
