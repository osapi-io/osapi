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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/node/certificate"
)

// NewCertificateProcessor returns a ProcessorFunc that handles certificate-related operations.
func NewCertificateProcessor(
	certProvider certificate.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		if certProvider == nil {
			return nil, fmt.Errorf("certificate provider not available")
		}

		// Extract base operation from dotted operation (e.g., "ca.list" -> "ca")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "ca":
			return processCertificateCAOperation(certProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported certificate operation: %s", req.Operation)
		}
	}
}

// processCertificateCAOperation dispatches certificate CA sub-operations.
func processCertificateCAOperation(
	certProvider certificate.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract sub-operation: "ca.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid certificate CA operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processCertificateCAList(ctx, certProvider, logger)
	case "create":
		return processCertificateCACreate(ctx, certProvider, logger, jobRequest)
	case "update":
		return processCertificateCAUpdate(ctx, certProvider, logger, jobRequest)
	case "delete":
		return processCertificateCADelete(ctx, certProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported certificate CA operation: %s", jobRequest.Operation)
	}
}

// processCertificateCAList lists all CA certificates.
func processCertificateCAList(
	ctx context.Context,
	certProvider certificate.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing certificate.ca.List")

	entries, err := certProvider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

// processCertificateCACreate creates a new CA certificate.
func processCertificateCACreate(
	ctx context.Context,
	certProvider certificate.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry certificate.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal certificate CA create data: %w", err)
	}

	logger.Debug("executing certificate.ca.Create",
		slog.String("name", entry.Name),
	)

	result, err := certProvider.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processCertificateCAUpdate updates an existing CA certificate.
func processCertificateCAUpdate(
	ctx context.Context,
	certProvider certificate.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry certificate.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal certificate CA update data: %w", err)
	}

	logger.Debug("executing certificate.ca.Update",
		slog.String("name", entry.Name),
	)

	result, err := certProvider.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processCertificateCADelete deletes a CA certificate.
func processCertificateCADelete(
	ctx context.Context,
	certProvider certificate.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal certificate CA delete data: %w", err)
	}

	logger.Debug("executing certificate.ca.Delete",
		slog.String("name", data.Name),
	)

	result, err := certProvider.Delete(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}
