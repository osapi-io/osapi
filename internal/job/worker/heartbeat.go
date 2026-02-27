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

package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// heartbeatInterval is the interval between heartbeat refreshes.
const heartbeatInterval = 10 * time.Second

// startHeartbeat writes the initial registration, spawns a goroutine that
// refreshes the entry on a ticker, and deregisters on ctx.Done().
func (w *Worker) startHeartbeat(
	ctx context.Context,
	hostname string,
) {
	if w.registryKV == nil {
		return
	}

	ttl := w.appConfig.NATS.Registry.TTL
	key := registryKey(hostname)

	w.writeRegistration(ctx, hostname)

	w.logger.Info(
		"registered in worker registry",
		slog.String("hostname", hostname),
		slog.String("key", key),
		slog.String("ttl", ttl),
	)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.deregister(hostname)
				return
			case <-ticker.C:
				w.writeRegistration(ctx, hostname)

				w.logger.Info(
					"heartbeat refreshed",
					slog.String("hostname", hostname),
					slog.String("key", key),
					slog.String("next_in", heartbeatInterval.String()),
				)
			}
		}
	}()
}

// writeRegistration marshals a WorkerRegistration and puts it to the registry KV.
func (w *Worker) writeRegistration(
	ctx context.Context,
	hostname string,
) {
	reg := job.WorkerRegistration{
		Hostname:     hostname,
		Labels:       w.appConfig.Job.Worker.Labels,
		RegisteredAt: time.Now(),
	}

	data, err := json.Marshal(reg)
	if err != nil {
		w.logger.Warn(
			"failed to marshal worker registration",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := registryKey(hostname)
	if _, err := w.registryKV.Put(ctx, key, data); err != nil {
		w.logger.Warn(
			"failed to write worker registration",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// deregister deletes the worker's registration key on clean shutdown.
func (w *Worker) deregister(
	hostname string,
) {
	key := registryKey(hostname)
	if err := w.registryKV.Delete(context.Background(), key); err != nil {
		w.logger.Warn(
			"failed to deregister worker",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	w.logger.Info(
		"worker deregistered",
		slog.String("hostname", hostname),
		slog.String("key", key),
	)
}

// registryKey returns the KV key for a worker's registration entry.
func registryKey(
	hostname string,
) string {
	return "workers." + job.SanitizeHostname(hostname)
}
