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

package enrollment_test

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

	"github.com/osapi-io/osapi/internal/agent/pki"
	"github.com/osapi-io/osapi/internal/controller/enrollment"
	enrollMocks "github.com/osapi-io/osapi/internal/controller/enrollment/mocks"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type KeyStorePublicTestSuite struct {
	suite.Suite

	ctx       context.Context
	mockCtrl  *gomock.Controller
	mockNC    *enrollMocks.MockNATSSubscriber
	mockKV    *jobMocks.MockKeyValue
	mockPKI   *enrollMocks.MockPKIProvider
	watcher   *enrollment.Watcher
	fixedTime time.Time
	pubKey    ed25519.PublicKey
}

func (s *KeyStorePublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = enrollMocks.NewMockNATSSubscriber(s.mockCtrl)
	s.mockKV = jobMocks.NewMockKeyValue(s.mockCtrl)
	s.mockPKI = enrollMocks.NewMockPKIProvider(s.mockCtrl)
	s.fixedTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	s.pubKey = make(ed25519.PublicKey, ed25519.PublicKeySize)

	enrollment.SetNowFn(func() time.Time { return s.fixedTime })

	s.watcher = enrollment.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV,
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
	s.watcher = enrollment.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV,
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *KeyStorePublicTestSuite) TearDownSubTest() {
	enrollment.ResetNowFn()
}

func (s *KeyStorePublicTestSuite) TestLookupAgentKey() {
	tests := []struct {
		name         string
		machineID    string
		setupMock    func()
		validateFunc func(*enrollment.AcceptedAgent, error)
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
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(record *enrollment.AcceptedAgent, err error) {
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
					Get(gomock.Any(), "accepted.machine-404").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(record *enrollment.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, enrollment.ErrAgentKeyNotFound)
				s.NotErrorIs(err, enrollment.ErrAgentKeyStoreUnavailable)
			},
		},
		{
			name:      "reports an unreadable store as unavailable",
			machineID: "machine-001",
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(nil, errors.New("kv error"))
			},
			validateFunc: func(record *enrollment.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, enrollment.ErrAgentKeyStoreUnavailable)
				s.NotErrorIs(err, enrollment.ErrAgentKeyNotFound)
			},
		},
		{
			name:      "reports an undecodable record as unavailable",
			machineID: "machine-001",
			setupMock: func() {
				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return([]byte("bad json"))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(record *enrollment.AcceptedAgent, err error) {
				s.Require().Error(err)
				s.Nil(record)
				s.Require().ErrorIs(err, enrollment.ErrAgentKeyStoreUnavailable)
				s.NotErrorIs(err, enrollment.ErrAgentKeyNotFound)
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

// TestLookupAgentKeyCaches proves the second lookup does not read the store:
// the Get is expected exactly once.
func (s *KeyStorePublicTestSuite) TestLookupAgentKeyCaches() {
	entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		Get(gomock.Any(), "accepted.machine-001").
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
		Get(gomock.Any(), "accepted.machine-001").
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
					Delete(gomock.Any(), "accepted.machine-001").
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
					Delete(gomock.Any(), "accepted.machine-001").
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

// TestRemoveAgentKeyInvalidatesCache proves a removed agent stops verifying
// immediately rather than when a cache entry would have expired: the store is
// read again after the removal.
func (s *KeyStorePublicTestSuite) TestRemoveAgentKeyInvalidatesCache() {
	first := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	first.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		Get(gomock.Any(), "accepted.machine-001").
		Return(first, nil)

	_, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	s.mockKV.EXPECT().
		Delete(gomock.Any(), "accepted.machine-001").
		Return(nil)
	s.Require().NoError(s.watcher.RemoveAgentKey(s.ctx, "machine-001"))

	s.mockKV.EXPECT().
		Get(gomock.Any(), "accepted.machine-001").
		Return(nil, jetstream.ErrKeyNotFound)

	_, err = s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().ErrorIs(err, enrollment.ErrAgentKeyNotFound)
}

// TestRecordAgentKeyInvalidatesCache proves a re-acceptance is visible to the
// next lookup rather than shadowed by the previously cached record.
func (s *KeyStorePublicTestSuite) TestRecordAgentKeyInvalidatesCache() {
	stale := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
	stale.EXPECT().Value().Return(s.makeAcceptedJSON("machine-001", "web-01"))
	s.mockKV.EXPECT().
		Get(gomock.Any(), "accepted.machine-001").
		Return(stale, nil)

	_, err := s.watcher.LookupAgentKey(s.ctx, "machine-001")
	s.Require().NoError(err)

	s.mockKV.EXPECT().
		Put(gomock.Any(), "accepted.machine-001", gomock.Any()).
		Return(uint64(2), nil)
	s.Require().NoError(enrollment.ExportRecordAgentKey(
		s.ctx,
		s.watcher,
		enrollment.PendingAgent{
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
		Get(gomock.Any(), "accepted.machine-001").
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

	record := enrollment.AcceptedAgent{
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
