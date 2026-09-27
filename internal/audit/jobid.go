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

package audit

import (
	"context"
	"sync"
)

// jobIDSinkKey is the context key the sink is carried under.
type jobIDSinkKey struct{}

// jobIDSink collects the jobs a single request created. The audit middleware puts
// one in the request context before the handler runs, and the job client fills it
// in as it dispatches, which is how an entry learns a job ID without thirty
// handlers each remembering to report one.
//
// It is mutable state shared between the request goroutine and whatever the
// client does, so it is guarded.
type jobIDSink struct {
	mu sync.Mutex
	id string
}

// WithJobIDSink returns a context that collects the job ID a request creates.
func WithJobIDSink(
	ctx context.Context,
) context.Context {
	return context.WithValue(ctx, jobIDSinkKey{}, &jobIDSink{})
}

// RecordJobID notes the job a request created. Called by the job client for every
// dispatch; a context with no sink — a background task, an agent, a test — is
// left alone.
//
// The first job wins. A broadcast is one job across many agents, and a request
// that somehow dispatches twice is better described by the job it started with
// than by the last one it happened to make.
func RecordJobID(
	ctx context.Context,
	jobID string,
) {
	sink, ok := ctx.Value(jobIDSinkKey{}).(*jobIDSink)
	if !ok || jobID == "" {
		return
	}

	sink.mu.Lock()
	defer sink.mu.Unlock()

	if sink.id == "" {
		sink.id = jobID
	}
}

// JobIDFrom returns the job a request created, or the empty string when it
// created none.
func JobIDFrom(
	ctx context.Context,
) string {
	sink, ok := ctx.Value(jobIDSinkKey{}).(*jobIDSink)
	if !ok {
		return ""
	}

	sink.mu.Lock()
	defer sink.mu.Unlock()

	return sink.id
}
