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
	"fmt"
	"log/slog"
	"strings"
)

// Start starts a systemd service. If the service is already active, it
// returns Changed: false without taking action.
func (d *Debian) Start(
	_ context.Context,
	name string,
) (*ActionResult, error) {
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("service: start: %w", err)
	}

	unitName := managedPrefix + name + ".service"

	d.logger.Debug("executing service.Start", slog.String("name", unitName))

	output, _ := d.execManager.RunCmd("systemctl", []string{"is-active", unitName})
	if strings.TrimSpace(output) == "active" {
		return &ActionResult{Name: name, Changed: false}, nil
	}

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"start", unitName}); err != nil {
		return nil, fmt.Errorf("service: start: %w", err)
	}

	return &ActionResult{Name: name, Changed: true}, nil
}

// Stop stops a systemd service. If the service is already inactive, it
// returns Changed: false without taking action.
func (d *Debian) Stop(
	_ context.Context,
	name string,
) (*ActionResult, error) {
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("service: stop: %w", err)
	}

	unitName := managedPrefix + name + ".service"

	d.logger.Debug("executing service.Stop", slog.String("name", unitName))

	output, _ := d.execManager.RunCmd("systemctl", []string{"is-active", unitName})
	if strings.TrimSpace(output) != "active" {
		return &ActionResult{Name: name, Changed: false}, nil
	}

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"stop", unitName}); err != nil {
		return nil, fmt.Errorf("service: stop: %w", err)
	}

	return &ActionResult{Name: name, Changed: true}, nil
}

// Restart restarts a systemd service. Always returns Changed: true on success.
func (d *Debian) Restart(
	_ context.Context,
	name string,
) (*ActionResult, error) {
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("service: restart: %w", err)
	}

	unitName := managedPrefix + name + ".service"

	d.logger.Debug("executing service.Restart", slog.String("name", unitName))

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"restart", unitName}); err != nil {
		return nil, fmt.Errorf("service: restart: %w", err)
	}

	return &ActionResult{Name: name, Changed: true}, nil
}

// Enable enables a systemd service. If the service is already enabled, it
// returns Changed: false without taking action.
func (d *Debian) Enable(
	_ context.Context,
	name string,
) (*ActionResult, error) {
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("service: enable: %w", err)
	}

	unitName := managedPrefix + name + ".service"

	d.logger.Debug("executing service.Enable", slog.String("name", unitName))

	output, _ := d.execManager.RunCmd("systemctl", []string{"is-enabled", unitName})
	if strings.TrimSpace(output) == "enabled" {
		return &ActionResult{Name: name, Changed: false}, nil
	}

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"enable", unitName}); err != nil {
		return nil, fmt.Errorf("service: enable: %w", err)
	}

	return &ActionResult{Name: name, Changed: true}, nil
}

// Disable disables a systemd service. If the service is already disabled, it
// returns Changed: false without taking action.
func (d *Debian) Disable(
	_ context.Context,
	name string,
) (*ActionResult, error) {
	if err := validateName(name); err != nil {
		return nil, fmt.Errorf("service: disable: %w", err)
	}

	unitName := managedPrefix + name + ".service"

	d.logger.Debug("executing service.Disable", slog.String("name", unitName))

	output, _ := d.execManager.RunCmd("systemctl", []string{"is-enabled", unitName})
	if strings.TrimSpace(output) != "enabled" {
		return &ActionResult{Name: name, Changed: false}, nil
	}

	if _, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"disable", unitName}); err != nil {
		return nil, fmt.Errorf("service: disable: %w", err)
	}

	return &ActionResult{Name: name, Changed: true}, nil
}
