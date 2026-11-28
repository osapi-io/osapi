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

package admission

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// Seen tells whether an admitted machine still reports: the agents in the
// registry now, and the last report of each machine.
type Seen interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
	AgentsLastSeen(ctx context.Context) (map[string]time.Time, error)
}

// Delete removes a machine from admission, in an admission.delete span: its
// key, and each record of it. Unlike revoke, it keeps no record, so a machine
// that asks again is new. It is for a machine that goes away on purpose,
// such as a VM that is rebuilt.
func (w *Watcher) Delete(
	ctx context.Context,
	machineID string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "admission.delete",
		attribute.String("osapi.agent.machine_id", machineID),
	)

	hostname := w.hostnameOf(ctx, machineID)
	if err := w.deleteMachine(ctx, machineID); err != nil {
		return opspan.End(span, err)
	}
	w.event(ctx, machineID, hostname, "deleted", "deleted by an operator")

	return opspan.End(span, nil)
}

// deleteMachine removes the admitted key and each other record of
// machineID. A machine with no record is ErrNotFound.
func (w *Watcher) deleteMachine(
	ctx context.Context,
	machineID string,
) error {
	if _, err := w.get(ctx, machineID); err != nil {
		return err
	}
	if err := w.RemoveAgentKey(ctx, machineID); err != nil {
		return err
	}
	// RemoveAgentKey deleted the admitted record; these are the others.
	for _, prefix := range []string{kvPrefix, rejectedKVPrefix, revokedKVPrefix} {
		if err := w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, prefix+machineID); err != nil {
			return fmt.Errorf("delete record of %s: %w", machineID, err)
		}
	}

	return nil
}

// DeleteStale deletes each admitted machine that has not reported for
// after: it is not in the registry, and its last report, or its admission
// when it never reported, is older than after. It returns the machine IDs
// that it deleted. When the registry or the last reports do not read, it
// deletes nothing: a machine that cannot be seen is not a machine that is
// gone.
func (w *Watcher) DeleteStale(
	ctx context.Context,
	seen Seen,
	after time.Duration,
) ([]string, error) {
	agents, err := seen.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	last, err := seen.AgentsLastSeen(ctx)
	if err != nil {
		return nil, fmt.Errorf("read last seen: %w", err)
	}
	records, err := w.list(ctx)
	if err != nil {
		return nil, err
	}

	online := make(map[string]bool, len(agents))
	for _, a := range agents {
		online[a.MachineID] = true
	}

	now := nowFn()
	var deleted []string
	for _, r := range records {
		if r.State != types.AdmissionAdmitted || online[r.MachineID] {
			continue
		}
		at := last[r.MachineID]
		if at.IsZero() {
			at = r.AdmittedAt
		}
		if at.IsZero() || now.Sub(at) < after {
			continue
		}
		if err := w.deleteMachine(ctx, r.MachineID); err != nil {
			w.logger.Warn("failed to delete a stale machine",
				slog.String("machine_id", r.MachineID),
				slog.String("error", err.Error()),
			)

			continue
		}
		w.logger.Info("deleted a machine that has not reported",
			slog.String("machine_id", r.MachineID),
			slog.String("hostname", r.Hostname),
			slog.Time("last_seen", at),
			slog.Duration("delete_after", after),
		)
		w.event(ctx, r.MachineID, r.Hostname, "deleted",
			"deleted by delete_after: not seen since "+at.UTC().Format(time.RFC3339))
		w.record(ctx, audit.Change(ctx, audit.ActorSystem, "delete_after", "agent.delete",
			r.MachineID, http.StatusOK,
			"hostname="+r.Hostname+" last_seen="+at.UTC().Format(time.RFC3339)))
		deleted = append(deleted, r.MachineID)
	}

	return deleted, nil
}

// staleInterval is how often RunStaleCleanup looks. The agents write their
// last report once a minute, so a shorter interval sees nothing new.
var staleInterval = time.Minute

// RunStaleCleanup runs DeleteStale every staleInterval until ctx is done.
func (w *Watcher) RunStaleCleanup(
	ctx context.Context,
	seen Seen,
	after time.Duration,
) error {
	w.logger.Info("stale machine cleanup started", slog.Duration("delete_after", after))

	ticker := time.NewTicker(staleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.cleanOnce(ctx, seen, after)
		}
	}
}

// cleanOnce is one pass of the stale cleanup, in an admission.stale_cleanup
// span of its own: a timer starts it, not a request.
func (w *Watcher) cleanOnce(
	ctx context.Context,
	seen Seen,
	after time.Duration,
) {
	ctx, span := opspan.Root(ctx, tracerName, "admission.stale_cleanup")
	deleted, err := w.DeleteStale(ctx, seen, after)
	span.SetAttributes(attribute.Int("osapi.admission.deleted", len(deleted)))
	if opspan.End(span, err) != nil {
		w.logger.Warn("stale machine cleanup skipped", slog.String("error", err.Error()))
	}
}
