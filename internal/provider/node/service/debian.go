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
	"fmt"
	"log/slog"
	"regexp"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/file"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9_@.-]+$`)

// Compile-time check: Debian must satisfy Provider.
var _ Provider = (*Debian)(nil)

// Compile-time check: Debian must satisfy FactsSetter.
var _ provider.FactsSetter = (*Debian)(nil)

// Debian implements the Provider interface for Debian-family systems.
// It delegates unit file writes to a FileDeployer for SHA tracking and
// idempotency. Service control operations use systemctl via exec.Manager.
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
		logger:       logger.With(slog.String("subsystem", "provider.service")),
		fs:           fs,
		fileDeployer: fileDeployer,
		stateKV:      stateKV,
		execManager:  execManager,
		hostname:     hostname,
	}
}

// validateName checks that a service name is safe for use.
func validateName(
	name string,
) error {
	if name == "" {
		return fmt.Errorf("invalid service name: empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf(
			"invalid service name %q: must match %s",
			name,
			validName.String(),
		)
	}

	return nil
}
