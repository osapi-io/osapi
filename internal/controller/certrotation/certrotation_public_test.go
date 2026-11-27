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

package certrotation_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/certrotation/mocks"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

const bucket = "state"

var errBoom = errors.New("boom")

type CertRotationPublicTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	kv       *mocks.MockKV
	admitted *mocks.MockAdmitted
	seen     *mocks.MockSeen
	keys     *pki.Manager
	src      pki.CertSource
	now      time.Time
}

func TestCertRotationPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CertRotationPublicTestSuite))
}

func (s *CertRotationPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.kv = mocks.NewMockKV(s.ctrl)
	s.admitted = mocks.NewMockAdmitted(s.ctrl)
	s.seen = mocks.NewMockSeen(s.ctrl)
	dir := s.T().TempDir()
	s.keys = pki.New(osfs.NewWithNoIdm(), dir, "controller")
	s.Require().NoError(s.keys.LoadOrGenerate())
	_, err := config.ServerTLS{Enabled: true}.WithSelfSigned(dir, "nats")
	s.Require().NoError(err)
	cert, key := config.SelfSignedFiles(dir, "nats")
	s.src = pki.CertSource{Name: "nats", CertFile: cert, KeyFile: key, SelfSigned: true}
	s.now = time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
}

func (s *CertRotationPublicTestSuite) SetupSubTest() {
	s.SetupTest()
}

// rotation is a Rotation with the suite's doubles, and admitted and seen
// unless bare.
func (s *CertRotationPublicTestSuite) rotation(
	bare bool,
) *certrotation.Rotation {
	cfg := certrotation.Config{
		KV: s.kv, Bucket: bucket, Keys: s.keys, Source: s.src,
		Grace: time.Hour, Now: func() time.Time { return s.now },
	}
	if !bare {
		cfg.Admitted, cfg.Seen = s.admitted, s.seen
	}

	return certrotation.New(cfg)
}

// current is the PEM certificate that the server serves.
func (s *CertRotationPublicTestSuite) current() []byte {
	data, err := os.ReadFile(s.src.CertFile)
	s.Require().NoError(err)

	return data
}

// makeNext makes the next certificate in its files, and returns it.
func (s *CertRotationPublicTestSuite) makeNext() []byte {
	cert, key := certrotation.NextFiles(s.src)
	s.Require().NoError(config.MakeSelfSigned(cert, key))
	data, err := os.ReadFile(cert)
	s.Require().NoError(err)

	return data
}

// record is rec, signed with the controller's key, as KVGet returns it at
// revision 3.
func (s *CertRotationPublicTestSuite) record(
	rec pki.CertRecord,
	priv ed25519.PrivateKey,
) {
	rec.Sign(priv)
	data, err := json.Marshal(rec)
	s.Require().NoError(err)
	entry := jobmocks.NewMockKeyValueEntry(s.ctrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	entry.EXPECT().Revision().Return(uint64(3)).AnyTimes()
	s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).Return(entry, nil)
}

// rotating is a record of a rotation in progress that started at started.
func (s *CertRotationPublicTestSuite) rotating(
	next []byte,
	started time.Time,
) pki.CertRecord {
	return pki.CertRecord{
		Serial: 4, Name: "nats", Current: s.current(), Next: next,
		StartedAt: started.Format(time.RFC3339Nano),
	}
}

// agents are the admitted agents that the switch waits for: a with next,
// b as hasB, a pending one, and one that has not reported for a day.
func (s *CertRotationPublicTestSuite) agents(
	next []byte,
	hasB bool,
) {
	fp := pki.CertFingerprint(next)
	s.admitted.EXPECT().List(gomock.Any()).Return([]admission.Record{
		{MachineID: "a", Hostname: "web-01", State: types.AdmissionAdmitted},
		{MachineID: "b", Hostname: "web-02", State: types.AdmissionAdmitted},
		{MachineID: "p", Hostname: "new", State: types.AdmissionPending},
		{MachineID: "gone", Hostname: "old", State: types.AdmissionAdmitted},
	}, nil)
	bNext := ""
	if hasB {
		bNext = fp
	}
	s.seen.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
		{MachineID: "a", NATSCertNext: fp},
		{MachineID: "b", NATSCertNext: bNext},
	}, nil)
	s.seen.EXPECT().AgentsLastSeen(gomock.Any()).Return(map[string]time.Time{
		"gone": s.now.Add(-24 * time.Hour),
	}, nil)
}

// verify checks a signature with the controller's key.
func (s *CertRotationPublicTestSuite) verify(
	data []byte,
	signature []byte,
) bool {
	return ed25519.Verify(s.keys.PublicKey(), data, signature)
}

func (s *CertRotationPublicTestSuite) TestPreviousKey() {
	s.Run("a record that the controller's previous key signed is its own", func() {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		s.Require().NoError(err)
		s.keys.SetPreviousControllerPublicKey(pub)
		s.record(pki.CertRecord{Serial: 5, Name: "nats", Current: []byte("x")}, priv)
		got, err := s.rotation(true).Status(context.Background())
		s.Require().NoError(err)
		s.Equal(pki.CertFingerprint([]byte("x")), got.Current)
	})
}

func (s *CertRotationPublicTestSuite) TestNextFiles() {
	cert, key := certrotation.NextFiles(
		pki.CertSource{CertFile: "/k/nats.crt", KeyFile: "/k/nats.key"},
	)
	s.Equal("/k/nats.next.crt", cert)
	s.Equal("/k/nats.next.key", key)
}

func (s *CertRotationPublicTestSuite) TestStart() {
	s.Run("when no rotation is in progress, it makes the next certificate and writes it", func() {
		var written pki.CertRecord
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		s.kv.EXPECT().KVCreate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
				s.Require().NoError(json.Unmarshal(data, &written))
				return 1, nil
			})
		got, err := s.rotation(true).Start(context.Background())
		s.Require().NoError(err)
		next, _ := certrotation.NextFiles(s.src)
		data, err := os.ReadFile(next)
		s.Require().NoError(err)
		s.Equal(pki.CertFingerprint(data), got.Next)
		s.Equal(pki.CertFingerprint(s.current()), got.Current)
		s.Equal(s.now.Add(time.Hour), got.SwitchBy)
		s.Empty(got.Agents, "a rotation of the CLI waits for no agents")
		s.True(written.Rotating())
		s.Equal(uint64(1), written.Serial)
		s.True(written.SignedWith(s.verify))
	})
	s.Run("when a rotation is in progress, it shows it and makes nothing", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now), s.keys.PrivateKey())
		s.agents(next, false)
		got, err := s.rotation(false).Start(context.Background())
		s.Require().NoError(err)
		s.Equal(pki.CertFingerprint(next), got.Next)
		s.Equal([]certrotation.Agent{
			{MachineID: "a", Hostname: "web-01", HasNext: true},
			{MachineID: "b", Hostname: "web-02"},
		}, got.Agents)
	})
	s.Run("when the record is not the controller's, it starts over the record", func() {
		_, other, _ := ed25519.GenerateKey(rand.Reader)
		s.record(s.rotating([]byte("forged"), s.now), other)
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(4), nil)
		got, err := s.rotation(true).Start(context.Background())
		s.Require().NoError(err)
		s.NotEmpty(got.Next)
	})
	s.Run("when the certificate is the operator's, it refuses", func() {
		s.src.SelfSigned = false
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorIs(err, certrotation.ErrOperatorCert)
	})
	s.Run("when the record does not read, it is an error", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).Return(nil, errBoom)
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorIs(err, errBoom)
	})
	s.Run("when the record is not JSON, it starts over the record", func() {
		entry := jobmocks.NewMockKeyValueEntry(s.ctrl)
		entry.EXPECT().Value().Return([]byte("nope"))
		entry.EXPECT().Revision().Return(uint64(3))
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).Return(entry, nil)
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(4), nil)
		got, err := s.rotation(true).Start(context.Background())
		s.Require().NoError(err)
		s.NotEmpty(got.Next)
	})
	s.Run("when the next certificate does not read, it is an error", func() {
		old := certrotation.SetMakeCert(func(string, string) error { return nil })
		defer func() { certrotation.SetMakeCert(old) }()
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorContains(err, "read the next certificate")
	})
	s.Run("when the certificate is not on this host, it says where to rotate it", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		s.Require().NoError(os.Remove(s.src.CertFile))
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorContains(err, "rotate it on the host of the NATS server")
	})
	s.Run("when the next certificate is not made, it is an error", func() {
		old := certrotation.SetMakeCert(func(string, string) error { return errBoom })
		defer func() { certrotation.SetMakeCert(old) }()
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorIs(err, errBoom)
	})
	s.Run("when the record does not write, it is an error", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		s.kv.EXPECT().KVCreate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any()).
			Return(uint64(0), errBoom)
		_, err := s.rotation(true).Start(context.Background())
		s.ErrorIs(err, errBoom)
	})
	s.Run("when the agents do not list, it is an error", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now), s.keys.PrivateKey())
		s.admitted.EXPECT().List(gomock.Any()).Return(nil, errBoom)
		_, err := s.rotation(false).Start(context.Background())
		s.ErrorIs(err, errBoom)
	})
}

func (s *CertRotationPublicTestSuite) TestStatus() {
	s.Run("with no record, it is the certificate that the server serves", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		got, err := s.rotation(true).Status(context.Background())
		s.Require().NoError(err)
		s.Equal(pki.CertFingerprint(s.current()), got.Current)
		s.Empty(got.Next)
	})
	s.Run("with no record and no certificate, it is an error", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		s.Require().NoError(os.Remove(s.src.CertFile))
		_, err := s.rotation(true).Status(context.Background())
		s.ErrorIs(err, os.ErrNotExist)
	})
	s.Run("with a switch, it is the new certificate", func() {
		s.record(pki.CertRecord{Serial: 5, Name: "nats", Current: []byte("x")}, s.keys.PrivateKey())
		got, err := s.rotation(true).Status(context.Background())
		s.Require().NoError(err)
		s.Equal(pki.CertFingerprint([]byte("x")), got.Current)
	})
	s.Run("when the record does not read, it is an error", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).Return(nil, errBoom)
		_, err := s.rotation(true).Status(context.Background())
		s.ErrorIs(err, errBoom)
	})
}

func (s *CertRotationPublicTestSuite) TestWaitingForErrors() {
	for _, tc := range []struct {
		name  string
		setup func()
	}{
		{
			name: "when the agents do not list",
			setup: func() {
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().ListAgents(gomock.Any()).Return(nil, errBoom)
			},
		},
		{
			name: "when the last reports do not read",
			setup: func() {
				s.admitted.EXPECT().List(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().ListAgents(gomock.Any()).Return(nil, nil)
				s.seen.EXPECT().AgentsLastSeen(gomock.Any()).Return(nil, errBoom)
			},
		},
	} {
		s.Run(tc.name, func() {
			next := s.makeNext()
			s.record(s.rotating(next, s.now), s.keys.PrivateKey())
			tc.setup()
			_, err := s.rotation(false).Status(context.Background())
			s.ErrorIs(err, errBoom)
		})
	}
}

func (s *CertRotationPublicTestSuite) TestStep() {
	s.Run("with no rotation, it does nothing", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			Return(nil, jetstream.ErrKeyNotFound)
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.False(got)
	})
	s.Run("when an agent does not trust the next certificate yet, it waits", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now), s.keys.PrivateKey())
		s.agents(next, false)
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.False(got)
	})
	s.Run("when each agent trusts it, it renames it into place and writes the switch", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now), s.keys.PrivateKey())
		s.agents(next, true)
		var written pki.CertRecord
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte, _ uint64) (uint64, error) {
				s.Require().NoError(json.Unmarshal(data, &written))
				return 4, nil
			})
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.True(got)
		s.Equal(next, s.current(), "the next certificate is in place")
		cert, key := certrotation.NextFiles(s.src)
		s.NoFileExists(cert)
		s.NoFileExists(key)
		s.False(written.Rotating())
		s.Equal(next, written.Current)
		s.Equal(uint64(5), written.Serial)
		s.True(written.SignedWith(s.verify))
	})
	s.Run("when the grace period ended, it switches without the agents", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(4), nil)
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.True(got)
	})
	s.Run("when a switch stopped after its renames, it writes the switch", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		cert, key := certrotation.NextFiles(s.src)
		s.Require().NoError(os.Rename(key, s.src.KeyFile))
		s.Require().NoError(os.Rename(cert, s.src.CertFile))
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(4), nil)
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.True(got)
	})
	s.Run("when a switch stopped between its renames, it renames the certificate", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		_, key := certrotation.NextFiles(s.src)
		s.Require().NoError(os.Rename(key, s.src.KeyFile))
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(4), nil)
		got, err := s.rotation(false).Step(context.Background())
		s.Require().NoError(err)
		s.True(got)
		s.Equal(next, s.current(), "the next certificate is in place")
	})
	s.Run("when the next key is gone and the key in place is not its pair, it is an error", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		_, key := certrotation.NextFiles(s.src)
		s.Require().NoError(os.Remove(key))
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorContains(err, "switch to the next certificate")
	})
	s.Run("when the next certificate is gone, it is an error", func() {
		s.record(s.rotating([]byte("next"), s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorContains(err, "read the next certificate")
	})
	s.Run("when the next certificate is not the record's, it is an error", func() {
		s.makeNext()
		s.record(s.rotating([]byte("other"), s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorContains(err, "is not the one of the record")
	})
	for _, file := range []string{"key", "cert"} {
		s.Run("when the "+file+" does not rename, it is an error", func() {
			next := s.makeNext()
			s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
			target := s.src.KeyFile
			if file == "cert" {
				target = s.src.CertFile
			}
			s.Require().NoError(os.Remove(target))
			s.Require().NoError(os.MkdirAll(filepath.Join(target, "busy"), 0o700))
			_, err := s.rotation(false).Step(context.Background())
			s.ErrorContains(err, "switch to the next certificate")
		})
	}
	s.Run("when the switch does not write, it is an error", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now.Add(-2*time.Hour)), s.keys.PrivateKey())
		s.kv.EXPECT().KVUpdate(gomock.Any(), bucket, pki.NATSCertStateKey, gomock.Any(), uint64(3)).
			Return(uint64(0), errBoom)
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorIs(err, errBoom)
	})
	s.Run("when the record does not read, it is an error", func() {
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).Return(nil, errBoom)
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorIs(err, errBoom)
	})
	s.Run("when the agents do not list, it is an error", func() {
		next := s.makeNext()
		s.record(s.rotating(next, s.now), s.keys.PrivateKey())
		s.admitted.EXPECT().List(gomock.Any()).Return(nil, errBoom)
		_, err := s.rotation(false).Step(context.Background())
		s.ErrorIs(err, errBoom)
	})
}

func (s *CertRotationPublicTestSuite) TestRun() {
	s.Run("it steps until ctx is done, and logs a step that fails", func() {
		defer certrotation.SetInterval(certrotation.SetInterval(time.Millisecond))
		ctx, cancel := context.WithCancel(context.Background())
		s.kv.EXPECT().KVGet(gomock.Any(), bucket, pki.NATSCertStateKey).
			DoAndReturn(func(context.Context, string, string) (jetstream.KeyValueEntry, error) {
				cancel()
				return nil, errBoom
			}).MinTimes(1)
		s.NoError(certrotation.New(certrotation.Config{
			KV: s.kv, Bucket: bucket, Keys: s.keys, Source: s.src,
		}).Run(ctx))
	})
}
