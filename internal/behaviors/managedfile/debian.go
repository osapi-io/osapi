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

package managedfile

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/fsutil"
)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// debian keeps managed files on the host's filesystem and their records in
// the file state store, keyed by host and path.
type debian struct {
	fs       avfs.VFS
	state    jetstream.KeyValue
	exec     host.Exec
	hostname string
}

// NewDebian is the provider for a filesystem and a file state store. The
// hand-written domains that write managed files build one with it.
func NewDebian(
	fs avfs.VFS,
	state jetstream.KeyValue,
	exec host.Exec,
	hostname string,
) Provider {
	return &debian{fs: fs, state: state, exec: exec, hostname: hostname}
}

// newDebian is the provider on Debian and its family. It needs the file state
// store, and declines a host without one.
func newDebian(
	h host.Host,
) (Provider, error) {
	if h.FileState == nil {
		return nil, behavior.ErrUnsupported
	}

	return NewDebian(h.FS, h.FileState, h.Exec, h.Hostname), nil
}

// Write makes f.Path hold f.Content with its mode and ownership. Whether it
// changes anything is decided by the file on disk rather than by the record,
// so a file edited by hand is rewritten instead of reported unchanged.
func (d *debian) Write(
	ctx context.Context,
	f File,
) (bool, error) {
	mode := f.Mode
	if mode == 0 {
		mode = 0o644
	}

	changed := false
	if onDisk, err := d.fs.ReadFile(f.Path); err != nil || !bytes.Equal(onDisk, f.Content) {
		if err := d.fs.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
			return false, fmt.Errorf("create directory: %w", err)
		}
		// Atomically: a reader of the directory never sees half a file.
		if err := fsutil.WriteFileAtomic(d.fs, f.Path, f.Content, mode); err != nil {
			return false, fmt.Errorf("write file: %w", err)
		}
		changed = true
	}

	info, err := d.fs.Stat(f.Path)
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}
	if info.Mode().Perm() != mode.Perm() {
		if err := d.fs.Chmod(f.Path, mode); err != nil {
			return false, fmt.Errorf("set mode: %w", err)
		}
		changed = true
	}

	owned, err := d.own(ctx, f)
	if err != nil {
		return false, err
	}
	if !changed && !owned {
		return false, nil
	}

	state := State{
		Path:       f.Path,
		Mode:       fmt.Sprintf("%04o", mode.Perm()),
		Owner:      f.Owner,
		Group:      f.Group,
		DeployedAt: time.Now().UTC().Format(time.RFC3339),
		Metadata:   f.Metadata,
	}
	if !f.Secret {
		state.SHA256 = checksum(f.Content)
	}
	data, err := marshalJSON(state)
	if err != nil {
		return false, fmt.Errorf("marshal state: %w", err)
	}
	if _, err := d.state.Put(ctx, StateKey(d.hostname, f.Path), data); err != nil {
		return false, fmt.Errorf("update state: %w", err)
	}

	return true, nil
}

// own gives f.Path f's owner and group when it does not have them, and
// reports whether it had to. The current owner comes from stat and the change
// from chown with sudo, since the agent is not root.
func (d *debian) own(
	ctx context.Context,
	f File,
) (bool, error) {
	if f.Owner == "" && f.Group == "" {
		return false, nil
	}

	out, err := d.exec.RunCmd(ctx, "stat", []string{"-c", "%U:%G", "--", f.Path})
	if err != nil {
		return false, fmt.Errorf("read ownership: %w", err)
	}
	owner, group, _ := strings.Cut(strings.TrimSpace(out), ":")
	if (f.Owner == "" || f.Owner == owner) && (f.Group == "" || f.Group == group) {
		return false, nil
	}

	spec := f.Owner
	if f.Group != "" {
		spec += ":" + f.Group
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "chown", []string{"--", spec, f.Path}); err != nil {
		return false, fmt.Errorf("set ownership: %w", err)
	}

	return true, nil
}

// Remove deletes path when its record says osapi manages it. A file osapi
// has no record of, or has already stopped managing, is left alone.
func (d *debian) Remove(
	ctx context.Context,
	path string,
) (bool, error) {
	key := StateKey(d.hostname, path)
	entry, err := d.state.Get(ctx, key)
	if err != nil {
		return false, nil
	}

	var state State
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return false, nil
	}
	if state.UndeployedAt != "" {
		return false, nil
	}

	changed := false
	if _, err := d.fs.Stat(path); err == nil {
		if err := d.fs.Remove(path); err != nil {
			return false, fmt.Errorf("remove file: %w", err)
		}
		changed = true
	}

	state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := marshalJSON(state)
	if err != nil {
		return false, fmt.Errorf("marshal state: %w", err)
	}
	if _, err := d.state.Put(ctx, key, data); err != nil {
		return false, fmt.Errorf("update state: %w", err)
	}

	return changed, nil
}

// Read is the record of path.
func (d *debian) Read(
	ctx context.Context,
	path string,
) (Record, error) {
	entry, err := d.state.Get(ctx, StateKey(d.hostname, path))
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, ErrNotManaged)
	}

	var state State
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, ErrUnreadable)
	}

	return Record{
		Path:       path,
		SHA256:     state.SHA256,
		Metadata:   state.Metadata,
		Undeployed: state.UndeployedAt != "",
	}, nil
}

func checksum(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}
