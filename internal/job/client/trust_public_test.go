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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type RegistrationTrustPublicTestSuite struct {
	suite.Suite

	ctx      context.Context
	mockCtrl *gomock.Controller
	mockNC   *jobmocks.MockNATSClient
	pub      ed25519.PublicKey
	priv     ed25519.PrivateKey
}

func (s *RegistrationTrustPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.pub, s.priv = pub, priv
}

func (s *RegistrationTrustPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// registry makes the registry bucket hold exactly this registration.
func (s *RegistrationTrustPublicTestSuite) registry(
	reg job.AgentRegistration,
) {
	s.T().Helper()

	data, err := json.Marshal(reg)
	s.Require().NoError(err)

	s.mockNC.EXPECT().
		KVKeys(gomock.Any(), "registry").
		Return([]string{"agents." + reg.MachineID}, nil)

	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data)
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
		Return(entry, nil)
}

func (s *RegistrationTrustPublicTestSuite) signed(
	reg job.AgentRegistration,
) job.AgentRegistration {
	s.T().Helper()

	reg.Signature = ed25519.Sign(s.priv, job.RegistrationSigningBytes(&reg))

	return reg
}

func (s *RegistrationTrustPublicTestSuite) base() job.AgentRegistration {
	return job.AgentRegistration{
		MachineID:    "machine-001",
		Hostname:     "web-01",
		State:        "Ready",
		RegisteredAt: time.Now(),
	}
}

func (s *RegistrationTrustPublicTestSuite) TestListAgentsTrust() {
	tests := []struct {
		name         string
		reg          func() job.AgentRegistration
		setupStore   func(*jobmocks.MockAgentKeyStore)
		noStore      bool
		validateFunc func(job.AgentInfo)
	}{
		{
			name: "a registration signed by the stored key is authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: s.pub,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.True(a.Verified)
				s.True(a.KeyStored)
			},
		},
		{
			name: "an unsigned registration is not authoritative",
			reg:  func() job.AgentRegistration { return s.base() },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: s.pub,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.True(a.KeyStored, "the key is stored; the registration just is not signed")
			},
		},
		{
			name: "a registration signed by another key is not authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				other, _, _ := ed25519.GenerateKey(rand.Reader)
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: other,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
			},
		},
		{
			name: "an agent with no stored key is visible but not authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseKeyUnknown)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.False(a.KeyStored)
				s.Equal("web-01", a.Hostname, "still listed, so a rollout can be staged")
			},
		},
		{
			name: "an unreadable store never yields an authoritative registration",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseStoreUnavailable)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
			},
		},
		{
			name:    "with no store wired no registration is trusted",
			reg:     func() job.AgentRegistration { return s.base() },
			noStore: true,
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.False(a.KeyStored)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reg := tt.reg()
			s.registry(reg)

			c, err := client.New(slog.Default(), s.mockNC, &client.Options{
				Timeout:        30 * time.Second,
				JobsBucket:     "jobs",
				RegistryBucket: "registry",
				StreamName:     "JOBS",
			})
			s.Require().NoError(err)

			if !tt.noStore {
				store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
				tt.setupStore(store)
				c.SetAgentKeyStore(store)
			}

			agents, err := c.ListAgents(s.ctx)
			s.Require().NoError(err)
			s.Require().Len(agents, 1)

			tt.validateFunc(agents[0])
		})
	}
}

// newClient is a job client of the mocked NATS, with store.
func (s *RegistrationTrustPublicTestSuite) newClient(
	store client.AgentKeyStore,
) *client.Client {
	s.T().Helper()

	c, err := client.New(slog.Default(), s.mockNC, &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		RegistryBucket: "registry",
		StreamName:     "JOBS",
	})
	s.Require().NoError(err)
	c.SetAgentKeyStore(store)

	return c
}

// registryOf makes the registry bucket hold regs, for each read.
func (s *RegistrationTrustPublicTestSuite) registryOf(
	regs ...job.AgentRegistration,
) {
	s.T().Helper()

	keys := make([]string, 0, len(regs))
	for _, reg := range regs {
		keys = append(keys, "agents."+reg.MachineID)
		data, err := json.Marshal(reg)
		s.Require().NoError(err)
		entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
		entry.EXPECT().Value().Return(data).AnyTimes()
		s.mockNC.EXPECT().
			KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
			Return(entry, nil).AnyTimes()
	}
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(keys, nil).AnyTimes()
}

// TestHostnameFollows is a host whose hostname changed: its record follows a
// signed registration, unless another admitted machine that is online holds
// the name (GHSA-j73r).
func (s *RegistrationTrustPublicTestSuite) TestHostnameFollows() {
	renamed := func() job.AgentRegistration {
		reg := s.base()
		reg.Hostname = "web-02"

		return s.signed(reg)
	}
	record := func(machineID, hostname string) *client.AgentKey {
		return &client.AgentKey{MachineID: machineID, Hostname: hostname, PublicKey: s.pub}
	}
	tests := []struct {
		name         string
		regs         func() []job.AgentRegistration
		setupStore   func(*jobmocks.MockAgentKeyStore)
		wantVerified bool
	}{
		{
			name: "when no other machine holds the new name, the record follows it",
			regs: func() []job.AgentRegistration { return []job.AgentRegistration{renamed()} },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").Return(nil)
			},
			wantVerified: true,
		},
		{
			name: "when another admitted machine online holds the name, the claim is refused",
			regs: func() []job.AgentRegistration {
				holder := job.AgentRegistration{
					MachineID: "machine-002",
					Hostname:  "web-02",
					State:     "Ready",
				}

				return []job.AgentRegistration{renamed(), holder}
			},
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-002").
					Return(record("machine-002", "web-02"), nil).AnyTimes()
			},
		},
		{
			name: "when the record cannot be renamed, the claim is refused",
			regs: func() []job.AgentRegistration { return []job.AgentRegistration{renamed()} },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").
					Return(errors.New("kv: down"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mockCtrl = gomock.NewController(s.T())
			s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
			s.registryOf(tc.regs()...)
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			tc.setupStore(store)
			c := s.newClient(store)

			agents, err := c.ListAgents(s.ctx)

			s.Require().NoError(err)
			s.Equal(tc.wantVerified, agents[0].Verified)
		})
	}
}

// TestHostnameRegistryUnreadable is a rename while the registry cannot be
// read for the check: the claim is refused, not taken on a guess.
func (s *RegistrationTrustPublicTestSuite) TestHostnameRegistryUnreadable() {
	reg := s.base()
	reg.Hostname = "web-02"
	data, err := json.Marshal(s.signed(reg))
	s.Require().NoError(err)
	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	gomock.InOrder(
		s.mockNC.EXPECT().
			KVKeys(gomock.Any(), "registry").
			Return([]string{"agents.machine-001"}, nil),
		s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(nil, errors.New("kv: down")),
	)
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents.machine-001").
		Return(entry, nil).
		AnyTimes()
	store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	store.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
		Return(&client.AgentKey{MachineID: "machine-001", Hostname: "web-01", PublicKey: s.pub}, nil).
		AnyTimes()

	agents, err := s.newClient(store).ListAgents(s.ctx)

	s.Require().NoError(err)
	s.False(agents[0].Verified)
}

// TestHostnameEntryUnreadable is a registry entry that cannot be read
// during the check: it is skipped, and the rename goes on.
func (s *RegistrationTrustPublicTestSuite) TestHostnameEntryUnreadable() {
	reg := s.base()
	reg.Hostname = "web-02"
	data, err := json.Marshal(s.signed(reg))
	s.Require().NoError(err)
	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").
		Return([]string{"agents.machine-001", "agents.machine-002"}, nil).AnyTimes()
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents.machine-001").
		Return(entry, nil).
		AnyTimes()
	s.mockNC.EXPECT().KVGet(gomock.Any(), "registry", "agents.machine-002").
		Return(nil, errors.New("kv: gone")).AnyTimes()
	store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	store.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
		Return(&client.AgentKey{MachineID: "machine-001", Hostname: "web-01", PublicKey: s.pub}, nil).
		AnyTimes()
	store.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").Return(nil)

	agents, err := s.newClient(store).ListAgents(s.ctx)

	s.Require().NoError(err)
	s.True(agents[0].Verified)
}

func TestRegistrationTrustPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RegistrationTrustPublicTestSuite))
}
