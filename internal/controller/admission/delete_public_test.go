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
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/job"
)

// fakeSeen is the registry and the last reports that DeleteStale reads.
type fakeSeen struct {
	online  []string
	last    map[string]time.Time
	listErr error
	seenErr error
	calls   atomic.Int32
}

func (f *fakeSeen) ListAgents(context.Context) ([]job.AgentInfo, error) {
	f.calls.Add(1)
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]job.AgentInfo, 0, len(f.online))
	for _, id := range f.online {
		out = append(out, job.AgentInfo{MachineID: id})
	}

	return out, nil
}

func (f *fakeSeen) AgentsLastSeen(context.Context) (map[string]time.Time, error) {
	return f.last, f.seenErr
}

// deleteFailsKV is a bucket whose deletes under prefix fail.
type deleteFailsKV struct {
	fakeKV
	prefix string
}

func (f *deleteFailsKV) KVDelete(ctx context.Context, b string, key string) error {
	if strings.HasPrefix(key, f.prefix) {
		return errors.New("delete failed")
	}

	return f.fakeKV.KVDelete(ctx, b, key)
}

// admitted stores an admitted record of machineID, admitted at at.
func (s *StorePublicTestSuite) admitted(
	machineID string,
	at time.Time,
) {
	s.put(admission.AcceptedKVPrefix()+machineID, admission.AcceptedAgent{
		MachineID: machineID, Hostname: "web-" + machineID, Fingerprint: s.fp, AcceptedAt: at,
	})
}

func (s *StorePublicTestSuite) TestDelete() {
	s.Run("when the machine has records removes its key and each record", func() {
		s.admitted("m-1", s.at)
		s.put(admission.RevokedKVPrefix()+"m-1", admission.Record{MachineID: "m-1"})
		s.put(admission.RejectedKVPrefix()+"m-1", admission.Record{MachineID: "m-1"})
		s.pending("m-1", "web-01")
		s.admitted("m-2", s.at)

		s.Require().NoError(s.watcher.Delete(context.Background(), "m-1"))

		_, err := s.watcher.Get(context.Background(), "m-1")
		s.ErrorIs(err, admission.ErrNotFound)
		_, err = s.watcher.LookupAgentKey(context.Background(), "m-1")
		s.ErrorIs(err, admission.ErrAgentKeyNotFound)
		_, err = s.watcher.Get(context.Background(), "m-2")
		s.NoError(err)
	})

	s.Run("when a machine is deleted opens admission.delete", func() {
		s.spanOf("admission.delete", func(w *admission.Watcher, ctx context.Context) error {
			return w.Delete(ctx, "m-1")
		})
	})

	s.Run("when the machine has no record returns ErrNotFound", func() {
		s.ErrorIs(s.watcher.Delete(context.Background(), "m-x"), admission.ErrNotFound)
	})

	for _, prefix := range []string{admission.AcceptedKVPrefix(), admission.RevokedKVPrefix()} {
		s.Run("when the "+prefix+" record cannot be deleted returns the error", func() {
			kv := &deleteFailsKV{fakeKV{ctrl: s.ctrl, data: map[string][]byte{}}, prefix}
			data, _ := json.Marshal(admission.AcceptedAgent{MachineID: "m-1"})
			kv.data[admission.AcceptedKVPrefix()+"m-1"] = data
			w := admission.NewWatcher(slog.Default(), s.nc, kv, "enrollment", nil, false, "osapi")

			s.ErrorContains(w.Delete(context.Background(), "m-1"), "delete failed")
		})
	}
}

func (s *StorePublicTestSuite) TestDeleteStale() {
	after := time.Hour

	s.Run("when an admitted machine has not reported for after deletes it", func() {
		s.admitted("gone", s.at.Add(-48*time.Hour))
		s.admitted("old", s.at.Add(-48*time.Hour))
		s.admitted("new", s.at.Add(-time.Minute))
		s.admitted("up", s.at.Add(-48*time.Hour))
		s.admitted("recent", s.at.Add(-48*time.Hour))
		s.admitted("unknown", time.Time{})
		s.pending("asks", "web-asks")
		seen := &fakeSeen{
			online: []string{"up"},
			last: map[string]time.Time{
				"gone":   s.at.Add(-2 * time.Hour),
				"recent": s.at.Add(-time.Minute),
			},
		}

		log := &joinLog{}
		s.watcher.SetAudit(log)
		deleted, err := s.watcher.DeleteStale(context.Background(), seen, after)

		s.Require().NoError(err)
		s.Len(log.entries, 2)
		for _, e := range log.entries {
			s.Equal(audit.ActorSystem, e.Actor)
			s.Equal("delete_after", e.User)
			s.Equal("agent.delete", e.Action)
		}
		s.ElementsMatch(
			[]string{"gone", "old"},
			deleted,
		)
		for _, id := range []string{"up", "recent", "new", "unknown", "asks"} {
			_, err := s.watcher.Get(context.Background(), id)
			s.NoError(err)
		}
	})

	s.Run("when the registry does not read deletes nothing", func() {
		s.admitted("gone", s.at.Add(-48*time.Hour))

		_, err := s.watcher.DeleteStale(
			context.Background(),
			&fakeSeen{listErr: errors.New("kv down")},
			after,
		)

		s.ErrorContains(err, "list agents")
		_, err = s.watcher.Get(context.Background(), "gone")
		s.NoError(err)
	})

	s.Run("when the last reports do not read deletes nothing", func() {
		s.admitted("gone", s.at.Add(-48*time.Hour))

		_, err := s.watcher.DeleteStale(
			context.Background(),
			&fakeSeen{seenErr: errors.New("kv down")},
			after,
		)

		s.ErrorContains(err, "read last seen")
	})

	s.Run("when the records do not read returns the error", func() {
		s.kv.keysErr = errors.New("kv down")

		_, err := s.watcher.DeleteStale(context.Background(), &fakeSeen{}, after)

		s.Error(err)
	})

	s.Run("when a delete fails logs it and goes on", func() {
		kv := &deleteFailsKV{
			fakeKV{ctrl: s.ctrl, data: map[string][]byte{}},
			admission.AcceptedKVPrefix(),
		}
		data, _ := json.Marshal(
			admission.AcceptedAgent{MachineID: "m-1", AcceptedAt: s.at.Add(-48 * time.Hour)},
		)
		kv.data[admission.AcceptedKVPrefix()+"m-1"] = data
		w := admission.NewWatcher(slog.Default(), s.nc, kv, "enrollment", nil, false, "osapi")

		deleted, err := w.DeleteStale(context.Background(), &fakeSeen{}, after)

		s.NoError(err)
		s.Empty(deleted)
	})
}

func (s *StorePublicTestSuite) TestRunStaleCleanup() {
	s.Run("when it runs looks each interval and stops with the context", func() {
		admission.SetStaleInterval(time.Millisecond)
		defer admission.ResetStaleInterval()
		s.admitted("gone", s.at.Add(-48*time.Hour))
		seen := &fakeSeen{listErr: errors.New("kv down")}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error)

		go func() { done <- s.watcher.RunStaleCleanup(ctx, seen, time.Hour) }()
		s.Eventually(func() bool { return seen.calls.Load() >= 2 }, time.Second, time.Millisecond)
		cancel()

		s.NoError(<-done)
	})
}
