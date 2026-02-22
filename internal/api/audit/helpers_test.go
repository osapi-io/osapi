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

package audit_test

import (
	"context"
	"log/slog"

	auditapi "github.com/retr0h/osapi/internal/api/audit"
	auditstore "github.com/retr0h/osapi/internal/audit"
)

// fakeStore is a simple in-memory audit store for handler tests.
type fakeStore struct {
	// Write
	writeErr error

	// Get
	getEntry *auditstore.Entry
	getErr   error

	// List
	listEntries []auditstore.Entry
	listTotal   int
	listErr     error
}

func (f *fakeStore) Write(
	_ context.Context,
	_ auditstore.Entry,
) error {
	return f.writeErr
}

func (f *fakeStore) Get(
	_ context.Context,
	_ string,
) (*auditstore.Entry, error) {
	return f.getEntry, f.getErr
}

func (f *fakeStore) List(
	_ context.Context,
	_ int,
	_ int,
) ([]auditstore.Entry, int, error) {
	return f.listEntries, f.listTotal, f.listErr
}

func (f *fakeStore) reset() {
	f.writeErr = nil
	f.getEntry = nil
	f.getErr = nil
	f.listEntries = nil
	f.listTotal = 0
	f.listErr = nil
}

// newTestAuditHandler creates an audit handler for integration tests.
func newTestAuditHandler(
	logger *slog.Logger,
	store auditstore.Store,
) *auditapi.Audit {
	return auditapi.New(logger, store)
}
