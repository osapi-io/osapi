// Copyright (c) 2025 John Dewey

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
	"log/slog"

	"github.com/retr0h/osapi/internal/job"
)

// Start starts the worker without blocking. Call Stop to shut down.
func (w *Worker) Start() {
	w.ctx, w.cancel = context.WithCancel(context.Background())

	w.logger.Info("starting job worker")

	// Determine worker hostname (GetWorkerHostname always succeeds)
	hostname, _ := job.GetWorkerHostname(w.appConfig.Job.Worker.Hostname)

	w.logger.Info(
		"worker configuration",
		slog.String("hostname", hostname),
		slog.String("queue_group", w.appConfig.Job.Worker.QueueGroup),
		slog.Int("max_jobs", w.appConfig.Job.Worker.MaxJobs),
		slog.Any("labels", w.appConfig.Job.Worker.Labels),
	)

	// Register in worker registry and start heartbeat keepalive.
	w.startHeartbeat(w.ctx, hostname)

	// Start consuming messages for different job types.
	// Each consume function spawns goroutines tracked by w.wg.
	_ = w.consumeQueryJobs(w.ctx, hostname)
	_ = w.consumeModifyJobs(w.ctx, hostname)

	w.logger.Info("job worker started successfully")
}

// Stop gracefully shuts down the worker, waiting for in-flight jobs
// to finish or the context deadline to expire.
func (w *Worker) Stop(
	ctx context.Context,
) {
	w.logger.Info("job worker shutting down")
	w.cancel()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("job worker stopped gracefully")
	case <-ctx.Done():
		w.logger.Warn("job worker shutdown timed out")
	}
}
