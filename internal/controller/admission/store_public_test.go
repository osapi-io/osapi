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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
	jobMocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// fakeKV is an in-memory enrollment bucket, so a test checks what the store
// leaves in it rather than the calls it makes.
type fakeKV struct {
	ctrl    *gomock.Controller
	data    map[string][]byte
	keysErr error
	// failGet and failPut fail a read or a write of a key with this prefix.
	failGet string
	failPut string
	// beforeGet runs at the start of each read, with the key.
	beforeGet func(key string)
}

func (f *fakeKV) KVGet(_ context.Context, _ string, key string) (jetstream.KeyValueEntry, error) {
	v, ok := f.data[key]
	if f.beforeGet != nil {
		f.beforeGet(key)
	}
	if f.failGet != "" && strings.HasPrefix(key, f.failGet) {
		return nil, errors.New("kv: down")
	}
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}
	e := jobMocks.NewMockKeyValueEntry(f.ctrl)
	e.EXPECT().Value().Return(v).AnyTimes()

	return e, nil
}

func (f *fakeKV) KVPut(_ context.Context, _ string, key string, value []byte) (uint64, error) {
	if f.failPut != "" && strings.HasPrefix(key, f.failPut) {
		return 0, errors.New("kv: full")
	}
	f.data[key] = value

	return 1, nil
}

func (f *fakeKV) KVDelete(_ context.Context, _ string, key string) error {
	delete(f.data, key)

	return nil
}

func (f *fakeKV) KVKeys(_ context.Context, _ string, _ ...string) ([]string, error) {
	if f.keysErr != nil {
		return nil, f.keysErr
	}
	keys := make([]string, 0, len(f.data))
	for k := range f.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}

type StorePublicTestSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	kv      *fakeKV
	nc      *admissionMocks.MockNATSSubscriber
	watcher *admission.Watcher
	key     ed25519.PublicKey
	fp      string
	at      time.Time
}

func TestStorePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StorePublicTestSuite))
}

func (s *StorePublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.kv = &fakeKV{ctrl: s.ctrl, data: map[string][]byte{}}
	s.nc = admissionMocks.NewMockNATSSubscriber(s.ctrl)
	s.nc.EXPECT().PublishCore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	pkiProvider := admissionMocks.NewMockPKIProvider(s.ctrl)
	pkiProvider.EXPECT().
		PublicKey().
		Return(make(ed25519.PublicKey, ed25519.PublicKeySize)).
		AnyTimes()
	s.watcher = admission.NewWatcher(
		slog.Default(),
		s.nc,
		s.kv,
		"enrollment",
		pkiProvider,
		false,
		"osapi",
	)
	s.key, _, _ = ed25519.GenerateKey(nil)
	s.fp = pki.FingerprintOfKeys(
		s.key,
		"BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
	)
	s.at = time.Date(2026, 11, 19, 1, 0, 0, 0, time.UTC)
	admission.SetNowFn(func() time.Time { return s.at })
}

func (s *StorePublicTestSuite) TearDownSubTest() {
	admission.ResetNowFn()
	admission.ResetMarshalFn()
}

// put stores v as JSON under key.
func (s *StorePublicTestSuite) put(
	key string,
	v any,
) {
	data, err := json.Marshal(v)
	s.Require().NoError(err)
	s.kv.data[key] = data
}

// pending stores a request from machineID with s.key.
func (s *StorePublicTestSuite) pending(
	machineID, hostname string,
) {
	s.put(admission.KVPrefix()+machineID, admission.PendingAgent{
		MachineID:     machineID,
		Hostname:      hostname,
		PublicKey:     s.key,
		EncryptionKey: "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
		Fingerprint:   "SHA256:self-reported",
		RequestedAt:   s.at,
	})
}

func (s *StorePublicTestSuite) TestList() {
	s.Run(
		"when the bucket has each state returns one record per machine in hostname order",
		func() {
			s.pending("m-new", "web-03")
			s.put(admission.AcceptedKVPrefix()+"m-ok", admission.AcceptedAgent{
				MachineID: "m-ok", Hostname: "web-01", Fingerprint: "SHA256:ok", AcceptedAt: s.at,
				EncryptionKey: "XADMITTED",
			})
			s.put(
				"rejected.m-no",
				admission.Record{MachineID: "m-no", Hostname: "web-02", Reason: "unknown host"},
			)
			// A machine revoked, then requesting again with a new key.
			s.put(
				"revoked.m-back",
				admission.Record{
					MachineID:   "m-back",
					Hostname:    "web-04",
					Fingerprint: "SHA256:old",
				},
			)
			s.pending("m-back", "web-04")
			// Neither a record nor readable.
			s.kv.data["other.x"] = []byte("{}")
			s.kv.data["accepted.bad"] = []byte("{")
			s.kv.data["enrollment.bad"] = []byte("{")
			s.kv.data["rejected.bad"] = []byte("{")

			got, err := s.watcher.List(context.Background())

			s.Require().NoError(err)
			s.Require().Len(got, 4)
			s.Equal([]string{"web-01", "web-02", "web-03", "web-04"},
				[]string{got[0].Hostname, got[1].Hostname, got[2].Hostname, got[3].Hostname})
			s.Equal(types.AdmissionAdmitted, got[0].State)
			s.Equal(s.at, got[0].AdmittedAt)
			s.Equal("XADMITTED", got[0].EncryptionKey)
			s.Equal(types.AdmissionRejected, got[1].State)
			s.Equal("unknown host", got[1].Reason)
			s.Equal(types.AdmissionPending, got[2].State)
			s.Equal(s.fp, got[2].Fingerprint)
			s.Equal(
				"BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
				got[2].EncryptionKey,
			)
			s.Empty(got[2].Previous)
			s.Equal(types.AdmissionPending, got[3].State)
			s.Equal(types.AdmissionRevoked, got[3].Previous)
			s.Equal("SHA256:old", got[3].PreviousFingerprint)
		},
	)

	s.Run("when two machines share a hostname orders them by machine ID", func() {
		s.pending("m-b", "web-01")
		s.pending("m-a", "web-01")

		got, err := s.watcher.List(context.Background())

		s.Require().NoError(err)
		s.Equal("m-a", got[0].MachineID)
		s.Equal("m-b", got[1].MachineID)
	})

	s.Run("when the bucket cannot be listed returns the error", func() {
		s.kv.keysErr = errors.New("kv down")

		_, err := s.watcher.List(context.Background())

		s.ErrorContains(err, "list admission keys: kv down")
	})
}

func (s *StorePublicTestSuite) TestGet() {
	s.Run("when the machine has a record returns it", func() {
		s.put(
			admission.AcceptedKVPrefix()+"m-ok",
			admission.AcceptedAgent{MachineID: "m-ok", Hostname: "web-01"},
		)

		got, err := s.watcher.Get(context.Background(), "m-ok")

		s.Require().NoError(err)
		s.Equal(types.AdmissionAdmitted, got.State)
	})

	s.Run("when the machine has no record returns ErrNotFound", func() {
		_, err := s.watcher.Get(context.Background(), "m-none")

		s.ErrorIs(err, admission.ErrNotFound)
	})
}

func (s *StorePublicTestSuite) TestAdmitWithoutEncryptionKey() {
	s.Run("when the request has no encryption key it has no fingerprint to admit", func() {
		s.put(admission.KVPrefix()+"m-old", admission.PendingAgent{
			MachineID: "m-old",
			Hostname:  "web-09",
			PublicKey: s.key,
		})

		got, err := s.watcher.Get(context.Background(), "m-old")
		s.Require().NoError(err)
		s.Empty(got.Fingerprint)

		err = s.watcher.Admit(context.Background(), "m-old", pki.FingerprintOf(s.key))
		s.ErrorIs(err, admission.ErrFingerprintMismatch)
	})
}

func (s *StorePublicTestSuite) TestAdmit() {
	s.Run(
		"when the fingerprint matches the key admits the machine and clears earlier decisions",
		func() {
			s.pending("m-1", "web-01")
			s.put("rejected.m-1", admission.Record{MachineID: "m-1"})
			s.put("revoked.m-1", admission.Record{MachineID: "m-1"})

			s.Require().NoError(s.watcher.Admit(context.Background(), "m-1", s.fp))

			got, err := s.watcher.Get(context.Background(), "m-1")
			s.Require().NoError(err)
			s.Equal(types.AdmissionAdmitted, got.State)
			s.Equal(s.fp, got.Fingerprint)
			s.NotContains(s.kv.data, "rejected.m-1")
			s.NotContains(s.kv.data, "revoked.m-1")
			s.NotContains(s.kv.data, admission.KVPrefix()+"m-1")
		},
	)

	s.Run("when the fingerprint differs admits nothing", func() {
		s.pending("m-1", "web-01")

		err := s.watcher.Admit(context.Background(), "m-1", "SHA256:self-reported")

		s.ErrorIs(err, admission.ErrFingerprintMismatch)
		s.Contains(s.kv.data, admission.KVPrefix()+"m-1")
		s.NotContains(s.kv.data, admission.AcceptedKVPrefix()+"m-1")
	})

	s.Run("when the fingerprint is empty admits nothing", func() {
		s.pending("m-1", "web-01")

		s.ErrorIs(
			s.watcher.Admit(context.Background(), "m-1", ""),
			admission.ErrFingerprintMismatch,
		)
	})

	s.Run("when the machine has no request returns ErrNoPendingAgent", func() {
		s.ErrorIs(s.watcher.Admit(context.Background(), "m-1", s.fp), admission.ErrNoPendingAgent)
	})

	s.Run("when publishing the answer fails returns the error", func() {
		s.pending("m-1", "web-01")
		nc := admissionMocks.NewMockNATSSubscriber(s.ctrl)
		nc.EXPECT().
			PublishCore(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("nats down"))
		pkiProvider := admissionMocks.NewMockPKIProvider(s.ctrl)
		pkiProvider.EXPECT().PublicKey().Return(make(ed25519.PublicKey, ed25519.PublicKeySize))
		w := admission.NewWatcher(
			slog.Default(),
			nc,
			s.kv,
			"enrollment",
			pkiProvider,
			false,
			"osapi",
		)

		s.ErrorContains(w.Admit(context.Background(), "m-1", s.fp), "nats down")
	})
}

func (s *StorePublicTestSuite) TestReject() {
	s.Run("when the machine has a request rejects it and keeps a record", func() {
		s.pending("m-1", "web-01")

		s.Require().NoError(s.watcher.Reject(context.Background(), "m-1", "unknown host"))

		got, err := s.watcher.Get(context.Background(), "m-1")
		s.Require().NoError(err)
		s.Equal(types.AdmissionRejected, got.State)
		s.Equal("unknown host", got.Reason)
		s.Equal(s.fp, got.Fingerprint)
		s.Equal(s.at, got.DecidedAt)
	})

	s.Run("when the machine has no request returns ErrNoPendingAgent", func() {
		s.ErrorIs(s.watcher.Reject(context.Background(), "m-1", "x"), admission.ErrNoPendingAgent)
	})

	s.Run("when publishing the answer fails keeps no record", func() {
		s.pending("m-1", "web-01")
		nc := admissionMocks.NewMockNATSSubscriber(s.ctrl)
		nc.EXPECT().
			PublishCore(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("nats down"))
		w := admission.NewWatcher(slog.Default(), nc, s.kv, "enrollment", nil, false, "osapi")

		s.ErrorContains(w.Reject(context.Background(), "m-1", "x"), "nats down")
		s.NotContains(s.kv.data, "rejected.m-1")
	})

	s.Run("when the record does not marshal returns the error", func() {
		s.pending("m-1", "web-01")
		calls := 0
		admission.SetMarshalFn(func(v any) ([]byte, error) {
			calls++
			if calls > 1 {
				return nil, errors.New("bad json")
			}
			return json.Marshal(v)
		})

		s.ErrorContains(
			s.watcher.Reject(context.Background(), "m-1", "x"),
			"marshal rejected record m-1: bad json",
		)
	})
}

func (s *StorePublicTestSuite) TestRevoke() {
	s.Run("when the machine is admitted removes its key and keeps a record", func() {
		s.put(admission.AcceptedKVPrefix()+"m-1", admission.AcceptedAgent{
			MachineID: "m-1", Hostname: "web-01", Fingerprint: s.fp, AcceptedAt: s.at,
		})

		s.Require().NoError(s.watcher.Revoke(context.Background(), "m-1"))

		got, err := s.watcher.Get(context.Background(), "m-1")
		s.Require().NoError(err)
		s.Equal(types.AdmissionRevoked, got.State)
		s.Equal(s.fp, got.Fingerprint)
		s.Equal(s.at, got.AdmittedAt)
		_, err = s.watcher.LookupAgentKey(context.Background(), "m-1")
		s.ErrorIs(err, admission.ErrAgentKeyNotFound)
	})

	s.Run("when the machine is not admitted returns ErrNoAcceptedAgent", func() {
		s.ErrorIs(s.watcher.Revoke(context.Background(), "m-1"), admission.ErrNoAcceptedAgent)
	})
}

// failingKV is a bucket whose writes and deletes fail.
type failingKV struct{ fakeKV }

func (f *failingKV) KVPut(context.Context, string, string, []byte) (uint64, error) {
	return 0, errors.New("put failed")
}

func (f *failingKV) KVDelete(context.Context, string, string) error {
	return errors.New("delete failed")
}

// putFailsKV is a bucket whose writes under prefix fail.
type putFailsKV struct {
	fakeKV
	prefix string
}

func (f *putFailsKV) KVPut(ctx context.Context, b string, key string, v []byte) (uint64, error) {
	if len(key) >= len(f.prefix) && key[:len(f.prefix)] == f.prefix {
		return 0, errors.New("put failed")
	}

	return f.fakeKV.KVPut(ctx, b, key, v)
}

func (s *StorePublicTestSuite) TestWriteFailures() {
	s.Run("when the key cannot be removed revoke returns the error", func() {
		kv := &failingKV{fakeKV{ctrl: s.ctrl, data: map[string][]byte{}}}
		data, _ := json.Marshal(admission.AcceptedAgent{MachineID: "m-1"})
		kv.data[admission.AcceptedKVPrefix()+"m-1"] = data
		w := admission.NewWatcher(slog.Default(), s.nc, kv, "enrollment", nil, false, "osapi")

		s.ErrorContains(w.Revoke(context.Background(), "m-1"), "delete failed")
	})

	s.Run("when the revoked record cannot be stored revoke returns the error", func() {
		kv := &putFailsKV{fakeKV{ctrl: s.ctrl, data: map[string][]byte{}}, "revoked."}
		data, _ := json.Marshal(admission.AcceptedAgent{MachineID: "m-1"})
		kv.data[admission.AcceptedKVPrefix()+"m-1"] = data
		w := admission.NewWatcher(slog.Default(), s.nc, kv, "enrollment", nil, false, "osapi")

		s.ErrorContains(
			w.Revoke(context.Background(), "m-1"),
			"store revoked record m-1: put failed",
		)
	})

	s.Run("when the record cannot be stored reject returns the error", func() {
		s.pending("m-1", "web-01")
		kv := &failingKV{fakeKV{ctrl: s.ctrl, data: s.kv.data}}
		w := admission.NewWatcher(slog.Default(), s.nc, kv, "enrollment", nil, false, "osapi")

		s.ErrorContains(w.Reject(context.Background(), "m-1", "x"), "delete failed")
	})
}

// spanOf runs call on a watcher whose KV fails, under a request span, and
// checks that call opens span as a child of the request, marks it failed,
// and gives the KV calls its context.
func (s *StorePublicTestSuite) spanOf(
	name string,
	call func(w *admission.Watcher, ctx context.Context) error,
) {
	recorder := tracetest.NewSpanRecorder()
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	defer otel.SetTracerProvider(previous)

	ctrl := gomock.NewController(s.T())
	kv := admissionMocks.NewMockKVStore(ctrl)
	var inside []context.Context
	kv.EXPECT().KVKeys(gomock.Any(), "enrollment").
		DoAndReturn(func(ctx context.Context, _ string, _ ...string) ([]string, error) {
			inside = append(inside, ctx)
			return nil, errors.New("kv down")
		}).AnyTimes()
	kv.EXPECT().KVGet(gomock.Any(), "enrollment", gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _ string) (jetstream.KeyValueEntry, error) {
			inside = append(inside, ctx)
			return nil, errors.New("kv down")
		}).
		AnyTimes()
	w := admission.NewWatcher(slog.Default(), admissionMocks.NewMockNATSSubscriber(ctrl),
		kv, "enrollment", admissionMocks.NewMockPKIProvider(ctrl), false, "osapi")

	ctx, request := otel.Tracer("test").Start(context.Background(), "POST /api/admission")
	s.Error(call(w, ctx))
	request.End()

	var span sdktrace.ReadOnlySpan
	for _, sp := range recorder.Ended() {
		if sp.Name() == name {
			span = sp
		}
	}
	s.Require().NotNil(span)
	s.Equal(request.SpanContext().SpanID(), span.Parent().SpanID())
	s.Equal(codes.Error, span.Status().Code)
	s.Require().NotEmpty(inside)
	for _, c := range inside {
		s.Equal(span.SpanContext().SpanID(), trace.SpanContextFromContext(c).SpanID())
	}
}

// TestOperations: each operation of store.go opens its span.
func (s *StorePublicTestSuite) TestOperations() {
	tests := []struct {
		name string
		span string
		call func(w *admission.Watcher, ctx context.Context) error
	}{
		{
			name: "when the records are listed opens admission.list",
			span: "admission.list",
			call: func(w *admission.Watcher, ctx context.Context) error {
				_, err := w.List(ctx)
				return err
			},
		},
		{
			name: "when a record is read opens admission.get",
			span: "admission.get",
			call: func(w *admission.Watcher, ctx context.Context) error {
				_, err := w.Get(ctx, "m-1")
				return err
			},
		},
		{
			name: "when a machine is admitted opens admission.admit",
			span: "admission.admit",
			call: func(w *admission.Watcher, ctx context.Context) error {
				return w.Admit(ctx, "m-1", "SHA256:ab")
			},
		},
		{
			name: "when a machine is rejected opens admission.reject",
			span: "admission.reject",
			call: func(w *admission.Watcher, ctx context.Context) error {
				return w.Reject(ctx, "m-1", "no")
			},
		},
		{
			name: "when a machine is revoked opens admission.revoke",
			span: "admission.revoke",
			call: func(w *admission.Watcher, ctx context.Context) error {
				return w.Revoke(ctx, "m-1")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.spanOf(tc.span, tc.call)
		})
	}
}
