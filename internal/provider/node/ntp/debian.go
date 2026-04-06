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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/avfs/avfs"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

const (
	sourcesDir  = "/etc/chrony/sources.d"
	sourcesFile = "/etc/chrony/sources.d/osapi-ntp.sources"
)

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems
// using chrony for NTP management.
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	fs          avfs.VFS
	execManager exec.Manager
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) *Debian {
	return &Debian{
		logger:      logger.With(slog.String("subsystem", "provider.ntp")),
		fs:          fs,
		execManager: execManager,
	}
}

// Get returns current NTP sync status and configured servers by
// running chronyc tracking and chronyc sources -c.
func (d *Debian) Get(
	_ context.Context,
) (*Status, error) {
	trackingOutput, err := d.execManager.RunCmd("chronyc", []string{"tracking"})
	if err != nil {
		return nil, fmt.Errorf("ntp: chronyc tracking: %w", err)
	}

	status := parseTracking(trackingOutput)

	sourcesOutput, err := d.execManager.RunCmd("chronyc", []string{"sources", "-c"})
	if err != nil {
		return nil, fmt.Errorf("ntp: chronyc sources: %w", err)
	}

	status.Servers = parseSources(sourcesOutput)

	return status, nil
}

// Create deploys a managed NTP server configuration via a chrony
// drop-in file. Fails if the config file already exists.
func (d *Debian) Create(
	_ context.Context,
	config Config,
) (*CreateResult, error) {
	if _, err := d.fs.Stat(sourcesFile); err == nil {
		return &CreateResult{Changed: false}, nil
	}

	content := generateContent(config.Servers)

	if mkErr := d.fs.MkdirAll(sourcesDir, 0o755); mkErr != nil {
		return nil, fmt.Errorf("ntp: create directory: %w", mkErr)
	}

	if writeErr := d.fs.WriteFile(sourcesFile, content, 0o644); writeErr != nil {
		return nil, fmt.Errorf("ntp: write file: %w", writeErr)
	}

	d.reloadSources()

	d.logger.Info(
		"ntp config deployed",
		slog.Int("servers", len(config.Servers)),
		slog.Bool("changed", true),
	)

	return &CreateResult{Changed: true}, nil
}

// Update replaces the managed NTP server configuration. Fails if the
// config file does not exist. Idempotent: returns Changed false when
// the content SHA matches.
func (d *Debian) Update(
	_ context.Context,
	config Config,
) (*UpdateResult, error) {
	existing, err := d.fs.ReadFile(sourcesFile)
	if err != nil {
		return nil, fmt.Errorf("ntp: config not managed")
	}

	content := generateContent(config.Servers)
	newSHA := computeSHA256(content)
	oldSHA := computeSHA256(existing)

	if newSHA == oldSHA {
		d.logger.Debug(
			"ntp config unchanged, skipping deploy",
			slog.String("path", sourcesFile),
		)

		return &UpdateResult{Changed: false}, nil
	}

	if writeErr := d.fs.WriteFile(sourcesFile, content, 0o644); writeErr != nil {
		return nil, fmt.Errorf("ntp: write file: %w", writeErr)
	}

	d.reloadSources()

	d.logger.Info(
		"ntp config updated",
		slog.Int("servers", len(config.Servers)),
		slog.Bool("changed", true),
	)

	return &UpdateResult{Changed: true}, nil
}

// Delete removes the managed NTP server configuration. Fails if the
// config file does not exist.
func (d *Debian) Delete(
	_ context.Context,
) (*DeleteResult, error) {
	if _, err := d.fs.Stat(sourcesFile); err != nil {
		return nil, fmt.Errorf("ntp: config not managed")
	}

	if removeErr := d.fs.Remove(sourcesFile); removeErr != nil {
		return nil, fmt.Errorf("ntp: remove file: %w", removeErr)
	}

	d.reloadSources()

	d.logger.Info(
		"ntp config removed",
		slog.Bool("changed", true),
	)

	return &DeleteResult{Changed: true}, nil
}

// reloadSources runs chronyc reload sources. Failures are logged as
// warnings but do not fail the operation.
func (d *Debian) reloadSources() {
	if _, err := d.execManager.RunPrivilegedCmd("chronyc", []string{"reload", "sources"}); err != nil {
		d.logger.Warn(
			"chronyc reload sources failed",
			slog.String("error", err.Error()),
		)
	}
}

// generateContent builds the chrony sources file content from a list
// of server addresses.
func generateContent(
	servers []string,
) []byte {
	var b strings.Builder

	for _, s := range servers {
		b.WriteString("server ")
		b.WriteString(s)
		b.WriteString(" iburst\n")
	}

	return []byte(b.String())
}

// parseTracking extracts sync status fields from chronyc tracking output.
func parseTracking(
	output string,
) *Status {
	status := &Status{}

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

// computeSHA256 returns the hex-encoded SHA-256 hash of the given data.
func computeSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
