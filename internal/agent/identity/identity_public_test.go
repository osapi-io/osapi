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

package identity_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent/identity"
)

type GetIdentityPublicTestSuite struct {
	suite.Suite
}

func (suite *GetIdentityPublicTestSuite) TearDownSubTest() {
	identity.ResetGetMachineIDFn()
}

func (suite *GetIdentityPublicTestSuite) TestGetIdentity() {
	tests := []struct {
		name           string
		configHostname string
		machineIDFn    func(avfs.VFS) (string, error)
		validateFunc   func(*identity.Identity, error)
	}{
		{
			name:           "when config hostname is provided and machine-id succeeds",
			configHostname: "my-host",
			machineIDFn: func(_ avfs.VFS) (string, error) {
				return "abc123", nil
			},
			validateFunc: func(got *identity.Identity, err error) {
				require.NoError(suite.T(), err)
				require.NotNil(suite.T(), got)
				assert.Equal(suite.T(), "abc123", got.MachineID)
				assert.Equal(suite.T(), "my-host", got.Hostname)
			},
		},
		{
			name:           "when config hostname is empty falls back to system hostname",
			configHostname: "",
			machineIDFn: func(_ avfs.VFS) (string, error) {
				return "abc123", nil
			},
			validateFunc: func(got *identity.Identity, err error) {
				require.NoError(suite.T(), err)
				require.NotNil(suite.T(), got)
				assert.Equal(suite.T(), "abc123", got.MachineID)
				assert.NotEmpty(suite.T(), got.Hostname)
			},
		},
		{
			name:           "when machine-id resolution fails",
			configHostname: "my-host",
			machineIDFn: func(_ avfs.VFS) (string, error) {
				return "", fmt.Errorf("read machine-id: file not found")
			},
			validateFunc: func(got *identity.Identity, err error) {
				require.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "machine-id")
				assert.Nil(suite.T(), got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			fs := memfs.New()

			if tc.machineIDFn != nil {
				identity.SetGetMachineIDFn(tc.machineIDFn)
			}

			tc.validateFunc(identity.GetIdentity(fs, tc.configHostname))
		})
	}
}

func TestGetIdentityPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(GetIdentityPublicTestSuite))
}

// runner is a gohai command runner that answers ioreg with out and err.
type runner struct {
	out []byte
	err error
}

func (r runner) Execute(
	_ context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	if name != "ioreg" || len(args) == 0 {
		return nil, fmt.Errorf("unexpected command %s %v", name, args)
	}

	return r.out, r.err
}

// PlatformPublicTestSuite tests the read of the machine ID through gohai.
type PlatformPublicTestSuite struct {
	suite.Suite
}

func (suite *PlatformPublicTestSuite) TearDownSubTest() {
	identity.ResetOSPlatform()
}

func (suite *PlatformPublicTestSuite) TestDefaultGetMachineID() {
	const uuid = "4C4C4544-0042-3110-8030-B4C04F333232"
	ioreg := []byte(`+-o J316sAP  <class IOPlatformExpertDevice>
    {
      "IOPlatformUUID" = "` + uuid + `"
    }
`)

	tests := []struct {
		name     string
		platform string
		files    map[string]string
		runner   runner
		wantID   string
		wantErr  string
	}{
		{
			name:     "when platform is linux reads /etc/machine-id",
			platform: "linux",
			files:    map[string]string{"/etc/machine-id": "abc123def456\n"},
			wantID:   "abc123def456",
		},
		{
			name:     "when /etc/machine-id has white space around the ID",
			platform: "linux",
			files:    map[string]string{"/etc/machine-id": "  abc123def456  \n"},
			wantID:   "abc123def456",
		},
		{
			name:     "when /etc/machine-id is missing reads the dbus machine ID",
			platform: "linux",
			files:    map[string]string{"/var/lib/dbus/machine-id": "fedcba987654\n"},
			wantID:   "fedcba987654",
		},
		{
			name:     "when the host has no machine ID",
			platform: "linux",
			files:    map[string]string{"/etc/machine-id": "   \n"},
			wantErr:  "read the machine ID: the host has none",
		},
		{
			name:     "when platform is darwin reads the IOPlatformUUID",
			platform: "darwin",
			runner:   runner{out: ioreg},
			wantID:   uuid,
		},
		{
			name:     "when ioreg fails",
			platform: "darwin",
			runner:   runner{err: errors.New("exec: ioreg: not found")},
			wantErr:  "read the machine ID: ioreg IOPlatformExpertDevice: exec: ioreg: not found",
		},
		{
			name:     "when the ioreg output has no IOPlatformUUID",
			platform: "darwin",
			runner:   runner{out: []byte("no uuid here")},
			wantErr:  "read the machine ID: the host has none",
		},
		{
			name:     "when platform is unsupported returns error",
			platform: "windows",
			wantErr:  "unsupported platform: windows",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			identity.SetOSPlatform(tc.platform)
			defer identity.SetMachineIDRunner(tc.runner)()
			fs := memfs.New()
			for path, data := range tc.files {
				require.NoError(suite.T(), fs.MkdirAll("/etc", 0o755))
				require.NoError(suite.T(), fs.MkdirAll("/var/lib/dbus", 0o755))
				require.NoError(suite.T(), avfs.WriteFile(fs, path, []byte(data), 0o644))
			}

			got, err := identity.ExportDefaultGetMachineID(fs)
			if tc.wantErr != "" {
				assert.EqualError(suite.T(), err, tc.wantErr)

				return
			}
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tc.wantID, got)
		})
	}
}

func (suite *PlatformPublicTestSuite) TestThisHost() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		suite.T().Skipf("osapi runs on linux and darwin, not %s", runtime.GOOS)
	}
	identity.ResetOSPlatform()

	got, err := identity.ExportDefaultGetMachineID(osfs.NewWithNoIdm())
	require.NoError(suite.T(), err, "gohai reads this host's machine ID")
	assert.NotEmpty(suite.T(), got)
}

func TestPlatformPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PlatformPublicTestSuite))
}
