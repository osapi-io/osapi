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

// Start starts the worker and runs until the context is canceled.
func (w *Worker) Start(
	ctx context.Context,
) {
	w.logger.Info("starting job worker")
	w.run(ctx)
	w.logger.Info("job worker stopped")
}

// run contains the main worker loop.
func (w *Worker) run(
	ctx context.Context,
) {
	w.logger.Info("job worker started successfully")

	// Determine worker hostname (GetWorkerHostname always succeeds)
	hostname, _ := job.GetWorkerHostname(w.appConfig.Worker.Hostname)

	w.logger.Info(
		"worker configuration",
		slog.String("hostname", hostname),
		slog.String("queue_group", w.appConfig.Worker.QueueGroup),
		slog.Int("max_jobs", w.appConfig.Worker.MaxJobs),
	)

	// Start consuming messages for different job types
	w.wg.Add(2)

	// Consumer for query jobs (read operations)
	go func() {
		defer w.wg.Done()
		// consumeQueryJobs handles errors internally and always returns nil
		_ = w.consumeQueryJobs(ctx, hostname)
	}()

	// Consumer for modify jobs (write operations)
	go func() {
		defer w.wg.Done()
		// consumeModifyJobs handles errors internally and always returns nil
		_ = w.consumeModifyJobs(ctx, hostname)
	}()

	// Wait for cancellation
	<-ctx.Done()
	w.logger.Info("job worker shutting down")
}
