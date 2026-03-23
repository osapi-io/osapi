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

package cron_test

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// errorStatFs wraps an afero.Fs and returns a configurable error from Stat.
type errorStatFs struct {
	afero.Fs
	statErr error
}

func (e *errorStatFs) Stat(
	_ string,
) (fs.FileInfo, error) {
	return nil, e.statErr
}

// errorRemoveFs wraps an afero.Fs and returns a configurable error from Remove.
type errorRemoveFs struct {
	afero.Fs
	removeErr error
}

func (e *errorRemoveFs) Remove(
	_ string,
) error {
	return e.removeErr
}

// errorWriteFileFs wraps an afero.Fs and returns a configurable error from
// OpenFile with write flags (used by afero.WriteFile internally).
type errorWriteFileFs struct {
	afero.Fs
	createErr error
}

func (e *errorWriteFileFs) OpenFile(
	name string,
	flag int,
	perm fs.FileMode,
) (afero.File, error) {
	// Only fail on write-mode opens; allow read-only opens so Exists works.
	if flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0 || flag&os.O_CREATE != 0 {
		return nil, e.createErr
	}

	return e.Fs.OpenFile(name, flag, perm)
}

// errorReadFileFs wraps an afero.Fs and returns a configurable error from Open
// (used by afero.ReadFile internally), while allowing Stat to succeed.
type errorReadFileFs struct {
	afero.Fs
	openErr error
}

func (e *errorReadFileFs) Open(
	_ string,
) (afero.File, error) {
	return nil, e.openErr
}

func (e *errorReadFileFs) OpenFile(
	_ string,
	_ int,
	_ fs.FileMode,
) (afero.File, error) {
	return nil, e.openErr
}

type DebianPublicTestSuite struct {
	suite.Suite

	logger   *slog.Logger
	fs       afero.Fs
	provider *cron.Debian
}

func (suite *DebianPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.fs = afero.NewMemMapFs()
	_ = suite.fs.MkdirAll("/etc/cron.d", 0o755)
	_ = suite.fs.MkdirAll("/etc/cron.hourly", 0o755)
	_ = suite.fs.MkdirAll("/etc/cron.daily", 0o755)
	_ = suite.fs.MkdirAll("/etc/cron.weekly", 0o755)
	_ = suite.fs.MkdirAll("/etc/cron.monthly", 0o755)
	suite.provider = cron.NewDebianProvider(suite.logger, suite.fs)
}

func (suite *DebianPublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]cron.Entry, error)
	}{
		{
			name:  "when directory is empty",
			setup: func() {},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Empty(entries)
			},
		},
		{
			name: "when managed files exist",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/cleanup",
					[]byte("# Managed by osapi\n0 2 * * * nobody /usr/bin/cleanup\n"),
					0o644,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 2)
				suite.Equal("cron.d", entries[0].Source)
			},
		},
		{
			name: "when non-managed files are skipped",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/managed",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/manual",
					[]byte("*/10 * * * * root /usr/bin/manual\n"),
					0o644,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 1)
				suite.Equal("managed", entries[0].Name)
			},
		},
		{
			name: "when directories are skipped",
			setup: func() {
				_ = suite.fs.MkdirAll("/etc/cron.d/subdir", 0o755)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 1)
				suite.Equal("backup", entries[0].Name)
			},
		},
		{
			name: "when ReadDir fails",
			setup: func() {
				// Replace the provider with one pointing at a non-existent dir.
				badFs := afero.NewMemMapFs()
				suite.provider = cron.NewDebianProvider(suite.logger, badFs)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entries)
				suite.Contains(err.Error(), "list cron entries")
			},
		},
		{
			name: "when periodic directory does not exist ReadDir is skipped",
			setup: func() {
				// Remove one periodic dir so ReadDir fails for it.
				noMonthlyFs := afero.NewMemMapFs()
				_ = noMonthlyFs.MkdirAll("/etc/cron.d", 0o755)
				_ = noMonthlyFs.MkdirAll("/etc/cron.hourly", 0o755)
				_ = noMonthlyFs.MkdirAll("/etc/cron.daily", 0o755)
				_ = noMonthlyFs.MkdirAll("/etc/cron.weekly", 0o755)
				// /etc/cron.monthly intentionally missing
				_ = afero.WriteFile(
					noMonthlyFs,
					"/etc/cron.daily/backup",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/bin/backup\n"),
					0o755,
				)
				suite.provider = cron.NewDebianProvider(suite.logger, noMonthlyFs)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 1)
				suite.Equal("daily", entries[0].Source)
			},
		},
		{
			name: "when periodic entries exist",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte(
						"#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate /etc/logrotate.conf\n",
					),
					0o755,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.hourly/health-check",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/bin/health-check\n"),
					0o755,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 2)
				// hourly comes before daily in ordering
				suite.Equal("hourly", entries[0].Source)
				suite.Equal("hourly", entries[0].Interval)
				suite.Equal("/usr/bin/health-check", entries[0].Command)
				suite.Equal("daily", entries[1].Source)
				suite.Equal("daily", entries[1].Interval)
			},
		},
		{
			name: "when mixed cron.d and periodic entries exist",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.weekly/gc",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/bin/gc\n"),
					0o755,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 2)
				suite.Equal("backup", entries[0].Name)
				suite.Equal("cron.d", entries[0].Source)
				suite.Equal("gc", entries[1].Name)
				suite.Equal("weekly", entries[1].Source)
			},
		},
		{
			name: "when non-managed periodic files are skipped",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/manual",
					[]byte("#!/bin/sh\n/usr/bin/manual\n"),
					0o755,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Empty(entries)
			},
		},
		{
			name: "when periodic directory has subdirectory",
			setup: func() {
				_ = suite.fs.MkdirAll("/etc/cron.daily/subdir", 0o755)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/managed",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/bin/run\n"),
					0o755,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 1)
				suite.Equal("managed", entries[0].Name)
			},
		},
		{
			name: "when periodic file has invalid format",
			setup: func() {
				// Only two lines instead of required three.
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/short",
					[]byte("#!/bin/sh\n# Managed by osapi"),
					0o755,
				)
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Empty(entries)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			entries, err := suite.provider.List()

			tc.validateFunc(entries, err)
		})
	}
}

func (suite *DebianPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		entryName    string
		setup        func()
		validateFunc func(*cron.Entry, error)
	}{
		{
			name:      "when file exists in cron.d",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", entry.Name)
				suite.Equal("*/5 * * * *", entry.Schedule)
				suite.Equal("root", entry.User)
				suite.Equal("/usr/bin/backup", entry.Command)
				suite.Equal("cron.d", entry.Source)
			},
		},
		{
			name:      "when file exists in periodic directory",
			entryName: "logrotate",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte(
						"#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate /etc/logrotate.conf\n",
					),
					0o755,
				)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("logrotate", entry.Name)
				suite.Equal("daily", entry.Interval)
				suite.Equal("daily", entry.Source)
				suite.Equal("/usr/sbin/logrotate /etc/logrotate.conf", entry.Command)
				suite.Empty(entry.Schedule)
				suite.Empty(entry.User)
			},
		},
		{
			name:      "when file does not exist anywhere",
			entryName: "nonexistent",
			setup:     func() {},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:      "when name is invalid",
			entryName: "bad name",
			setup:     func() {},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "invalid cron entry name")
			},
		},
		{
			name:      "when file has only header line",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi"),
					0o644,
				)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				// cron.d fails, then periodic dirs searched, then "not found".
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:      "when cron line has too few fields",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n* * * * * root\n"),
					0o644,
				)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				// cron.d fails, then periodic dirs searched, then "not found".
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:      "when cron.d entry is not managed and periodic not found",
			entryName: "manual",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/manual",
					[]byte("*/10 * * * * root /usr/bin/manual\n"),
					0o644,
				)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "not found")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			entry, err := suite.provider.Get(tc.entryName)

			tc.validateFunc(entry, err)
		})
	}
}

func (suite *DebianPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		entry        cron.Entry
		setup        func()
		validateFunc func(*cron.CreateResult, error)
	}{
		{
			name: "when creating a new schedule entry",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.d/backup")
				suite.NoError(readErr)
				suite.Equal(
					"# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n",
					string(content),
				)
			},
		},
		{
			name: "when creating an interval entry",
			entry: cron.Entry{
				Name:     "logrotate",
				Interval: "daily",
				Command:  "/usr/sbin/logrotate /etc/logrotate.conf",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("logrotate", result.Name)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.daily/logrotate")
				suite.NoError(readErr)
				suite.Equal(
					"#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate /etc/logrotate.conf\n",
					string(content),
				)

				// Verify executable permissions.
				info, statErr := suite.fs.Stat("/etc/cron.daily/logrotate")
				suite.NoError(statErr)
				suite.Equal(fs.FileMode(0o755), info.Mode().Perm())
			},
		},
		{
			name: "when creating hourly interval entry",
			entry: cron.Entry{
				Name:     "health-check",
				Interval: "hourly",
				Command:  "/usr/bin/health-check",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.hourly/health-check")
				suite.NoError(readErr)
				suite.Contains(string(content), "#!/bin/sh")
				suite.Contains(string(content), "# Managed by osapi")
				suite.Contains(string(content), "/usr/bin/health-check")
			},
		},
		{
			name: "when creating weekly interval entry",
			entry: cron.Entry{
				Name:     "gc",
				Interval: "weekly",
				Command:  "/usr/bin/gc",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				exists, _ := afero.Exists(suite.fs, "/etc/cron.weekly/gc")
				suite.True(exists)
			},
		},
		{
			name: "when creating monthly interval entry",
			entry: cron.Entry{
				Name:     "report",
				Interval: "monthly",
				Command:  "/usr/bin/report",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				exists, _ := afero.Exists(suite.fs, "/etc/cron.monthly/report")
				suite.True(exists)
			},
		},
		{
			name: "when interval entry already exists",
			entry: cron.Entry{
				Name:     "logrotate",
				Interval: "daily",
				Command:  "/usr/sbin/logrotate",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate\n"),
					0o755,
				)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already exists")
			},
		},
		{
			name: "when file already exists",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already exists")
			},
		},
		{
			name: "when name exists in different directory (cron.d vs daily)",
			entry: cron.Entry{
				Name:     "backup",
				Interval: "daily",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n0 2 * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already exists")
			},
		},
		{
			name: "when name exists in different directory (daily vs cron.d)",
			entry: cron.Entry{
				Name:     "logrotate",
				Schedule: "0 0 * * *",
				User:     "root",
				Command:  "/usr/sbin/logrotate",
			},
			setup: func() {
				_ = suite.fs.MkdirAll("/etc/cron.daily", 0o755)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate\n"),
					0o755,
				)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already exists")
			},
		},
		{
			name: "when name contains slash",
			entry: cron.Entry{
				Name:     "bad/name",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "contains '/'")
			},
		},
		{
			name: "when name contains dot-dot",
			entry: cron.Entry{
				Name:     "bad..name",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "contains '..'")
			},
		},
		{
			name: "when name has invalid characters",
			entry: cron.Entry{
				Name:     "bad name",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "must match")
			},
		},
		{
			name: "when user is empty defaults to root",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.d/backup")
				suite.NoError(readErr)
				suite.Contains(string(content), "root /usr/bin/backup")
			},
		},
		{
			name: "when WriteFile fails during create",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				// Use a ReadOnlyFs to make writes fail.
				base := afero.NewMemMapFs()
				_ = base.MkdirAll("/etc/cron.d", 0o755)
				readOnly := afero.NewReadOnlyFs(base)
				suite.provider = cron.NewDebianProvider(suite.logger, readOnly)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "create cron entry")
			},
		},
		{
			name: "when WriteFile fails",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				errFs := &errorWriteFileFs{
					Fs:        suite.fs,
					createErr: errors.New("write error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "create cron entry")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Create(tc.entry)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *DebianPublicTestSuite) TestUpdate() {
	tests := []struct {
		name         string
		entry        cron.Entry
		setup        func()
		validateFunc func(*cron.UpdateResult, error)
	}{
		{
			name: "when content changes",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 2 * * *",
				User:     "root",
				Command:  "/usr/bin/backup --full",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.d/backup")
				suite.NoError(readErr)
				suite.Equal(
					"# Managed by osapi\n0 2 * * * root /usr/bin/backup --full\n",
					string(content),
				)
			},
		},
		{
			name: "when content matches",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.False(result.Changed)
			},
		},
		{
			name: "when updating periodic entry with changed content",
			entry: cron.Entry{
				Name:     "logrotate",
				Interval: "daily",
				Command:  "/usr/sbin/logrotate --force",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate\n"),
					0o755,
				)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := afero.ReadFile(suite.fs, "/etc/cron.daily/logrotate")
				suite.NoError(readErr)
				suite.Equal(
					"#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate --force\n",
					string(content),
				)
			},
		},
		{
			name: "when updating periodic entry with matching content",
			entry: cron.Entry{
				Name:     "logrotate",
				Interval: "daily",
				Command:  "/usr/sbin/logrotate",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate\n"),
					0o755,
				)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name: "when name is invalid",
			entry: cron.Entry{
				Name:     "bad name",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid cron entry name")
			},
		},
		{
			name: "when file does not exist",
			entry: cron.Entry{
				Name:     "nonexistent",
				Schedule: "*/5 * * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "does not exist")
			},
		},
		{
			name: "when Exists check fails",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 2 * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				errFs := &errorStatFs{
					Fs:      suite.fs,
					statErr: errors.New("stat error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "does not exist")
			},
		},
		{
			name: "when ReadFile fails",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 2 * * *",
				User:     "root",
				Command:  "/usr/bin/backup",
			},
			setup: func() {
				// Write a file through the real fs so Stat/Exists succeeds,
				// then wrap with a fs that fails Open so ReadFile errors.
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				errFs := &errorReadFileFs{
					Fs:      suite.fs,
					openErr: errors.New("read error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "update cron entry")
			},
		},
		{
			name: "when WriteFile fails on changed content",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 3 * * *",
				User:     "root",
				Command:  "/usr/bin/backup --new",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				errFs := &errorWriteFileFs{
					Fs:        suite.fs,
					createErr: errors.New("write error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "update cron entry")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Update(tc.entry)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *DebianPublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		entryName    string
		setup        func()
		validateFunc func(*cron.DeleteResult, error)
	}{
		{
			name:      "when file exists in cron.d",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
			},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)

				exists, _ := afero.Exists(suite.fs, "/etc/cron.d/backup")
				suite.False(exists)
			},
		},
		{
			name:      "when file exists in periodic directory",
			entryName: "logrotate",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.daily/logrotate",
					[]byte("#!/bin/sh\n# Managed by osapi\n/usr/sbin/logrotate\n"),
					0o755,
				)
			},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("logrotate", result.Name)
				suite.True(result.Changed)

				exists, _ := afero.Exists(suite.fs, "/etc/cron.daily/logrotate")
				suite.False(exists)
			},
		},
		{
			name:      "when file does not exist",
			entryName: "nonexistent",
			setup:     func() {},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("nonexistent", result.Name)
				suite.False(result.Changed)
			},
		},
		{
			name:      "when Exists check fails",
			entryName: "backup",
			setup: func() {
				errFs := &errorStatFs{
					Fs:      suite.fs,
					statErr: errors.New("stat error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name:      "when Remove fails",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0o644,
				)
				errFs := &errorRemoveFs{
					Fs:        suite.fs,
					removeErr: errors.New("remove error"),
				}
				suite.provider = cron.NewDebianProvider(suite.logger, errFs)
			},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "delete cron entry")
			},
		},
		{
			name:      "when name is invalid",
			entryName: "bad name",
			setup:     func() {},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid cron entry name")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Delete(tc.entryName)

			tc.validateFunc(result, err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDebianPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianPublicTestSuite))
}
