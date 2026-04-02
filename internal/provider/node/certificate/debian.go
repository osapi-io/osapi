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

package certificate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/file"
)

const (
	systemCADir   = "/usr/share/ca-certificates"
	customCADir   = "/usr/local/share/ca-certificates"
	managedPrefix = "osapi-"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Compile-time check: Debian must satisfy Provider.
var _ Provider = (*Debian)(nil)

// Compile-time check: Debian must satisfy FactsSetter.
var _ provider.FactsSetter = (*Debian)(nil)

// Debian implements the Provider interface for Debian-family systems.
// It delegates file writes to a FileDeployer for SHA tracking and
// idempotency. Managed files are identified by their presence in the
// file-state KV with an osapi- filename prefix.
type Debian struct {
	provider.FactsAware
	logger       *slog.Logger
	fs           avfs.VFS
	fileDeployer file.Deployer
	stateKV      jetstream.KeyValue
	execManager  exec.Manager
	hostname     string
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	fileDeployer file.Deployer,
	stateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
) *Debian {
	return &Debian{
		logger:       logger.With(slog.String("subsystem", "provider.certificate")),
		fs:           fs,
		fileDeployer: fileDeployer,
		stateKV:      stateKV,
		execManager:  execManager,
		hostname:     hostname,
	}
}

// Create deploys a new CA certificate via the file provider.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := certFilePath(entry.Name)

	if _, err := d.fs.Stat(filePath); err == nil {
		return nil, fmt.Errorf("certificate %q already exists", entry.Name)
	}

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName: entry.Object,
		Path:       filePath,
		Mode:       "0644",
		Metadata:   map[string]string{"source": "custom"},
	})
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	if result.Changed {
		if err := d.updateCACertificates(); err != nil {
			return nil, fmt.Errorf("create certificate: %w", err)
		}
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Update redeploys an existing CA certificate via the file provider.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := certFilePath(entry.Name)

	if _, err := d.fs.Stat(filePath); err != nil {
		return nil, fmt.Errorf("certificate %q does not exist", entry.Name)
	}

	// If no new object was specified, preserve the current one.
	if entry.Object == "" {
		existing := d.buildEntryFromState(ctx, entry.Name, filePath)
		if existing == nil {
			return nil, fmt.Errorf(
				"update certificate: failed to read existing state for %q",
				entry.Name,
			)
		}
		entry.Object = existing.Object
	}

	result, err := d.fileDeployer.Deploy(ctx, file.DeployRequest{
		ObjectName: entry.Object,
		Path:       filePath,
		Mode:       "0644",
		Metadata:   map[string]string{"source": "custom"},
	})
	if err != nil {
		return nil, fmt.Errorf("update certificate: %w", err)
	}

	if result.Changed {
		if err := d.updateCACertificates(); err != nil {
			return nil, fmt.Errorf("update certificate: %w", err)
		}
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: result.Changed,
	}, nil
}

// Delete undeploys a CA certificate via the file provider.
func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath := certFilePath(name)

	if _, err := d.fs.Stat(filePath); err != nil {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	result, err := d.fileDeployer.Undeploy(ctx, file.UndeployRequest{
		Path: filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("delete certificate: %w", err)
	}

	if result.Changed {
		if err := d.updateCACertificates(); err != nil {
			return nil, fmt.Errorf("delete certificate: %w", err)
		}
	}

	return &DeleteResult{
		Name:    name,
		Changed: result.Changed,
	}, nil
}

// certFilePath returns the file path for a managed certificate.
func certFilePath(
	name string,
) string {
	return customCADir + "/" + managedPrefix + name + ".crt"
}

// updateCACertificates runs update-ca-certificates to rebuild the trust store.
func (d *Debian) updateCACertificates() error {
	_, err := d.execManager.RunPrivilegedCmd("update-ca-certificates", nil)
	if err != nil {
		return fmt.Errorf("update-ca-certificates: %w", err)
	}

	return nil
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
		Source: "custom",
	}
}

// validateName checks that a certificate name is safe for use as a filename.
func validateName(
	name string,
) error {
	if name == "" {
		return fmt.Errorf("invalid certificate name: empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf(
			"invalid certificate name %q: must match %s",
			name,
			validName.String(),
		)
	}

	return nil
}
