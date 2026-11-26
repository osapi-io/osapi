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

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// advisory is the max deliveries advisory that NATS writes to the dead
// letter stream: io.nats.jetstream.advisory.v1.max_deliver.
type advisory struct {
	Timestamp  time.Time `json:"timestamp"`
	Consumer   string    `json:"consumer"`
	StreamSeq  uint64    `json:"stream_seq"`
	Deliveries uint64    `json:"deliveries"`
}

// ListDeadLetters is up to limit of the newest dead letters, newest first,
// each with the job that it carried, and the count of them all, in a
// job.deadletter.list span.
func (c *Client) ListDeadLetters(
	ctx context.Context,
	limit int,
) ([]job.DeadLetter, int, error) {
	ctx, span := opspan.Start(ctx, tracerName, "job.deadletter.list",
		attribute.Int("osapi.limit", limit))
	out, total, err := c.listDeadLetters(ctx, limit)
	span.SetAttributes(attribute.Int("osapi.count", len(out)))

	return out, total, opspan.End(span, err)
}

// listDeadLetters reads the dead letter stream from its last message back.
// A sequence that the stream no longer has, such as one past its max age,
// is skipped.
func (c *Client) listDeadLetters(
	ctx context.Context,
	limit int,
) ([]job.DeadLetter, int, error) {
	dlq := c.streamName + "-DLQ"
	info, err := c.natsClient.StreamInfo(ctx, dlq)
	if err != nil {
		return nil, 0, fmt.Errorf("read the dead letter stream %s: %w", dlq, err)
	}
	out := []job.DeadLetter{}
	first, last := info.State.FirstSeq, info.State.LastSeq
	for seq := last; seq >= first && seq > 0 && len(out) < limit; seq-- {
		msg, err := c.natsClient.StreamGetMsg(ctx, dlq, seq)
		if errors.Is(err, jetstream.ErrMsgNotFound) {
			continue
		}
		if err != nil {
			return nil, 0, fmt.Errorf("read dead letter %d: %w", seq, err)
		}
		var a advisory
		if err := json.Unmarshal(msg.Data, &a); err != nil {
			return nil, 0, fmt.Errorf("dead letter %d: %w", seq, err)
		}
		dl := job.DeadLetter{
			Seq: seq, Time: a.Timestamp, Consumer: a.Consumer,
			Deliveries: a.Deliveries, StreamSeq: a.StreamSeq,
		}
		// The job stream's message is the job's ID. A message that the
		// stream no longer has leaves the ID empty.
		if orig, err := c.natsClient.StreamGetMsg(ctx, c.streamName, a.StreamSeq); err == nil {
			dl.JobID = string(orig.Data)
		}
		out = append(out, dl)
	}

	return out, int(info.State.Msgs), nil
}
