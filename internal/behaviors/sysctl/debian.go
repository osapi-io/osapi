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
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/provider"
)

// ProcDir is where the kernel shows its parameters, one file each, the dots
// of a parameter's name being the directories of its path.
const ProcDir = "/proc/sys"

// DefaultDir is where osapi writes its kernel parameter files unless
// behaviors.sysctl.dir says otherwise.
const DefaultDir = "/etc/sysctl.d"

// validKey matches sysctl parameter names: letters, digits, dots,
// underscores, and hyphens. Dots separate namespace components (e.g.
// net.ipv4.ip_forward). Slashes, spaces, and other characters are
// rejected so the key cannot be used for path traversal when building
// the conf file path. The first character must be a letter or digit so
// the key cannot be mistaken for a `sysctl` command-line option (e.g.
// -p) when passed as an argument.
var validKey = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
// It writes sysctl conf files directly to its directory, /etc/sysctl.d/ by
// default, and tracks state in the file-state KV for idempotency. Unlike the
// cron meta-provider, sysctl generates trivial content inline rather than
// fetching from the NATS Object Store.
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	fs          avfs.VFS
	execManager host.Exec
	// files writes the conf files and keeps their records: the managedfile
	// behavior, which sysctl requires.
	files managedfile.Provider
	// dir is where the conf files are written: behaviors.sysctl.dir.
	dir string
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	stateKV jetstream.KeyValue,
	execManager host.Exec,
	hostname string,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.sysctl")),
		fs:          fs,
		execManager: execManager,
		files:       managedfile.NewDebian(fs, stateKV, execManager, hostname),
		dir:         DefaultDir,
	}
}

// WithFiles sets the managed file provider the conf files are written
// through. The sysctl behavior passes the one the agent built for the
// managedfile behavior it requires.
func (d *Debian) WithFiles(
	files managedfile.Provider,
) *Debian {
	d.files = files

	return d
}

// WithDir sets the directory the conf files are written to, from
// behaviors.sysctl.dir. An empty dir keeps DefaultDir.
func (d *Debian) WithDir(
	dir string,
) *Debian {
	if dir != "" {
		d.dir = dir
	}

	return d
}

// Create deploys a new sysctl conf file and applies it. Idempotent on content
// rather than on existence: a drop-in already holding what was asked for
// reports Changed: false, and one that has been edited, deleted, or holds a
// different value is rewritten, because the operation's job is to make the
// setting true.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if entry.Key == "" {
		return nil, fmt.Errorf("sysctl create: key must not be empty")
	}

	if err := validateKey(entry.Key); err != nil {
		return nil, fmt.Errorf("sysctl create: %w", err)
	}

	if entry.Value == "" {
		return nil, fmt.Errorf("sysctl create: value must not be empty")
	}

	if err := validateValue(entry.Value); err != nil {
		return nil, fmt.Errorf("sysctl create: %w", err)
	}

	// Whether the file is recorded as managed does not decide anything. The
	// record says what osapi last wrote, which is exactly what drift makes
	// untrue, so deploy compares against the file on disk instead.
	return d.deploy(ctx, entry, "create")
}

// Update sets a parameter and manages it, whether or not osapi managed it
// before, as Create does. Idempotent: if the content hasn't changed, Changed
// is false.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if entry.Key == "" {
		return nil, fmt.Errorf("sysctl update: key must not be empty")
	}

	if err := validateKey(entry.Key); err != nil {
		return nil, fmt.Errorf("sysctl update: %w", err)
	}

	if entry.Value == "" {
		return nil, fmt.Errorf("sysctl update: value must not be empty")
	}

	if err := validateValue(entry.Value); err != nil {
		return nil, fmt.Errorf("sysctl update: %w", err)
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
	confPath := d.confPath(entry.Key)

	// The file on disk, not the record, decides whether anything changes, so
	// a drop-in edited by hand is rewritten rather than reported unchanged.
	result, err := managedfile.Update(ctx, d.files, managedfile.UpdateInput{
		Path:    confPath,
		Content: entry.Key + " = " + entry.Value + "\n",
		Metadata: map[string]string{
			"key":   entry.Key,
			"value": entry.Value,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sysctl %s: %w", opName, err)
	}
	if !result.Changed {
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

	// Apply the sysctl conf file.
	if _, execErr := d.execManager.RunPrivilegedCmd(ctx, "sysctl", []string{"-p", confPath}); execErr != nil {
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

	if err := validateKey(key); err != nil {
		return nil, fmt.Errorf("sysctl delete: %w", err)
	}

	confPath := d.confPath(key)

	// Only a file osapi recorded is removed; anything else is left alone.
	result, err := managedfile.Delete(ctx, d.files, managedfile.DeleteInput{Path: confPath})
	if err != nil {
		return nil, fmt.Errorf("sysctl delete: %w", err)
	}
	changed := result.Changed

	// Reload sysctl defaults.
	if changed {
		if _, execErr := d.execManager.RunPrivilegedCmd(ctx, "sysctl", []string{"--system"}); execErr != nil {
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

// List returns every kernel parameter the host has, with its current value
// from ProcDir, marking the ones osapi manages. A prefix, such as net.ipv4,
// narrows it to that parameter and the ones under it. A managed parameter the
// kernel does not have right now, such as one whose module is not loaded, is
// listed with the value osapi wrote. A parameter whose file cannot be read,
// such as one only root may read, is left out, as sysctl -a leaves it out.
func (d *Debian) List(
	ctx context.Context,
	prefix string,
) ([]Entry, error) {
	if prefix != "" {
		if err := validateKey(prefix); err != nil {
			return nil, fmt.Errorf("sysctl list: %w", err)
		}
	}

	managed, err := d.managed(ctx)
	if err != nil {
		return nil, err
	}

	byKey := map[string]Entry{}
	err = d.fs.WalkDir(ProcDir, func(path string, e fs.DirEntry, walkErr error) error {
		// What cannot be read, /proc/sys itself included on a host without
		// it, is left out rather than failing the list, as sysctl -a does.
		if walkErr != nil || e.IsDir() {
			return nil
		}
		key := strings.ReplaceAll(strings.TrimPrefix(path, ProcDir+"/"), "/", ".")
		if !under(key, prefix) {
			return nil
		}
		data, readErr := d.fs.ReadFile(path)
		if readErr != nil {
			return nil
		}
		_, isManaged := managed[key]
		byKey[key] = Entry{Key: key, Value: value(data), Managed: isManaged}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sysctl list: %w", err)
	}

	for key, v := range managed {
		if _, ok := byKey[key]; !ok && under(key, prefix) {
			byKey[key] = Entry{Key: key, Value: v, Managed: true}
		}
	}

	result := make([]Entry, 0, len(byKey))
	for _, e := range byKey {
		result = append(result, e)
	}
	slices.SortFunc(result, func(a, b Entry) int { return strings.Compare(a.Key, b.Key) })

	return result, nil
}

// Get returns one kernel parameter's current value, and whether osapi manages
// it. It is not found only when the kernel has no such parameter and osapi
// does not manage it.
func (d *Debian) Get(
	ctx context.Context,
	key string,
) (*Entry, error) {
	if key == "" {
		return nil, fmt.Errorf("sysctl get: key must not be empty")
	}

	if err := validateKey(key); err != nil {
		return nil, fmt.Errorf("sysctl get: %w", err)
	}

	record, err := managedfile.Get(ctx, d.files, managedfile.GetInput{Path: d.confPath(key)})
	if errors.Is(err, managedfile.ErrUnreadable) {
		return nil, fmt.Errorf("sysctl entry %q: failed to read state", key)
	}
	isManaged := err == nil && !record.Undeployed

	data, readErr := d.fs.ReadFile(ProcDir + "/" + strings.ReplaceAll(key, ".", "/"))
	switch {
	case readErr == nil:
		return &Entry{Key: key, Value: value(data), Managed: isManaged}, nil
	case isManaged:
		return &Entry{Key: key, Value: record.Metadata["value"], Managed: true}, nil
	case errors.Is(readErr, fs.ErrNotExist):
		return nil, fmt.Errorf("sysctl entry %q: %w", key, behavior.ErrNotFound)
	default:
		return nil, fmt.Errorf("sysctl entry %q: %w", key, readErr)
	}
}

// managed is every parameter osapi manages, with the value it wrote: one
// osapi-<key>.conf file each in the directory, whose record is deployed.
func (d *Debian) managed(
	ctx context.Context,
) (map[string]string, error) {
	managed := map[string]string{}
	dirEntries, err := d.fs.ReadDir(d.dir)
	if errors.Is(err, fs.ErrNotExist) {
		// No directory yet means osapi has written nothing.
		return managed, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sysctl list: %w", err)
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !strings.HasPrefix(dirEntry.Name(), "osapi-") {
			continue
		}
		if entry := d.buildEntryFromState(ctx, d.dir+"/"+dirEntry.Name()); entry != nil {
			managed[entry.Key] = entry.Value
		}
	}

	return managed, nil
}

// under reports whether key is prefix or a parameter under it.
func under(
	key string,
	prefix string,
) bool {
	return prefix == "" || key == prefix || strings.HasPrefix(key, prefix+".")
}

// value is a parameter's file content as one line: a value of several fields,
// which the kernel separates with tabs, is joined with single spaces, the way
// it is written in a conf file.
func value(
	data []byte,
) string {
	return strings.Join(strings.Fields(string(data)), " ")
}

// buildEntryFromState creates an Entry from file-state KV metadata.
// Returns nil if the entry is not managed or has been undeployed.
func (d *Debian) buildEntryFromState(
	ctx context.Context,
	path string,
) *Entry {
	record, err := managedfile.Get(ctx, d.files, managedfile.GetInput{Path: path})
	if err != nil || record.Undeployed {
		return nil
	}

	key, ok := record.Metadata["key"]
	if !ok {
		return nil
	}

	return &Entry{
		Key:   key,
		Value: record.Metadata["value"],
	}
}

// confPath returns the conf file path for a sysctl key, such as
// /etc/sysctl.d/osapi-net.ipv4.ip_forward.conf.
func (d *Debian) confPath(
	key string,
) string {
	return d.dir + "/osapi-" + key + ".conf"
}

// validateKey checks that a sysctl key is safe for use in a file path.
func validateKey(
	key string,
) error {
	if !validKey.MatchString(key) {
		return fmt.Errorf(
			"invalid sysctl key %q: must match %s",
			key,
			validKey.String(),
		)
	}

	return nil
}

// validateValue checks that a sysctl value contains no line breaks, which
// would let it inject additional lines into the conf file. The value
// itself is not included in the error since it may be long.
func validateValue(
	value string,
) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("value must not contain line breaks")
	}

	return nil
}
