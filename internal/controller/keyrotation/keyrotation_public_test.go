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

package keyrotation_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/keyrotation"
	"github.com/osapi-io/osapi/internal/controller/keyrotation/mocks"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

const (
	bucket = "agent-state"
	key    = pki.ControllerKeyStateKey
)

type KeyRotationPublicTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	kv       *mocks.MockKV
	admitted *mocks.MockAdmitted
	seen     *mocks.MockSeen
	fs       avfs.VFS
	keys     *pki.Manager
	now      time.Time
}

func (s *KeyRotationPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.kv = mocks.NewMockKV(s.ctrl)
	s.admitted = mocks.NewMockAdmitted(s.ctrl)
	s.seen = mocks.NewMockSeen(s.ctrl)
	s.fs = memfs.New()
	s.keys = pki.New(s.fs, "/ctl", "controller")
	s.Require().NoError(s.keys.LoadOrGenerate())
	s.keys.SetControllerPublicKey(s.keys.PublicKey())
	s.now = time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
}

// rotation is a Rotation over the suite's doubles and keys.
func (s *KeyRotationPublicTestSuite) rotation() *keyrotation.Rotation {
	return keyrotation.New(keyrotation.Config{
		KV: s.kv, Bucket: bucket, Keys: s.keys, Admitted: s.admitted, Seen: s.seen,
		Grace: time.Hour, Logger: slog.New(slog.DiscardHandler),
		Now: func() time.Time { return s.now },
	})
}

// entry is a KV entry that holds rec at revision.
func (s *KeyRotationPublicTestSuite) entry(
	rec pki.ControllerKeyRecord,
	revision uint64,
) jetstream.KeyValueEntry {
	data, err := json.Marshal(rec)
	s.Require().NoError(err)
	e := jobmocks.NewMockKeyValueEntry(s.ctrl)
	e.EXPECT().Value().Return(data).AnyTimes()
	e.EXPECT().Revision().Return(revision).AnyTimes()

	return e
}

// raw is a KV entry that holds data.
func (s *KeyRotationPublicTestSuite) raw(
	data []byte,
) jetstream.KeyValueEntry {
	e := jobmocks.NewMockKeyValueEntry(s.ctrl)
	e.EXPECT().Value().Return(data).AnyTimes()
	e.EXPECT().Revision().Return(uint64(1)).AnyTimes()

	return e
}

// next makes the next key on the suite's file system, as a Start makes it.
func (s *KeyRotationPublicTestSuite) next() *pki.Manager {
	next, err := s.keys.Next()
	s.Require().NoError(err)

	return next
}

// rotating is a record of a rotation to next that the current key signed.
func (s *KeyRotationPublicTestSuite) rotating(
	next ed25519.PublicKey,
) pki.ControllerKeyRecord {
	rec := pki.ControllerKeyRecord{
		Serial: 1, Current: s.keys.PublicKey(), Next: next, StartedAt: s.now,
	}
	rec.Sign(s.keys.PrivateKey())

	return rec
}

// foreign is a key that the controller does not have.
func (s *KeyRotationPublicTestSuite) foreign() (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	return pub, priv
}

// agents expects the admitted agents and the registry: an admitted
// agent web-01 that reports next, or not.
func (s *KeyRotationPublicTestSuite) agents(
	reports string,
) {
	s.admitted.EXPECT().List(gomock.Any()).Return([]admission.Record{
		{MachineID: "m-1", Hostname: "web-01", State: types.AdmissionAdmitted},
	}, nil)
	s.seen.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
		{MachineID: "m-1", Hostname: "web-01", ControllerKeyNext: reports},
	}, nil)
	s.seen.EXPECT().AgentsLastSeen(gomock.Any()).Return(map[string]time.Time{}, nil)
}

// failing is fs, where the call fn fails.
func failing(
	fs avfs.VFS,
	fn avfs.FnVFS,
) avfs.VFS {
	vfs := failfs.New(fs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, got avfs.FnVFS, _ *failfs.FailParam) error {
		if got == fn {
			return errors.New("injected")
		}

		return nil
	})

	return vfs
}

func (s *KeyRotationPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		setup        func()
		wantErr      string
		validateFunc func(keyrotation.Status)
	}{
		{
			name: "with no record, it makes the next key and writes the first record",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
				s.kv.EXPECT().KVCreate(gomock.Any(), bucket, key, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
						var rec pki.ControllerKeyRecord
						s.Require().NoError(json.Unmarshal(data, &rec))
						s.True(rec.SignedBy(s.keys.PublicKey()))
						s.True(rec.Rotating())
						s.Equal(uint64(1), rec.Serial)

						return 1, nil
					})
				s.agents("")
			},
			validateFunc: func(st keyrotation.Status) {
				s.Equal(s.keys.Fingerprint(), st.Current)
				s.NotEmpty(st.Next)
				s.Equal(s.now.Add(time.Hour), st.SwitchBy)
				s.Equal([]keyrotation.Agent{{MachineID: "m-1", Hostname: "web-01"}}, st.Agents)
			},
		},
		{
			name: "after a switch, it updates the record at its revision and keeps its retired keys",
			setup: func() {
				old, _ := s.foreign()
				retired, _ := s.foreign()
				rec := pki.ControllerKeyRecord{
					Serial:   4,
					Current:  s.keys.PublicKey(),
					Previous: old,
					Retired:  [][]byte{retired},
				}
				rec.Sign(s.keys.PrivateKey())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 7), nil)
				s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(7)).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte, _ uint64) (uint64, error) {
						var got pki.ControllerKeyRecord
						s.Require().NoError(json.Unmarshal(data, &got))
						s.Equal(uint64(5), got.Serial)
						s.Equal(old, ed25519.PublicKey(got.Previous))
						s.Equal([][]byte{retired}, got.Retired)

						return 8, nil
					})
				s.agents("")
			},
			validateFunc: func(st keyrotation.Status) {
				s.NotEmpty(st.Next)
			},
		},
		{
			name: "a rotation in progress is not started again",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.agents(next.Fingerprint())
			},
			validateFunc: func(st keyrotation.Status) {
				s.Equal(
					[]keyrotation.Agent{{MachineID: "m-1", Hostname: "web-01", HasNext: true}},
					st.Agents,
				)
			},
		},
		{
			name: "a record that the controller's key did not sign is not the controller's",
			setup: func() {
				pub, priv := s.foreign()
				rec := pki.ControllerKeyRecord{Serial: 9, Current: pub, Next: pub}
				rec.Sign(priv)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 3), nil)
				s.kv.EXPECT().
					KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(3)).
					Return(uint64(4), nil)
				s.agents("")
			},
			validateFunc: func(st keyrotation.Status) {
				s.NotEmpty(st.Next)
			},
		},
		{
			name: "a record that does not read",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, errors.New("timeout"))
			},
			wantErr: "read the controller key record: timeout",
		},
		{
			name: "a record that is not JSON",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.raw([]byte("{")), nil)
			},
			wantErr: "the controller key record: ",
		},
		{
			name: "a next key that is not made",
			setup: func() {
				s.keys = pki.New(failing(s.fs, avfs.FnMkdirAll), "/ctl", "controller")
				s.Require().NoError(s.keys.Load())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
			},
			wantErr: "make the next controller key: ",
		},
		{
			name: "a record that does not write",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
				s.kv.EXPECT().
					KVCreate(gomock.Any(), bucket, key, gomock.Any()).
					Return(uint64(0), errors.New("full"))
			},
			wantErr: "write the controller key record: full",
		},
		{
			name: "the agents that do not list",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
				s.kv.EXPECT().
					KVCreate(gomock.Any(), bucket, key, gomock.Any()).
					Return(uint64(1), nil)
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, errors.New("down"))
			},
			wantErr: "list the admitted agents: down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()
			st, err := s.rotation().Start(context.Background())
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(st)
		})
	}
}

func (s *KeyRotationPublicTestSuite) TestStatus() {
	tests := []struct {
		name         string
		setup        func()
		wantErr      string
		validateFunc func(keyrotation.Status)
	}{
		{
			name: "with no rotation, it is the current key only",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(st keyrotation.Status) {
				s.Equal(keyrotation.Status{Current: s.keys.Fingerprint()}, st)
			},
		},
		{
			name: "the agents that the switch waits for: admitted, and online or seen within the grace period",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.admitted.EXPECT().List(gomock.Any()).Return([]admission.Record{
					{MachineID: "online", Hostname: "web-01", State: types.AdmissionAdmitted},
					{MachineID: "switched", Hostname: "web-02", State: types.AdmissionAdmitted},
					{MachineID: "recent", Hostname: "web-03", State: types.AdmissionAdmitted},
					{MachineID: "gone", Hostname: "web-04", State: types.AdmissionAdmitted},
					{MachineID: "never", Hostname: "web-05", State: types.AdmissionAdmitted},
					{MachineID: "pending", Hostname: "web-06", State: types.AdmissionPending},
				}, nil)
				s.seen.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
					{MachineID: "online", ControllerKeyNext: next.Fingerprint()},
					{MachineID: "switched", ControllerKey: next.Fingerprint()},
				}, nil)
				s.seen.EXPECT().AgentsLastSeen(gomock.Any()).Return(map[string]time.Time{
					"recent": s.now.Add(-time.Minute),
					"gone":   s.now.Add(-2 * time.Hour),
				}, nil)
			},
			validateFunc: func(st keyrotation.Status) {
				s.Equal([]keyrotation.Agent{
					{MachineID: "online", Hostname: "web-01", HasNext: true},
					{MachineID: "switched", Hostname: "web-02", HasNext: true},
					{MachineID: "recent", Hostname: "web-03"},
				}, st.Agents)
			},
		},
		{
			name: "a record that does not read",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, errors.New("timeout"))
			},
			wantErr: "timeout",
		},
		{
			name: "the registry that does not list",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("down"))
			},
			wantErr: "list the agents: down",
		},
		{
			name: "the last reports that do not read",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().ListAgents(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().AgentsLastSeen(gomock.Any()).Return(nil, errors.New("down"))
			},
			wantErr: "read when the agents last reported: down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()
			st, err := s.rotation().Status(context.Background())
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(st)
		})
	}
}

func (s *KeyRotationPublicTestSuite) TestStatusWithNoAgents() {
	s.Run("a rotation of the CLI, with no registry, lists no agents", func() {
		next := s.next()
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
			Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
		r := keyrotation.New(keyrotation.Config{KV: s.kv, Bucket: bucket, Keys: s.keys})
		st, err := r.Status(context.Background())
		s.Require().NoError(err)
		s.Equal(next.Fingerprint(), st.Next)
		s.Nil(st.Agents)
	})
}

func (s *KeyRotationPublicTestSuite) TestStep() {
	tests := []struct {
		name         string
		setup        func()
		wantSwitched bool
		wantErr      string
		validateFunc func(old ed25519.PublicKey)
	}{
		{
			name: "it switches when each agent has the next key",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.agents(next.Fingerprint())
				s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(2)).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte, _ uint64) (uint64, error) {
						var rec pki.ControllerKeyRecord
						s.Require().NoError(json.Unmarshal(data, &rec))
						s.True(rec.SignedBy(next.PublicKey()))
						s.Equal(next.PublicKey(), ed25519.PublicKey(rec.Current))
						s.False(rec.Rotating())
						s.Equal(uint64(2), rec.Serial)

						return 3, nil
					})
			},
			wantSwitched: true,
			validateFunc: func(old ed25519.PublicKey) {
				s.NotEqual(old, s.keys.PublicKey())
				s.Equal(s.keys.PublicKey(), s.keys.ControllerPublicKey())
				s.Equal(old, s.keys.PreviousControllerPublicKey())
				// The next key moved in place of the current one.
				reloaded := pki.New(s.fs, "/ctl", "controller")
				s.Require().NoError(reloaded.Load())
				s.Equal(s.keys.PublicKey(), reloaded.PublicKey())
			},
		},
		{
			name: "it waits while an agent does not have the next key",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.agents("")
			},
			validateFunc: func(old ed25519.PublicKey) {
				s.Equal(old, s.keys.PublicKey())
			},
		},
		{
			name: "it switches when the grace period ended, with an agent that does not have it",
			setup: func() {
				next := s.next()
				rec := s.rotating(next.PublicKey())
				s.now = s.now.Add(2 * time.Hour)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)
				s.kv.EXPECT().
					KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(2)).
					Return(uint64(3), nil)
			},
			wantSwitched: true,
			validateFunc: func(old ed25519.PublicKey) {
				s.NotEqual(old, s.keys.PublicKey())
			},
		},
		{
			name: "with no rotation, it does nothing",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(old ed25519.PublicKey) {
				s.Equal(old, s.keys.PublicKey())
			},
		},
		{
			name: "a record that does not read",
			setup: func() {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, errors.New("timeout"))
			},
			wantErr: "timeout",
		},
		{
			name: "the agents that do not list",
			setup: func() {
				next := s.next()
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(s.entry(s.rotating(next.PublicKey()), 2), nil)
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, errors.New("down"))
			},
			wantErr: "list the admitted agents: down",
		},
		{
			name: "a next key on the host that is not the record's",
			setup: func() {
				s.next()
				other, _ := s.foreign()
				rec := s.rotating(other)
				s.now = s.now.Add(2 * time.Hour)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)
			},
			wantErr: "the next controller key on the host is not the one of the record",
		},
		{
			name: "a next key that does not read",
			setup: func() {
				other, _ := s.foreign()
				rec := s.rotating(other)
				s.now = s.now.Add(2 * time.Hour)
				s.keys = pki.New(failing(s.fs, avfs.FnMkdirAll), "/ctl", "controller")
				s.Require().NoError(s.keys.Load())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)
			},
			wantErr: "read the next controller key: ",
		},
		{
			name: "a next key that does not move in place",
			setup: func() {
				next := s.next()
				rec := s.rotating(next.PublicKey())
				s.now = s.now.Add(2 * time.Hour)
				s.keys = pki.New(failing(s.fs, avfs.FnRename), "/ctl", "controller")
				s.Require().NoError(s.keys.Load())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)
			},
			wantErr: "switch to the next controller key: ",
		},
		{
			name: "a switch whose record does not write",
			setup: func() {
				next := s.next()
				rec := s.rotating(next.PublicKey())
				s.now = s.now.Add(2 * time.Hour)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)
				s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(2)).
					Return(uint64(0), errors.New("full"))
			},
			wantErr: "write the controller key record: full",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()
			old := s.keys.PublicKey()
			switched, err := s.rotation().Step(context.Background())
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantSwitched, switched)
			tc.validateFunc(old)
		})
	}
}

func (s *KeyRotationPublicTestSuite) TestLoad() {
	tests := []struct {
		name         string
		setup        func() ed25519.PublicKey
		wantErr      string
		validateFunc func(want ed25519.PublicKey)
	}{
		{
			name: "after a switch, the previous key verifies again",
			setup: func() ed25519.PublicKey {
				old, _ := s.foreign()
				rec := pki.ControllerKeyRecord{
					Serial:   2,
					Current:  s.keys.PublicKey(),
					Previous: old,
				}
				rec.Sign(s.keys.PrivateKey())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 2), nil)

				return old
			},
			validateFunc: func(want ed25519.PublicKey) {
				s.Equal(want, s.keys.PreviousControllerPublicKey())
			},
		},
		{
			name: "after a switch, each retired key verifies the stored records again",
			setup: func() ed25519.PublicKey {
				old, _ := s.foreign()
				retired, _ := s.foreign()
				rec := pki.ControllerKeyRecord{
					Serial:   3,
					Current:  s.keys.PublicKey(),
					Previous: old,
					Retired:  [][]byte{retired},
				}
				rec.Sign(s.keys.PrivateKey())
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 3), nil)

				return retired
			},
			validateFunc: func(want ed25519.PublicKey) {
				s.Contains(s.keys.StoredRecordKeys(), want)
			},
		},
		{
			name: "a stopped switch retires the previous key in its record",
			setup: func() ed25519.PublicKey {
				oldPub, oldPriv := s.foreign()
				previous, _ := s.foreign()
				rec := pki.ControllerKeyRecord{
					Serial: 2, Current: oldPub, Next: s.keys.PublicKey(),
					Previous: previous, StartedAt: s.now,
				}
				rec.Sign(oldPriv)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 6), nil)
				s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(6)).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte, _ uint64) (uint64, error) {
						var done pki.ControllerKeyRecord
						s.Require().NoError(json.Unmarshal(data, &done))
						s.True(done.SignedBy(s.keys.PublicKey()))
						s.Equal([][]byte{previous}, done.Retired)

						return 7, nil
					})

				return previous
			},
			validateFunc: func(want ed25519.PublicKey) {
				s.Contains(s.keys.StoredRecordKeys(), want)
			},
		},
		{
			name: "a switch that stopped after its renames writes its record",
			setup: func() ed25519.PublicKey {
				oldPub, oldPriv := s.foreign()
				rec := pki.ControllerKeyRecord{
					Serial: 1, Current: oldPub, Next: s.keys.PublicKey(), StartedAt: s.now,
				}
				rec.Sign(oldPriv)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(s.entry(rec, 4), nil)
				s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, key, gomock.Any(), uint64(4)).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte, _ uint64) (uint64, error) {
						var done pki.ControllerKeyRecord
						s.Require().NoError(json.Unmarshal(data, &done))
						s.True(done.SignedBy(s.keys.PublicKey()))
						s.Equal(oldPub, ed25519.PublicKey(done.Previous))

						return 5, nil
					})

				return oldPub
			},
			validateFunc: func(want ed25519.PublicKey) {
				s.Equal(want, s.keys.PreviousControllerPublicKey())
			},
		},
		{
			name: "with no record, it changes nothing",
			setup: func() ed25519.PublicKey {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, jetstream.ErrKeyNotFound)

				return nil
			},
			validateFunc: func(want ed25519.PublicKey) {
				s.Equal(want, s.keys.PreviousControllerPublicKey())
			},
		},
		{
			name: "a record that does not read",
			setup: func() ed25519.PublicKey {
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, errors.New("timeout"))

				return nil
			},
			wantErr: "timeout",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			want := tc.setup()
			err := s.rotation().Load(context.Background())
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(want)
		})
	}
}

func (s *KeyRotationPublicTestSuite) TestRun() {
	old := keyrotation.SetInterval(time.Millisecond)
	defer keyrotation.SetInterval(old)

	tests := []struct {
		name  string
		setup func(cancel context.CancelFunc)
	}{
		{
			name: "it loads, steps until ctx ends, and logs what fails",
			setup: func(cancel context.CancelFunc) {
				first := s.kv.EXPECT().
					KVGet(gomock.Any(), bucket, key).
					Return(nil, errors.New("load"))
				second := s.kv.EXPECT().
					KVGet(gomock.Any(), bucket, key).
					Return(nil, errors.New("step")).
					After(first)
				s.kv.EXPECT().KVGet(gomock.Any(), bucket, key).After(second).
					DoAndReturn(func(context.Context, string, string) (jetstream.KeyValueEntry, error) {
						cancel()

						return nil, jetstream.ErrKeyNotFound
					}).
					AnyTimes()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tc.setup(cancel)
			s.NoError(s.rotation().Run(ctx))
		})
	}
}

func TestKeyRotationPublicTestSuite(t *testing.T) {
	suite.Run(t, new(KeyRotationPublicTestSuite))
}
