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
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/provider"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// debian keeps managed files on the host's filesystem and their records in
// the file state store, keyed by host and path.
type debian struct {
	fs       avfs.VFS
	state    provider.FileState
	bucket   string
	exec     host.Exec
	hostname string
	// stat are the arguments of stat that print a file's owner and group
	// as owner:group: GNU's on Linux, BSD's on macOS.
	stat []string
}

// gnuStat are the arguments of GNU stat that print owner:group.
var gnuStat = []string{"-c", "%U:%G"}

// NewDebian is the provider for a filesystem and a file state store. The
// hand-written domains that write managed files build one with it.
func NewDebian(
	fs avfs.VFS,
	state provider.FileState,
	bucket string,
	exec host.Exec,
	hostname string,
) Provider {
	return &debian{
		fs: fs, state: state, bucket: bucket, exec: exec, hostname: hostname, stat: gnuStat,
	}
}

// newDebian is the provider on Debian and its family. It needs the file state
// store, and declines a host without one.
func newDebian(
	h host.Host,
) (Provider, error) {
	if h.FileState == nil {
		return nil, behavior.ErrUnsupported
	}

	return NewDebian(h.FS, h.FileState, h.FileStateBucket, h.Exec, h.Hostname), nil
}

// Write makes f.Path hold f.Content with its mode and ownership. Whether it
// changes anything is decided by the file on disk rather than by the record,
// so a file edited by hand is rewritten instead of reported unchanged.
func (d *debian) Write(
	ctx context.Context,
	f File,
) (bool, error) {
	// No mode keeps the mode of a file that is there: 0644 could give more
	// access than somebody set. Only a new file gets 0644.
	mode := f.Mode
	if mode == 0 {
		mode = 0o644
		if info, err := d.fs.Stat(f.Path); err == nil {
			mode = info.Mode().Perm()
		}
	}

	changed := false
	if onDisk, err := d.fs.ReadFile(f.Path); err != nil || !bytes.Equal(onDisk, f.Content) {
		if err := d.fs.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
			return false, fmt.Errorf("create directory: %w", err)
		}
		// Atomically: a reader of the directory never sees half a file.
		if err := atomicfile.Write(ctx, d.fs, f.Path, f.Content, mode); err != nil {
			return false, fmt.Errorf("write file: %w", err)
		}
		changed = true
	}

	info, err := d.fs.Stat(f.Path)
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}
	if info.Mode().Perm() != mode.Perm() {
		if err := atomicfile.Chmod(ctx, d.fs, f.Path, mode); err != nil {
			return false, fmt.Errorf("set mode: %w", err)
		}
		changed = true
	}

	owned, err := d.own(ctx, f)
	if err != nil {
		return false, err
	}
	if !changed && !owned && !d.stale(ctx, f) {
		return false, nil
	}

	state := State{
		ObjectName:  f.Object,
		ContentType: f.ContentType,
		Path:        f.Path,
		Mode:        fmt.Sprintf("%04o", mode.Perm()),
		Owner:       f.Owner,
		Group:       f.Group,
		DeployedAt:  time.Now().UTC().Format(time.RFC3339),
		Metadata:    f.Metadata,
	}
	if !f.Secret {
		state.SHA256 = checksum(f.Content)
		state.ObjectSHA256 = f.ObjectSHA256
	}
	data, err := marshalJSON(state)
	if err != nil {
		return false, fmt.Errorf("marshal state: %w", err)
	}
	if _, err := d.state.KVPut(ctx, d.bucket, StateKey(d.hostname, f.Path), data); err != nil {
		return false, fmt.Errorf("update state: %w", err)
	}

	return true, nil
}

// stale reports whether osapi deployed f.Path with a record that does not
// describe f: other metadata, such as a cron entry whose schedule changed and
// whose file did not, or another object, such as a template whose object
// changed and renders the same file. The write then changes the record, so
// that a read gives the new metadata and the stale check the new object. A
// path that has no record, or an undeployed one, is not stale: a file with
// the content that osapi did not deploy stays as it is.
func (d *debian) stale(
	ctx context.Context,
	f File,
) bool {
	entry, err := d.state.KVGet(ctx, d.bucket, StateKey(d.hostname, f.Path))
	if err != nil {
		return false
	}
	var state State
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return false
	}

	objectSHA := f.ObjectSHA256
	if f.Secret {
		objectSHA = ""
	}

	return state.UndeployedAt == "" && (!maps.Equal(state.Metadata, f.Metadata) ||
		state.ObjectName != f.Object || state.ContentType != f.ContentType ||
		state.ObjectSHA256 != objectSHA)
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

	out, err := d.exec.RunCmd(ctx, "stat", append(slices.Clone(d.stat), "--", f.Path))
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
	entry, err := d.state.KVGet(ctx, d.bucket, key)
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
		if err := atomicfile.Remove(ctx, d.fs, path); err != nil {
			return false, fmt.Errorf("remove file: %w", err)
		}
		changed = true
	}

	state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := marshalJSON(state)
	if err != nil {
		return false, fmt.Errorf("marshal state: %w", err)
	}
	if _, err := d.state.KVPut(ctx, d.bucket, key, data); err != nil {
		return false, fmt.Errorf("update state: %w", err)
	}

	return changed, nil
}

// Read is the record of path.
func (d *debian) Read(
	ctx context.Context,
	path string,
) (Record, error) {
	entry, err := d.state.KVGet(ctx, d.bucket, StateKey(d.hostname, path))
	if err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, ErrNotManaged)
	}

	var state State
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, ErrUnreadable)
	}

	return Record{
		Path:         path,
		SHA256:       state.SHA256,
		Metadata:     state.Metadata,
		Undeployed:   state.UndeployedAt != "",
		ObjectSHA256: state.ObjectSHA256,
	}, nil
}

func checksum(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}
