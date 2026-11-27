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

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
)

type KeyStorePublicTestSuite struct {
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

func (s *KeyStorePublicTestSuite) SetupTest() {
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

func (s *KeyStorePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *KeyStorePublicTestSuite) SetupSubTest() {
	// A fresh watcher per subtest. The key cache lives on the Watcher, so a
	// record cached by an earlier case would satisfy a lookup that this case
	// expects to reach the store.
	s.watcher = admission.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV, "enrollment",
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *KeyStorePublicTestSuite) TearDownSubTest() {
	admission.ResetNowFn()
}

func (s *KeyStorePublicTestSuite) TestLookupAgentKey() {
	tests := []struct {
		name         string
		machineID    string
		setupMock    func()
		validateFunc func(*admission.AcceptedAgent, error)
	}{
		{
			name:      "returns the stored record",
			machineID: "machine-001",
			setupMock: func() {
				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(
					s.makeAcceptedJSON("machine-001", "web-01"),
				)
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(record *admission.AcceptedAgent, err error) {
				s.Require().NoError(err)
				s.Equal("machine-001", record.MachineID)
				s.Equal("web-01", record.Hostname)
				s.Equal(
					pki.FingerprintOf(s.pubKey),
					record.Fingerprint,
				)
			},
		},
		{
			name:      "reports a missing record as not found",
			machineID: "machine-404",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-404").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(record *admission.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, admission.ErrAgentKeyNotFound)
				s.NotErrorIs(err, admission.ErrAgentKeyStoreUnavailable)
			},
		},
		{
			name:      "reports an unreadable store as unavailable",
			machineID: "machine-001",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, errors.New("kv error"))
			},
			validateFunc: func(record *admission.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, admission.ErrAgentKeyStoreUnavailable)
				s.NotErrorIs(err, admission.ErrAgentKeyNotFound)
			},
		},
		{
			name:      "reports an undecodable record as unavailable",
			machineID: "machine-001",
			setupMock: func() {
				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return([]byte("bad json"))
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(record *admission.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, admission.ErrAgentKeyStoreUnavailable)
				s.NotErrorIs(err, admission.ErrAgentKeyNotFound)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()

			tc.validateFunc(s.watcher.LookupAgentKey(s.ctx, tc.machineID))
		})
	}
}

// TestRenameAgentKey records a new hostname on an accepted record, and
// keeps its machine ID and keys.
func (s *KeyStorePublicTestSuite) TestRenameAgentKey() {
	stored := func() {
		entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
		entry.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
		s.mockKV.EXPECT().
			KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
			Return(entry, nil)
	}
	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "when the record exists, it holds the new hostname",
			setupMock: func() {
				stored()
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
						var r admission.AcceptedAgent
						s.Require().NoError(json.Unmarshal(data, &r))
						s.Equal("db-01", r.Hostname)
						s.Equal("machine-001", r.MachineID)
						s.Equal(ed25519.PublicKey(s.pubKey), ed25519.PublicKey(r.PublicKey))

						return 1, nil
					})
			},
		},
		{
			name: "when there is no record, it is an error",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			wantErr: "machine-001",
		},
		{
			name: "when the record does not encode, it is an error",
			setupMock: func() {
				stored()
				admission.SetMarshalFn(
					func(any) ([]byte, error) { return nil, errors.New("encode") },
				)
			},
			wantErr: "marshal accepted agent machine-001",
		},
		{
			name: "when the store refuses the write, it is an error",
			setupMock: func() {
				stored()
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
					Return(uint64(0), errors.New("kv error"))
			},
			wantErr: "rename accepted agent machine-001",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer admission.ResetMarshalFn()
			tc.setupMock()

			err := s.watcher.RenameAgentKey(s.ctx, "machine-001", "db-01")
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.NoError(err)
		})
	}
}

// TestLookupAgentKeyCaches proves the second lookup does not read the store:
// the Get is expected exactly once.
func (s *KeyStorePublicTestSuite) TestLookupAgentKeyCaches() {
	entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(entry, nil).
		Times(1)

	first, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	second, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)
	s.Equal(first, second)
}

// TestLookupAgentKeyReturnsCopy proves a caller cannot mutate what the next
// lookup returns, since the cached record is shared.
func (s *KeyStorePublicTestSuite) TestLookupAgentKeyReturnsCopy() {
	entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(entry, nil).
		Times(1)

	first, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	first.Hostname = "attacker-01"

	second, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)
	s.Equal("web-01", second.Hostname)
}

func (s *KeyStorePublicTestSuite) TestRemoveAgentKey() {
	tests := []struct {
		name         string
		machineID    string
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name:      "removes the stored record",
			machineID: "machine-001",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "enrollment", "accepted.machine-001").
					Return(nil)
			},
			validateFunc: func(err error) {
				s.Require().NoError(err)
			},
		},
		{
			name:      "returns error when the delete fails",
			machineID: "machine-001",
			setupMock: func() {
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

			tc.validateFunc(s.watcher.RemoveAgentKey(s.ctx, tc.machineID))
		})
	}
}

// TestRemoveAgentKeyRevokesNATSUser: with a revoker, a removal cuts the NATS
// user that the machine joined with before it deletes the record, so a
// revocation that fails leaves the record for the removal to run again.
func (s *KeyStorePublicTestSuite) TestRemoveAgentKeyRevokesNATSUser() {
	const userKey = "UAUSER"
	withUser := func(natsUser string) []byte {
		data, err := json.Marshal(admission.AcceptedAgent{
			MachineID:   "machine-001",
			Hostname:    "web-01",
			PublicKey:   s.pubKey,
			NATSUserKey: natsUser,
		})
		s.Require().NoError(err)

		return data
	}
	lookup := func(value []byte, err error) {
		var e jetstream.KeyValueEntry
		if value != nil {
			m := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
			m.EXPECT().Value().Return(value)
			e = m
		}
		s.mockKV.EXPECT().
			KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
			Return(e, err)
	}
	deleted := func() {
		s.mockKV.EXPECT().
			KVDelete(gomock.Any(), "enrollment", "accepted.machine-001").
			Return(nil)
	}

	tests := []struct {
		name    string
		setup   func(r *admissionMocks.MockRevoker)
		wantErr string
	}{
		{
			name: "revokes the machine's NATS user, then removes the record",
			setup: func(r *admissionMocks.MockRevoker) {
				lookup(withUser(userKey), nil)
				r.EXPECT().Revoke(gomock.Any(), userKey).Return(nil)
				deleted()
			},
		},
		{
			name: "a machine with no NATS user of its own revokes nothing",
			setup: func(*admissionMocks.MockRevoker) {
				lookup(withUser(""), nil)
				deleted()
			},
		},
		{
			name: "a machine that is not admitted revokes nothing",
			setup: func(*admissionMocks.MockRevoker) {
				lookup(nil, jetstream.ErrKeyNotFound)
				deleted()
			},
		},
		{
			name: "a lookup that fails keeps the record",
			setup: func(*admissionMocks.MockRevoker) {
				lookup(nil, errors.New("kv down"))
			},
			wantErr: "kv down",
		},
		{
			name: "a revocation that fails keeps the record",
			setup: func(r *admissionMocks.MockRevoker) {
				lookup(withUser(userKey), nil)
				r.EXPECT().Revoke(gomock.Any(), userKey).Return(errors.New("no record"))
			},
			wantErr: "revoke the NATS user of machine-001: no record",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := admissionMocks.NewMockRevoker(s.mockCtrl)
			s.watcher.SetRevoker(r)
			tc.setup(r)

			err := s.watcher.RemoveAgentKey(s.ctx, "machine-001")
			if tc.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.wantErr)

				return
			}
			s.NoError(err)
		})
	}
}

// TestRemoveAgentKeyInvalidatesCache proves a removed agent stops verifying
// immediately rather than when a cache entry would have expired: the store is
// read again after the removal.
func (s *KeyStorePublicTestSuite) TestRemoveAgentKeyInvalidatesCache() {
	first := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	first.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(first, nil)

	_, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	s.mockKV.EXPECT().
		KVDelete(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(nil)
	s.Require().NoError(s.watcher.RemoveAgentKey(s.ctx, "machine-001"))

	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(nil, jetstream.ErrKeyNotFound)

	_, err = s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().ErrorIs(err, admission.ErrAgentKeyNotFound)
}

// TestRecordAgentKeyInvalidatesCache proves a re-acceptance is visible to the
// next lookup rather than shadowed by the previously cached record.
func (s *KeyStorePublicTestSuite) TestRecordAgentKeyInvalidatesCache() {
	stale := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	stale.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(stale, nil)

	_, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	s.mockKV.EXPECT().
		KVPut(gomock.Any(), "enrollment", "accepted.machine-001", gomock.Any()).
		Return(uint64(2), nil)
	s.Require().NoError(admission.ExportRecordAgentKey(
		s.ctx,
		s.watcher,
		admission.PendingAgent{
			MachineID: "machine-001",
			Hostname:  "web-01-renamed",
			PublicKey: s.pubKey,
		},
	))

	fresh := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	fresh.EXPECT().Value().Return(
		s.makeAcceptedJSON("machine-001", "web-01-renamed"),
	)
	s.mockKV.EXPECT().
		KVGet(gomock.Any(), "enrollment", "accepted.machine-001").
		Return(fresh, nil)

	record, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)
	s.Equal("web-01-renamed", record.Hostname)
}

// makeAcceptedJSON creates serialized AcceptedAgent JSON.
func (s *KeyStorePublicTestSuite) makeAcceptedJSON(
	machineID string,
	hostname string,
) []byte {
	s.T().Helper()

	record := admission.AcceptedAgent{
		MachineID:   machineID,
		Hostname:    hostname,
		PublicKey:   s.pubKey,
		Fingerprint: pki.FingerprintOf(s.pubKey),
		AcceptedAt:  s.fixedTime,
	}

	data, err := json.Marshal(record)
	s.Require().NoError(err)

	return data
}

func TestKeyStorePublicTestSuite(
	t *testing.T,
) {
	// Not parallel: this package's tests swap package-level seams, and the
	// rotation suite already runs in parallel. A second parallel suite races
	// it over those globals.
	suite.Run(t, new(KeyStorePublicTestSuite))
}
