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
	"maps"
	"slices"
	"sync"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// jobIDSinkKey is the context key the sink is carried under.
type jobIDSinkKey struct{}

// jobIDSink collects the jobs a single request created. The audit middleware puts
// one in the request context before the handler runs, and the job client fills it
// in as it dispatches, which is how an entry learns a job ID without thirty
// handlers each remembering to report one.
//
// The job client also gives it the steps of each host that answers the job,
// so an entry records the chain of behavior calls that the request made.
//
// It is mutable state shared between the request goroutine and whatever the
// client does, so it is guarded.
type jobIDSink struct {
	mu    sync.Mutex
	id    string
	steps map[string][]client.Step
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

// RecordSteps notes the steps of one host that answered the job of a request.
// Called by the job client for each response that it collects, also when the
// job failed on the host. Only the steps of the job that RecordJobID recorded
// are kept, and a context with no sink is left alone.
func RecordSteps(
	ctx context.Context,
	jobID string,
	hostname string,
	steps []client.Step,
) {
	sink, ok := ctx.Value(jobIDSinkKey{}).(*jobIDSink)
	if !ok || len(steps) == 0 {
		return
	}

	sink.mu.Lock()
	defer sink.mu.Unlock()

	if sink.id == "" || sink.id != jobID {
		return
	}

	if sink.steps == nil {
		sink.steps = make(map[string][]client.Step)
	}
	sink.steps[hostname] = steps
}

// StepsFrom returns the steps of each host that answered the job of a
// request, sorted by hostname, or nil when no host made a call.
func StepsFrom(
	ctx context.Context,
) []HostSteps {
	sink, ok := ctx.Value(jobIDSinkKey{}).(*jobIDSink)
	if !ok {
		return nil
	}

	sink.mu.Lock()
	defer sink.mu.Unlock()

	if len(sink.steps) == 0 {
		return nil
	}

	out := make([]HostSteps, 0, len(sink.steps))
	for _, hostname := range slices.Sorted(maps.Keys(sink.steps)) {
		out = append(out, HostSteps{Hostname: hostname, Steps: sink.steps[hostname]})
	}

	return out
}
