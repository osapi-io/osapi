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

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// Undeploy removes a deployed file from disk. The object store entry
// is preserved. The file-state KV is updated to record the undeploy
// timestamp.
func (p *Service) Undeploy(
	ctx context.Context,
	req UndeployRequest,
) (*UndeployResult, error) {
	_, err := p.fs.Stat(req.Path)
	if err != nil {
		p.logger.Debug(
			"file not on disk, nothing to undeploy",
			slog.String("path", req.Path),
		)

		return &UndeployResult{
			Changed: false,
			Path:    req.Path,
		}, nil
	}

	if err := p.fs.Remove(req.Path); err != nil {
		return nil, fmt.Errorf("failed to remove file %q: %w", req.Path, err)
	}

	stateKey := BuildStateKey(p.hostname, req.Path)

	entry, err := p.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(entry.Value(), &state); unmarshalErr == nil {
			state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)

			stateBytes, marshalErr := marshalJSON(state)
			if marshalErr == nil {
				_, _ = p.stateKV.Put(ctx, stateKey, stateBytes)
			}
		}
	}

	p.logger.Info(
		"file undeployed",
		slog.String("path", req.Path),
		slog.Bool("changed", true),
	)

	return &UndeployResult{
		Changed: true,
		Path:    req.Path,
	}, nil
}
