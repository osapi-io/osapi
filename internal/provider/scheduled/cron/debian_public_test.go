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
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

type DebianPublicTestSuite struct {
	suite.Suite

	logger   *slog.Logger
	fs       afero.Fs
	provider *cron.Debian
}

func (suite *DebianPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.fs = afero.NewMemMapFs()
	_ = suite.fs.MkdirAll("/etc/cron.d", 0755)
	suite.provider = cron.NewDebianProvider(suite.logger, suite.fs)
}

func (suite *DebianPublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]cron.CronEntry, error)
	}{
		{
			name:  "when directory is empty",
			setup: func() {},
			validateFunc: func(
				entries []cron.CronEntry,
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
					0644,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/cleanup",
					[]byte("# Managed by osapi\n0 2 * * * nobody /usr/bin/cleanup\n"),
					0644,
				)
			},
			validateFunc: func(
				entries []cron.CronEntry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 2)
			},
		},
		{
			name: "when non-managed files are skipped",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/managed",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0644,
				)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/manual",
					[]byte("*/10 * * * * root /usr/bin/manual\n"),
					0644,
				)
			},
			validateFunc: func(
				entries []cron.CronEntry,
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
				_ = suite.fs.MkdirAll("/etc/cron.d/subdir", 0755)
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0644,
				)
			},
			validateFunc: func(
				entries []cron.CronEntry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 1)
				suite.Equal("backup", entries[0].Name)
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
		validateFunc func(*cron.CronEntry, error)
	}{
		{
			name:      "when file exists",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0644,
				)
			},
			validateFunc: func(
				entry *cron.CronEntry,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", entry.Name)
				suite.Equal("*/5 * * * *", entry.Schedule)
				suite.Equal("root", entry.User)
				suite.Equal("/usr/bin/backup", entry.Command)
			},
		},
		{
			name:      "when file does not exist",
			entryName: "nonexistent",
			setup:     func() {},
			validateFunc: func(
				entry *cron.CronEntry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "nonexistent")
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
		entry        cron.CronEntry
		setup        func()
		validateFunc func(*cron.CreateResult, error)
	}{
		{
			name: "when creating a new entry",
			entry: cron.CronEntry{
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
				suite.Equal("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n", string(content))
			},
		},
		{
			name: "when file already exists",
			entry: cron.CronEntry{
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
					0644,
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
			entry: cron.CronEntry{
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
			entry: cron.CronEntry{
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
			name: "when user is empty defaults to root",
			entry: cron.CronEntry{
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
		entry        cron.CronEntry
		setup        func()
		validateFunc func(*cron.UpdateResult, error)
	}{
		{
			name: "when content changes",
			entry: cron.CronEntry{
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
					0644,
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
			entry: cron.CronEntry{
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
					0644,
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
			name: "when file does not exist",
			entry: cron.CronEntry{
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
			name:      "when file exists",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.fs,
					"/etc/cron.d/backup",
					[]byte("# Managed by osapi\n*/5 * * * * root /usr/bin/backup\n"),
					0644,
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
