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

package cron

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

const (
	cronDir       = "/etc/cron.d"
	managedHeader = "# Managed by osapi"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Debian implements the Provider interface for Debian-family systems.
type Debian struct {
	logger *slog.Logger
	fs     afero.Fs
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	fs afero.Fs,
) *Debian {
	return &Debian{
		logger: logger,
		fs:     fs,
	}
}

// List returns all osapi-managed cron entries from /etc/cron.d/.
func (d *Debian) List() ([]Entry, error) {
	entries, err := afero.ReadDir(d.fs, cronDir)
	if err != nil {
		return nil, fmt.Errorf("list cron entries: %w", err)
	}

	var result []Entry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		cronEntry, err := d.readCronFile(entry.Name())
		if err != nil {
			continue
		}

		result = append(result, *cronEntry)
	}

	return result, nil
}

// Get returns a single cron entry by name.
func (d *Debian) Get(
	name string,
) (*Entry, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	return d.readCronFile(name)
}

// Create writes a new cron drop-in file.
func (d *Debian) Create(
	entry Entry,
) (*CreateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := cronDir + "/" + entry.Name

	exists, err := afero.Exists(d.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("create cron entry: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("cron entry %q already exists", entry.Name)
	}

	content := buildFileContent(entry)
	if err := afero.WriteFile(d.fs, filePath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("create cron entry: %w", err)
	}

	return &CreateResult{
		Name:    entry.Name,
		Changed: true,
	}, nil
}

// Update overwrites an existing cron drop-in file.
func (d *Debian) Update(
	entry Entry,
) (*UpdateResult, error) {
	if err := validateName(entry.Name); err != nil {
		return nil, err
	}

	filePath := cronDir + "/" + entry.Name

	exists, err := afero.Exists(d.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("update cron entry: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("cron entry %q does not exist", entry.Name)
	}

	newContent := buildFileContent(entry)

	current, err := afero.ReadFile(d.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("update cron entry: %w", err)
	}

	if string(current) == newContent {
		return &UpdateResult{
			Name:    entry.Name,
			Changed: false,
		}, nil
	}

	if err := afero.WriteFile(d.fs, filePath, []byte(newContent), 0o644); err != nil {
		return nil, fmt.Errorf("update cron entry: %w", err)
	}

	return &UpdateResult{
		Name:    entry.Name,
		Changed: true,
	}, nil
}

// Delete removes a cron drop-in file.
func (d *Debian) Delete(
	name string,
) (*DeleteResult, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	filePath := cronDir + "/" + name

	exists, err := afero.Exists(d.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("delete cron entry: %w", err)
	}
	if !exists {
		return &DeleteResult{
			Name:    name,
			Changed: false,
		}, nil
	}

	if err := d.fs.Remove(filePath); err != nil {
		return nil, fmt.Errorf("delete cron entry: %w", err)
	}

	return &DeleteResult{
		Name:    name,
		Changed: true,
	}, nil
}

// readCronFile reads and parses a single cron drop-in file.
// Returns nil if the file is not managed by osapi.
func (d *Debian) readCronFile(
	name string,
) (*Entry, error) {
	filePath := cronDir + "/" + name

	data, err := afero.ReadFile(d.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("read cron entry %q: %w", name, err)
	}

	content := string(data)
	if !strings.HasPrefix(content, managedHeader) {
		return nil, fmt.Errorf("cron entry %q is not managed by osapi", name)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("cron entry %q has invalid format", name)
	}

	// Parse the cron line: SCHEDULE USER COMMAND
	cronLine := lines[1]
	parts := strings.Fields(cronLine)
	if len(parts) < 7 {
		return nil, fmt.Errorf("cron entry %q has invalid cron line", name)
	}

	schedule := strings.Join(parts[:5], " ")
	user := parts[5]
	command := strings.Join(parts[6:], " ")

	return &Entry{
		Name:     name,
		Schedule: schedule,
		User:     user,
		Command:  command,
	}, nil
}

// validateName checks that a cron entry name is safe for use as a filename.
func validateName(
	name string,
) error {
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid cron entry name %q: contains '..'", name)
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("invalid cron entry name %q: contains '/'", name)
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid cron entry name %q: must match %s", name, validName.String())
	}

	return nil
}

// buildFileContent generates the content for a cron drop-in file.
func buildFileContent(
	entry Entry,
) string {
	user := entry.User
	if user == "" {
		user = "root"
	}

	return fmt.Sprintf("%s\n%s %s %s\n", managedHeader, entry.Schedule, user, entry.Command)
}
