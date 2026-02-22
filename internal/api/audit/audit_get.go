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
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/api/audit/gen"
)

// GetAuditLogByID returns a single audit log entry by ID.
func (a *Audit) GetAuditLogByID(
	ctx context.Context,
	request gen.GetAuditLogByIDRequestObject,
) (gen.GetAuditLogByIDResponseObject, error) {
	id := request.Id.String()

	entry, err := a.Store.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			errMsg := "audit entry not found"
			return gen.GetAuditLogByID404JSONResponse{Error: &errMsg}, nil
		}

		a.logger.Error(
			"failed to get audit entry",
			slog.String("error", err.Error()),
			slog.String("id", id),
		)
		errMsg := "failed to get audit entry"
		return gen.GetAuditLogByID500JSONResponse{Error: &errMsg}, nil
	}

	return gen.GetAuditLogByID200JSONResponse{
		Entry: mapEntryToGen(*entry),
	}, nil
}
