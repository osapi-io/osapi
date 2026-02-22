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

// Package export provides pluggable audit log export with pagination.
package export

import (
	"context"
	"fmt"
	"log/slog"
)

// ProgressFunc is called after each batch with the running exported count and total.
type ProgressFunc func(exported int, total int)

// Run paginates through audit entries and writes each to the exporter.
func Run(
	ctx context.Context,
	logger *slog.Logger,
	fetcher Fetcher,
	exporter Exporter,
	batchSize int,
	onProgress ProgressFunc,
) (*Result, error) {
	if err := exporter.Open(ctx); err != nil {
		return nil, fmt.Errorf("opening exporter: %w", err)
	}

	defer func() {
		if closeErr := exporter.Close(ctx); closeErr != nil {
			logger.Error("closing exporter", slog.String("error", closeErr.Error()))
		}
	}()

	result := &Result{}
	offset := 0

	for {
		entries, total, err := fetcher(ctx, batchSize, offset)
		if err != nil {
			return result, fmt.Errorf("fetching entries at offset %d: %w", offset, err)
		}

		result.TotalEntries = total

		for _, entry := range entries {
			if err := exporter.Write(ctx, entry); err != nil {
				return result, fmt.Errorf("writing entry: %w", err)
			}
			result.ExportedEntries++
		}

		if onProgress != nil {
			onProgress(result.ExportedEntries, total)
		}

		offset += len(entries)
		if offset >= total || len(entries) == 0 {
			break
		}
	}

	return result, nil
}
