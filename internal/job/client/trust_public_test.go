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
	mockKV   *jobmocks.MockKeyValue
	pub      ed25519.PublicKey
	priv     ed25519.PrivateKey
}

func (s *RegistrationTrustPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.pub, s.priv = pub, priv
}

func (s *RegistrationTrustPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// registry returns a registry KV holding exactly this registration.
func (s *RegistrationTrustPublicTestSuite) registry(
	reg job.AgentRegistration,
) *jobmocks.MockKeyValue {
	s.T().Helper()

	data, err := json.Marshal(reg)
	s.Require().NoError(err)

	kv := jobmocks.NewMockKeyValue(s.mockCtrl)
	kv.EXPECT().Keys(gomock.Any()).Return([]string{"agents." + reg.MachineID}, nil)

	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data)
	kv.EXPECT().
		Get(gomock.Any(), "agents."+reg.MachineID).
		Return(entry, nil)

	return kv
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
			name: "an accepted agent may not answer for another host",
			reg: func() job.AgentRegistration {
				reg := s.base()
				reg.Hostname = "web-01"

				return s.signed(reg)
			},
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				// Signature is valid, but this machine enrolled as evil-01.
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "evil-01",
						PublicKey: s.pub,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.True(a.KeyStored)
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
			name:    "with no store wired every registration is trusted as before",
			reg:     func() job.AgentRegistration { return s.base() },
			noStore: true,
			validateFunc: func(a job.AgentInfo) {
				s.True(a.Verified)
				s.False(a.KeyStored)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reg := tt.reg()
			registryKV := s.registry(reg)

			c, err := client.New(slog.Default(), s.mockNC, &client.Options{
				Timeout:    30 * time.Second,
				KVBucket:   s.mockKV,
				RegistryKV: registryKV,
				StreamName: "JOBS",
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

func TestRegistrationTrustPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RegistrationTrustPublicTestSuite))
}
