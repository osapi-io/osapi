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
	"fmt"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// AuditService provides audit log operations.
type AuditService struct {
	client *gen.ClientWithResponses
}

// List retrieves audit log entries with pagination.
func (s *AuditService) List(
	ctx context.Context,
	limit int,
	offset int,
) (*Response[AuditList], error) {
	params := &gen.GetAuditLogsParams{
		Limit:  &limit,
		Offset: &offset,
	}

	resp, err := s.client.GetAuditLogsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(auditListFromGen(resp.JSON200), resp.Body), nil
}

// Get retrieves a single audit log entry by ID.
func (s *AuditService) Get(
	ctx context.Context,
	id string,
) (*Response[AuditEntry], error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.GetAuditLogByIDWithResponse(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("get audit log %s: %w", id, err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(auditEntryFromGen(resp.JSON200.Entry), resp.Body), nil
}

// Export retrieves all audit log entries for export.
func (s *AuditService) Export(
	ctx context.Context,
) (*Response[AuditList], error) {
	resp, err := s.client.GetAuditExportWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("export audit logs: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(auditListFromGen(resp.JSON200), resp.Body), nil
}
