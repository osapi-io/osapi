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
	"errors"
	"log/slog"

	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

func (s *AgentPublicTestSuite) TestApplyConfig() {
	s.Run("a config that cannot be read keeps the labels", func() {
		s.appConfig.Agent.Labels = map[string]string{"group": "db"}
		a := s.buildAgent()
		s.mockJobClient.EXPECT().
			GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{}, errors.New("nats down"))
		labels, version := agent.ExportApplyConfig(a, "m-1")
		s.Equal(map[string]string{"group": "db"}, labels)
		s.Zero(version)
	})

	s.Run("the version that it applied changes nothing", func() {
		s.appConfig.Agent.Labels = map[string]string{"group": "db"}
		a := s.buildAgent()
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").Return(job.AgentConfig{}, nil)
		labels, _ := agent.ExportApplyConfig(a, "m-1")
		s.Equal(map[string]string{"group": "db"}, labels)
	})

	s.Run("new labels lie over the file's, and the timeline says so", func() {
		s.appConfig.Agent.MaxJobs = 10
		s.appConfig.Agent.Labels = map[string]string{"group": "db", "dc": "east"}
		a := s.buildAgent()
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}}, nil)
		s.mockJobClient.EXPECT().
			WriteAgentTimelineEvent(gomock.Any(), "m-1", gomock.Any(), "config_applied",
				"config version 2, labels dc=east,group=web, max jobs 10, log level info").
			Return(nil)
		labels, version := agent.ExportApplyConfig(a, "m-1")
		s.Equal(map[string]string{"group": "web", "dc": "east"}, labels)
		s.Equal(2, version)
	})

	s.Run("max jobs replaces the file's, and a version without it gives the file's back", func() {
		s.appConfig.Agent.MaxJobs = 10
		s.appConfig.Agent.Labels = map[string]string{"group": "web"}
		a := s.buildAgent()
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 1, Labels: map[string]string{"group": "web"}, MaxJobs: 2}, nil)
		s.mockJobClient.EXPECT().
			WriteAgentTimelineEvent(gomock.Any(), "m-1", gomock.Any(), "config_applied",
				"config version 1, labels group=web, max jobs 2, log level info").
			Return(nil)
		agent.ExportApplyConfig(a, "m-1")
		s.Equal(2, agent.ExportMaxJobs(a))

		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}}, nil)
		s.mockJobClient.EXPECT().
			WriteAgentTimelineEvent(gomock.Any(), "m-1", gomock.Any(), "config_applied",
				"config version 2, labels group=web, max jobs 10, log level info").
			Return(nil)
		agent.ExportApplyConfig(a, "m-1")
		s.Equal(10, agent.ExportMaxJobs(a))
	})

	s.Run(
		"a log level replaces the file's at once, and a version without it gives it back",
		func() {
			s.appConfig.Agent.MaxJobs = 10
			s.appConfig.Agent.Labels = map[string]string{"group": "web"}
			a := s.buildAgent()
			level := new(slog.LevelVar)
			a.SetLogLevel(level)
			s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
				Return(job.AgentConfig{Version: 1, Labels: map[string]string{"group": "web"}, LogLevel: types.LogDebug}, nil)
			s.mockJobClient.EXPECT().
				WriteAgentTimelineEvent(gomock.Any(), "m-1", gomock.Any(), "config_applied",
					"config version 1, labels group=web, max jobs 10, log level debug").
				Return(nil)
			agent.ExportApplyConfig(a, "m-1")
			s.Equal(slog.LevelDebug, level.Level())

			s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
				Return(job.AgentConfig{Version: 2, Labels: map[string]string{"group": "web"}}, nil)
			s.mockJobClient.EXPECT().
				WriteAgentTimelineEvent(gomock.Any(), "m-1", gomock.Any(), "config_applied",
					"config version 2, labels group=web, max jobs 10, log level info").
				Return(nil)
			agent.ExportApplyConfig(a, "m-1")
			s.Equal(slog.LevelInfo, level.Level())
		},
	)

	s.Run("a new version with the same labels writes no event", func() {
		s.appConfig.Agent.Labels = map[string]string{"group": "web"}
		a := s.buildAgent()
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 3, Labels: map[string]string{"group": "web"}}, nil)
		_, version := agent.ExportApplyConfig(a, "m-1")
		s.Equal(3, version)
	})

	s.Run("an agent with no labels of its own takes the config's", func() {
		s.appConfig.Agent.Labels = nil
		a := s.buildAgent()
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 1, Labels: map[string]string{"group": "web"}}, nil)
		s.mockJobClient.EXPECT().
			WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		labels, _ := agent.ExportApplyConfig(a, "m-1")
		s.Equal(map[string]string{"group": "web"}, labels)
	})

	s.Run("running consumers start again on the new labels", func() {
		s.appConfig.Agent.Labels = map[string]string{"group": "db"}
		a := s.buildAgent()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		consumerCtx, consumerCancel := context.WithCancel(ctx)
		agent.SetAgentLifecycle(ctx, consumerCtx, a, cancel, consumerCancel)
		s.mockJobClient.EXPECT().GetAgentConfig(gomock.Any(), "m-1").
			Return(job.AgentConfig{Version: 1, Labels: map[string]string{"group": "web"}}, nil)
		s.mockJobClient.EXPECT().
			WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		s.mockJobClient.EXPECT().
			CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
			Return(nil).
			AnyTimes()
		s.mockJobClient.EXPECT().
			ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
			Return(context.Canceled).
			AnyTimes()
		labels, _ := agent.ExportApplyConfig(a, "m-1")
		s.Equal(map[string]string{"group": "web"}, labels)
		s.Error(consumerCtx.Err(), "the consumers of the old labels stop")
		cancel()
		agent.WaitAgentWG(a)
	})
}
