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
	"github.com/osapi-io/osapi/internal/job/client"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type KeyStoreAdapterPublicTestSuite struct {
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

func (s *KeyStoreAdapterPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = enrollMocks.NewMockNATSSubscriber(s.mockCtrl)
	s.mockKV = jobMocks.NewMockKeyValue(s.mockCtrl)
	s.mockPKI = enrollMocks.NewMockPKIProvider(s.mockCtrl)
	s.fixedTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	s.pubKey = make(ed25519.PublicKey, ed25519.PublicKeySize)

	s.watcher = enrollment.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV,
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *KeyStoreAdapterPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *KeyStoreAdapterPublicTestSuite) SetupSubTest() {
	// A fresh watcher per subtest so a record cached by an earlier case does
	// not satisfy a lookup this case expects to reach the store.
	s.watcher = enrollment.NewWatcher(
		slog.Default(),
		s.mockNC,
		s.mockKV,
		s.mockPKI,
		false,
		"osapi",
	)
}

func (s *KeyStoreAdapterPublicTestSuite) TestLookupAgentKey() {
	tests := []struct {
		name         string
		machineID    string
		setupMock    func()
		validateFunc func(*client.AgentKey, error)
	}{
		{
			name:      "returns the stored key and the enrolled hostname",
			machineID: "machine-001",
			setupMock: func() {
				record := enrollment.AcceptedAgent{
					MachineID:   "machine-001",
					Hostname:    "web-01",
					PublicKey:   s.pubKey,
					Fingerprint: pki.FingerprintOf(s.pubKey),
					AcceptedAt:  s.fixedTime,
				}
				data, err := json.Marshal(record)
				s.Require().NoError(err)

				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(data)
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(key *client.AgentKey, err error) {
				s.Require().NoError(err)
				s.Equal("machine-001", key.MachineID)
				s.Equal("web-01", key.Hostname)
				s.Equal(ed25519.PublicKey(s.pubKey), key.PublicKey)
			},
		},
		{
			name:      "maps a missing record to the client's unknown-key cause",
			machineID: "machine-404",
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-404").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(key *client.AgentKey, err error) {
				s.Require().Error(err)
				s.Nil(key)
				s.Require().ErrorIs(err, client.ErrResponseKeyUnknown)
				s.NotErrorIs(err, client.ErrResponseStoreUnavailable)
			},
		},
		{
			name:      "maps an unreadable store to the client's unavailable cause",
			machineID: "machine-001",
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(nil, errors.New("kv error"))
			},
			validateFunc: func(key *client.AgentKey, err error) {
				s.Require().Error(err)
				s.Nil(key)
				s.Require().ErrorIs(err, client.ErrResponseStoreUnavailable)
				s.NotErrorIs(err, client.ErrResponseKeyUnknown)
			},
		},
		{
			name:      "maps an undecodable record to the unavailable cause",
			machineID: "machine-001",
			setupMock: func() {
				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return([]byte("bad json"))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(key *client.AgentKey, err error) {
				s.Require().Error(err)
				s.Nil(key)
				s.Require().ErrorIs(err, client.ErrResponseStoreUnavailable)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			tt.validateFunc(
				s.watcher.KeyStore().LookupAgentKey(s.ctx, tt.machineID),
			)
		})
	}
}

func TestKeyStoreAdapterPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeyStoreAdapterPublicTestSuite))
}
