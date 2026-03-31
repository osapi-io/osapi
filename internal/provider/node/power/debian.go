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

package power

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// minDelayMinutes is the minimum delay before shutdown/reboot in minutes.
// This ensures the agent has time to write the job result back to NATS,
// send the response to the API, and complete graceful shutdown before
// the system goes down.
const minDelayMinutes = 1

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
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
		logger:      logger.With(slog.String("subsystem", "provider.power")),
		execManager: execManager,
	}
}

// Reboot schedules a system reboot with the specified delay.
func (d *Debian) Reboot(
	ctx context.Context,
	opts Opts,
) (*Result, error) {
	return d.executePowerAction(ctx, "reboot", "-r", opts)
}

// Shutdown schedules a system shutdown with the specified delay.
func (d *Debian) Shutdown(
	ctx context.Context,
	opts Opts,
) (*Result, error) {
	return d.executePowerAction(ctx, "shutdown", "-h", opts)
}

// executePowerAction runs the shutdown command with the given flag (-r or -h).
// Uses `shutdown <flag> +N` where N is in minutes (minimum 1 minute).
func (d *Debian) executePowerAction(
	_ context.Context,
	action string,
	flag string,
	opts Opts,
) (*Result, error) {
	delayMinutes := max(opts.Delay/60, minDelayMinutes)
	if opts.Delay > 0 && opts.Delay < 60 {
		delayMinutes = minDelayMinutes
	}

	if opts.Message != "" {
		d.logger.Info(
			"scheduling power action",
			slog.String("action", action),
			slog.Int("delay_minutes", delayMinutes),
			slog.String("message", opts.Message),
		)
	}

	args := []string{flag, fmt.Sprintf("+%d", delayMinutes)}
	if opts.Message != "" {
		args = append(args, opts.Message)
	}

	if _, err := d.execManager.RunCmd("shutdown", args); err != nil {
		return nil, fmt.Errorf("power: %s: %w", action, err)
	}

	return &Result{
		Action:  action,
		Delay:   delayMinutes * 60,
		Changed: true,
	}, nil
}
