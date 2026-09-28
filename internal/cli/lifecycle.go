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

package cli

import (
	"context"
	"time"
)

// Lifecycle represents a long-running server or agent.
type Lifecycle interface {
	// Start starts the server without blocking.
	Start()
	// Stop gracefully shuts down the server.
	Stop(ctx context.Context)
}

// ShutdownTimeout bounds everything a shutdown does: stopping the server and
// then every cleanup function, against one clock.
//
// The cleanups are where the network is: a metrics provider and an OTLP exporter
// both flush on the way out, and an unreachable collector makes that flush block.
// Handing them a context.Background() — which is what they used to get — is how
// SIGTERM stops terminating the process.
const ShutdownTimeout = 10 * time.Second

// RunServer blocks until ctx is cancelled, then shuts down the server and runs
// the cleanup functions, all within ShutdownTimeout.
//
// Cleanup receives the same context the server did, so one budget covers the
// whole shutdown rather than each phase having its own and the total being
// unbounded. A cleanup that ignores its context can still overrun; the ones here
// pass it to the calls that block.
func RunServer(
	ctx context.Context,
	server Lifecycle,
	cleanupFns ...func(context.Context),
) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		ShutdownTimeout,
	)
	defer cancel()

	server.Stop(shutdownCtx)

	for _, fn := range cleanupFns {
		fn(shutdownCtx)
	}
}
