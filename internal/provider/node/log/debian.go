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

package log

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems
// using journalctl for log querying.
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	execManager exec.Manager
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	execManager exec.Manager,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.log")),
		execManager: execManager,
	}
}

// Query returns journal entries with optional filtering.
func (d *Debian) Query(
	_ context.Context,
	opts QueryOpts,
) ([]Entry, error) {
	d.logger.Debug("executing log.Query")
	args := buildArgs(opts)

	output, err := d.execManager.RunCmd("journalctl", args)
	if err != nil {
		return nil, fmt.Errorf("log: query: %w", err)
	}

	return parseJournalOutput(output, d.logger), nil
}

// QueryUnit returns journal entries for a specific systemd unit.
func (d *Debian) QueryUnit(
	_ context.Context,
	unit string,
	opts QueryOpts,
) ([]Entry, error) {
	d.logger.Debug(
		"executing log.QueryUnit",
		slog.String("unit", unit),
	)
	args := buildUnitArgs(unit, opts)

	output, err := d.execManager.RunCmd("journalctl", args)
	if err != nil {
		return nil, fmt.Errorf("log: query unit: %w", err)
	}

	return parseJournalOutput(output, d.logger), nil
}

// ListSources returns unique syslog identifiers from the journal.
func (d *Debian) ListSources(
	_ context.Context,
) ([]string, error) {
	d.logger.Debug("executing log.ListSources")

	output, err := d.execManager.RunCmd("journalctl", []string{"--field=SYSLOG_IDENTIFIER"})
	if err != nil {
		return nil, fmt.Errorf("log: list sources: %w", err)
	}

	return parseSources(output), nil
}
