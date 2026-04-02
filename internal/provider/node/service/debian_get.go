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
	"strconv"
	"strings"
)

// Get returns a single systemd service by name using systemctl show.
func (d *Debian) Get(
	_ context.Context,
	name string,
) (*Info, error) {
	d.logger.Debug(
		"executing service.Get",
		slog.String("name", name),
	)

	if err := validateName(name); err != nil {
		return nil, err
	}

	output, err := d.execManager.RunCmd("systemctl", []string{
		"show",
		name,
		"--property=ActiveState,UnitFileState,Description,MainPID",
		"--no-pager",
	})
	if err != nil {
		return nil, fmt.Errorf("service: get: %w", err)
	}

	props, err := parseProperties(output)
	if err != nil {
		return nil, fmt.Errorf("service: get: %w", err)
	}

	pid, _ := strconv.Atoi(props["MainPID"])

	info := &Info{
		Name:        name,
		Status:      props["ActiveState"],
		Enabled:     props["UnitFileState"] == "enabled",
		Description: props["Description"],
		PID:         pid,
	}

	return info, nil
}

// parseProperties parses key=value output from systemctl show.
func parseProperties(
	output string,
) (map[string]string, error) {
	props := make(map[string]string)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"malformed property line: %q",
				line,
			)
		}

		props[parts[0]] = parts[1]
	}

	return props, nil
}
