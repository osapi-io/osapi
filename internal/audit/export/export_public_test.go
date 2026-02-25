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

package export_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	gen "github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/audit/export"
)

type ExportPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	logger *slog.Logger
}

func (suite *ExportPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.logger = slog.Default()
}

func (suite *ExportPublicTestSuite) newEntry(
	user string,
) gen.AuditEntry {
	return gen.AuditEntry{
		Id:           uuid.New(),
		Timestamp:    time.Date(2026, 2, 21, 10, 30, 0, 0, time.UTC),
		User:         user,
		Roles:        []string{"admin"},
		Method:       "GET",
		Path:         "/system/hostname",
		SourceIp:     "127.0.0.1",
		ResponseCode: 200,
		DurationMs:   42,
	}
}

func (suite *ExportPublicTestSuite) TestRun() {
	tests := []struct {
		name         string
		fetcher      export.Fetcher
		exporter     *mockExporter
		batchSize    int
		validateFunc func(exp *mockExporter, result *export.Result, err error)
	}{
		{
			name: "when no entries returns zero counts",
			fetcher: func(_ context.Context, _, _ int) ([]gen.AuditEntry, int, error) {
				return nil, 0, nil
			},
			exporter:  &mockExporter{},
			batchSize: 100,
			validateFunc: func(exp *mockExporter, result *export.Result, err error) {
				suite.NoError(err)
				suite.Equal(0, result.TotalEntries)
				suite.Equal(0, result.ExportedEntries)
				suite.True(exp.opened)
				suite.True(exp.closed)
			},
		},
		{
			name: "when single page exports all entries",
			fetcher: func(_ context.Context, _, _ int) ([]gen.AuditEntry, int, error) {
				return []gen.AuditEntry{
					suite.newEntry("alice@example.com"),
					suite.newEntry("bob@example.com"),
				}, 2, nil
			},
			exporter:  &mockExporter{},
			batchSize: 100,
			validateFunc: func(exp *mockExporter, result *export.Result, err error) {
				suite.NoError(err)
				suite.Equal(2, result.TotalEntries)
				suite.Equal(2, result.ExportedEntries)
				suite.Len(exp.entries, 2)
				suite.Equal("alice@example.com", exp.entries[0].User)
				suite.Equal("bob@example.com", exp.entries[1].User)
			},
		},
		{
			name: "when multi-page paginates correctly",
			fetcher: newPagedFetcher([][]gen.AuditEntry{
				{suite.newEntry("alice@example.com"), suite.newEntry("bob@example.com")},
				{suite.newEntry("charlie@example.com")},
			}, 3),
			exporter:  &mockExporter{},
			batchSize: 2,
			validateFunc: func(exp *mockExporter, result *export.Result, err error) {
				suite.NoError(err)
				suite.Equal(3, result.TotalEntries)
				suite.Equal(3, result.ExportedEntries)
				suite.Len(exp.entries, 3)
			},
		},
		{
			name: "when fetcher errors returns partial result",
			fetcher: func(_ context.Context, _, offset int) ([]gen.AuditEntry, int, error) {
				if offset > 0 {
					return nil, 0, fmt.Errorf("connection lost")
				}
				return []gen.AuditEntry{suite.newEntry("alice@example.com")}, 3, nil
			},
			exporter:  &mockExporter{},
			batchSize: 1,
			validateFunc: func(_ *mockExporter, result *export.Result, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "fetching entries at offset 1")
				suite.Contains(err.Error(), "connection lost")
				suite.Equal(1, result.ExportedEntries)
				suite.Equal(3, result.TotalEntries)
			},
		},
		{
			name: "when write errors returns partial result",
			fetcher: func(_ context.Context, _, _ int) ([]gen.AuditEntry, int, error) {
				return []gen.AuditEntry{suite.newEntry("alice@example.com")}, 1, nil
			},
			exporter:  &mockExporter{writeErr: fmt.Errorf("disk full")},
			batchSize: 100,
			validateFunc: func(_ *mockExporter, result *export.Result, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "writing entry")
				suite.Equal(0, result.ExportedEntries)
			},
		},
		{
			name: "when open errors returns nil result",
			fetcher: func(_ context.Context, _, _ int) ([]gen.AuditEntry, int, error) {
				return nil, 0, nil
			},
			exporter:  &mockExporter{openErr: fmt.Errorf("permission denied")},
			batchSize: 100,
			validateFunc: func(_ *mockExporter, result *export.Result, err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "opening exporter")
				suite.Nil(result)
			},
		},
		{
			name: "when close errors logs warning but returns result",
			fetcher: func(_ context.Context, _, _ int) ([]gen.AuditEntry, int, error) {
				return nil, 0, nil
			},
			exporter:  &mockExporter{closeErr: fmt.Errorf("close failed")},
			batchSize: 100,
			validateFunc: func(_ *mockExporter, result *export.Result, err error) {
				suite.NoError(err)
				suite.Equal(0, result.TotalEntries)
				suite.Equal(0, result.ExportedEntries)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result, err := export.Run(
				suite.ctx,
				suite.logger,
				tc.fetcher,
				tc.exporter,
				tc.batchSize,
				nil,
			)
			tc.validateFunc(tc.exporter, result, err)
		})
	}
}

func (suite *ExportPublicTestSuite) TestRunProgress() {
	tests := []struct {
		name         string
		fetcher      export.Fetcher
		batchSize    int
		validateFunc func(calls []progressCall)
	}{
		{
			name: "when multi-page calls progress after each batch",
			fetcher: newPagedFetcher([][]gen.AuditEntry{
				{suite.newEntry("alice@example.com"), suite.newEntry("bob@example.com")},
				{suite.newEntry("charlie@example.com")},
			}, 3),
			batchSize: 2,
			validateFunc: func(calls []progressCall) {
				suite.Require().Len(calls, 2)
				suite.Equal(2, calls[0].exported)
				suite.Equal(3, calls[0].total)
				suite.Equal(3, calls[1].exported)
				suite.Equal(3, calls[1].total)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var calls []progressCall
			onProgress := func(exported int, total int) {
				calls = append(calls, progressCall{exported: exported, total: total})
			}

			_, err := export.Run(
				suite.ctx,
				suite.logger,
				tc.fetcher,
				&mockExporter{},
				tc.batchSize,
				onProgress,
			)
			suite.NoError(err)
			tc.validateFunc(calls)
		})
	}
}

func TestExportPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ExportPublicTestSuite))
}

// mockExporter implements export.Exporter for testing.
type mockExporter struct {
	opened   bool
	closed   bool
	entries  []gen.AuditEntry
	openErr  error
	writeErr error
	closeErr error
}

func (m *mockExporter) Open(
	_ context.Context,
) error {
	if m.openErr != nil {
		return m.openErr
	}
	m.opened = true
	return nil
}

func (m *mockExporter) Write(
	_ context.Context,
	entry gen.AuditEntry,
) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockExporter) Close(
	_ context.Context,
) error {
	m.closed = true
	return m.closeErr
}

type progressCall struct {
	exported int
	total    int
}

// newPagedFetcher creates a fetcher that returns pages of entries based on offset.
func newPagedFetcher(
	pages [][]gen.AuditEntry,
	total int,
) export.Fetcher {
	return func(
		_ context.Context,
		limit int,
		offset int,
	) ([]gen.AuditEntry, int, error) {
		_ = limit
		pageIdx := 0
		remaining := offset
		for pageIdx < len(pages) && remaining >= len(pages[pageIdx]) {
			remaining -= len(pages[pageIdx])
			pageIdx++
		}
		if pageIdx >= len(pages) {
			return nil, total, nil
		}
		return pages[pageIdx], total, nil
	}
}
