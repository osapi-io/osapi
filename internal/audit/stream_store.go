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
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"go.opentelemetry.io/otel/attribute"

	"github.com/nats-io/nats.go/jetstream"
)

// ensure StreamStore implements Store at compile time.
var _ Store = (*StreamStore)(nil)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// fetchTimeout is the longest wait for one page of a read of the stream.
const fetchTimeout = 5 * time.Second

// readPage is the most entries that one read of the stream asks for. A large
// export is many reads, and each one has its own fetchTimeout.
const readPage = 500

// Stream is the access to NATS that StreamStore uses: it publishes each entry
// to the audit stream, and reads the stream back. A *natsclient.Client is one.
type Stream interface {
	Publish(ctx context.Context, subject string, data []byte) error
	StreamInfo(ctx context.Context, name string) (*jetstream.StreamInfo, error)
	StreamLastMsg(ctx context.Context, name string, subject string) (*jetstream.RawStreamMsg, error)
	StreamRead(ctx context.Context, name string, from uint64, count int) ([]jetstream.Msg, error)
}

// StreamStore implements Store backed by a NATS JetStream stream.
type StreamStore struct {
	logger  *slog.Logger
	nc      Stream
	stream  string
	subject string
}

// NewStreamStore creates a new StreamStore.
func NewStreamStore(
	logger *slog.Logger,
	nc Stream,
	stream string,
	subject string,
) *StreamStore {
	return &StreamStore{
		logger:  logger.With(slog.String("subsystem", "controller.audit.store")),
		nc:      nc,
		stream:  stream,
		subject: subject,
	}
}

// Write persists an audit entry by publishing to the stream.
func (s *StreamStore) Write(
	ctx context.Context,
	entry Entry,
) error {
	data, err := marshalJSON(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	subject := s.subject + "." + entry.ID
	if err := s.nc.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish audit entry: %w", err)
	}

	return nil
}

// Get retrieves a single audit entry by ID.
func (s *StreamStore) get(
	ctx context.Context,
	id string,
) (*Entry, error) {
	subject := s.subject + "." + id

	msg, err := s.nc.StreamLastMsg(ctx, s.stream, subject)
	if err != nil {
		return nil, fmt.Errorf("get audit entry %s: %w: %w", id, ErrNotFound, err)
	}

	var entry Entry
	if err := json.Unmarshal(msg.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal audit entry: %w", err)
	}

	return &entry, nil
}

// List retrieves audit entries with pagination, newest-first.
// Returns the entries, total count, and any error.
func (s *StreamStore) list(
	ctx context.Context,
	limit int,
	offset int,
) ([]Entry, int, error) {
	info, err := s.nc.StreamInfo(ctx, s.stream)
	if err != nil {
		return nil, 0, fmt.Errorf("get stream info: %w", err)
	}

	total := int(info.State.Msgs)
	if total == 0 {
		return []Entry{}, 0, nil
	}

	if offset >= total {
		return []Entry{}, total, nil
	}

	// For newest-first pagination, fetch the page that ends offset
	// entries before the newest. The newest is LastSeq, not
	// FirstSeq+Msgs-1: a deleted entry leaves a gap in the sequences, and
	// then the two differ. A page that spans a gap has fewer entries.
	count := limit
	if offset+count > total {
		count = total - offset
	}

	startSeq := info.State.LastSeq - uint64(offset) - uint64(count) + 1

	entries, err := s.read(ctx, startSeq, info.State.LastSeq, count)
	if err != nil {
		return nil, 0, err
	}

	// Reverse for newest-first ordering.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, total, nil
}

// ListAll retrieves all audit entries without pagination, newest-first.
func (s *StreamStore) listAll(
	ctx context.Context,
) ([]Entry, error) {
	info, err := s.nc.StreamInfo(ctx, s.stream)
	if err != nil {
		return nil, fmt.Errorf("get stream info: %w", err)
	}

	total := int(info.State.Msgs)
	if total == 0 {
		return []Entry{}, nil
	}

	entries, err := s.read(ctx, info.State.FirstSeq, info.State.LastSeq, total)
	if err != nil {
		return nil, err
	}

	// Reverse for newest-first ordering.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

// read reads at most count entries of the stream, oldest first, from the
// sequence from up to the sequence last: the last of the stream when the
// caller read its info. It reads in pages of readPage. A read that stops
// before last is an error, so an export of part of the log never looks like
// all of it. An entry that does not decode is left out, with a warning.
func (s *StreamStore) read(
	ctx context.Context,
	from uint64,
	last uint64,
	count int,
) ([]Entry, error) {
	entries := make([]Entry, 0, count)
	next := from
	var seen uint64

	for got := 0; got < count && seen < last; {
		pageCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
		msgs, err := s.nc.StreamRead(pageCtx, s.stream, next, min(count-got, readPage))
		cancel()

		for _, msg := range msgs {
			meta, merr := msg.Metadata()
			if merr != nil {
				return nil, fmt.Errorf("fetch audit entries: %w", merr)
			}
			seen = meta.Sequence.Stream
			if seen > last {
				break
			}
			got++

			var entry Entry
			if err := json.Unmarshal(msg.Data(), &entry); err != nil {
				s.logger.Warn(
					"failed to unmarshal audit entry",
					slog.Uint64("sequence", seen),
					slog.String("error", err.Error()),
				)

				continue
			}

			entries = append(entries, entry)
		}

		if seen >= last {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(
				"fetch audit entries: the read stopped at sequence %d of %d: %w",
				seen, last, err)
		}
		if len(msgs) == 0 {
			return nil, fmt.Errorf(
				"fetch audit entries: the read stopped at sequence %d of %d",
				seen, last)
		}
		next = seen + 1
	}

	return entries, nil
}

// Get returns one audit entry, in an audit.get span.
func (s *StreamStore) Get(
	ctx context.Context,
	id string,
) (*Entry, error) {
	ctx, span := opspan.Start(ctx, tracerName, "audit.get",
		attribute.String("osapi.audit.id", id),
	)
	entry, err := s.get(ctx, id)

	return entry, opspan.End(span, err)
}

// List returns a page of audit entries and the total, in an audit.list
// span.
func (s *StreamStore) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]Entry, int, error) {
	ctx, span := opspan.Start(ctx, tracerName, "audit.list")
	entries, total, err := s.list(ctx, limit, offset)
	if err == nil {
		span.SetAttributes(attribute.Int("osapi.audit.count", len(entries)))
	}

	return entries, total, opspan.End(span, err)
}

// ListAll returns every audit entry, for an export, in an audit.export span.
func (s *StreamStore) ListAll(
	ctx context.Context,
) ([]Entry, error) {
	ctx, span := opspan.Start(ctx, tracerName, "audit.export")
	entries, err := s.listAll(ctx)
	if err == nil {
		span.SetAttributes(attribute.Int("osapi.audit.count", len(entries)))
	}

	return entries, opspan.End(span, err)
}

// tracerName names the spans of the audit store in a trace.
const tracerName = "osapi-audit"
