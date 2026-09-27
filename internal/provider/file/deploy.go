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

	"github.com/osapi-io/osapi/internal/fsutil"
	"github.com/osapi-io/osapi/internal/job"
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

	// When ContentType is not explicitly set, resolve it from the object's
	// metadata header. This allows meta providers (service, certificate) to
	// omit ContentType and have the file provider respect the uploader's intent.
	contentType := req.ContentType
	if contentType == "" {
		contentType = "raw"

		info, infoErr := p.objStore.GetInfo(ctx, req.ObjectName)
		if infoErr == nil && info.Headers != nil {
			if ct := info.Headers.Get("Osapi-Content-Type"); ct != "" {
				contentType = ct
			}
		}
	}

	if contentType == "template" {
		content, err = p.renderTemplate(content, req.Vars)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
	}

	sha := computeSHA256(content)
	stateKey := BuildStateKey(p.hostname, req.Path)

	entry, err := p.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(entry.Value(), &state); unmarshalErr == nil {
			if state.SHA256 == sha {
				if _, statErr := p.fs.Stat(req.Path); statErr == nil {
					// The content is what it should be, which says nothing about
					// the permissions or the ownership: a request that changes
					// only the mode would otherwise never take effect.
					changed, permErr := p.applyPermissions(ctx, req, state)
					if permErr != nil {
						return nil, permErr
					}

					p.logger.Debug(
						"file content unchanged",
						slog.String("path", req.Path),
						slog.String("sha256", sha),
						slog.Bool("permissions_changed", changed),
					)

					if changed {
						if err := p.putState(ctx, stateKey, req, sha, contentType); err != nil {
							return nil, err
						}
					}

					return &DeployResult{
						Changed: changed,
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

	dir := p.fs.Dir(req.Path)
	if err := p.fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	if err := fsutil.WriteFileAtomic(p.fs, req.Path, content, mode); err != nil {
		return nil, fmt.Errorf("failed to write file %q: %w", req.Path, err)
	}

	if err := p.applyOwnership(ctx, req.Path, req.Owner, req.Group); err != nil {
		return nil, err
	}

	if err := p.putState(ctx, stateKey, req, sha, contentType); err != nil {
		return nil, err
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

// applyPermissions brings an already-correct file's mode and ownership in line
// with the request, and reports whether anything had to change.
//
// The mode is compared against the file on disk, so a mode changed by anything
// is corrected. The ownership is compared against the last deploy's state,
// because the uid and gid a name resolves to are not readable through the
// filesystem abstraction this provider writes through: ownership changed outside
// osapi is therefore not detected, while a request that changes the owner or the
// group is applied.
func (p *Service) applyPermissions(
	ctx context.Context,
	req DeployRequest,
	state job.FileState,
) (bool, error) {
	changed := false

	// An absent mode means "leave the permissions alone", not 0644: the default
	// belongs to a file being created, and applying it to one already on disk
	// would quietly widen permissions someone else set.
	if req.Mode != "" {
		want := parseFileMode(req.Mode)

		info, err := p.fs.Stat(req.Path)
		if err != nil {
			return false, fmt.Errorf("failed to stat file %q: %w", req.Path, err)
		}

		if info.Mode().Perm() != want.Perm() {
			if err := p.fs.Chmod(req.Path, want); err != nil {
				return false, fmt.Errorf("failed to set mode on %q: %w", req.Path, err)
			}

			changed = true
		}
	}

	if req.Owner != state.Owner || req.Group != state.Group {
		if err := p.applyOwnership(ctx, req.Path, req.Owner, req.Group); err != nil {
			return false, err
		}

		changed = true
	}

	return changed, nil
}

// applyOwnership sets the file's owner and group when the request names either.
// It runs chown rather than calling the filesystem, because the agent reaches
// privileged operations through sudo and is not itself root.
func (p *Service) applyOwnership(
	ctx context.Context,
	path string,
	owner string,
	group string,
) error {
	if owner == "" && group == "" {
		return nil
	}

	spec := owner
	if group != "" {
		spec = owner + ":" + group
	}

	if _, err := p.execManager.RunPrivilegedCmd(ctx, "chown", []string{spec, path}); err != nil {
		return fmt.Errorf("failed to set ownership %q on %q: %w", spec, path, err)
	}

	return nil
}

// putState records what was deployed, so the next deploy can tell what changed.
func (p *Service) putState(
	ctx context.Context,
	stateKey string,
	req DeployRequest,
	sha string,
	contentType string,
) error {
	state := job.FileState{
		ObjectName:  req.ObjectName,
		Path:        req.Path,
		SHA256:      sha,
		Mode:        req.Mode,
		Owner:       req.Owner,
		Group:       req.Group,
		DeployedAt:  time.Now().UTC().Format(time.RFC3339),
		ContentType: contentType,
		Metadata:    req.Metadata,
	}

	stateBytes, err := marshalJSON(state)
	if err != nil {
		return fmt.Errorf("failed to marshal file state: %w", err)
	}

	if _, err := p.stateKV.Put(ctx, stateKey, stateBytes); err != nil {
		return fmt.Errorf("failed to update file state: %w", err)
	}

	return nil
}

// computeSHA256 returns the hex-encoded SHA-256 hash of the given data.
func computeSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

// BuildStateKey returns the KV key for a file's deploy state.
// Format: <hostname>.<sha256-of-path>.
func BuildStateKey(
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
