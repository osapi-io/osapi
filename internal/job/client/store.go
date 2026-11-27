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
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// codeStoreFailed is the JetStream error code of a write that the stream's
// store refused, such as a write past a bucket's max_bytes. The nats.go
// jetstream package names no constant for it.
const codeStoreFailed jetstream.ErrorCode = 10077

// evictShare is the share of the bucket's jobs that one eviction removes,
// of the finished ones, oldest first: a third. One job at a time would
// free room for one write, and the next write would be full again. A third
// takes a full bucket below the 80% at which health status warns, so the
// warning clears after a removal.
const evictShare = 3

// storeFull reports whether err is a write that the jobs bucket refused
// because it is at its max_bytes. A KV bucket discards new writes, not old
// ones, so it stays full until something removes old messages.
func storeFull(
	err error,
) bool {
	if err == nil {
		return false
	}
	var apiErr *jetstream.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode == codeStoreFailed {
		return true
	}

	return strings.Contains(err.Error(), "maximum bytes exceeded")
}

// putJob writes a job's record to the jobs bucket. When the bucket is
// full, it removes the oldest finished jobs and writes again. When nothing
// can be removed, the error wraps job.ErrStoreFull, and the log has the
// cause.
func (c *Client) putJob(
	ctx context.Context,
	key string,
	value []byte,
) (uint64, error) {
	rev, err := c.natsClient.KVPut(ctx, c.jobsBucket, key, value)
	if !storeFull(err) {
		return rev, err
	}
	if c.purger == nil {
		return 0, c.full(ctx, err)
	}

	// One write at a time frees room. A write that waited here tries
	// again first, because the write before it may have freed room.
	c.evictMu.Lock()
	defer c.evictMu.Unlock()
	if rev, err = c.natsClient.KVPut(ctx, c.jobsBucket, key, value); !storeFull(err) {
		return rev, err
	}

	evicted, evictErr := c.evictFinished(ctx)
	switch {
	case evictErr != nil:
		return 0, c.full(ctx, evictErr)
	case evicted == 0:
		return 0, c.full(ctx, fmt.Errorf("no finished job to remove: %w", err))
	}

	rev, err = c.natsClient.KVPut(ctx, c.jobsBucket, key, value)
	if storeFull(err) {
		return 0, c.full(ctx, err)
	}

	return rev, err
}

// full logs a write that the jobs bucket refused, with its cause, and
// returns an error that wraps job.ErrStoreFull.
func (c *Client) full(
	ctx context.Context,
	cause error,
) error {
	c.logger.ErrorContext(
		ctx, "the job store is full",
		slog.String("bucket", c.jobsBucket),
		slog.String("error", cause.Error()),
	)

	return fmt.Errorf("%w: %w", job.ErrStoreFull, cause)
}

// finished reports whether a job with status has ended on every host, so
// removing it loses no work that is in flight.
func finished(
	status types.JobStatus,
) bool {
	switch status {
	case types.JobStatusCompleted,
		types.JobStatusFailed,
		types.JobStatusPartialFailure,
		types.JobStatusSkipped:
		return true
	default:
		return false
	}
}

// evictFinished removes the oldest finished jobs, a third of the jobs in
// the bucket, and at least one: each job's record, its status events and
// its responses. A job that is submitted or processing stays. It returns
// how many jobs it removed.
func (c *Client) evictFinished(
	ctx context.Context,
) (int, error) {
	keys, err := c.natsClient.KVKeys(ctx, c.jobsBucket)
	if err != nil {
		return 0, fmt.Errorf("list the jobs to remove: %w", err)
	}

	// Newest first, so the oldest are at the end.
	ids, statuses := computeStatusFromKeyNames(keys)
	want := max(len(ids)/evictShare, 1)
	stream := "KV_" + c.jobsBucket
	prefix := "$KV." + c.jobsBucket + "."

	evicted := 0
	for i := len(ids) - 1; i >= 0 && evicted < want; i-- {
		id := ids[i]
		info, ok := statuses[id]
		if !ok || !finished(info.Status) {
			continue
		}
		for _, subject := range []string{
			prefix + "status." + id + ".>",
			prefix + "responses." + id + ".>",
			prefix + "jobs." + id,
		} {
			if err := c.purger.PurgeSubject(ctx, stream, subject); err != nil {
				return evicted, fmt.Errorf("remove job %s: %w", id, err)
			}
		}
		evicted++
	}

	c.logger.WarnContext(
		ctx, "the job store is full: removed the oldest finished jobs",
		slog.Int("removed", evicted),
		slog.Int("jobs", len(ids)),
	)

	return evicted, nil
}
