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

	"github.com/nats-io/nats.go/jetstream"
)

// ensure StreamStore implements Store at compile time.
var _ Store = (*StreamStore)(nil)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// fetchTimeout is the maximum wait time for fetching messages from a consumer.
const fetchTimeout = 5 * time.Second

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
func (s *StreamStore) Get(
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
func (s *StreamStore) List(
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

	// For newest-first pagination, we need to compute which sequence
	// range to fetch. Messages are stored oldest=FirstSeq to
	// newest=FirstSeq+Msgs-1. To get the N-th newest page:
	//   newestSeq = FirstSeq + Msgs - 1
	//   startSeq  = newestSeq - offset - limit + 1
	// We fetch `limit` messages starting at startSeq, then reverse.
	newestSeq := info.State.FirstSeq + info.State.Msgs - 1
	count := limit
	if offset+count > total {
		count = total - offset
	}

	startSeq := newestSeq - uint64(offset) - uint64(count) + 1

	entries, err := s.read(ctx, startSeq, count)
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
func (s *StreamStore) ListAll(
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

	entries, err := s.read(ctx, 0, total)
	if err != nil {
		return nil, err
	}

	// Reverse for newest-first ordering.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

// read reads count entries of the stream from the sequence from, or from the
// first when from is 0, oldest first. An entry that does not decode is left
// out. When the read stops with an error after some entries, the entries are
// kept and the error is logged; with no entry, the error is returned.
func (s *StreamStore) read(
	ctx context.Context,
	from uint64,
	count int,
) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	msgs, err := s.nc.StreamRead(ctx, s.stream, from, count)
	if err != nil && len(msgs) == 0 {
		return nil, fmt.Errorf("fetch audit entries: %w", err)
	}
	if err != nil {
		s.logger.Warn("read of the audit stream stopped", slog.String("error", err.Error()))
	}

	entries := make([]Entry, 0, len(msgs))
	for _, msg := range msgs {
		var entry Entry
		if err := json.Unmarshal(msg.Data(), &entry); err != nil {
			s.logger.Warn(
				"failed to unmarshal audit entry",
				slog.String("error", err.Error()),
			)

			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
