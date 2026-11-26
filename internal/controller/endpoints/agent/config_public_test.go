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
package agent_test

import (
	"context"
	"time"

	"github.com/osapi-io/osapi/internal/controller/endpoints/agent/mocks"

	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	"github.com/osapi-io/osapi/internal/job"
)

func (s *MethodsPublicTestSuite) TestConfig() {
	s.Run("list and get report the version that the controller holds, and behind", func() {
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().ListAgents(gomock.Any()).
			Return([]job.AgentInfo{{MachineID: "m-1", Hostname: "server1", ConfigVersion: 2}}, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{Version: 3}, nil)
		got, err := agent.List(context.Background(), s.deps(), agent.ListInput{})
		s.Require().NoError(err)
		s.Equal(3, got.Items[0].ConfigWanted)
		s.True(got.Items[0].ConfigBehind)

		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
			Return(&job.AgentInfo{MachineID: "m-1", Hostname: "server1", ConfigVersion: 3}, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{Version: 3}, nil)
		one, err := agent.Get(context.Background(), s.deps(), agent.GetInput{Hostname: "server1"})
		s.Require().NoError(err)
		s.False(one.ConfigBehind)

		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
			Return(&job.AgentInfo{MachineID: "m-1", Hostname: "server1"}, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{}, errBoom)
		one, err = agent.Get(context.Background(), s.deps(), agent.GetInput{Hostname: "server1"})
		s.Require().NoError(err, "a config that cannot be read leaves it out")
		s.Zero(one.ConfigWanted)
	})

	ctx := context.Background()
	host := &job.AgentInfo{MachineID: "m-1", Hostname: "server1"}
	at := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	s.Run("get: the config that the controller holds", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(host, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}, UpdatedAt: at}, nil)
		got, err := agent.GetConfig(ctx, s.deps(), agent.ConfigGetInput{Hostname: "server1"})
		s.Require().NoError(err)
		s.Equal(2, got.Version)
		s.Equal(at, *got.UpdatedAt)
	})
	s.Run("get: an agent with none has version 0, and no time", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(host, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{}, nil)
		got, err := agent.GetConfig(ctx, s.deps(), agent.ConfigGetInput{Hostname: "server1"})
		s.Require().NoError(err)
		s.Zero(got.Version)
		s.Nil(got.UpdatedAt)
	})
	s.Run("get: an agent that is not there", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(nil, missing())
		_, err := agent.GetConfig(ctx, s.deps(), agent.ConfigGetInput{Hostname: "server1"})
		s.ErrorIs(err, agent.ErrNotFound)
	})
	s.Run("get: a config that cannot be read", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(host, nil)
		s.jobs.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{}, errBoom)
		_, err := agent.GetConfig(ctx, s.deps(), agent.ConfigGetInput{Hostname: "server1"})
		s.ErrorIs(err, errBoom)
	})
	s.Run("update: the next version", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		labels := map[string]string{"group": "web"}
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(host, nil)
		s.jobs.EXPECT().SetAgentConfig(gomock.Any(), "m-1", labels).
			Return(job.AgentConfig{Version: 3, Labels: labels, UpdatedAt: at}, nil)
		got, err := agent.SetConfig(ctx, s.deps(), agent.ConfigSetInput{
			Hostname: "server1", Body: agent.AgentConfigRequest{Labels: labels},
		})
		s.Require().NoError(err)
		s.Equal(3, got.Version)
	})
	s.Run("update: an agent that is not there", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(nil, missing())
		_, err := agent.SetConfig(ctx, s.deps(), agent.ConfigSetInput{Hostname: "server1"})
		s.ErrorIs(err, agent.ErrNotFound)
	})
	s.Run("update: a config that cannot be written", func() {
		// A mock with no default read of the config.
		s.jobs = mocks.NewMockJobs(s.ctrl)
		s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(host, nil)
		s.jobs.EXPECT().
			SetAgentConfig(gomock.Any(), "m-1", gomock.Any()).
			Return(job.AgentConfig{}, errBoom)
		_, err := agent.SetConfig(ctx, s.deps(), agent.ConfigSetInput{Hostname: "server1"})
		s.ErrorIs(err, errBoom)
	})
}
