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
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package client_test

import (
	"context"
	"errors"
	"strings"
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

type StorePublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
	nc       *jobmocks.MockNATSClient
	purger   *jobmocks.MockStorePurger
}

func TestStorePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StorePublicTestSuite))
}

func (s *StorePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.nc = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.purger = jobmocks.NewMockStorePurger(s.mockCtrl)
}

func (s *StorePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// errFull is the error that NATS gives for a write past a bucket's
// max_bytes.
var errFull = &jetstream.APIError{
	Code:        503,
	ErrorCode:   10077,
	Description: "maximum bytes exceeded",
}

// jobKey matches the put of a job's record, not of its status events.
var jobKey = gomock.Cond(func(key string) bool { return strings.HasPrefix(key, "jobs.") })

// statusKey matches the put of a job's first status event.
var statusKey = gomock.Cond(func(key string) bool { return strings.HasPrefix(key, "status.") })

// The bucket, oldest first: a job that an agent runs now, a finished job,
// a job no agent has answered, and a newer failed one. A third of four
// jobs is one job, and the job that runs stays.
var bucketKeys = []string{
	"jobs.running",
	"status.running.started.web-01.0",
	"jobs.old",
	"status.old.completed.web-01.1",
	"responses.old.web-01",
	"jobs.pending",
	"jobs.new",
	"status.new.failed.web-01.2",
}

func (s *StorePublicTestSuite) newClient(
	purger client.StorePurger,
) *client.Client {
	c, err := newTrusted(s.nc, &client.Options{
		Timeout:    time.Second,
		JobsBucket: "jobs",
		StreamName: "JOBS",
		Purger:     purger,
	})
	s.Require().NoError(err)

	return c
}

// sent expects the status event and the notification of a job whose
// record was stored.
func (s *StorePublicTestSuite) sent() {
	s.nc.EXPECT().KVPut(gomock.Any(), "jobs", statusKey, gomock.Any()).Return(uint64(2), nil)
	s.nc.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
}

func (s *StorePublicTestSuite) TestCreateJobWhenTheStoreIsFull() {
	tests := []struct {
		name       string
		noPurger   bool
		setupMocks func()
		wantErr    string
	}{
		{
			name:     "with no purger a full store is ErrStoreFull",
			noPurger: true,
			setupMocks: func() {
				s.nc.EXPECT().
					KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
					Return(uint64(0), errFull)
			},
			wantErr: "maximum bytes exceeded",
		},
		{
			name: "a write that another write made room for removes nothing",
			setupMocks: func() {
				gomock.InOrder(
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(0), errFull),
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(1), nil),
				)
				s.sent()
			},
		},
		{
			name: "removes the oldest finished job and writes again",
			setupMocks: func() {
				gomock.InOrder(
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(0), errFull),
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(0), errFull),
					s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return(bucketKeys, nil),
					s.purger.EXPECT().
						PurgeSubject(gomock.Any(), "KV_jobs", "$KV.jobs.status.old.>").
						Return(nil),
					s.purger.EXPECT().
						PurgeSubject(gomock.Any(), "KV_jobs", "$KV.jobs.responses.old.>").
						Return(nil),
					s.purger.EXPECT().
						PurgeSubject(gomock.Any(), "KV_jobs", "$KV.jobs.jobs.old").
						Return(nil),
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(3), nil),
				)
				s.sent()
			},
		},
		{
			name: "a store of jobs that are not finished is ErrStoreFull",
			setupMocks: func() {
				s.nc.EXPECT().KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
					Return(uint64(0), errors.New("nats: maximum bytes exceeded")).Times(2)
				s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return([]string{"jobs.pending"}, nil)
			},
			wantErr: "no finished job to remove",
		},
		{
			name: "a list that fails is ErrStoreFull",
			setupMocks: func() {
				s.nc.EXPECT().
					KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
					Return(uint64(0), errFull).
					Times(2)
				s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return(nil, errors.New("keys down"))
			},
			wantErr: "list the jobs to remove: keys down",
		},
		{
			name: "a purge that fails is ErrStoreFull",
			setupMocks: func() {
				s.nc.EXPECT().
					KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
					Return(uint64(0), errFull).
					Times(2)
				s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return(bucketKeys, nil)
				s.purger.EXPECT().PurgeSubject(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("purge refused"))
			},
			wantErr: "remove job old: purge refused",
		},
		{
			name: "a store still full after the removal is ErrStoreFull",
			setupMocks: func() {
				s.nc.EXPECT().
					KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
					Return(uint64(0), errFull).
					Times(3)
				s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return(bucketKeys, nil)
				s.purger.EXPECT().
					PurgeSubject(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					Times(3)
			},
			wantErr: "maximum bytes exceeded",
		},
		{
			name: "a write that fails for another reason after the removal is that error",
			setupMocks: func() {
				gomock.InOrder(
					s.nc.EXPECT().
						KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(0), errFull).
						Times(2),
					s.nc.EXPECT().KVKeys(gomock.Any(), "jobs").Return(bucketKeys, nil),
					s.purger.EXPECT().
						PurgeSubject(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(3),
					s.nc.EXPECT().KVPut(gomock.Any(), "jobs", jobKey, gomock.Any()).
						Return(uint64(0), errors.New("nats down")),
				)
			},
			wantErr: "nats down",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var purger client.StorePurger = s.purger
			if tt.noPurger {
				purger = nil
			}
			c := s.newClient(purger)
			tt.setupMocks()

			_, err := c.CreateJob(context.Background(), types.JobTypeQuery, map[string]interface{}{
				"type": "node.hostname.get",
				"data": map[string]interface{}{},
			}, "web-01")

			if tt.wantErr == "" {
				s.NoError(err)

				return
			}
			s.ErrorContains(err, tt.wantErr)
			if tt.wantErr != "nats down" {
				s.ErrorIs(err, job.ErrStoreFull)
			} else {
				s.NotErrorIs(err, job.ErrStoreFull)
			}
		})
	}
}
