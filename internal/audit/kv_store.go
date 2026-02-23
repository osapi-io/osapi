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
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/nats-io/nats.go/jetstream"
)

// ensure KVStore implements Store at compile time.
var _ Store = (*KVStore)(nil)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// KVStore implements Store backed by a NATS KeyValue bucket.
type KVStore struct {
	kv     jetstream.KeyValue
	logger *slog.Logger
}

// NewKVStore creates a new KVStore.
func NewKVStore(
	logger *slog.Logger,
	kv jetstream.KeyValue,
) *KVStore {
	return &KVStore{
		kv:     kv,
		logger: logger,
	}
}

// Write persists an audit entry to the KV bucket.
func (s *KVStore) Write(
	ctx context.Context,
	entry Entry,
) error {
	data, err := marshalJSON(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	if _, err := s.kv.Put(ctx, entry.ID, data); err != nil {
		return fmt.Errorf("put audit entry: %w", err)
	}

	return nil
}

// Get retrieves a single audit entry by ID.
func (s *KVStore) Get(
	ctx context.Context,
	id string,
) (*Entry, error) {
	kve, err := s.kv.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get audit entry: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(kve.Value(), &entry); err != nil {
		return nil, fmt.Errorf("unmarshal audit entry: %w", err)
	}

	return &entry, nil
}

// List retrieves audit entries with pagination.
func (s *KVStore) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]Entry, int, error) {
	keys, err := s.kv.Keys(ctx)
	if err != nil {
		// jetstream.ErrNoKeysFound means the bucket is empty
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return []Entry{}, 0, nil
		}
		return nil, 0, fmt.Errorf("list audit keys: %w", err)
	}

	total := len(keys)

	// Sort descending (newest first — ULIDs/UUIDs with timestamp prefix sort naturally)
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	// Apply pagination
	if offset >= total {
		return []Entry{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	pageKeys := keys[offset:end]

	entries := make([]Entry, 0, len(pageKeys))
	for _, key := range pageKeys {
		kve, err := s.kv.Get(ctx, key)
		if err != nil {
			s.logger.Warn(
				"failed to get audit entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		var entry Entry
		if err := json.Unmarshal(kve.Value(), &entry); err != nil {
			s.logger.Warn(
				"failed to unmarshal audit entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, total, nil
}

// ListAll retrieves all audit entries without pagination.
func (s *KVStore) ListAll(
	ctx context.Context,
) ([]Entry, error) {
	keys, err := s.kv.Keys(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("list audit keys: %w", err)
	}

	// Sort descending (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	entries := make([]Entry, 0, len(keys))
	for _, key := range keys {
		kve, err := s.kv.Get(ctx, key)
		if err != nil {
			s.logger.Warn(
				"failed to get audit entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		var entry Entry
		if err := json.Unmarshal(kve.Value(), &entry); err != nil {
			s.logger.Warn(
				"failed to unmarshal audit entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
