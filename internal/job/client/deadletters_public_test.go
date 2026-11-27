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
	"strconv"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type DeadLettersPublicTestSuite struct {
	suite.Suite
	nats *jobmocks.MockNATSClient
	jobs *client.Client
}

func TestDeadLettersPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DeadLettersPublicTestSuite))
}

func (s *DeadLettersPublicTestSuite) SetupSubTest() {
	s.nats = jobmocks.NewMockNATSClient(gomock.NewController(s.T()))
	var err error
	s.jobs, err = newTrusted(s.nats, &client.Options{JobsBucket: "jobs", StreamName: "JOBS"})
	s.Require().NoError(err)
}

// state is the dead letter stream with its first and last sequences, and
// msgs messages.
func (s *DeadLettersPublicTestSuite) state(first, last, msgs uint64) {
	s.nats.EXPECT().StreamInfo(gomock.Any(), "JOBS-DLQ").Return(&jetstream.StreamInfo{
		State: jetstream.StreamState{FirstSeq: first, LastSeq: last, Msgs: msgs},
	}, nil)
}

// advisory is the dead letter at seq, for the job stream's message streamSeq.
func (s *DeadLettersPublicTestSuite) advisory(seq, streamSeq uint64) {
	data := []byte(`{"type":"io.nats.jetstream.advisory.v1.max_deliver",` +
		`"timestamp":"2026-11-26T10:00:00Z","stream":"JOBS","consumer":"jobs-agent",` +
		`"stream_seq":` + strconv.FormatUint(streamSeq, 10) + `,"deliveries":5}`)
	s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS-DLQ", seq).
		Return(&jetstream.RawStreamMsg{Data: data}, nil)
}

func (s *DeadLettersPublicTestSuite) TestListDeadLetters() {
	s.Run("the newest first, each with its job, up to the limit", func() {
		s.state(1, 3, 3)
		s.advisory(3, 30)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS", uint64(30)).
			Return(&jetstream.RawStreamMsg{Data: []byte("job-3")}, nil)
		s.advisory(2, 20)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS", uint64(20)).
			Return(nil, jetstream.ErrMsgNotFound)

		got, total, err := s.jobs.ListDeadLetters(context.Background(), 2)
		s.Require().NoError(err)
		s.Equal(3, total)
		s.Require().Len(got, 2)
		s.Equal(uint64(3), got[0].Seq)
		s.Equal("job-3", got[0].JobID)
		s.Equal("jobs-agent", got[0].Consumer)
		s.Equal(uint64(5), got[0].Deliveries)
		s.Equal(time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC), got[0].Time)
		s.Empty(got[1].JobID, "a job whose message is gone has no ID")
	})
	s.Run("a sequence that the stream no longer has is skipped", func() {
		s.state(1, 2, 1)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS-DLQ", uint64(2)).
			Return(nil, jetstream.ErrMsgNotFound)
		s.advisory(1, 10)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS", uint64(10)).
			Return(&jetstream.RawStreamMsg{Data: []byte("job-1")}, nil)

		got, _, err := s.jobs.ListDeadLetters(context.Background(), 10)
		s.Require().NoError(err)
		s.Require().Len(got, 1)
		s.Equal(uint64(1), got[0].Seq)
	})
	s.Run("an empty stream is none", func() {
		s.state(0, 0, 0)
		got, total, err := s.jobs.ListDeadLetters(context.Background(), 10)
		s.Require().NoError(err)
		s.Empty(got)
		s.Zero(total)
	})
	s.Run("a stream that cannot be read is the error", func() {
		s.nats.EXPECT().StreamInfo(gomock.Any(), "JOBS-DLQ").Return(nil, errors.New("nats down"))
		_, _, err := s.jobs.ListDeadLetters(context.Background(), 10)
		s.ErrorContains(err, "read the dead letter stream JOBS-DLQ: nats down")
	})
	s.Run("a dead letter that cannot be read is the error", func() {
		s.state(1, 1, 1)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS-DLQ", uint64(1)).
			Return(nil, errors.New("timeout"))
		_, _, err := s.jobs.ListDeadLetters(context.Background(), 10)
		s.ErrorContains(err, "read dead letter 1: timeout")
	})
	s.Run("a dead letter that is not an advisory is skipped, not the error", func() {
		s.state(1, 2, 2)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS-DLQ", uint64(2)).
			Return(&jetstream.RawStreamMsg{Data: []byte("not json")}, nil)
		s.advisory(1, 10)
		s.nats.EXPECT().StreamGetMsg(gomock.Any(), "JOBS", uint64(10)).
			Return(&jetstream.RawStreamMsg{Data: []byte("job-1")}, nil)
		got, _, err := s.jobs.ListDeadLetters(context.Background(), 10)
		s.Require().NoError(err)
		s.Require().Len(got, 1)
		s.Equal("job-1", got[0].JobID)
	})
}
