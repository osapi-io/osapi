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

package admission_test

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
)

type WatcherPublicTestSuite struct {
	suite.Suite

	ctx       context.Context
	mockCtrl  *gomock.Controller
	mockNC    *admissionMocks.MockNATSSubscriber
	mockKV    *admissionMocks.MockKVStore
	mockPKI   *admissionMocks.MockPKIProvider
	watcher   *admission.Watcher
	fixedTime time.Time
	pubKey    ed25519.PublicKey
}

func (s *WatcherPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = admissionMocks.NewMockNATSSubscriber(s.mockCtrl)
	s.mockKV = admissionMocks.NewMockKVStore(s.mockCtrl)
	s.mockPKI = admissionMocks.NewMockPKIProvider(s.mockCtrl)
	s.fixedTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	s.pubKey = make(ed25519.PublicKey, ed25519.PublicKeySize)

	admission.SetNowFn(func() time.Time { return s.fixedTime })

	s.watcher = admission.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV, "enrollment",
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *WatcherPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *WatcherPublicTestSuite) TearDownSubTest() {
	admission.ResetKeyCache(s.watcher)
	admission.ResetMarshalFn()
	admission.ResetUnmarshalFn()
	admission.ResetNowFn()
}

func (s *WatcherPublicTestSuite) TestHandleEnrollmentRequest() {
	tests := []struct {
		name         string
		setupMock    func()
		msg          *nats.Msg
		validateFunc func()
	}{
		{
			name: "stores pending agent in KV",
			setupMock: func() {
				// A new request: no pending record yet.
				// The machine is not admitted with a different key.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
					DoAndReturn(func(
						_ context.Context, _ string,
						_ string,
						value []byte,
					) (uint64, error) {
						var stored admission.PendingAgent
						err := json.Unmarshal(value, &stored)
						s.Require().NoError(err, "stored value must be valid JSON")
						s.Equal("machine-001", stored.MachineID)
						s.Equal("web-01", stored.Hostname)
						s.Equal(
							testFingerprint,
							stored.Fingerprint,
							"the fingerprint of the keys, not the one the request reports",
						)
						s.Equal(testEncryptionKey, stored.EncryptionKey)
						s.False(
							stored.RequestedAt.IsZero(),
							"requested_at must be non-zero",
						)

						return uint64(1), nil
					})
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			name:      "logs warning on unmarshal failure",
			setupMock: func() {},
			msg:       &nats.Msg{Data: []byte("invalid json")},
		},
		{
			// An agent older than the controller sends no encryption key.
			// Nothing is stored, so it cannot be admitted without one.
			name:      "drops a request with no encryption key",
			setupMock: func() {},
			msg: func() *nats.Msg {
				data, err := json.Marshal(pki.EnrollmentRequest{
					MachineID: "machine-001",
					Hostname:  "web-01",
					PublicKey: []byte("test-pubkey"),
				})
				s.Require().NoError(err)
				return &nats.Msg{Data: data}
			}(),
		},
		{
			name:      "drops a request whose encryption key is not a curve key",
			setupMock: func() {},
			msg: func() *nats.Msg {
				data, err := json.Marshal(pki.EnrollmentRequest{
					MachineID:     "machine-001",
					Hostname:      "web-01",
					PublicKey:     []byte("test-pubkey"),
					EncryptionKey: "UABC",
				})
				s.Require().NoError(err)
				return &nats.Msg{Data: data}
			}(),
		},
		{
			name: "a request with a different key for an admitted machine is pending for the operator",
			setupMock: func() {
				other := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				other.EXPECT().Value().Return(func() []byte {
					b, _ := json.Marshal(admission.AcceptedAgent{
						MachineID: "machine-001",
						PublicKey: []byte("victim-pubkey"),
						Fingerprint: pki.FingerprintOfKeys(
							[]byte("victim-pubkey"),
							testEncryptionKey,
						),
					})

					return b
				}())
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(other, nil)
				// It is pending, so that an operator compares its fingerprint
				// with the host and admits it.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
					Return(uint64(1), nil)
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			name: "refuses a request when the accepted record cannot be read",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, errors.New("kv down"))
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			name: "logs warning on marshal failure",
			setupMock: func() {
				// The machine is not admitted with a different key.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				admission.SetMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			name: "logs warning on KV put failure",
			setupMock: func() {
				// A new request: no pending record yet.
				// The machine is not admitted with a different key.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
					Return(uint64(0), errors.New("kv error"))
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			admission.ExportHandleEnrollmentRequest(s.ctx, s.watcher, tc.msg)
		})
	}
}

func (s *WatcherPublicTestSuite) TestHandleEnrollmentRequestAutoAdmit() {
	autoWatcher := admission.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV, "enrollment",
		s.mockPKI,
		true,
		"osapi",
	)

	tests := []struct {
		name      string
		setupMock func()
		msg       *nats.Msg
	}{
		{
			name: "auto-admits on successful store",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				// Store in KV.
				// A new request: no pending record yet.
				// The machine is not admitted with a different key.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
					Return(uint64(1), nil)

				// AcceptAgent: Get from KV.
				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				// AcceptAgent: Publish response — verify the acceptance payload.
				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					DoAndReturn(func(_ context.Context,
						_ string,
						data []byte,
					) error {
						var resp struct {
							Accepted            bool   `json:"accepted"`
							ControllerPublicKey []byte `json:"controller_public_key"`
						}
						err := json.Unmarshal(data, &resp)
						s.Require().NoError(err, "response must be valid JSON")
						s.True(resp.Accepted, "response must indicate acceptance")
						s.NotEmpty(
							resp.ControllerPublicKey,
							"response must include controller public key",
						)

						return nil
					})

				// AcceptAgent: Delete from KV.
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(1), nil)
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			// A request carries no proof that it comes from the machine, so
			// auto_admit does not let it replace an admitted key.
			name: "auto_admit does not admit a different key for an admitted machine",
			setupMock: func() {
				other := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				other.EXPECT().Value().Return(func() []byte {
					b, _ := json.Marshal(admission.AcceptedAgent{
						MachineID: "machine-001",
						PublicKey: []byte("victim-pubkey"),
						Fingerprint: pki.FingerprintOfKeys(
							[]byte("victim-pubkey"),
							testEncryptionKey,
						),
					})

					return b
				}())
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(other, nil)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
					Return(uint64(1), nil)
				// No response is published and the accepted record is not
				// written: the mocks fail on a call that is not expected.
			},
			msg: s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"),
		},
		{
			name: "logs warning when auto-admit fails",
			setupMock: func() {
				// The machine is not admitted with a different key.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-002").
					Return(nil, jetstream.ErrKeyNotFound)
				// Store in KV.
				// A new request: no pending record yet.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-002").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "enrollment.machine-002", gomock.Any()).
					Return(uint64(1), nil)

				// AcceptAgent: Get from KV fails.
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-002").
					Return(nil, errors.New("kv error"))
			},
			msg: s.makeEnrollmentMsg("machine-002", "web-02", "SHA256:def456"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			log := &joinLog{}
			autoWatcher.SetAudit(log)
			admission.ExportHandleEnrollmentRequest(s.ctx, autoWatcher, tc.msg)
			if tc.name == "auto-admits on successful store" {
				s.Require().Len(log.entries, 2, "the request and the admission are audited")
				s.Equal(audit.ActorAgent, log.entries[0].Actor)
				s.Equal("agent.enroll.request", log.entries[0].Action)
				e := log.entries[1]
				s.Equal(audit.ActorSystem, e.Actor)
				s.Equal("auto_admit", e.User)
				s.Equal("agent.admit", e.Action)
				s.Equal("machine-001", e.Target)
			}
		})
	}
}

// TestEnrollmentRequestRepeats: a pending agent asks again at each
// heartbeat. Only a new request, or one with a new key, is audited
// (osapi-io/osapi#1453).
func (s *WatcherPublicTestSuite) TestEnrollmentRequestRepeats() {
	stored := func(fp string) jetstream.KeyValueEntry {
		e := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
		e.EXPECT().Value().Return(s.makePendingJSON("machine-001", "web-01", fp))

		return e
	}
	tests := []struct {
		name    string
		entry   func() jetstream.KeyValueEntry
		audited bool
	}{
		{name: "the same key again is not audited", entry: func() jetstream.KeyValueEntry {
			return stored(testFingerprint)
		}, audited: false},
		{name: "a new key is audited", entry: func() jetstream.KeyValueEntry {
			return stored("SHA256:old")
		}, audited: true},
		{name: "a record that is not JSON is a new request", entry: func() jetstream.KeyValueEntry {
			e := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
			e.EXPECT().Value().Return([]byte("{"))

			return e
		}, audited: true},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			log := &joinLog{}
			s.watcher.SetAudit(log)
			defer s.watcher.SetAudit(nil)
			// The machine is not admitted with a different key.
			s.mockKV.EXPECT().
				KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
				Return(nil, jetstream.ErrKeyNotFound)
			s.mockKV.EXPECT().
				KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
				Return(tc.entry(), nil)
			s.mockKV.EXPECT().
				KVPut(gomock.Any(), "enrollment", "enrollment.machine-001", gomock.Any()).
				Return(uint64(1), nil)
			admission.ExportHandleEnrollmentRequest(
				s.ctx, s.watcher, s.makeEnrollmentMsg("machine-001", "web-01", "SHA256:abc123"))
			if tc.audited {
				s.Require().Len(log.entries, 1)
				s.Equal("agent.enroll.request", log.entries[0].Action)
				s.Equal("machine-001", log.entries[0].User)
			} else {
				s.Empty(log.entries)
			}
		})
	}
}

func (s *WatcherPublicTestSuite) TestAcceptAgent() {
	tests := []struct {
		name         string
		machineID    string
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name:      "accepts pending agent and publishes response",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					DoAndReturn(func(
						_ context.Context, _ string,
						_ string,
						value []byte,
					) (uint64, error) {
						var stored admission.AcceptedAgent
						s.Require().NoError(json.Unmarshal(value, &stored))
						s.Equal("machine-001", stored.MachineID)
						s.Equal("web-01", stored.Hostname)
						s.Equal(
							pki.FingerprintOfKeys([]byte("test-pubkey"), testEncryptionKey),
							stored.Fingerprint,
							"the fingerprint is recomputed over both keys, not taken from the request",
						)
						s.Equal(testEncryptionKey, stored.EncryptionKey)
						s.Equal(s.fixedTime, stored.AcceptedAt)

						return uint64(1), nil
					})
			},
			validateFunc: func(err error) {
				s.Require().NoError(err)
			},
		},
		{
			name:      "returns error when pending agent not found",
			machineID: "missing",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.missing").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "get pending agent missing")
			},
		},
		{
			name:      "returns error on unmarshal failure",
			machineID: "machine-001",
			setupMock: func() {
				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte("bad json"))
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "unmarshal pending agent machine-001")
			},
		},
		{
			name:      "returns error on marshal failure",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(1), nil)

				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)

				// The stored record marshals first; fail the response marshal.
				calls := 0
				admission.SetMarshalFn(func(v any) ([]byte, error) {
					calls++
					if calls == 1 {
						return json.Marshal(v)
					}

					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "marshal acceptance response")
			},
		},
		{
			name:      "returns error on publish failure",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(1), nil)

				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(errors.New("publish error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "publish acceptance for machine-001")
			},
		},
		{
			name:      "returns error on delete failure",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(1), nil)

				s.mockPKI.EXPECT().PublicKey().Return(s.pubKey)
				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(errors.New("delete error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "delete pending agent machine-001")
			},
		},
		{
			name:      "returns error when the key store rejects the record",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(0), errors.New("kv error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "store accepted agent machine-001")
			},
		},
		{
			name:      "returns error when the record fails to marshal",
			machineID: "machine-001",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				admission.SetMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "marshal accepted agent machine-001")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			tc.validateFunc(s.watcher.AcceptAgent(s.ctx, tc.machineID))
		})
	}
}

func (s *WatcherPublicTestSuite) TestRejectAgent() {
	tests := []struct {
		name         string
		machineID    string
		reason       string
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name:      "rejects pending agent and publishes response",
			machineID: "machine-001",
			reason:    "not authorized",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, data []byte) error {
						var resp pki.EnrollmentResponse
						s.Require().NoError(json.Unmarshal(data, &resp))
						s.False(resp.Accepted)
						s.Equal("not authorized", resp.Reason)
						return nil
					})

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil)
			},
			validateFunc: func(err error) {
				s.Require().NoError(err)
			},
		},
		{
			name:      "returns error when pending agent not found",
			machineID: "missing",
			reason:    "denied",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.missing").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "get pending agent missing")
			},
		},
		{
			name:      "returns error on unmarshal failure",
			machineID: "machine-001",
			reason:    "denied",
			setupMock: func() {
				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte("bad"))
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "unmarshal pending agent machine-001")
			},
		},
		{
			name:      "returns error on marshal failure",
			machineID: "machine-001",
			reason:    "denied",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				admission.SetMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "marshal rejection response")
			},
		},
		{
			name:      "returns error on publish failure",
			machineID: "machine-001",
			reason:    "denied",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(errors.New("publish error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "publish rejection for machine-001")
			},
		},
		{
			name:      "returns error on delete failure",
			machineID: "machine-001",
			reason:    "denied",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(errors.New("delete error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "delete pending agent machine-001")
			},
		},
		{
			name:      "returns error when removing the stored key fails",
			machineID: "machine-001",
			reason:    "denied",
			setupMock: func() {
				pendingData := s.makePendingJSON("machine-001", "web-01", "SHA256:abc123")

				mockEntry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(pendingData)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(mockEntry, nil)

				s.mockNC.EXPECT().
					PublishCore(gomock.Any(), "osapi.enroll.response.machine-001", gomock.Any()).
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "enrollment.machine-001").
					Return(nil)

				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(errors.New("delete error"))
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "remove accepted agent machine-001")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			tc.validateFunc(s.watcher.RejectAgent(s.ctx, tc.machineID, tc.reason))
		})
	}
}

func (s *WatcherPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		setupMock    func()
		wantErr      bool
		validateFunc func(error)
	}{
		{
			name: "subscribes and blocks until context cancelled",
			setupMock: func() {
				s.mockNC.EXPECT().
					Subscribe(gomock.Any(), "osapi.enroll.request", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, cb natsclient.MsgHandler) (*nats.Subscription, error) {
						// Invoke the callback to cover the closure body inside Start.
						// The msg contains invalid JSON, so handleEnrollmentRequest
						// logs a warning and returns without side effects.
						cb(context.Background(), &nats.Msg{Data: []byte("invalid")})
						return &nats.Subscription{}, nil
					})
			},
			validateFunc: func(err error) {
				s.Require().NoError(err)
			},
		},
		{
			name: "returns error on subscribe failure",
			setupMock: func() {
				s.mockNC.EXPECT().
					Subscribe(gomock.Any(), "osapi.enroll.request", gomock.Any()).
					Return(nil, errors.New("subscribe error"))
			},
			wantErr: true,
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Contains(err.Error(), "subscribe to enrollment requests")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()

			if tc.wantErr {
				tc.validateFunc(s.watcher.Start(s.ctx))
			} else {
				ctx, cancel := context.WithCancel(s.ctx)
				errCh := make(chan error, 1)

				go func() {
					errCh <- s.watcher.Start(ctx)
				}()

				cancel()

				tc.validateFunc(<-errCh)
			}
		})
	}
}

func (s *WatcherPublicTestSuite) TestEnrollSubject() {
	tests := []struct {
		name         string
		namespace    string
		suffix       string
		validateFunc func(string)
	}{
		{
			name:      "with namespace",
			namespace: "osapi",
			suffix:    "enroll.request",
			validateFunc: func(got string) {
				s.Equal("osapi.enroll.request", got)
			},
		},
		{
			name:      "without namespace",
			namespace: "",
			suffix:    "enroll.request",
			validateFunc: func(got string) {
				s.Equal("enroll.request", got)
			},
		},
		{
			name:      "response subject with namespace",
			namespace: "osapi",
			suffix:    "enroll.response.machine-001",
			validateFunc: func(got string) {
				s.Equal("osapi.enroll.response.machine-001", got)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := admission.EnrollSubject(tc.namespace, tc.suffix)
			tc.validateFunc(got)
		})
	}
}

func (s *WatcherPublicTestSuite) TestKVPrefix() {
	s.Equal("enrollment.", admission.KVPrefix())
	s.Equal("accepted.", admission.AcceptedKVPrefix())
}

// testEncryptionKey is a public P-256 key in pki.EncryptionKeyEncoding.
const testEncryptionKey = "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs"

// testFingerprint is the fingerprint of the keys of makeEnrollmentMsg.
var testFingerprint = pki.FingerprintOfKeys([]byte("test-pubkey"), testEncryptionKey)

// makeEnrollmentMsg creates a *nats.Msg with a serialized EnrollmentRequest.
func (s *WatcherPublicTestSuite) makeEnrollmentMsg(
	machineID string,
	hostname string,
	fingerprint string,
) *nats.Msg {
	s.T().Helper()

	req := pki.EnrollmentRequest{
		MachineID:     machineID,
		Hostname:      hostname,
		PublicKey:     []byte("test-pubkey"),
		EncryptionKey: testEncryptionKey,
		Fingerprint:   fingerprint,
	}

	data, err := json.Marshal(req)
	s.Require().NoError(err)

	return &nats.Msg{Data: data}
}

// makePendingJSON creates serialized PendingAgent JSON.
func (s *WatcherPublicTestSuite) makePendingJSON(
	machineID string,
	hostname string,
	fingerprint string,
) []byte {
	s.T().Helper()

	pending := admission.PendingAgent{
		MachineID:     machineID,
		Hostname:      hostname,
		PublicKey:     []byte("test-pubkey"),
		EncryptionKey: testEncryptionKey,
		Fingerprint:   fingerprint,
		RequestedAt:   s.fixedTime,
	}

	data, err := json.Marshal(pending)
	s.Require().NoError(err)

	return data
}

func TestWatcherPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(WatcherPublicTestSuite))
}
