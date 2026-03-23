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
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/file"
	filemocks "github.com/retr0h/osapi/internal/provider/file/mocks"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

const testHostname = "test-host"

// managedStateJSON returns a JSON-encoded FileState with no UndeployedAt,
// indicating the file is actively managed by osapi.
func managedStateJSON(
	objectName string,
	path string,
) []byte {
	state := job.FileState{
		ObjectName: objectName,
		Path:       path,
		SHA256:     "abc123",
		DeployedAt: "2026-01-01T00:00:00Z",
	}

	b, _ := json.Marshal(state)

	return b
}

type DebianPublicTestSuite struct {
	suite.Suite

	ctrl         *gomock.Controller
	logger       *slog.Logger
	memFs        afero.Fs
	mockDeployer *filemocks.MockFileDeployer
	mockStateKV  *jobmocks.MockKeyValue
	provider     *cron.Debian
}

func (suite *DebianPublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.memFs = afero.NewMemMapFs()
	suite.mockDeployer = filemocks.NewMockFileDeployer(suite.ctrl)
	suite.mockStateKV = jobmocks.NewMockKeyValue(suite.ctrl)

	_ = suite.memFs.MkdirAll("/etc/cron.d", 0o755)
	_ = suite.memFs.MkdirAll("/etc/cron.hourly", 0o755)
	_ = suite.memFs.MkdirAll("/etc/cron.daily", 0o755)
	_ = suite.memFs.MkdirAll("/etc/cron.weekly", 0o755)
	_ = suite.memFs.MkdirAll("/etc/cron.monthly", 0o755)

	suite.provider = cron.NewDebianProvider(
		suite.logger,
		suite.memFs,
		suite.mockDeployer,
		suite.mockStateKV,
		testHostname,
	)
}

func (suite *DebianPublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DebianPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		entry        cron.Entry
		setup        func()
		validateFunc func(*cron.CreateResult, error)
	}{
		{
			name: "when deploy succeeds",
			entry: cron.Entry{
				Name:   "backup",
				Object: "backup-script",
			},
			setup: func() {
				suite.mockDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(&file.DeployResult{Changed: true, Path: "/etc/cron.d/backup"}, nil)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)
			},
		},
		{
			name: "when entry already exists",
			entry: cron.Entry{
				Name:   "backup",
				Object: "backup-script",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("existing content"),
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
			name: "when deploy fails",
			entry: cron.Entry{
				Name:   "backup",
				Object: "backup-script",
			},
			setup: func() {
				suite.mockDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("deploy error"))
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
			name: "when invalid name",
			entry: cron.Entry{
				Name:   "bad name",
				Object: "backup-script",
			},
			setup: func() {},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid cron entry name")
			},
		},
		{
			name: "when interval entry",
			entry: cron.Entry{
				Name:     "logrotate",
				Object:   "logrotate-script",
				Interval: "daily",
			},
			setup: func() {
				suite.mockDeployer.EXPECT().
					Deploy(gomock.Any(), file.DeployRequest{
						ObjectName: "logrotate-script",
						Path:       "/etc/cron.daily/logrotate",
						Mode:       "0755",
					}).
					Return(&file.DeployResult{Changed: true, Path: "/etc/cron.daily/logrotate"}, nil)
			},
			validateFunc: func(
				result *cron.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("logrotate", result.Name)
				suite.True(result.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Create(context.Background(), tc.entry)

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
			name: "when deploy succeeds",
			entry: cron.Entry{
				Name:   "backup",
				Object: "backup-script-v2",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("existing content"),
					0o644,
				)
				suite.mockDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(&file.DeployResult{Changed: true, Path: "/etc/cron.d/backup"}, nil)
			},
			validateFunc: func(
				result *cron.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)
			},
		},
		{
			name: "when entry does not exist",
			entry: cron.Entry{
				Name:   "nonexistent",
				Object: "some-script",
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
			name: "when deploy fails",
			entry: cron.Entry{
				Name:   "backup",
				Object: "backup-script",
			},
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("existing content"),
					0o644,
				)
				suite.mockDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("deploy error"))
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

			result, err := suite.provider.Update(context.Background(), tc.entry)

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
			name:      "when undeploy succeeds",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("existing content"),
					0o644,
				)
				suite.mockDeployer.EXPECT().
					Undeploy(gomock.Any(), file.UndeployRequest{Path: "/etc/cron.d/backup"}).
					Return(&file.UndeployResult{Changed: true, Path: "/etc/cron.d/backup"}, nil)
			},
			validateFunc: func(
				result *cron.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", result.Name)
				suite.True(result.Changed)
			},
		},
		{
			name:      "when entry not found",
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
			name:      "when undeploy fails",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("existing content"),
					0o644,
				)
				suite.mockDeployer.EXPECT().
					Undeploy(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("undeploy error"))
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

			result, err := suite.provider.Delete(context.Background(), tc.entryName)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]cron.Entry, error)
	}{
		{
			name: "when managed entries exist",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("content"),
					0o644,
				)
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.daily/logrotate",
					[]byte("content"),
					0o755,
				)

				stateData := managedStateJSON("backup-script", "/etc/cron.d/backup")
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData).AnyTimes()

				dailyStateData := managedStateJSON("logrotate-script", "/etc/cron.daily/logrotate")
				dailyMockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				dailyMockEntry.EXPECT().Value().Return(dailyStateData).AnyTimes()

				// isManagedFile + buildEntryFromState = 2 Get calls per file.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil).
					Times(2) // backup: isManagedFile + buildEntryFromState
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(dailyMockEntry, nil).
					Times(2) // logrotate: isManagedFile + buildEntryFromState
			},
			validateFunc: func(
				entries []cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Len(entries, 2)
			},
		},
		{
			name: "when no managed entries",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/manual",
					[]byte("content"),
					0o644,
				)
				// stateKV.Get returns error => not managed.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found")).
					AnyTimes()
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
			name: "when cron dir read fails",
			setup: func() {
				// Replace provider with one pointing at a non-existent cron.d.
				badFs := afero.NewMemMapFs()
				suite.provider = cron.NewDebianProvider(
					suite.logger,
					badFs,
					suite.mockDeployer,
					suite.mockStateKV,
					testHostname,
				)
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
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			entries, err := suite.provider.List(context.Background())

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
			name:      "when managed entry found",
			entryName: "backup",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/backup",
					[]byte("content"),
					0o644,
				)
				stateData := managedStateJSON("backup-script", "/etc/cron.d/backup")
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData).AnyTimes()

				// isManagedFile + buildEntryFromState = 2 calls.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil).
					Times(2)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("backup", entry.Name)
				suite.Equal("cron.d", entry.Source)
				suite.Equal("backup-script", entry.Object)
			},
		},
		{
			name:      "when entry not managed",
			entryName: "manual",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.d/manual",
					[]byte("content"),
					0o644,
				)
				// stateKV.Get returns error => not managed.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.Error(err)
				suite.Nil(entry)
				suite.Contains(err.Error(), "not managed")
			},
		},
		{
			name:      "when entry not found",
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
			name:      "when managed entry in periodic directory",
			entryName: "logrotate",
			setup: func() {
				_ = afero.WriteFile(
					suite.memFs,
					"/etc/cron.daily/logrotate",
					[]byte("content"),
					0o755,
				)
				stateData := managedStateJSON("logrotate-script", "/etc/cron.daily/logrotate")
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData).AnyTimes()

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil).
					Times(2)
			},
			validateFunc: func(
				entry *cron.Entry,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("logrotate", entry.Name)
				suite.Equal("daily", entry.Source)
				suite.Equal("daily", entry.Interval)
				suite.Equal("logrotate-script", entry.Object)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			entry, err := suite.provider.Get(context.Background(), tc.entryName)

			tc.validateFunc(entry, err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDebianPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianPublicTestSuite))
}
