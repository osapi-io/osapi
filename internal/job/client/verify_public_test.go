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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type VerifyAgentResponsePublicTestSuite struct {
	suite.Suite

	ctx      context.Context
	mockCtrl *gomock.Controller
	store    *jobmocks.MockAgentKeyStore
}

func (s *VerifyAgentResponsePublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.store = jobmocks.NewMockAgentKeyStore(s.mockCtrl)
}

func (s *VerifyAgentResponsePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// signedResponse returns a response envelope signed by signer and naming
// machineID as the sender.
func (s *VerifyAgentResponsePublicTestSuite) signedResponse(
	signer client.PKISigner,
	machineID string,
	payload []byte,
) []byte {
	s.T().Helper()

	wrapped, err := client.ExportWrapInSignedEnvelope(signer, machineID, payload)
	s.Require().NoError(err)

	return wrapped
}

func (s *VerifyAgentResponsePublicTestSuite) TestVerifyAgentResponse() {
	tests := []struct {
		name         string
		setupData    func() []byte
		setupMock    func()
		validateFunc func([]byte, error)
	}{
		{
			name: "accepts a response signed by the agent's stored key",
			setupData: func() []byte {
				signer, pub := newSigner(s.mockCtrl)
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: pub,
					}, nil)

				return s.signedResponse(
					signer,
					"machine-001",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().NoError(err)

				var response job.Response
				s.Require().NoError(json.Unmarshal(payload, &response))
				s.Equal(job.StatusCompleted, response.Status)
				s.Equal("web-01", response.Hostname)
			},
		},
		{
			name: "rejects a response signed by a key that is not the agent's",
			setupData: func() []byte {
				signer, _ := newSigner(s.mockCtrl)

				// The stored record holds a different key entirely: this is
				// the forged response the advisory describes.
				otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: otherPub,
					}, nil)

				return s.signedResponse(
					signer,
					"machine-001",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseSignatureInvalid)
			},
		},
		{
			name: "rejects an accepted agent answering for another host",
			setupData: func() []byte {
				signer, pub := newSigner(s.mockCtrl)

				// Signature verifies, but this agent enrolled as evil-01
				// and is claiming to be web-01.
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-evil").
					Return(&client.AgentKey{
						MachineID: "machine-evil",
						Hostname:  "evil-01",
						PublicKey: pub,
					}, nil)

				return s.signedResponse(
					signer,
					"machine-evil",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseHostnameMismatch)
				s.Contains(err.Error(), "evil-01")
				s.Contains(err.Error(), "web-01")
			},
		},
		{
			name: "reports an agent with no stored key distinctly",
			setupData: func() []byte {
				signer, _ := newSigner(s.mockCtrl)
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseKeyUnknown)

				return s.signedResponse(
					signer,
					"machine-001",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseKeyUnknown)
				s.NotErrorIs(err, client.ErrResponseSignatureInvalid)
			},
		},
		{
			name: "reports an unreadable store distinctly",
			setupData: func() []byte {
				signer, _ := newSigner(s.mockCtrl)
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseStoreUnavailable)

				return s.signedResponse(
					signer,
					"machine-001",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseStoreUnavailable)
				s.NotErrorIs(err, client.ErrResponseKeyUnknown)
			},
		},
		{
			name: "rejects data that is not JSON",
			setupData: func() []byte {
				return []byte(`not json at all`)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseNotSigned)
			},
		},
		{
			name: "rejects an unsigned response",
			setupData: func() []byte {
				return []byte(`{"status":"completed","hostname":"web-01"}`)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseNotSigned)
			},
		},
		{
			name: "rejects an envelope that names no machine",
			setupData: func() []byte {
				signer, _ := newSigner(s.mockCtrl)

				// An agent that has not been upgraded signs without
				// stamping its identity: there is nothing to look up.
				return s.signedResponse(
					signer,
					"",
					[]byte(`{"status":"completed","hostname":"web-01"}`),
				)
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseNotSigned)
			},
		},
		{
			name: "rejects a verified payload that is not a response",
			setupData: func() []byte {
				signer, pub := newSigner(s.mockCtrl)
				s.store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: pub,
					}, nil)

				return s.signedResponse(signer, "machine-001", []byte(`[1,2,3]`))
			},
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseNotSigned)
				s.Contains(err.Error(), "not a response")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			data := tt.setupData()
			if tt.setupMock != nil {
				tt.setupMock()
			}

			tt.validateFunc(
				client.ExportVerifyAgentResponse(s.ctx, s.store, data),
			)
		})
	}
}

func TestVerifyAgentResponsePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(VerifyAgentResponsePublicTestSuite))
}

// TestAcceptableKeys covers the rotation window itself: which keys a message
// may be signed with, and when the replaced one stops counting.
func (s *VerifyAgentResponsePublicTestSuite) TestAcceptableKeys() {
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	current, _, _ := ed25519.GenerateKey(rand.Reader)
	old, _, _ := ed25519.GenerateKey(rand.Reader)

	tests := []struct {
		name         string
		key          *client.AgentKey
		validateFunc func([]ed25519.PublicKey)
	}{
		{
			name: "outside a rotation only the current key counts",
			key:  &client.AgentKey{PublicKey: current},
			validateFunc: func(keys []ed25519.PublicKey) {
				s.Len(keys, 1)
				s.Equal(current, keys[0])
			},
		},
		{
			name: "inside the window both keys count",
			key: &client.AgentKey{
				PublicKey:       current,
				SupersededKey:   old,
				SupersededUntil: now.Add(time.Hour),
			},
			validateFunc: func(keys []ed25519.PublicKey) {
				s.Len(keys, 2)
				s.Equal(current, keys[0])
				s.Equal(old, keys[1])
			},
		},
		{
			name: "once the window lapses the replaced key is gone",
			key: &client.AgentKey{
				PublicKey:       current,
				SupersededKey:   old,
				SupersededUntil: now.Add(-time.Second),
			},
			validateFunc: func(keys []ed25519.PublicKey) {
				s.Len(keys, 1)
				s.Equal(current, keys[0])
			},
		},
		{
			name: "a removed record offers no keys",
			key:  nil,
			validateFunc: func(keys []ed25519.PublicKey) {
				s.Empty(keys)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tt.key.AcceptableKeys(now))
		})
	}
}

// TestVerifyAgentResponseDuringRotation proves the window is honoured end to
// end, and that an expired key reads as a mismatch rather than as its own
// outcome.
func (s *VerifyAgentResponsePublicTestSuite) TestVerifyAgentResponseDuringRotation() {
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		until        time.Time
		validateFunc func([]byte, error)
	}{
		{
			name:  "a response signed with the replaced key verifies inside the window",
			until: now.Add(time.Hour),
			validateFunc: func(payload []byte, err error) {
				s.Require().NoError(err)
				s.NotEmpty(payload)
			},
		},
		{
			name:  "the same response is a mismatch once the window lapses",
			until: now.Add(-time.Second),
			validateFunc: func(payload []byte, err error) {
				s.Require().Error(err)
				s.Nil(payload)
				s.Require().ErrorIs(err, client.ErrResponseSignatureInvalid)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			client.SetNowFn(func() time.Time { return now })
			defer client.ResetNowFn()

			// The agent still signs with the key the rotation replaced.
			oldSigner, oldPub := newSigner(s.mockCtrl)
			currentPub, _, _ := ed25519.GenerateKey(rand.Reader)

			s.store.EXPECT().
				LookupAgentKey(gomock.Any(), "machine-001").
				Return(&client.AgentKey{
					MachineID:       "machine-001",
					Hostname:        "web-01",
					PublicKey:       currentPub,
					SupersededKey:   oldPub,
					SupersededUntil: tt.until,
				}, nil)

			data := s.signedResponse(
				oldSigner,
				"machine-001",
				[]byte(`{"status":"completed","hostname":"web-01"}`),
			)

			tt.validateFunc(client.ExportVerifyAgentResponse(s.ctx, s.store, data))
		})
	}
}
