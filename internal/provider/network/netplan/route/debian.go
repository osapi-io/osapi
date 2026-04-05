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

package route

import (
	"encoding/json"
	"log/slog"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// marshalJSON is a package-level variable for testing the marshal
// error path in route metadata serialization.
var marshalJSON = json.Marshal

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family
// systems. It writes Netplan YAML route files to /etc/netplan/ with an
// osapi- prefix and tracks state in the file-state KV for idempotency.
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
		logger:      logger.With(slog.String("subsystem", "provider.netplan.route")),
		fs:          fs,
		stateKV:     stateKV,
		execManager: execManager,
		hostname:    hostname,
	}
}
