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
		return nil, fmt.Errorf("file deploy: get object %q: %w", req.ObjectName, err)
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
			return nil, fmt.Errorf("file deploy: render template: %w", err)
		}
	}

	sha := computeSHA256(content)
	stateKey := BuildStateKey(p.hostname, req.Path)

	// What matters is the file on disk, not what the last deploy recorded. A
	// config edited by hand has the SHA the record remembers and the content
	// nobody asked for, and comparing against the record would call that
	// unchanged and walk away.
	onDisk, readErr := p.fs.ReadFile(req.Path)
	if readErr == nil && computeSHA256(onDisk) == sha {
		// The content is right, which says nothing about the permissions or the
		// ownership: a request that changes only the mode would otherwise never
		// take effect.
		changed, permErr := p.applyPermissions(ctx, req)
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

	mode, err := parseFileMode(req.Mode)
	if err != nil {
		return nil, err
	}

	dir := p.fs.Dir(req.Path)
	if err := p.fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("file deploy: create directory %q: %w", dir, err)
	}

	if err := fsutil.WriteFileAtomic(p.fs, req.Path, content, mode); err != nil {
		return nil, fmt.Errorf("file deploy: write file %q: %w", req.Path, err)
	}

	if _, err := p.enforceOwnership(ctx, req); err != nil {
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
// Both are compared against the file itself rather than against what the last
// deploy recorded, so a mode or an owner changed by anything at all is corrected.
// What osapi remembers is a record of what it did, never evidence of what is
// there now.
func (p *Service) applyPermissions(
	ctx context.Context,
	req DeployRequest,
) (bool, error) {
	changed := false

	// An absent mode means "leave the permissions alone", not 0644: the default
	// belongs to a file being created, and applying it to one already on disk
	// would quietly widen permissions someone else set.
	if req.Mode != "" {
		want, err := parseFileMode(req.Mode)
		if err != nil {
			return false, err
		}

		info, err := p.fs.Stat(req.Path)
		if err != nil {
			return false, fmt.Errorf("file deploy: stat file %q: %w", req.Path, err)
		}

		if info.Mode().Perm() != want.Perm() {
			if err := p.fs.Chmod(req.Path, want); err != nil {
				return false, fmt.Errorf("file deploy: set mode on %q: %w", req.Path, err)
			}

			changed = true
		}
	}

	ownershipChanged, err := p.enforceOwnership(ctx, req)
	if err != nil {
		return false, err
	}

	return changed || ownershipChanged, nil
}

// enforceOwnership applies the requested owner and group when the file does not
// already have them, and reports whether it had to.
func (p *Service) enforceOwnership(
	ctx context.Context,
	req DeployRequest,
) (bool, error) {
	if req.Owner == "" && req.Group == "" {
		return false, nil
	}

	wantUID, err := resolveID(req.Owner, lookupUID)
	if err != nil {
		return false, fmt.Errorf("file deploy: resolve owner for %q: %w", req.Path, err)
	}

	wantGID, err := resolveID(req.Group, lookupGID)
	if err != nil {
		return false, fmt.Errorf("file deploy: resolve group for %q: %w", req.Path, err)
	}

	matches, err := ownershipMatches(req.Path, wantUID, wantGID)
	if err != nil {
		return false, fmt.Errorf("file deploy: read ownership of %q: %w", req.Path, err)
	}

	if matches {
		return false, nil
	}

	// chown rather than the filesystem, because the agent reaches privileged
	// operations through sudo and is not itself root.
	spec := req.Owner
	if req.Group != "" {
		spec = req.Owner + ":" + req.Group
	}

	if _, err := p.execManager.RunPrivilegedCmd(
		ctx,
		"chown",
		[]string{spec, req.Path},
	); err != nil {
		return false, fmt.Errorf(
			"file deploy: set ownership %q on %q: %w",
			spec,
			req.Path,
			err,
		)
	}

	return true, nil
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
		return fmt.Errorf("file deploy: marshal file state: %w", err)
	}

	if _, err := p.stateKV.Put(ctx, stateKey, stateBytes); err != nil {
		return fmt.Errorf("file deploy: update file state: %w", err)
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
//
// An absent mode means the 0644 a new file is created with. An unparsable one is
// an error: it used to become 0644 as well, which silently granted more than the
// request asked for and reported success for a mode nobody applied.
func parseFileMode(
	mode string,
) (os.FileMode, error) {
	if mode == "" {
		return 0o644, nil
	}

	parsed, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode %q: %w", mode, err)
	}

	return os.FileMode(parsed), nil
}
