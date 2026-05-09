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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/file"
)

const (
	unitDir       = "/etc/systemd/system"
	managedPrefix = "osapi-"
)

// Create deploys a new unit file via the file provider.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := unitFilePath(entry.Name)

	if _, err := d.fs.Stat(filePath); err == nil {
		return &CreateResult{
			Name:    entry.Name,
			Changed: false,
		}, nil
	}

	d.logger.Debug(
		"creating service unit file",
		slog.String("name", entry.Name),
	)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName: entry.Object,
		Path:       filePath,
		Mode:       "0644",
		Metadata:   map[string]string{"source": "custom"},
	})
	if err != nil {
		return nil, fmt.Errorf("service: create: %w", err)
	}

	if result.Changed {
		if err := d.daemonReload(); err != nil {
			return nil, fmt.Errorf("service: create: %w", err)
		}
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Update redeploys an existing unit file via the file provider.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := unitFilePath(entry.Name)

	if _, err := d.fs.Stat(filePath); err != nil {
		return nil, fmt.Errorf("service unit %q not managed", entry.Name)
	}

	// If no new object was specified, preserve the current one.
	if entry.Object == "" {
		existing := d.buildEntryFromState(ctx, entry.Name, filePath)
		if existing == nil {
			return nil, fmt.Errorf(
				"service: update: failed to read existing state for %q",
				entry.Name,
			)
		}
		entry.Object = existing.Object
	}

	d.logger.Debug(
		"updating service unit file",
		slog.String("name", entry.Name),
	)

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName: entry.Object,
		Path:       filePath,
		Mode:       "0644",
		Metadata:   map[string]string{"source": "custom"},
	})
	if err != nil {
		return nil, fmt.Errorf("service: update: %w", err)
	}

	if result.Changed {
		if err := d.daemonReload(); err != nil {
			return nil, fmt.Errorf("service: update: %w", err)
		}
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Delete undeploys a unit file via the file provider.
func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath := unitFilePath(name)

	if _, err := d.fs.Stat(filePath); err != nil {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	unitName := managedPrefix + name + ".service"

	// Best-effort stop and disable before removing the unit file.
	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"stop", unitName}); err != nil {
		d.logger.Warn(
			"failed to stop service before delete",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
	}

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"disable", unitName}); err != nil {
		d.logger.Warn(
			"failed to disable service before delete",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
	}

	d.logger.Debug(
		"deleting service unit file",
		slog.String("name", name),
	)

	result, err := d.fileDeployer.Undeploy(ctx, file.UndeployRequest{
		Path: filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("service: delete: %w", err)
	}

	if result.Changed {
		if err := d.daemonReload(); err != nil {
			return nil, fmt.Errorf("service: delete: %w", err)
		}
	}

	return &DeleteResult{
		Name:    name,
		Changed: result.Changed,
	}, nil
}

// unitFilePath returns the file path for a managed unit file.
func unitFilePath(
	name string,
) string {
	return unitDir + "/" + managedPrefix + name + ".service"
}

// daemonReload runs systemctl daemon-reload to pick up unit file changes.
func (d *Debian) daemonReload() error {
	_, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"daemon-reload"})
	if err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	return nil
}

// buildEntryFromState creates an Entry from file-state KV metadata.
func (d *Debian) buildEntryFromState(
	ctx context.Context,
	name string,
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

	return &Entry{
		Name:   name,
		Object: state.ObjectName,
	}
}
