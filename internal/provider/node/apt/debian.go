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

package apt

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

// dpkgQueryFormat is the format string for dpkg-query output.
const dpkgQueryFormat = "${Package}\t${Version}\t${binary:Summary}\t${db:Status-Abbrev}\t${Installed-Size}\n"

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems
// using dpkg-query and apt-get.
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
		logger:      logger.With(slog.String("subsystem", "provider.apt")),
		execManager: execManager,
	}
}

// List returns all installed packages by querying dpkg.
func (d *Debian) List(
	_ context.Context,
) ([]Package, error) {
	output, err := d.execManager.RunCmd(
		"dpkg-query",
		[]string{"-W", "-f", dpkgQueryFormat},
	)
	if err != nil {
		return nil, fmt.Errorf("package: list: %w", err)
	}

	return d.parsePackages(output), nil
}

// Get returns details for a single installed package.
func (d *Debian) Get(
	_ context.Context,
	name string,
) (*Package, error) {
	output, err := d.execManager.RunCmd(
		"dpkg-query",
		[]string{"-W", "-f", dpkgQueryFormat, name},
	)
	if err != nil {
		return nil, fmt.Errorf("package: get %q: %w", name, err)
	}

	pkgs := d.parsePackages(output)
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package: get %q: not found", name)
	}

	return &pkgs[0], nil
}

// Install installs a package by name using apt-get.
func (d *Debian) Install(
	_ context.Context,
	name string,
) (*Result, error) {
	_, err := d.execManager.RunCmd(
		"apt-get",
		[]string{"install", "-y", name},
	)
	if err != nil {
		return nil, fmt.Errorf("package: install %q: %w", name, err)
	}

	d.logger.Info(
		"package installed",
		slog.String("name", name),
	)

	return &Result{
		Name:    name,
		Changed: true,
	}, nil
}

// Remove removes a package by name using apt-get.
func (d *Debian) Remove(
	_ context.Context,
	name string,
) (*Result, error) {
	_, err := d.execManager.RunCmd(
		"apt-get",
		[]string{"remove", "-y", name},
	)
	if err != nil {
		return nil, fmt.Errorf("package: remove %q: %w", name, err)
	}

	d.logger.Info(
		"package removed",
		slog.String("name", name),
	)

	return &Result{
		Name:    name,
		Changed: true,
	}, nil
}

// Update refreshes the package index using apt-get update.
func (d *Debian) Update(
	_ context.Context,
) (*Result, error) {
	_, err := d.execManager.RunCmd(
		"apt-get",
		[]string{"update"},
	)
	if err != nil {
		return nil, fmt.Errorf("package: update: %w", err)
	}

	d.logger.Info("package index updated")

	return &Result{
		Changed: true,
	}, nil
}

// ListUpdates returns packages with available updates by parsing
// apt list --upgradable output.
func (d *Debian) ListUpdates(
	_ context.Context,
) ([]Update, error) {
	output, err := d.execManager.RunCmd(
		"apt",
		[]string{"list", "--upgradable"},
	)
	if err != nil {
		return nil, fmt.Errorf("package: list updates: %w", err)
	}

	return d.parseUpdates(output), nil
}

// parsePackages parses dpkg-query tab-separated output into Package
// slices. Only lines with status starting with "ii" (installed) are
// included. Installed-Size is converted from KB to bytes.
func (d *Debian) parsePackages(
	output string,
) []Package {
	var result []Package

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "\t", 5)
		if len(fields) < 5 {
			continue
		}

		status := strings.TrimSpace(fields[3])
		if !strings.HasPrefix(status, "ii") {
			continue
		}

		sizeKB, _ := strconv.ParseInt(strings.TrimSpace(fields[4]), 10, 64)

		result = append(result, Package{
			Name:        strings.TrimSpace(fields[0]),
			Version:     strings.TrimSpace(fields[1]),
			Description: strings.TrimSpace(fields[2]),
			Status:      "installed",
			Size:        sizeKB * 1024,
		})
	}

	return result
}

// parseUpdates parses apt list --upgradable output. Each line has the
// format: package/source version arch [upgradable from: oldversion]
// The first line ("Listing...") is skipped.
func (d *Debian) parseUpdates(
	output string,
) []Update {
	var result []Update

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}

		// Format: name/source version arch [upgradable from: oldversion]
		slashIdx := strings.Index(line, "/")
		if slashIdx < 0 {
			continue
		}

		name := line[:slashIdx]

		// Extract new version: between first space and second space.
		rest := line[slashIdx+1:]
		parts := strings.Fields(rest)

		if len(parts) < 6 {
			continue
		}

		// parts[0] = source, parts[1] = version, parts[2] = arch,
		// parts[3] = [upgradable, parts[4] = from:, parts[5] = oldversion]
		newVersion := parts[1]
		currentVersion := strings.TrimSuffix(parts[5], "]")

		result = append(result, Update{
			Name:           name,
			CurrentVersion: currentVersion,
			NewVersion:     newVersion,
		})
	}

	return result
}
