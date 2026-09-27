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
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/enrollment"
	enrollMocks "github.com/osapi-io/osapi/internal/controller/enrollment/mocks"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type KeyRotationPublicTestSuite struct {
	suite.Suite

	ctx       context.Context
	mockCtrl  *gomock.Controller
	mockNC    *enrollMocks.MockNATSSubscriber
	mockKV    *jobMocks.MockKeyValue
	mockPKI   *enrollMocks.MockPKIProvider
	watcher   *enrollment.Watcher
	fixedTime time.Time
	oldKey    ed25519.PublicKey
	newKey    ed25519.PublicKey
}

func (s *KeyRotationPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNC = enrollMocks.NewMockNATSSubscriber(s.mockCtrl)
	s.mockKV = jobMocks.NewMockKeyValue(s.mockCtrl)
	s.mockPKI = enrollMocks.NewMockPKIProvider(s.mockCtrl)
	s.fixedTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

	oldPub, _, err := ed25519.GenerateKey(nil)
	s.Require().NoError(err)
	newPub, _, err := ed25519.GenerateKey(nil)
	s.Require().NoError(err)
	s.oldKey, s.newKey = oldPub, newPub

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

func (s *KeyRotationPublicTestSuite) TearDownTest() {
	enrollment.ResetNowFn()
	s.mockCtrl.Finish()
}

// existing returns a KV entry holding the record a rotation replaces.
func (s *KeyRotationPublicTestSuite) existing() []byte {
	s.T().Helper()

	data, err := json.Marshal(enrollment.AcceptedAgent{
		MachineID:  "machine-001",
		Hostname:   "web-01",
		PublicKey:  s.oldKey,
		AcceptedAt: s.fixedTime.Add(-time.Hour),
	})
	s.Require().NoError(err)

	return data
}

func (s *KeyRotationPublicTestSuite) TestRecordAgentKeyRotation() {
	tests := []struct {
		name         string
		grace        time.Duration
		setupMock    func()
		validateFunc func(enrollment.AcceptedAgent)
	}{
		{
			name:  "a replacement keeps the outgoing key for the grace period",
			grace: 24 * time.Hour,
			setupMock: func() {
				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(s.existing())
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(stored enrollment.AcceptedAgent) {
				s.Equal(ed25519.PublicKey(s.newKey), stored.PublicKey)
				s.Equal(ed25519.PublicKey(s.oldKey), stored.SupersededKey)
				s.Equal(
					s.fixedTime.Add(24*time.Hour),
					stored.SupersededUntil,
					"the expiry is an instant, so a restart cannot extend it",
				)
			},
		},
		{
			name:  "with no grace configured the replacement takes effect at once",
			grace: 0,
			setupMock: func() {
				// No Get at all: nothing is carried forward, so there is
				// nothing to read.
			},
			validateFunc: func(stored enrollment.AcceptedAgent) {
				s.Equal(ed25519.PublicKey(s.newKey), stored.PublicKey)
				s.Empty(stored.SupersededKey)
				s.True(stored.SupersededUntil.IsZero())
			},
		},
		{
			name:  "re-accepting the same key is not a rotation",
			grace: 24 * time.Hour,
			setupMock: func() {
				data, err := json.Marshal(enrollment.AcceptedAgent{
					MachineID: "machine-001",
					Hostname:  "web-01",
					PublicKey: s.newKey,
				})
				s.Require().NoError(err)

				entry := jobMocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(data)
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(entry, nil)
			},
			validateFunc: func(stored enrollment.AcceptedAgent) {
				s.Empty(
					stored.SupersededKey,
					"nothing was replaced, so nothing needs a grace period",
				)
			},
		},
		{
			name:  "a first acceptance has nothing to supersede",
			grace: 24 * time.Hour,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "accepted.machine-001").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(stored enrollment.AcceptedAgent) {
				s.Empty(stored.SupersededKey)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.watcher = enrollment.NewWatcher(
				slog.Default(), s.mockNC, s.mockKV, s.mockPKI, false, "osapi",
			)
			s.watcher.SetRotationGracePeriod(tt.grace)
			tt.setupMock()

			var stored enrollment.AcceptedAgent

			s.mockKV.EXPECT().
				Put(gomock.Any(), "accepted.machine-001", gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					_ string,
					data []byte,
				) (uint64, error) {
					s.Require().NoError(json.Unmarshal(data, &stored))

					return uint64(1), nil
				})

			s.Require().NoError(enrollment.ExportRecordAgentKey(
				s.ctx,
				s.watcher,
				enrollment.PendingAgent{
					MachineID: "machine-001",
					Hostname:  "web-01",
					PublicKey: s.newKey,
				},
			))

			tt.validateFunc(stored)
		})
	}
}

func TestKeyRotationPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeyRotationPublicTestSuite))
}
