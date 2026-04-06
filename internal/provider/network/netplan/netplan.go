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

// Package netplan provides shared helpers for writing Netplan configuration
// files with validation, rollback, and file-state tracking.
package netplan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/file"
)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// ApplyConfig writes a Netplan configuration file to disk, validates it
// with `netplan generate`, applies it with `netplan apply`, and tracks
// the file state in the KV store. Returns (true, nil) when the file was
// written and applied, (false, nil) when the content is unchanged.
func ApplyConfig(
	ctx context.Context,
	logger *slog.Logger,
	fs avfs.VFS,
	stateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
	path string,
	content []byte,
	metadata map[string]string,
) (bool, error) {
	sha := ComputeSHA256(content)
	stateKey := file.BuildStateKey(hostname, path)

	// Check for idempotency: if SHA matches and file exists, skip.
	kvEntry, err := stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr == nil {
			if state.SHA256 == sha && state.UndeployedAt == "" {
				if _, statErr := fs.Stat(path); statErr == nil {
					logger.Debug(
						"netplan config unchanged, skipping deploy",
						slog.String("path", path),
					)

					return false, nil
				}
			}
		}
	}

	// Ensure the parent directory exists.
	dir := filepath.Dir(path)
	if mkErr := fs.MkdirAll(dir, 0o755); mkErr != nil {
		return false, fmt.Errorf("netplan apply: create directory: %w", mkErr)
	}

	// Write the config file.
	if writeErr := fs.WriteFile(path, content, 0o600); writeErr != nil {
		return false, fmt.Errorf("netplan apply: write file: %w", writeErr)
	}

	// Validate with netplan generate (validates without applying).
	if _, genErr := execManager.RunPrivilegedCmd("netplan", []string{"generate"}); genErr != nil {
		// Roll back: remove the invalid file.
		_ = fs.Remove(path)

		return false, fmt.Errorf(
			"netplan validate failed (file rolled back): %w",
			genErr,
		)
	}

	// Apply the configuration.
	if _, applyErr := execManager.RunPrivilegedCmd("netplan", []string{"apply"}); applyErr != nil {
		return false, fmt.Errorf("netplan apply: %w", applyErr)
	}

	// Persist state in KV.
	state := job.FileState{
		Path:       path,
		SHA256:     sha,
		Mode:       "0600",
		DeployedAt: time.Now().UTC().Format(time.RFC3339),
		Metadata:   metadata,
	}

	stateBytes, marshalErr := marshalJSON(state)
	if marshalErr != nil {
		return false, fmt.Errorf("netplan apply: marshal state: %w", marshalErr)
	}

	if _, putErr := stateKV.Put(ctx, stateKey, stateBytes); putErr != nil {
		return false, fmt.Errorf("netplan apply: update state: %w", putErr)
	}

	logger.Info(
		"netplan config deployed",
		slog.String("path", path),
		slog.String("sha256", sha),
	)

	return true, nil
}

// RemoveConfig removes a Netplan configuration file from disk, applies
// the change with `netplan apply`, and marks the file state as undeployed
// in the KV store. Returns (true, nil) when the file was removed and
// applied, (false, nil) when the file does not exist.
func RemoveConfig(
	ctx context.Context,
	logger *slog.Logger,
	fs avfs.VFS,
	stateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
	path string,
) (bool, error) {
	// Check if file exists — if not, nothing to do.
	if _, err := fs.Stat(path); err != nil {
		return false, nil
	}

	// Remove the file.
	if removeErr := fs.Remove(path); removeErr != nil {
		return false, fmt.Errorf("netplan remove: remove file: %w", removeErr)
	}

	// Apply the configuration change.
	if _, applyErr := execManager.RunPrivilegedCmd("netplan", []string{"apply"}); applyErr != nil {
		return false, fmt.Errorf("netplan remove: apply: %w", applyErr)
	}

	// Best-effort state cleanup: mark as undeployed in KV.
	stateKey := file.BuildStateKey(hostname, path)

	kvEntry, err := stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr == nil {
			state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)

			stateBytes, marshalErr := marshalJSON(state)
			if marshalErr == nil {
				if _, putErr := stateKV.Put(ctx, stateKey, stateBytes); putErr != nil {
					logger.Warn(
						"netplan remove: failed to update state",
						slog.String("path", path),
						slog.String("error", putErr.Error()),
					)
				}
			}
		}
	}

	logger.Info(
		"netplan config removed",
		slog.String("path", path),
	)

	return true, nil
}

// ComputeSHA256 returns the hex-encoded SHA-256 hash of the given data.
func ComputeSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
