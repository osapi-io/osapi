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
)

// systemctlUnit represents a single unit from systemctl list-units JSON output.
type systemctlUnit struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

// systemctlUnitFile represents a single unit file from systemctl list-unit-files JSON output.
type systemctlUnitFile struct {
	UnitFile string `json:"unit_file"`
	State    string `json:"state"`
	Preset   string `json:"preset"`
}

// List returns all systemd services by merging list-units and list-unit-files output.
func (d *Debian) List(
	_ context.Context,
) ([]Info, error) {
	d.logger.Debug("executing service.List")

	unitsJSON, err := d.execManager.RunCmd("systemctl", []string{
		"list-units",
		"--type=service",
		"--all",
		"--no-pager",
		"--output=json",
	})
	if err != nil {
		return nil, fmt.Errorf("service: list: %w", err)
	}

	var units []systemctlUnit
	if err := json.Unmarshal([]byte(unitsJSON), &units); err != nil {
		return nil, fmt.Errorf("service: list: parse units: %w", err)
	}

	enabledMap := d.buildEnabledMap()

	result := make([]Info, 0, len(units))
	for _, u := range units {
		info := Info{
			Name:        u.Unit,
			Status:      u.Active,
			Enabled:     enabledMap[u.Unit],
			Description: u.Description,
		}

		result = append(result, info)
	}

	return result, nil
}

// buildEnabledMap runs systemctl list-unit-files and returns a map of
// service name to enabled status. Errors are logged and result in an
// empty map (all services default to enabled=false).
func (d *Debian) buildEnabledMap() map[string]bool {
	unitFilesJSON, err := d.execManager.RunCmd("systemctl", []string{
		"list-unit-files",
		"--type=service",
		"--no-pager",
		"--output=json",
	})
	if err != nil {
		d.logger.Debug(
			"failed to list unit files, enabled status will be unavailable",
			slog.String("error", err.Error()),
		)

		return map[string]bool{}
	}

	var unitFiles []systemctlUnitFile
	if err := json.Unmarshal([]byte(unitFilesJSON), &unitFiles); err != nil {
		d.logger.Debug(
			"failed to parse unit files, enabled status will be unavailable",
			slog.String("error", err.Error()),
		)

		return map[string]bool{}
	}

	enabledMap := make(map[string]bool, len(unitFiles))
	for _, uf := range unitFiles {
		enabledMap[uf.UnitFile] = uf.State == "enabled"
	}

	return enabledMap
}
