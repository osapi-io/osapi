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

package sysctl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/file"
)

const sysctlDir = "/etc/sysctl.d"

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
// It writes sysctl conf files directly to /etc/sysctl.d/ and tracks
// state in the file-state KV for idempotency. Unlike the cron
// meta-provider, sysctl generates trivial content inline rather than
// fetching from the NATS Object Store.
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	fs          avfs.VFS
	stateKV     jetstream.KeyValue
	execManager exec.Manager
	hostname    string
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	stateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.sysctl")),
		fs:          fs,
		stateKV:     stateKV,
		execManager: execManager,
		hostname:    hostname,
	}
}

// Create deploys a new sysctl conf file and applies it. Fails if the
// key is already managed. Idempotent: if the content hasn't changed
// since the last deploy, the file is not rewritten and Changed is false.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if entry.Key == "" {
		return nil, fmt.Errorf("sysctl create: key must not be empty")
	}

	if entry.Value == "" {
		return nil, fmt.Errorf("sysctl create: value must not be empty")
	}

	confPath := confPath(entry.Key)
	stateKey := file.BuildStateKey(d.hostname, confPath)

	// Already managed — nothing to do.
	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr == nil {
			if state.UndeployedAt == "" {
				return &CreateResult{
					Key:     entry.Key,
					Changed: false,
				}, nil
			}
		}
	}

	return d.deploy(ctx, entry, "create")
}

// Update redeploys an existing managed sysctl conf file. Fails if the
// key is not currently managed. Idempotent: if the content hasn't
// changed, Changed is false.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if entry.Key == "" {
		return nil, fmt.Errorf("sysctl update: key must not be empty")
	}

	if entry.Value == "" {
		return nil, fmt.Errorf("sysctl update: value must not be empty")
	}

	confPath := confPath(entry.Key)
	stateKey := file.BuildStateKey(d.hostname, confPath)

	// Check that the key is managed.
	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf(
			"sysctl update: key %q not managed",
			entry.Key,
		)
	}

	var state job.FileState
	if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr == nil {
		if state.UndeployedAt != "" {
			return nil, fmt.Errorf(
				"sysctl update: key %q not managed",
				entry.Key,
			)
		}
	}

	result, deployErr := d.deploy(ctx, entry, "update")
	if deployErr != nil {
		return nil, deployErr
	}

	return (*UpdateResult)(result), nil
}

// deploy is the shared implementation for Create and Update.
func (d *Debian) deploy(
	ctx context.Context,
	entry Entry,
	opName string,
) (*CreateResult, error) {
	content := []byte(entry.Key + " = " + entry.Value + "\n")
	confPath := confPath(entry.Key)
	sha := computeSHA256(content)

	stateKey := file.BuildStateKey(d.hostname, confPath)

	// Check for idempotency: if SHA matches and file exists, skip.
	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err == nil {
		var state job.FileState
		if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr == nil {
			if state.SHA256 == sha && state.UndeployedAt == "" {
				if _, statErr := d.fs.Stat(confPath); statErr == nil {
					d.logger.Debug(
						"sysctl entry unchanged, skipping deploy",
						slog.String("key", entry.Key),
						slog.String("path", confPath),
					)

					return &CreateResult{
						Key:     entry.Key,
						Changed: false,
					}, nil
				}
			}
		}
	}

	// Ensure the directory exists.
	if mkErr := d.fs.MkdirAll(sysctlDir, 0o755); mkErr != nil {
		return nil, fmt.Errorf("sysctl %s: create directory: %w", opName, mkErr)
	}

	// Write the conf file.
	if writeErr := d.fs.WriteFile(confPath, content, 0o644); writeErr != nil {
		return nil, fmt.Errorf("sysctl %s: write file: %w", opName, writeErr)
	}

	// Persist state in KV.
	state := job.FileState{
		Path:       confPath,
		SHA256:     sha,
		Mode:       "0644",
		DeployedAt: time.Now().UTC().Format(time.RFC3339),
		Metadata: map[string]string{
			"key":   entry.Key,
			"value": entry.Value,
		},
	}

	stateBytes, marshalErr := marshalJSON(state)
	if marshalErr != nil {
		return nil, fmt.Errorf("sysctl %s: marshal state: %w", opName, marshalErr)
	}

	if _, putErr := d.stateKV.Put(ctx, stateKey, stateBytes); putErr != nil {
		return nil, fmt.Errorf("sysctl %s: update state: %w", opName, putErr)
	}

	// Apply the sysctl conf file.
	if _, execErr := d.execManager.RunPrivilegedCmd("sysctl", []string{"-p", confPath}); execErr != nil {
		d.logger.Warn(
			"sysctl apply failed after deploy",
			slog.String("key", entry.Key),
			slog.String("path", confPath),
			slog.String("error", execErr.Error()),
		)
	}

	d.logger.Info(
		"sysctl entry deployed",
		slog.String("key", entry.Key),
		slog.String("path", confPath),
		slog.Bool("changed", true),
	)

	return &CreateResult{
		Key:     entry.Key,
		Changed: true,
	}, nil
}

// Delete removes a managed sysctl conf file and reloads defaults.
func (d *Debian) Delete(
	ctx context.Context,
	key string,
) (*DeleteResult, error) {
	if key == "" {
		return nil, fmt.Errorf("sysctl delete: key must not be empty")
	}

	confPath := confPath(key)
	stateKey := file.BuildStateKey(d.hostname, confPath)

	// Check if managed.
	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return &DeleteResult{
			Key:     key,
			Changed: false,
		}, nil
	}

	var state job.FileState
	if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr != nil {
		return &DeleteResult{
			Key:     key,
			Changed: false,
		}, nil
	}

	if state.UndeployedAt != "" {
		return &DeleteResult{
			Key:     key,
			Changed: false,
		}, nil
	}

	// Remove the file from disk.
	changed := false

	if _, statErr := d.fs.Stat(confPath); statErr == nil {
		if removeErr := d.fs.Remove(confPath); removeErr != nil {
			return nil, fmt.Errorf("sysctl delete: remove file: %w", removeErr)
		}

		changed = true
	}

	// Update KV state with undeploy timestamp.
	state.UndeployedAt = time.Now().UTC().Format(time.RFC3339)

	stateBytes, marshalErr := marshalJSON(state)
	if marshalErr != nil {
		return nil, fmt.Errorf("sysctl delete: marshal state: %w", marshalErr)
	}

	if _, putErr := d.stateKV.Put(ctx, stateKey, stateBytes); putErr != nil {
		return nil, fmt.Errorf("sysctl delete: update state: %w", putErr)
	}

	// Reload sysctl defaults.
	if changed {
		if _, execErr := d.execManager.RunPrivilegedCmd("sysctl", []string{"--system"}); execErr != nil {
			d.logger.Warn(
				"sysctl reload failed after delete",
				slog.String("key", key),
				slog.String("error", execErr.Error()),
			)
		}
	}

	d.logger.Info(
		"sysctl entry removed",
		slog.String("key", key),
		slog.String("path", confPath),
		slog.Bool("changed", changed),
	)

	return &DeleteResult{
		Key:     key,
		Changed: changed,
	}, nil
}

// List returns all osapi-managed sysctl entries with current runtime values.
func (d *Debian) List(
	ctx context.Context,
) ([]Entry, error) {
	dirEntries, err := d.fs.ReadDir(sysctlDir)
	if err != nil {
		return nil, fmt.Errorf("list sysctl entries: %w", err)
	}

	var result []Entry

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		if !strings.HasPrefix(dirEntry.Name(), "osapi-") {
			continue
		}

		path := sysctlDir + "/" + dirEntry.Name()

		entry := d.buildEntryFromState(ctx, path)
		if entry == nil {
			continue
		}

		// Read current runtime value.
		runtimeValue := d.readRuntimeValue(entry.Key)
		if runtimeValue != "" {
			entry.Value = runtimeValue
		}

		result = append(result, *entry)
	}

	return result, nil
}

// Get returns a single sysctl entry by key with current runtime value.
func (d *Debian) Get(
	ctx context.Context,
	key string,
) (*Entry, error) {
	if key == "" {
		return nil, fmt.Errorf("sysctl get: key must not be empty")
	}

	confPath := confPath(key)
	stateKey := file.BuildStateKey(d.hostname, confPath)

	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf("sysctl entry %q: not found", key)
	}

	var state job.FileState
	if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr != nil {
		return nil, fmt.Errorf("sysctl entry %q: failed to read state", key)
	}

	if state.UndeployedAt != "" {
		return nil, fmt.Errorf("sysctl entry %q: not found", key)
	}

	entry := &Entry{
		Key: key,
	}

	if state.Metadata != nil {
		entry.Value = state.Metadata["value"]
	}

	// Read current runtime value.
	runtimeValue := d.readRuntimeValue(key)
	if runtimeValue != "" {
		entry.Value = runtimeValue
	}

	return entry, nil
}

// buildEntryFromState creates an Entry from file-state KV metadata.
// Returns nil if the entry is not managed or has been undeployed.
func (d *Debian) buildEntryFromState(
	ctx context.Context,
	path string,
) *Entry {
	stateKey := file.BuildStateKey(d.hostname, path)

	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return nil
	}

	var state job.FileState
	if err := json.Unmarshal(kvEntry.Value(), &state); err != nil {
		return nil
	}

	if state.UndeployedAt != "" {
		return nil
	}

	if state.Metadata == nil {
		return nil
	}

	key, ok := state.Metadata["key"]
	if !ok {
		return nil
	}

	return &Entry{
		Key:   key,
		Value: state.Metadata["value"],
	}
}

// readRuntimeValue reads the current runtime value for a sysctl key
// using `sysctl -n`. Returns empty string on error.
func (d *Debian) readRuntimeValue(
	key string,
) string {
	output, err := d.execManager.RunCmd("sysctl", []string{"-n", key})
	if err != nil {
		return ""
	}

	return strings.TrimSpace(output)
}

// confPath returns the conf file path for a sysctl key.
// Dots in the key are preserved: net.ipv4.ip_forward becomes
// /etc/sysctl.d/osapi-net.ipv4.ip_forward.conf.
func confPath(
	key string,
) string {
	return sysctlDir + "/osapi-" + key + ".conf"
}

// computeSHA256 returns the hex-encoded SHA-256 hash of the given data.
func computeSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
