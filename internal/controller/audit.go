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

package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// auditStreams is the part of a NATS client that the audit store uses.
type auditStreams interface {
	StreamCreateOrUpdate(ctx context.Context, config jetstream.StreamConfig) error
	audit.Stream
}

// newAuditStore makes the audit stream and a store on it. It returns nil if
// cfg names no stream, because then the controller keeps no audit log.
func newAuditStore(
	ctx context.Context,
	log *slog.Logger,
	nc auditStreams,
	namespace string,
	cfg config.NATSAudit,
) (*audit.StreamStore, error) {
	if cfg.Stream == "" {
		return nil, nil
	}

	if err := nc.StreamCreateOrUpdate(ctx, bus.AuditStream(namespace, cfg)); err != nil {
		return nil, fmt.Errorf("create audit stream: %w", err)
	}

	return audit.NewStreamStore(
		log,
		nc,
		job.ApplyNamespaceToInfraName(namespace, cfg.Stream),
		job.ApplyNamespaceToSubjects(namespace, cfg.Subject),
	), nil
}
