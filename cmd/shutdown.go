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

package cmd

import (
	"context"
	"sync"
	"time"
)

// shutdownTimeout bounds everything that a shutdown does: the stop of each
// component, then each cleanup, with one context.
//
// The cleanups flush the metrics provider and the OTLP exporter. An
// unreachable collector makes that flush block. One bounded context means that
// SIGTERM always ends the process.
const shutdownTimeout = 10 * time.Second

// component is a part of the process that a start command runs: the controller,
// the agent or the NATS server.
type component interface {
	// Start starts the component without blocking.
	Start()
	// Stop asks the component to stop.
	Stop(ctx context.Context)
}

// serve starts each component and blocks until ctx ends. Then it stops the
// components at the same time, so that one slow component does not use the
// time of the others, and runs each cleanup in order. The stops and the cleanups get
// one context that ends after shutdownTimeout.
func serve(
	ctx context.Context,
	components []component,
	cleanups ...func(context.Context),
) {
	for _, c := range components {
		c.Start()
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		shutdownTimeout,
	)
	defer cancel()

	var wg sync.WaitGroup
	for _, c := range components {
		wg.Go(func() { c.Stop(shutdownCtx) })
	}
	wg.Wait()

	for _, cleanup := range cleanups {
		cleanup(shutdownCtx)
	}
}
