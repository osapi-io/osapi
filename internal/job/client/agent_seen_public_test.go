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

package client_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type AgentSeenPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	nc   *jobmocks.MockNATSClient
}

func TestAgentSeenPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentSeenPublicTestSuite))
}

func (s *AgentSeenPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.nc = jobmocks.NewMockNATSClient(s.ctrl)
}

func (s *AgentSeenPublicTestSuite) client(
	stateBucket string,
) *client.Client {
	c, err := client.New(slog.Default(), s.nc, &client.Options{
		Timeout: time.Second, JobsBucket: "jobs", StateBucket: stateBucket,
	})
	s.Require().NoError(err)

	return c
}

func (s *AgentSeenPublicTestSuite) TestWriteAgentSeen() {
	s.Run("when the state bucket is set writes the time under the machine ID", func() {
		var stored []byte
		s.nc.EXPECT().KVPut(gomock.Any(), "state", "seen.m-1", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, v []byte) (uint64, error) {
				stored = v
				return 1, nil
			})

		s.Require().NoError(s.client("state").WriteAgentSeen(context.Background(), "m-1"))
		var t time.Time
		s.Require().NoError(t.UnmarshalText(stored))
		s.WithinDuration(time.Now(), t, 5*time.Second)
	})

	s.Run("when the write fails returns the error", func() {
		s.nc.EXPECT().
			KVPut(gomock.Any(), "state", "seen.m-1", gomock.Any()).
			Return(uint64(0), errors.New("kv down"))

		s.ErrorContains(
			s.client("state").WriteAgentSeen(context.Background(), "m-1"),
			"write last seen: kv down",
		)
	})

	s.Run("when there is no state bucket returns an error", func() {
		s.ErrorContains(s.client("").WriteAgentSeen(context.Background(), "m-1"), "not configured")
	})
}

func (s *AgentSeenPublicTestSuite) TestAgentsLastSeen() {
	at := time.Date(2026, 11, 19, 1, 0, 0, 0, time.UTC)

	s.Run(
		"when the bucket has times returns them by machine ID, leaving out those that do not read",
		func() {
			s.nc.EXPECT().KVKeys(gomock.Any(), "state", "seen.>").
				Return([]string{"seen.m-1", "seen.bad", "seen.gone"}, nil)
			good := jobmocks.NewMockKeyValueEntry(s.ctrl)
			text, _ := at.MarshalText()
			good.EXPECT().Value().Return(text)
			s.nc.EXPECT().KVGet(gomock.Any(), "state", "seen.m-1").Return(good, nil)
			bad := jobmocks.NewMockKeyValueEntry(s.ctrl)
			bad.EXPECT().Value().Return([]byte("yesterday"))
			s.nc.EXPECT().KVGet(gomock.Any(), "state", "seen.bad").Return(bad, nil)
			s.nc.EXPECT().
				KVGet(gomock.Any(), "state", "seen.gone").
				Return(nil, errors.New("not found"))

			got, err := s.client("state").AgentsLastSeen(context.Background())

			s.Require().NoError(err)
			s.Equal(map[string]time.Time{"m-1": at}, got)
		},
	)

	s.Run("when the keys cannot be listed returns the error", func() {
		s.nc.EXPECT().KVKeys(gomock.Any(), "state", "seen.>").Return(nil, errors.New("kv down"))

		_, err := s.client("state").AgentsLastSeen(context.Background())

		s.ErrorContains(err, "list last seen: kv down")
	})

	s.Run("when there is no state bucket returns none", func() {
		got, err := s.client("").AgentsLastSeen(context.Background())

		s.Require().NoError(err)
		s.Empty(got)
	})
}
