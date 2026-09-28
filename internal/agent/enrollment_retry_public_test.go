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
	"fmt"
	"log/slog"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/agent/pki"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// EnrollmentRetryPublicTestSuite covers the republish that keeps a lost enrollment
// request from leaving an agent Pending forever.
type EnrollmentRetryPublicTestSuite struct {
	suite.Suite
}

// pendingAgent builds an agent with a loaded keypair and the state the test wants.
func (s *EnrollmentRetryPublicTestSuite) pendingAgent(
	state string,
	publisher *agentMocks.MockNATSPublisher,
) *agent.Agent {
	fs := memfs.New()
	cfg := config.Config{
		Agent: config.AgentConfig{
			PKI: config.AgentPKI{Enabled: true, KeyDir: "/keys"},
		},
	}

	a := agent.New(
		fs, cfg, slog.Default(),
		nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, publisher,
	)

	m := pki.New(fs, "/keys", "agent")
	s.Require().NoError(m.LoadOrGenerate())
	agent.SetAgentPKIManager(a, m)
	agent.SetAgentState(a, state)
	agent.SetAgentMachineID(a, "machine-1")

	return a
}

func (s *EnrollmentRetryPublicTestSuite) TestRetryEnrollmentIfPending() {
	tests := []struct {
		name      string
		state     string
		setupMock func(*agentMocks.MockNATSPublisher)
	}{
		{
			name:  "a pending agent asks again",
			state: job.AgentStatePending,
			setupMock: func(publisher *agentMocks.MockNATSPublisher) {
				// The controller keys acceptance by machine ID, so asking again
				// overwrites the pending entry rather than adding one.
				publisher.EXPECT().
					PublishCore(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
		},
		{
			name:  "a pending agent whose republish fails keeps trying",
			state: job.AgentStatePending,
			setupMock: func(publisher *agentMocks.MockNATSPublisher) {
				publisher.EXPECT().
					PublishCore(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("no route to controller")).
					Times(1)
			},
		},
		{
			name:  "a ready agent asks for nothing",
			state: job.AgentStateReady,
			setupMock: func(publisher *agentMocks.MockNATSPublisher) {
				publisher.EXPECT().
					PublishCore(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
		{
			name:  "a cordoned agent asks for nothing either",
			state: job.AgentStateCordoned,
			setupMock: func(publisher *agentMocks.MockNATSPublisher) {
				publisher.EXPECT().
					PublishCore(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			publisher := agentMocks.NewMockNATSPublisher(ctrl)

			tt.setupMock(publisher)

			a := s.pendingAgent(tt.state, publisher)

			agent.ExportRetryEnrollmentIfPending(a)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestEnrollmentRetryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EnrollmentRetryPublicTestSuite))
}
