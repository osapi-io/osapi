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

package ntp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// The file that osapi writes the servers to: chrony reads each file of
// sourcesDir, and chronyc reload sources makes it read them again.
const (
	sourcesDir  = "/etc/chrony/sources.d"
	sourcesFile = sourcesDir + "/osapi-ntp.sources"
)

// debian reads chrony's state with chronyc and writes the servers through
// managedfile.
type debian struct {
	logger *slog.Logger
	exec   host.Exec
	files  managedfile.Provider
}

// Get reads chronyc tracking and chronyc sources.
func (d *debian) Get(
	ctx context.Context,
) (Status, error) {
	tracking, err := d.exec.RunCmd(ctx, "chronyc", []string{"tracking"})
	if err != nil {
		return Status{}, fmt.Errorf("ntp get: chronyc tracking: %w", err)
	}
	status := parseTracking(tracking)

	sources, err := d.exec.RunCmd(ctx, "chronyc", []string{"sources", "-c"})
	if err != nil {
		return Status{}, fmt.Errorf("ntp get: chronyc sources: %w", err)
	}
	status.Servers = parseSources(sources)

	return status, nil
}

// Create writes the servers, and reloads chrony's sources when the file
// changed.
func (d *debian) Create(
	ctx context.Context,
	servers []string,
) (bool, error) {
	return d.write(ctx, "create", servers)
}

// Update writes the servers when osapi manages them already.
func (d *debian) Update(
	ctx context.Context,
	servers []string,
) (bool, error) {
	record, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
		managedfile.GetInput{Path: sourcesFile})
	switch {
	case errors.Is(err, managedfile.ErrNotManaged), err == nil && record.Undeployed:
		return false, fmt.Errorf("ntp update: osapi manages no servers: %w", behavior.ErrNotFound)
	case err != nil:
		return false, fmt.Errorf("ntp update: %w", err)
	}

	return d.write(ctx, "update", servers)
}

// Delete removes the file, and reloads chrony's sources when it was there.
func (d *debian) Delete(
	ctx context.Context,
) (bool, error) {
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: sourcesFile})
	if err != nil {
		return false, fmt.Errorf("ntp delete: %w", err)
	}
	if result.Changed {
		d.reload(ctx)
	}

	return result.Changed, nil
}

// write makes sourcesFile hold the servers.
func (d *debian) write(
	ctx context.Context,
	op string,
	servers []string,
) (bool, error) {
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
		managedfile.UpdateInput{Path: sourcesFile, Content: content(servers)})
	if err != nil {
		return false, fmt.Errorf("ntp %s: %w", op, err)
	}
	if result.Changed {
		d.reload(ctx)
	}

	return result.Changed, nil
}

// reload runs chronyc reload sources. A failure is a warning: the file is
// written, and chrony reads it at its next start.
func (d *debian) reload(
	ctx context.Context,
) {
	if _, err := d.exec.RunPrivilegedCmd(
		ctx,
		"chronyc",
		[]string{"reload", "sources"},
	); err != nil {
		d.logger.Warn("chronyc reload sources failed", slog.String("error", err.Error()))
	}
}

// content is the sources file for servers.
func content(
	servers []string,
) string {
	var b strings.Builder
	for _, s := range servers {
		b.WriteString("server " + s + " iburst\n")
	}

	return b.String()
}

// parseTracking extracts sync status fields from chronyc tracking output.
func parseTracking(
	output string,
) Status {
	status := Status{}

	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Leap status":
			status.Synchronized = value == "Normal"
		case "Stratum":
			if n, err := strconv.Atoi(value); err == nil {
				status.Stratum = n
			}
		case "System time":
			status.Offset = parseOffset(value)
		case "Reference ID":
			status.CurrentSource = parseReferenceID(value)
		}
	}

	return status
}

// parseOffset extracts a human-readable offset from the System time line.
// Input: "0.000003422 seconds fast of NTP time" → "+0.000003422s"
// Input: "0.000001834 seconds slow of NTP time" → "-0.000001834s"
func parseOffset(
	value string,
) string {
	fields := strings.Fields(value)
	if len(fields) < 3 {
		return ""
	}

	sign := "+"
	if fields[2] == "slow" {
		sign = "-"
	}

	return sign + fields[0] + "s"
}

// refIDRegexp matches the parenthesized hostname in a Reference ID line.
var refIDRegexp = regexp.MustCompile(`\(([^)]+)\)`)

// parseReferenceID extracts the hostname from parens in the Reference ID line.
// Input: "A29FC801 (time.cloudflare.com)" → "time.cloudflare.com"
func parseReferenceID(
	value string,
) string {
	m := refIDRegexp.FindStringSubmatch(value)
	if len(m) < 2 {
		return ""
	}

	return m[1]
}

// parseSources extracts server names from chronyc sources -c CSV output.
// Each line has comma-separated fields; the Name is at index 2.
func parseSources(
	output string,
) []string {
	var servers []string

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			continue
		}

		name := strings.TrimSpace(fields[2])
		if name != "" {
			servers = append(servers, name)
		}
	}

	return servers
}
