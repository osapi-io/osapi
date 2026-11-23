// Copyright (c) 2025 John Dewey

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

	"github.com/osapi-io/osapi/internal/job"
)

// processJobOperation handles the actual job processing based on category and operation.
func (a *Agent) processJobOperation(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	a.logger.Debug(
		"dispatching to provider",
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	return a.registry.Dispatch(ctx, jobRequest)
}

// NewNodeProcessor returns a ProcessorFunc for the node category. Each node
// operation is a behavior now, so no hand-written one is left, and it refuses
// each operation.
func NewNodeProcessor(
	_ *slog.Logger,
) ProcessorFunc {
	return func(_ context.Context, req job.Request) (json.RawMessage, error) {
		return nil, fmt.Errorf("unsupported node operation: %s", req.Operation)
	}
}
