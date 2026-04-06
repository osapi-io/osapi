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

package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/file"
)

const cronDir = "/etc/cron.d"

// periodicDirs maps interval names to their directory paths.
var periodicDirs = map[string]string{
	"hourly":  "/etc/cron.hourly",
	"daily":   "/etc/cron.daily",
	"weekly":  "/etc/cron.weekly",
	"monthly": "/etc/cron.monthly",
}

// periodicIntervals is the ordered list of periodic interval names.
var periodicIntervals = []string{"hourly", "daily", "weekly", "monthly"}

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Compile-time check: Debian must satisfy FactsSetter.
var _ provider.FactsSetter = (*Debian)(nil)

// Debian implements the Provider interface for Debian-family systems.
// It delegates file writes to a FileDeployer for SHA tracking and
// idempotency. Managed files are identified by their presence in the
// file-state KV, not by file content headers.
type Debian struct {
	provider.FactsAware
	logger       *slog.Logger
	fs           avfs.VFS
	fileDeployer file.Deployer
	stateKV      jetstream.KeyValue
	hostname     string
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	fileDeployer file.Deployer,
	stateKV jetstream.KeyValue,
	hostname string,
) *Debian {
	return &Debian{
		logger:       logger,
		fs:           fs,
		fileDeployer: fileDeployer,
		stateKV:      stateKV,
		hostname:     hostname,
	}
}

// List returns all osapi-managed cron entries from /etc/cron.d/ and
// /etc/cron.{hourly,daily,weekly,monthly}/.
func (d *Debian) List(
	ctx context.Context,
) ([]Entry, error) {
	var result []Entry

	// Scan /etc/cron.d/ for schedule-based entries.
	cronDirEntries, err := d.fs.ReadDir(cronDir)
	if err != nil {
		return nil, fmt.Errorf("list cron entries: %w", err)
	}

	for _, dirEntry := range cronDirEntries {
		if dirEntry.IsDir() {
			continue
		}

		path := cronDir + "/" + dirEntry.Name()
		if !d.isManagedFile(ctx, path) {
			continue
		}

		entry := d.buildEntryFromState(ctx, dirEntry.Name(), path, "cron.d")
		if entry != nil {
			result = append(result, *entry)
		}
	}

	// Scan periodic directories for interval-based entries.
	for _, interval := range periodicIntervals {
		dir := periodicDirs[interval]

		dirEntries, err := d.fs.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, dirEntry := range dirEntries {
			if dirEntry.IsDir() {
				continue
			}

			path := dir + "/" + dirEntry.Name()
			if !d.isManagedFile(ctx, path) {
				continue
			}

			entry := d.buildEntryFromState(ctx, dirEntry.Name(), path, interval)
			if entry != nil {
				result = append(result, *entry)
			}
		}
	}

	return result, nil
}

// Get returns a single cron entry by name, searching all directories.
func (d *Debian) Get(
	ctx context.Context,
	name string,
) (*Entry, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath, _ := d.findEntryPath(name)
	if filePath == "" {
		return nil, fmt.Errorf("cron entry %q: not found", name)
	}

	if !d.isManagedFile(ctx, filePath) {
		return nil, fmt.Errorf("cron entry %q is not managed by osapi", name)
	}

	source := d.sourceForPath(filePath)

	entry := d.buildEntryFromState(ctx, name, filePath, source)
	if entry == nil {
		return nil, fmt.Errorf("cron entry %q: failed to read state", name)
	}

	return entry, nil
}

// Create deploys a new cron entry via the file provider.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	// Already managed — nothing to do.
	if existingPath, _ := d.findEntryPath(entry.Name); existingPath != "" {
		return &CreateResult{
			Name:    entry.Name,
			Changed: false,
		}, nil
	}

	filePath, perm := d.entryFilePath(entry)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        fmt.Sprintf("%04o", perm),
		ContentType: entry.ContentType,
		Vars:        entry.Vars,
		Metadata:    buildCronMetadata(entry),
	})
	if err != nil {
		return nil, fmt.Errorf("create cron entry: %w", err)
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Update redeploys an existing cron entry via the file provider.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath, perm := d.findEntryPath(entry.Name)
	if filePath == "" {
		return nil, fmt.Errorf("cron entry %q not managed", entry.Name)
	}

	// If no new object was specified, preserve the current one.
	if entry.Object == "" {
		existing := d.buildEntryFromState(ctx, entry.Name, filePath, "")
		if existing == nil {
			return nil, fmt.Errorf(
				"update cron entry: failed to read existing state for %q",
				entry.Name,
			)
		}
		entry.Object = existing.Object
	}

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName:  entry.Object,
		Path:        filePath,
		Mode:        fmt.Sprintf("%04o", perm),
		ContentType: entry.ContentType,
		Vars:        entry.Vars,
		Metadata:    buildCronMetadata(entry),
	})
	if err != nil {
		return nil, fmt.Errorf("update cron entry: %w", err)
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Delete undeploys a cron entry via the file provider.
func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath, _ := d.findEntryPath(name)
	if filePath == "" {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	result, err := d.fileDeployer.Undeploy(ctx, file.UndeployRequest{
		Path: filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("delete cron entry: %w", err)
	}

	return &DeleteResult{
		Name:    name,
		Changed: result.Changed,
	}, nil
}

// entryFilePath returns the file path and permissions for a new entry.
func (d *Debian) entryFilePath(
	entry Entry,
) (string, uint32) {
	if entry.Interval != "" {
		dir, ok := periodicDirs[entry.Interval]
		if ok {
			return dir + "/" + entry.Name, 0o755
		}
	}

	return cronDir + "/" + entry.Name, 0o644
}

// findEntryPath searches all directories for an existing entry by name.
// Returns the file path and permissions, or empty string if not found.
func (d *Debian) findEntryPath(
	name string,
) (string, uint32) {
	// Check /etc/cron.d/ first.
	cronPath := cronDir + "/" + name
	if _, err := d.fs.Stat(cronPath); err == nil {
		return cronPath, 0o644
	}

	// Check periodic directories.
	for _, interval := range periodicIntervals {
		dir := periodicDirs[interval]
		periodicPath := dir + "/" + name
		if _, err := d.fs.Stat(periodicPath); err == nil {
			return periodicPath, 0o755
		}
	}

	return "", 0
}

// isManagedFile checks if the file at path has a file-state KV entry,
// indicating it was deployed by osapi.
func (d *Debian) isManagedFile(
	ctx context.Context,
	path string,
) bool {
	stateKey := file.BuildStateKey(d.hostname, path)
	entry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return false
	}

	// Check that the state entry is not undeployed.
	var state job.FileState
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return false
	}

	return state.UndeployedAt == ""
}

// buildEntryFromState creates an Entry from file-state KV metadata.
func (d *Debian) buildEntryFromState(
	ctx context.Context,
	name string,
	path string,
	source string,
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

	entry := &Entry{
		Name:   name,
		Object: state.ObjectName,
		Source: source,
	}

	// Restore domain-specific fields from metadata.
	if state.Metadata != nil {
		entry.Schedule = state.Metadata["schedule"]
		entry.Interval = state.Metadata["interval"]
		entry.User = state.Metadata["user"]
	}

	return entry
}

// sourceForPath returns the source string for a given file path.
func (d *Debian) sourceForPath(
	path string,
) string {
	for _, interval := range periodicIntervals {
		dir := periodicDirs[interval]
		if len(path) > len(dir) && path[:len(dir)] == dir {
			return interval
		}
	}

	return "cron.d"
}

// buildCronMetadata returns the metadata map for a cron entry.
func buildCronMetadata(
	entry Entry,
) map[string]string {
	m := make(map[string]string)
	if entry.Schedule != "" {
		m["schedule"] = entry.Schedule
	}
	if entry.Interval != "" {
		m["interval"] = entry.Interval
	}
	if entry.User != "" {
		m["user"] = entry.User
	}

	return m
}

// validateName checks that a cron entry name is safe for use as a filename.
func validateName(
	name string,
) error {
	if name == "" {
		return fmt.Errorf("invalid cron entry name: empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid cron entry name %q: must match %s", name, validName.String())
	}

	return nil
}
