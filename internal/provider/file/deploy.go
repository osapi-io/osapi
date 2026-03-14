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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/job"
)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// Deploy writes file content to the target path with the specified
// permissions. It uses SHA-256 checksums for idempotency: if the
// content hasn't changed since the last deploy, the file is not
// rewritten and changed is false.
func (p *Service) Deploy(
	ctx context.Context,
	req DeployRequest,
) (*DeployResult, error) {
	content, err := p.objStore.GetBytes(ctx, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %q: %w", req.ObjectName, err)
	}

	if req.ContentType == "template" {
		content, err = p.renderTemplate(content, req.Vars)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
	}

	sha := computeSHA256(content)
	stateKey := buildStateKey(p.hostname, req.Path)

	entry, err := p.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(entry.Value(), &state); unmarshalErr == nil {
			if state.SHA256 == sha {
				if _, statErr := p.fs.Stat(req.Path); statErr == nil {
					p.logger.Debug(
						"file unchanged, skipping deploy",
						slog.String("path", req.Path),
						slog.String("sha256", sha),
					)

					return &DeployResult{
						Changed: false,
						SHA256:  sha,
						Path:    req.Path,
					}, nil
				}

				p.logger.Debug(
					"file missing from disk, redeploying",
					slog.String("path", req.Path),
					slog.String("sha256", sha),
				)
			}
		}
	}

	mode := parseFileMode(req.Mode)

	if err := afero.WriteFile(p.fs, req.Path, content, mode); err != nil {
		return nil, fmt.Errorf("failed to write file %q: %w", req.Path, err)
	}

	state := job.FileState{
		ObjectName:  req.ObjectName,
		Path:        req.Path,
		SHA256:      sha,
		Mode:        req.Mode,
		Owner:       req.Owner,
		Group:       req.Group,
		DeployedAt:  time.Now().UTC().Format(time.RFC3339),
		ContentType: req.ContentType,
	}

	stateBytes, err := marshalJSON(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file state: %w", err)
	}

	if _, err := p.stateKV.Put(ctx, stateKey, stateBytes); err != nil {
		return nil, fmt.Errorf("failed to update file state: %w", err)
	}

	p.logger.Info(
		"file deployed",
		slog.String("path", req.Path),
		slog.String("sha256", sha),
		slog.Bool("changed", true),
	)

	return &DeployResult{
		Changed: true,
		SHA256:  sha,
		Path:    req.Path,
	}, nil
}

// computeSHA256 returns the hex-encoded SHA-256 hash of the given data.
func computeSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

// buildStateKey returns the KV key for a file's deploy state.
// Format: <hostname>.<sha256-of-path>.
func buildStateKey(
	hostname string,
	path string,
) string {
	pathHash := computeSHA256([]byte(path))

	return hostname + "." + pathHash
}

// parseFileMode parses a string file mode (e.g., "0644") into an os.FileMode.
// Returns 0644 as the default if the string is empty or invalid.
func parseFileMode(
	mode string,
) os.FileMode {
	if mode == "" {
		return 0o644
	}

	parsed, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0o644
	}

	return os.FileMode(parsed)
}
