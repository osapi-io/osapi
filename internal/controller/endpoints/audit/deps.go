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

	auditstore "github.com/osapi-io/osapi/internal/audit"
)

// Deps is what the audit group's handlers use.
type Deps struct {
	// Store reads the audit log. Nil when nats.audit.stream is not set; each
	// method then answers 503.
	Store Store
}

// Store reads entries of the audit log.
type Store interface {
	Get(ctx context.Context, id string) (*auditstore.Entry, error)
	List(ctx context.Context, limit int, offset int) ([]auditstore.Entry, int, error)
	ListAll(ctx context.Context) ([]auditstore.Entry, error)
}
