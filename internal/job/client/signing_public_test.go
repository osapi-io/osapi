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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// newSigner returns a generated PKISigner mock backed by a real ed25519 key
// pair. Sign delegates to the real implementation, so a signature it produces
// verifies against the public key returned alongside it.
func newSigner(
	ctrl *gomock.Controller,
) (*jobmocks.MockPKISigner, ed25519.PublicKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	m := jobmocks.NewMockPKISigner(ctrl)
	m.EXPECT().
		Sign(gomock.Any()).
		DoAndReturn(func(data []byte) []byte {
			return ed25519.Sign(priv, data)
		}).
		AnyTimes()
	m.EXPECT().Fingerprint().Return("SHA256:test-fingerprint").AnyTimes()
	m.EXPECT().ControllerPublicKey().Return(nil).AnyTimes()
	m.EXPECT().StoredRecordKeys().Return(nil).AnyTimes()

	return m, pub
}

// newSignerWithControllerKey is newSigner with the signer's own public key
// reported as the controller's, for paths that verify controller-signed
// messages against it.
func newSignerWithControllerKey(
	ctrl *gomock.Controller,
) (*jobmocks.MockPKISigner, ed25519.PublicKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	m := jobmocks.NewMockPKISigner(ctrl)
	m.EXPECT().
		Sign(gomock.Any()).
		DoAndReturn(func(data []byte) []byte {
			return ed25519.Sign(priv, data)
		}).
		AnyTimes()
	m.EXPECT().Fingerprint().Return("SHA256:test-fingerprint").AnyTimes()
	m.EXPECT().ControllerPublicKey().Return(pub).AnyTimes()
	m.EXPECT().StoredRecordKeys().Return([]ed25519.PublicKey{pub}).AnyTimes()

	return m, pub
}

type SigningPublicTestSuite struct {
	suite.Suite

	ctx      context.Context
	mockCtrl *gomock.Controller
	store    *jobmocks.MockAgentKeyStore
}

func (s *SigningPublicTestSuite) TearDownSubTest() {
	client.ResetSigningMarshalFn()
}

func (s *SigningPublicTestSuite) TestWrapInSignedEnvelope() {
	tests := []struct {
		name         string
		payload      []byte
		setupFn      func()
		validateFunc func([]byte, error, ed25519.PublicKey)
	}{
		{
			name:    "when wrapping valid payload",
			payload: []byte(`{"id":"test-job","operation":{"type":"node.hostname.get"}}`),
			validateFunc: func(result []byte, err error, pubKey ed25519.PublicKey) {
				s.NoError(err)

				payload := []byte(`{"id":"test-job","operation":{"type":"node.hostname.get"}}`)

				var envelope job.SignedEnvelope
				s.NoError(json.Unmarshal(result, &envelope))
				s.Equal(payload, envelope.Payload)
				s.NotEmpty(envelope.Signature)
				s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)

				s.True(ed25519.Verify(pubKey, payload, envelope.Signature))
			},
		},
		{
			name:    "when wrapping empty payload",
			payload: []byte{},
			validateFunc: func(result []byte, err error, pubKey ed25519.PublicKey) {
				s.NoError(err)

				payload := []byte{}

				var envelope job.SignedEnvelope
				s.NoError(json.Unmarshal(result, &envelope))
				s.Equal(payload, envelope.Payload)
				s.NotEmpty(envelope.Signature)
				s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)

				s.True(ed25519.Verify(pubKey, payload, envelope.Signature))
			},
		},
		{
			name:    "when marshal fails returns error",
			payload: []byte(`{"id":"test"}`),
			setupFn: func() {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(_ []byte, err error, _ ed25519.PublicKey) {
				s.Error(err)
				s.Contains(err.Error(), "marshal signed envelope")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupFn != nil {
				tt.setupFn()
			}

			signer, pubKey := newSigner(gomock.NewController(s.T()))

			result, err := client.ExportWrapInSignedEnvelope(signer, "", tt.payload)

			tt.validateFunc(result, err, pubKey)
		})
	}
}

func (s *SigningPublicTestSuite) TestUnwrapSignedEnvelope() {
	signer, pubKey := newSigner(gomock.NewController(s.T()))
	wrapped, _ := client.ExportWrapInSignedEnvelope(signer, "", []byte(`{"id":"test"}`))
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)

	tests := []struct {
		name    string
		data    []byte
		pubKey  ed25519.PublicKey
		want    []byte
		wantErr error
		errText string
	}{
		{
			name:   "when the envelope verifies with the key, it is the payload",
			data:   wrapped,
			pubKey: pubKey,
			want:   []byte(`{"id":"test"}`),
		},
		{
			name:    "when there is no key, it is refused",
			data:    wrapped,
			wantErr: client.ErrNoSigner,
		},
		{
			name:    "when the key is a different one, it is refused",
			data:    wrapped,
			pubKey:  otherPub,
			errText: "invalid signature",
		},
		{
			name:    "when the value is plain JSON, it is refused",
			data:    []byte(`{"id":"test","operation":{"type":"node.hostname.get"}}`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
		{
			name:    "when the value is not JSON, it is refused",
			data:    []byte(`not json at all`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
		{
			name:    "when the envelope has empty fields, it is refused",
			data:    []byte(`{"payload":"","signature":"","fingerprint":""}`),
			pubKey:  pubKey,
			wantErr: client.ErrNotSigned,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			payload, err := client.ExportUnwrapSignedEnvelope(tt.data, tt.pubKey)

			switch {
			case tt.wantErr != nil:
				s.ErrorIs(err, tt.wantErr)
				s.Nil(payload)
			case tt.errText != "":
				s.ErrorContains(err, tt.errText)
				s.Nil(payload)
			default:
				s.Require().NoError(err)
				s.Equal(tt.want, payload)
			}
		})
	}
}

func (s *SigningPublicTestSuite) TestRoundTrip() {
	signer, pubKey := newSigner(gomock.NewController(s.T()))
	originalPayload := []byte(`{"id":"round-trip-test","status":"unprocessed"}`)

	wrapped, err := client.ExportWrapInSignedEnvelope(signer, "", originalPayload)
	s.NoError(err)

	unwrapped, err := client.ExportUnwrapSignedEnvelope(wrapped, pubKey)
	s.NoError(err)
	s.Equal(originalPayload, unwrapped)
}

func TestSigningPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SigningPublicTestSuite))
}

func (s *SigningPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.store = jobmocks.NewMockAgentKeyStore(s.mockCtrl)
}

func (s *SigningPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// signedResponse returns a response envelope signed by signer and naming
// machineID as the sender.
func (s *SigningPublicTestSuite) signedResponse(
	signer client.PKISigner,
	machineID string,
	payload []byte,
) []byte {
	s.T().Helper()

	wrapped, err := client.ExportWrapInSignedEnvelope(signer, machineID, payload)
	s.Require().NoError(err)

	return wrapped
}

func (s *SigningPublicTestSuite) TestVerifyAgentResponse() {
	tests := []struct {
		name         string
		setupData    func() []byte
		setupMock    func()
		validateFunc func(*job.Response, error)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(response)
				s.Equal(types.AgentJobStatusCompleted, response.Status)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
				s.Require().ErrorIs(err, client.ErrResponseSignatureInvalid)
			},
		},
		{
			name: "an answer shows under the record's hostname, not the one it claims",
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().NoError(err, "the machine's key signed it: the answer is its own")
				s.Equal("evil-01", response.Hostname, "it cannot put another host's name on it")
				s.Equal("machine-evil", response.MachineID)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
				s.Require().ErrorIs(err, client.ErrResponseStoreUnavailable)
				s.NotErrorIs(err, client.ErrResponseKeyUnknown)
			},
		},
		{
			name: "rejects data that is not JSON",
			setupData: func() []byte {
				return []byte(`not json at all`)
			},
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
				s.Require().ErrorIs(err, client.ErrResponseNotSigned)
			},
		},
		{
			name: "rejects an unsigned response",
			setupData: func() []byte {
				return []byte(`{"status":"completed","hostname":"web-01"}`)
			},
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
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
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
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

// TestAcceptableKeys covers the rotation window itself: which keys a message
// may be signed with, and when the replaced one stops counting.
func (s *SigningPublicTestSuite) TestAcceptableKeys() {
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
func (s *SigningPublicTestSuite) TestVerifyAgentResponseDuringRotation() {
	now := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		until        time.Time
		validateFunc func(*job.Response, error)
	}{
		{
			name:  "a response signed with the replaced key verifies inside the window",
			until: now.Add(time.Hour),
			validateFunc: func(response *job.Response, err error) {
				s.Require().NoError(err)
				s.NotEmpty(response)
			},
		},
		{
			name:  "the same response is a mismatch once the window lapses",
			until: now.Add(-time.Second),
			validateFunc: func(response *job.Response, err error) {
				s.Require().Error(err)
				s.Nil(response)
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

// TestFollowRename records the new name of a host that a hostname update
// renamed, so its later answers verify; any other answer changes nothing.
func (s *SigningPublicTestSuite) TestFollowRename() {
	update := job.OperationType(job.CategoryBehavior + ".hostname.update")
	renamed, err := json.Marshal(map[string]any{"name": "db-01", "changed": true})
	s.Require().NoError(err)
	tests := []struct {
		name      string
		req       *job.Request
		resp      *job.Response
		setupMock func(*jobmocks.MockAgentKeyStore)
	}{
		{
			name: "when a hostname update completes with a new name, the record takes it",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed,
				Hostname: "web-01", MachineID: "machine-001",
			},
			setupMock: func(m *jobmocks.MockAgentKeyStore) {
				m.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "db-01").Return(nil)
			},
		},
		{
			name: "when the store refuses the rename, the answer still stands",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed,
				Hostname: "web-01", MachineID: "machine-001",
			},
			setupMock: func(m *jobmocks.MockAgentKeyStore) {
				m.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "db-01").
					Return(errors.New("kv error"))
			},
		},
		{
			name: "when the job is another operation, nothing changes",
			req: &job.Request{
				Operation: job.OperationType(job.CategoryBehavior + ".hostname.get"),
			},
			resp: &job.Response{Status: types.AgentJobStatusCompleted, Data: renamed},
		},
		{
			name: "when the update failed, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{Status: types.AgentJobStatusFailed, Data: renamed},
		},
		{
			name: "when the answer has no name, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{Status: types.AgentJobStatusCompleted, Data: json.RawMessage(`{}`)},
		},
		{
			name: "when the name is the one it has, nothing changes",
			req:  &job.Request{Operation: update},
			resp: &job.Response{
				Status: types.AgentJobStatusCompleted, Data: renamed, Hostname: "db-01",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			if tt.setupMock != nil {
				tt.setupMock(store)
			}
			c, err := newTrusted(jobmocks.NewMockNATSClient(s.mockCtrl), &client.Options{
				Timeout: time.Second, JobsBucket: "jobs", StreamName: "JOBS",
			})
			s.Require().NoError(err)
			c.SetAgentKeyStore(store)

			client.ExportFollowRename(context.Background(), c, tt.req, tt.resp)
		})
	}
}
