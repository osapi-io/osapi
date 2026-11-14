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

package sysctl_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/sysctl"
	execmocks "github.com/osapi-io/osapi/internal/exec/mocks"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

const testHostname = "test-host"

// managedStateJSON returns a JSON-encoded FileState with sysctl metadata.
func managedStateJSON(
	key string,
	value string,
	path string,
) []byte {
	state := managedfile.State{
		Path:       path,
		SHA256:     "abc123",
		Mode:       "0644",
		DeployedAt: "2026-01-01T00:00:00Z",
		Metadata: map[string]string{
			"key":   key,
			"value": value,
		},
	}

	b, _ := json.Marshal(state)

	return b
}

type DebianPublicTestSuite struct {
	suite.Suite

	ctrl        *gomock.Controller
	logger      *slog.Logger
	memFs       avfs.VFS
	mockStateKV *jobmocks.MockKeyValue
	mockExec    *execmocks.MockManager
	provider    *sysctl.Debian
}

func (suite *DebianPublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.memFs = memfs.New()
	suite.mockStateKV = jobmocks.NewMockKeyValue(suite.ctrl)
	suite.mockExec = execmocks.NewMockManager(suite.ctrl)

	_ = suite.memFs.MkdirAll("/etc/sysctl.d", 0o755)

	suite.provider = sysctl.NewDebianProvider(
		suite.logger,
		suite.memFs,
		suite.mockStateKV,
		suite.mockExec,
		testHostname,
	)
}

func (suite *DebianPublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DebianPublicTestSuite) TearDownSubTest() {
}

func (suite *DebianPublicTestSuite) TestWithFiles() {
	tests := []struct {
		name string
	}{
		{name: "when given a managed file provider, the conf files go through it"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			otherKV := jobmocks.NewMockKeyValue(suite.ctrl)
			files := managedfile.NewDebian(suite.memFs, otherKV, nil, testHostname)
			p := suite.provider.WithFiles(files)
			otherKV.EXPECT().
				Put(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(uint64(1), nil)
			suite.mockExec.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "sysctl", gomock.Any()).
				Return("", nil)

			result, err := p.Create(
				context.Background(),
				sysctl.Entry{Key: "vm.swappiness", Value: "10"},
			)

			suite.NoError(err)
			suite.True(
				result.Changed,
				"the record went to the provider given, not the default store",
			)
		})
	}
}

func (suite *DebianPublicTestSuite) TestWithDir() {
	tests := []struct {
		name     string
		dir      string
		wantPath string
	}{
		{
			name:     "when behaviors.sysctl.dir is set the files go there",
			dir:      "/srv/sysctl.d",
			wantPath: "/srv/sysctl.d/osapi-vm.swappiness.conf",
		},
		{
			name:     "when it is empty the default stays",
			wantPath: sysctl.DefaultDir + "/osapi-vm.swappiness.conf",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			p := suite.provider.WithDir(tc.dir)
			suite.mockStateKV.EXPECT().
				Put(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(uint64(1), nil)
			suite.mockExec.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", tc.wantPath}).
				Return("", nil)

			_, err := p.Create(
				context.Background(),
				sysctl.Entry{Key: "vm.swappiness", Value: "10"},
			)
			suite.Require().NoError(err)

			content, readErr := suite.memFs.ReadFile(tc.wantPath)
			suite.Require().NoError(readErr)
			suite.Equal("vm.swappiness = 10\n", string(content))
		})
	}
}

func (suite *DebianPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		entry        sysctl.Entry
		setup        func()
		validateFunc func(*sysctl.CreateResult, error)
	}{
		{
			name: "when deploy succeeds",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("net.ipv4.ip_forward", result.Key)
				suite.True(result.Changed)

				// Verify file was written.
				content, readErr := suite.memFs.ReadFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				suite.NoError(readErr)
				suite.Equal("net.ipv4.ip_forward = 1\n", string(content))
			},
		},
		{
			name: "when the drop-in already holds what was asked for",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.NoError(suite.memFs.MkdirAll("/etc/sysctl.d", 0o755))
				suite.NoError(suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 1\n"),
					0o644,
				))
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.NotNil(result)
				suite.Equal("net.ipv4.ip_forward", result.Key)
				suite.False(result.Changed)
			},
		},
		{
			// The record says this key is managed and holds 0. The request asks
			// for 1. Deciding from the record reports no change and leaves the
			// host on 0, which is the defect this case exists to catch.
			name: "when a managed key is asked for a different value",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.NoError(suite.memFs.MkdirAll("/etc/sysctl.d", 0o755))
				suite.NoError(suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 0\n"),
					0o644,
				))
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := suite.memFs.ReadFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				suite.NoError(readErr)
				suite.Equal("net.ipv4.ip_forward = 1\n", string(content))
			},
		},
		{
			// Somebody edited the drop-in by hand. osapi owns the state, so the
			// next create overwrites it.
			name: "when the drop-in was edited by hand",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.NoError(suite.memFs.MkdirAll("/etc/sysctl.d", 0o755))
				suite.NoError(suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("# edited by hand\nnet.ipv4.ip_forward = 0\n"),
					0o644,
				))
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := suite.memFs.ReadFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				suite.NoError(readErr)
				suite.Equal("net.ipv4.ip_forward = 1\n", string(content))
			},
		},
		{
			// The record says managed, the file is gone. Restoring it is the
			// whole point of running create again.
			name: "when the drop-in was deleted",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)

				content, readErr := suite.memFs.ReadFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				suite.NoError(readErr)
				suite.Equal("net.ipv4.ip_forward = 1\n", string(content))
			},
		},
		{
			name: "when key is empty",
			entry: sysctl.Entry{
				Key:   "",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "key must not be empty")
			},
		},
		{
			name: "when value is empty",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not be empty")
			},
		},
		{
			name: "when key attempts path traversal",
			entry: sysctl.Entry{
				Key:   "../../etc/cron.d/pwn",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains a slash",
			entry: sysctl.Entry{
				Key:   "a/b",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains a space",
			entry: sysctl.Entry{
				Key:   "net.ipv4 ip_forward",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains an equals sign",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward=1",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key starts with a hyphen",
			entry: sysctl.Entry{
				Key:   "-pfoo",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key starts with a dot",
			entry: sysctl.Entry{
				Key:   ".hidden",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when value contains a newline",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1\nkernel.modprobe = /tmp/evil",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not contain line breaks")
			},
		},
		{
			name: "when value contains a carriage return",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1\r\nkernel.modprobe = /tmp/evil",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not contain line breaks")
			},
		},
		{
			name: "when a previously undeployed key is created again",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", gomock.Any()).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)
			},
		},
		{
			name: "when state put fails",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv put error"))
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "update state")
			},
		},
		{
			name: "when sysctl apply fails it still succeeds",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", gomock.Any()).
					Return("", errors.New("sysctl failed"))
			},
			validateFunc: func(
				result *sysctl.CreateResult,
				err error,
			) {
				suite.NoError(err)
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
		entry        sysctl.Entry
		setup        func()
		validateFunc func(*sysctl.UpdateResult, error)
	}{
		{
			name: "when update succeeds",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "0",
			},
			setup: func() {
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"-p", "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("net.ipv4.ip_forward", result.Key)
				suite.True(result.Changed)
			},
		},
		{
			name: "when key is empty",
			entry: sysctl.Entry{
				Key:   "",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "key must not be empty")
			},
		},
		{
			name: "when value is empty",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not be empty")
			},
		},
		{
			name: "when key attempts path traversal",
			entry: sysctl.Entry{
				Key:   "../../etc/cron.d/pwn",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains a slash",
			entry: sysctl.Entry{
				Key:   "a/b",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains a space",
			entry: sysctl.Entry{
				Key:   "net.ipv4 ip_forward",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when key contains an equals sign",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward=1",
				Value: "1",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when value contains a newline",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1\nkernel.modprobe = /tmp/evil",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not contain line breaks")
			},
		},
		{
			name: "when value contains a carriage return",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1\r\nkernel.modprobe = /tmp/evil",
			},
			setup: func() {},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "value must not contain line breaks")
			},
		},
		{
			name: "when content unchanged returns not changed",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "1",
			},
			setup: func() {
				// Write the file to disk so Stat succeeds.
				content := []byte("net.ipv4.ip_forward = 1\n")
				_ = suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					content,
					0o644,
				)

				// Create state with matching SHA.
			},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("net.ipv4.ip_forward", result.Key)
				suite.False(result.Changed)
			},
		},
		{
			name: "when deploy MkdirAll fails returns error",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "0",
			},
			setup: func() {
				baseFs := memfs.New()
				vfs := failfs.New(baseFs)
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnMkdirAll {
						return errors.New("mkdir failed")
					}

					return nil
				})
				suite.provider = sysctl.NewDebianProvider(
					suite.logger,
					vfs,
					suite.mockStateKV,
					suite.mockExec,
					testHostname,
				)
			},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "create directory")
			},
		},
		{
			name: "when deploy WriteFile fails returns error",
			entry: sysctl.Entry{
				Key:   "net.ipv4.ip_forward",
				Value: "0",
			},
			setup: func() {
				baseFs := memfs.New()
				_ = baseFs.MkdirAll("/etc/sysctl.d", 0o755)
				vfs := failfs.New(baseFs)
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnOpenFile {
						return errors.New("write failed")
					}

					return nil
				})
				suite.provider = sysctl.NewDebianProvider(
					suite.logger,
					vfs,
					suite.mockStateKV,
					suite.mockExec,
					testHostname,
				)
			},
			validateFunc: func(
				result *sysctl.UpdateResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "write file")
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
		key          string
		setup        func()
		validateFunc func(*sysctl.DeleteResult, error)
	}{
		{
			name: "when delete succeeds",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 1\n"),
					0o644,
				)
				stateData := managedStateJSON(
					"net.ipv4.ip_forward", "1",
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"--system"}).
					Return("", nil)
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.Equal("net.ipv4.ip_forward", result.Key)
				suite.True(result.Changed)
			},
		},
		{
			name:  "when key is empty",
			key:   "",
			setup: func() {},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "key must not be empty")
			},
		},
		{
			name:  "when key attempts path traversal",
			key:   "../../etc/cron.d/pwn",
			setup: func() {},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name:  "when key contains a slash",
			key:   "a/b",
			setup: func() {},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name:  "when key contains a space",
			key:   "net.ipv4 ip_forward",
			setup: func() {},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name:  "when key contains an equals sign",
			key:   "net.ipv4.ip_forward=1",
			setup: func() {},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid sysctl key")
			},
		},
		{
			name: "when not managed returns not changed",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name: "when state unmarshal fails returns not changed",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return([]byte("not-json"))

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name: "when already undeployed returns not changed",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				state := managedfile.State{
					Path:         "/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					SHA256:       "abc123",
					DeployedAt:   "2026-01-01T00:00:00Z",
					UndeployedAt: "2026-02-01T00:00:00Z",
					Metadata: map[string]string{
						"key":   "net.ipv4.ip_forward",
						"value": "1",
					},
				}
				stateBytes, _ := json.Marshal(state)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name: "when file not on disk but state exists",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				// No file on disk.
				stateData := managedStateJSON(
					"net.ipv4.ip_forward", "1",
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				// No sysctl --system call since changed is false.
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.False(result.Changed)
			},
		},
		{
			name: "when Remove fails returns error",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				baseFs := memfs.New()
				_ = baseFs.MkdirAll("/etc/sysctl.d", 0o755)
				_ = baseFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 1\n"),
					0o644,
				)
				vfs := failfs.New(baseFs)
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnRemove {
						return errors.New("remove failed")
					}

					return nil
				})
				suite.provider = sysctl.NewDebianProvider(
					suite.logger,
					vfs,
					suite.mockStateKV,
					suite.mockExec,
					testHostname,
				)
				stateData := managedStateJSON(
					"net.ipv4.ip_forward", "1",
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "remove file")
			},
		},
		{
			name: "when state put fails",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 1\n"),
					0o644,
				)
				stateData := managedStateJSON(
					"net.ipv4.ip_forward", "1",
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv put error"))
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "update state")
			},
		},
		{
			name: "when sysctl reload fails it still succeeds",
			key:  "net.ipv4.ip_forward",
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
					[]byte("net.ipv4.ip_forward = 1\n"),
					0o644,
				)
				stateData := managedStateJSON(
					"net.ipv4.ip_forward", "1",
					"/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf",
				)
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateData)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				suite.mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "sysctl", []string{"--system"}).
					Return("", errors.New("sysctl failed"))
			},
			validateFunc: func(
				result *sysctl.DeleteResult,
				err error,
			) {
				suite.NoError(err)
				suite.True(result.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Delete(context.Background(), tc.key)

			tc.validateFunc(result, err)
		})
	}
}

// failing is the provider on suite.memFs with every call fn makes on a path
// failing, so a test reaches the error it cannot set up on a real filesystem.
func (suite *DebianPublicTestSuite) failing(
	fn avfs.FnVFS,
	path string,
) *sysctl.Debian {
	vfs := failfs.New(suite.memFs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, f avfs.FnVFS, p *failfs.FailParam) error {
		if f == fn && p.Path == path {
			return errors.New("permission denied")
		}

		return nil
	})

	return sysctl.NewDebianProvider(
		suite.logger,
		vfs,
		suite.mockStateKV,
		suite.mockExec,
		testHostname,
	)
}

// kernel gives the fake /proc/sys the parameters and values given.
func (suite *DebianPublicTestSuite) kernel(
	params map[string]string,
) {
	for path, v := range params {
		_ = suite.memFs.MkdirAll(path[:strings.LastIndex(path, "/")], 0o755)
		_ = suite.memFs.WriteFile(path, []byte(v), 0o644)
	}
}

// manage records key as managed with value: its file in /etc/sysctl.d and
// a deployed record in the state store.
func (suite *DebianPublicTestSuite) manage(
	key string,
	value string,
) {
	path := "/etc/sysctl.d/osapi-" + key + ".conf"
	_ = suite.memFs.WriteFile(path, []byte(key+" = "+value+"\n"), 0o644)
	entry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
	entry.EXPECT().Value().Return(managedStateJSON(key, value, path)).AnyTimes()
	suite.mockStateKV.EXPECT().Get(gomock.Any(), gomock.Any()).Return(entry, nil).AnyTimes()
}

func (suite *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name    string
		setup   func()
		prefix  string
		want    []sysctl.Entry
		wantErr string
	}{
		{
			name: "when the kernel has parameters and osapi manages one",
			setup: func() {
				suite.kernel(map[string]string{
					"/proc/sys/kernel/panic":                 "0\n",
					"/proc/sys/net/ipv4/ip_forward":          "1\n",
					"/proc/sys/net/ipv4/ping_group_range":    "0\t2147483647\n",
					"/proc/sys/net/ipv4/conf/all/forwarding": "1\n",
				})
				suite.manage("net.ipv4.ip_forward", "1")
			},
			prefix: "net.ipv4",
			want: []sysctl.Entry{
				{Key: "net.ipv4.conf.all.forwarding", Value: "1"},
				{Key: "net.ipv4.ip_forward", Value: "1", Managed: true},
				{Key: "net.ipv4.ping_group_range", Value: "0 2147483647"},
			},
		},
		{
			name: "when /proc/sys is missing, only what osapi manages is listed",
			setup: func() {
				suite.manage("vm.swappiness", "10")
			},
			want: []sysctl.Entry{{Key: "vm.swappiness", Value: "10", Managed: true}},
		},
		{
			name: "when osapi's directory does not exist, nothing is managed",
			setup: func() {
				_ = suite.memFs.RemoveAll("/etc/sysctl.d")
				suite.kernel(map[string]string{"/proc/sys/kernel/panic": "0\n"})
			},
			want: []sysctl.Entry{{Key: "kernel.panic", Value: "0"}},
		},
		{
			name: "when files osapi did not write and directories are there, they are not managed",
			setup: func() {
				suite.kernel(map[string]string{"/proc/sys/kernel/panic": "10\n"})
				_ = suite.memFs.WriteFile(
					"/etc/sysctl.d/99-local.conf",
					[]byte("kernel.panic = 10\n"),
					0o644,
				)
				_ = suite.memFs.MkdirAll("/etc/sysctl.d/osapi-archive", 0o755)
			},
			want: []sysctl.Entry{{Key: "kernel.panic", Value: "10"}},
		},
		{
			name: "when a parameter cannot be read, it is left out",
			setup: func() {
				suite.kernel(map[string]string{
					"/proc/sys/kernel/panic":   "0\n",
					"/proc/sys/kernel/cad_pid": "1\n",
				})
				suite.provider = suite.failing(avfs.FnReadFile, "/proc/sys/kernel/cad_pid")
			},
			want: []sysctl.Entry{{Key: "kernel.panic", Value: "0"}},
		},
		{
			name: "when /proc/sys cannot be read",
			setup: func() {
				suite.kernel(map[string]string{"/proc/sys/kernel/panic": "0\n"})
				suite.provider = suite.failing(avfs.FnWalkDir, "/proc/sys")
			},
			wantErr: "sysctl list",
		},
		{
			name: "when osapi's directory cannot be read",
			setup: func() {
				suite.provider = suite.failing(avfs.FnReadDir, "/etc/sysctl.d")
			},
			wantErr: "sysctl list",
		},
		{
			name:    "when the prefix is not a parameter name",
			prefix:  "../etc",
			wantErr: "invalid sysctl key",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			got, err := suite.provider.List(context.Background(), tc.prefix)

			if tc.wantErr != "" {
				suite.ErrorContains(err, tc.wantErr)

				return
			}
			suite.Require().NoError(err)
			suite.Equal(tc.want, got)
		})
	}
}

func (suite *DebianPublicTestSuite) TestGet() {
	tests := []struct {
		name    string
		setup   func()
		key     string
		want    *sysctl.Entry
		wantErr string
	}{
		{
			name: "when osapi manages it, the kernel's value is returned",
			setup: func() {
				suite.kernel(map[string]string{"/proc/sys/vm/swappiness": "30\n"})
				suite.manage("vm.swappiness", "10")
			},
			key:  "vm.swappiness",
			want: &sysctl.Entry{Key: "vm.swappiness", Value: "30", Managed: true},
		},
		{
			name: "when the kernel's value cannot be read",
			setup: func() {
				suite.kernel(map[string]string{"/proc/sys/kernel/cad_pid": "1\n"})
				suite.mockStateKV.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(nil, jetstream.ErrKeyNotFound)
				suite.provider = suite.failing(avfs.FnReadFile, "/proc/sys/kernel/cad_pid")
			},
			key:     "kernel.cad_pid",
			wantErr: "permission denied",
		},
		{
			name: "when the record cannot be read",
			setup: func() {
				entry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				entry.EXPECT().Value().Return([]byte("not json"))
				suite.mockStateKV.EXPECT().Get(gomock.Any(), gomock.Any()).Return(entry, nil)
			},
			key:     "vm.swappiness",
			wantErr: "failed to read state",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			got, err := suite.provider.Get(context.Background(), tc.key)

			if tc.wantErr != "" {
				suite.ErrorContains(err, tc.wantErr)

				return
			}
			suite.Require().NoError(err)
			suite.Equal(tc.want, got)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
