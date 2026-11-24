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
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

type createInput struct {
	Name string                  `json:"name"`
	Env  map[string]secret.Value `json:"env,omitempty"`
}

type SealPublicTestSuite struct {
	suite.Suite

	ctx      context.Context
	mockCtrl *gomock.Controller
	mockNC   *jobmocks.MockNATSClient
	web      nkeys.KeyPair
	db       nkeys.KeyPair
}

func (s *SealPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)

	var err error
	s.web, err = nkeys.CreateCurveKeys()
	s.Require().NoError(err)
	s.db, err = nkeys.CreateCurveKeys()
	s.Require().NoError(err)
}

func (s *SealPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// registry makes the registry hold an agent for each registration.
func (s *SealPublicTestSuite) registry(
	regs ...job.AgentRegistration,
) {
	keys := make([]string, 0, len(regs))
	for _, reg := range regs {
		keys = append(keys, "agents."+reg.MachineID)
		data, _ := json.Marshal(reg)
		entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
		entry.EXPECT().Value().Return(data)
		s.mockNC.EXPECT().
			KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
			Return(entry, nil)
	}
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(keys, nil)
}

// store gives each machine ID its encryption key, and no record for any
// other.
func (s *SealPublicTestSuite) store(
	keys map[string]string,
) *jobmocks.MockAgentKeyStore {
	st := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	st.EXPECT().LookupAgentKey(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, id string) (*client.AgentKey, error) {
			key, ok := keys[id]
			if !ok {
				return nil, client.ErrResponseKeyUnknown
			}

			return &client.AgentKey{MachineID: id, EncryptionKey: key}, nil
		}).AnyTimes()

	return st
}

func (s *SealPublicTestSuite) newClient(
	store client.AgentKeyStore,
) *client.Client {
	c, err := client.New(slog.Default(), s.mockNC, &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		RegistryBucket: "registry",
		StreamName:     "JOBS",
	})
	s.Require().NoError(err)
	if store != nil {
		c.SetAgentKeyStore(store)
	}

	return c
}

func (s *SealPublicTestSuite) pub(
	kp nkeys.KeyPair,
) string {
	pub, err := kp.PublicKey()
	s.Require().NoError(err)

	return pub
}

// opened is the env of req's data, opened by own, the key of the agent id.
func (s *SealPublicTestSuite) opened(
	req *job.Request,
	id string,
	own nkeys.KeyPair,
) map[string]secret.Value {
	o, err := seal.Open(req.JobID, req.SealedKeys[id], req.SealedFrom, own)
	s.Require().NoError(err)
	var in createInput
	s.Require().NoError(json.Unmarshal(req.Data, &in))
	s.Require().NoError(secret.Map(&in, o.Value))

	return in.Env
}

func (s *SealPublicTestSuite) TestRequest() {
	web := job.AgentRegistration{
		MachineID: "m-web",
		Hostname:  "web-01",
		Labels:    map[string]string{"group": "web"},
	}
	db := job.AgentRegistration{MachineID: "m-db", Hostname: "db-01"}
	old := job.AgentRegistration{MachineID: "m-old", Hostname: "old-01"}
	input := createInput{Name: "app", Env: map[string]secret.Value{"TOKEN": "hunter2"}}

	tests := []struct {
		name         string
		store        func() client.AgentKeyStore
		setup        func()
		target       string
		resolved     string
		data         any
		wantErr      string
		validateFunc func(req *job.Request)
	}{
		{
			name:     "when the controller has no PKI, the job keeps each value",
			store:    func() client.AgentKeyStore { return nil },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			validateFunc: func(req *job.Request) {
				s.JSONEq(`{"name":"app","env":{"TOKEN":"hunter2"}}`, string(req.Data))
				s.Empty(req.SealedKeys)
				s.NotEmpty(req.JobID)
			},
		},
		{
			name: "when the data has no secret value, nothing is sealed",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			target:   "web-01",
			resolved: "m-web",
			data:     createInput{Name: "app"},
			validateFunc: func(req *job.Request) {
				s.JSONEq(`{"name":"app"}`, string(req.Data))
				s.Empty(req.SealedKeys)
			},
		},
		{
			name: "when the target is _all, the job key is sealed to each agent with a key",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": s.pub(s.web), "m-db": s.pub(s.db)})
			},
			setup:  func() { s.registry(web, db, old) },
			target: "_all",
			data:   input,
			validateFunc: func(req *job.Request) {
				s.NotContains(string(req.Data), "hunter2", "the job never has the value")
				s.Contains(string(req.Data), seal.Prefix)
				s.Len(req.SealedKeys, 2, "an agent with no key gets none")
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-web", s.web))
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-db", s.db))
				op := client.ExportOperation(req)
				s.Equal(req.SealedKeys, op["sealed_keys"])
				s.Equal(req.SealedFrom, op["sealed_from"])
			},
		},
		{
			name: "when the target is one host, the job key is sealed only to it",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": s.pub(s.web), "m-db": s.pub(s.db)})
			},
			setup:    func() { s.registry(web, db) },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			validateFunc: func(req *job.Request) {
				s.Equal([]string{"m-web"}, keysOf(req.SealedKeys))
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-web", s.web))
			},
		},
		{
			name: "when the agents cannot be listed, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			setup: func() {
				s.mockNC.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("nats: timeout"))
			},
			target:  "_all",
			data:    input,
			wantErr: "list the agents to seal the job to",
		},
		{
			name: "when the key of an agent is not an X25519 key, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": "not-a-key"})
			},
			setup:    func() { s.registry(web) },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			wantErr:  "seal the job key",
		},
		{
			name: "when the job key cannot be made, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			setup: func() {
				client.SetNewSealer(func(string) (*seal.Sealer, error) {
					return nil, errors.New("no entropy")
				})
			},
			target:  "web-01",
			data:    input,
			wantErr: "no entropy",
		},
		{
			name:    "when the data does not marshal, it is an error",
			store:   func() client.AgentKeyStore { return nil },
			target:  "web-01",
			data:    map[string]any{"bad": make(chan int)},
			wantErr: "marshal data",
		},
		{
			name: "when sealed data does not marshal, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			target: "web-01",
			data: struct {
				Token secret.Value `json:"token"`
				Bad   chan int     `json:"bad"`
			}{Token: "hunter2", Bad: make(chan int)},
			wantErr: "marshal data",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer client.ResetNewSealer()
			if tc.setup != nil {
				tc.setup()
			}
			c := s.newClient(tc.store())

			req, err := c.ExportRequest(s.ctx, job.TypeModify, tc.target, tc.resolved,
				job.CategoryBehavior, "docker.create", tc.data)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(job.TypeModify, req.Type)
			tc.validateFunc(req)
		})
	}
}

func keysOf(
	m map[string]string,
) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	return out
}

func TestSealPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SealPublicTestSuite))
}
