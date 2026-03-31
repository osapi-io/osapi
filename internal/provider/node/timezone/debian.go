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

package timezone

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems
// using timedatectl for timezone management.
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
		logger:      logger.With(slog.String("subsystem", "provider.timezone")),
		execManager: execManager,
	}
}

// Get returns the current system timezone by running timedatectl and date.
func (d *Debian) Get(
	_ context.Context,
) (*Info, error) {
	tzOutput, err := d.execManager.RunCmd(
		"timedatectl",
		[]string{"show", "-p", "Timezone", "--value"},
	)
	if err != nil {
		return nil, fmt.Errorf("timezone: timedatectl show: %w", err)
	}

	offsetOutput, err := d.execManager.RunCmd("date", []string{"+%:z"})
	if err != nil {
		return nil, fmt.Errorf("timezone: date offset: %w", err)
	}

	return &Info{
		Timezone:  strings.TrimSpace(tzOutput),
		UTCOffset: strings.TrimSpace(offsetOutput),
	}, nil
}

// Update sets the system timezone via timedatectl. Idempotent: returns
// Changed false when the timezone already matches.
func (d *Debian) Update(
	_ context.Context,
	timezone string,
) (*UpdateResult, error) {
	currentOutput, err := d.execManager.RunCmd(
		"timedatectl",
		[]string{"show", "-p", "Timezone", "--value"},
	)
	if err != nil {
		return nil, fmt.Errorf("timezone: timedatectl show: %w", err)
	}

	current := strings.TrimSpace(currentOutput)
	if current == timezone {
		d.logger.Debug(
			"timezone unchanged, skipping update",
			slog.String("timezone", timezone),
		)

		return &UpdateResult{
			Timezone: timezone,
			Changed:  false,
		}, nil
	}

	if _, setErr := d.execManager.RunCmd("timedatectl", []string{"set-timezone", timezone}); setErr != nil {
		return nil, fmt.Errorf("timezone: set-timezone: %w", setErr)
	}

	d.logger.Info(
		"timezone updated",
		slog.String("from", current),
		slog.String("to", timezone),
		slog.Bool("changed", true),
	)

	return &UpdateResult{
		Timezone: timezone,
		Changed:  true,
	}, nil
}
