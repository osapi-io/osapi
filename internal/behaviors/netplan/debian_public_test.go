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

package netplan_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	managedfilemocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// DebianPublicTestSuite covers what no example can: netplan generate
// refusing a file, which a fake host's commands do not do.
type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestRefusedFileIsRolledBack() {
	const path = netplan.Dir + "/osapi-eth1.yaml"
	tests := []struct {
		name      string
		previous  string
		expect    func(files *managedfilemocks.MockProvider)
		wantError string
	}{
		{
			name:     "when the file had content, it gets the content back",
			previous: "network:\n  version: 2\n",
			expect: func(files *managedfilemocks.MockProvider) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
				files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
					return string(f.Content) == "network:\n  version: 2\n"
				})).Return(true, nil)
			},
			wantError: "so its content was put back",
		},
		{
			name: "when the file is new, it is removed",
			expect: func(files *managedfilemocks.MockProvider) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
				files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
			},
			wantError: "so it was removed",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			if tc.previous != "" {
				s.Require().NoError(fs.WriteFile(path, []byte(tc.previous), 0o600))
			}
			files := managedfilemocks.NewMockProvider(ctrl)
			tc.expect(files)
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
				Return("", errors.New("bad yaml"))
			p, err := netplan.ExportNewDebian(host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			})
			s.Require().NoError(err)

			_, err = p.Apply(context.Background(), netplan.UpdateInput{Path: path, Content: "bad"})

			s.ErrorContains(err, tc.wantError)
			s.ErrorContains(err, "bad yaml")
		})
	}
}

// TestFailedStepIsInTheTree is a chain that fails after a step changed the
// host, as a job records it: the step tree marks the call that failed, and
// the write under it that changed the host. Nothing undoes a write that
// netplan generate accepted, so a failed netplan apply leaves the new file.
func (s *DebianPublicTestSuite) TestFailedStepIsInTheTree() {
	const path = netplan.Dir + "/osapi-dns-eth1.yaml"
	tests := []struct {
		name      string
		previous  string
		expect    func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec)
		wantError string
		wantSteps string
	}{
		{
			name: "when netplan apply fails, the file stays written",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("", nil)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"}).
					Return("", errors.New("link down"))
			},
			wantError: "netplan apply: link down",
			wantSteps: "netplan update (failed) → managedfile update (changed)",
		},
		{
			name:     "when netplan generate refuses the file, the put back is a step",
			previous: "network:\n  version: 2\n",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("", errors.New("bad yaml"))
			},
			wantError: "so its content was put back",
			wantSteps: "netplan update (failed) → (managedfile update (changed), managedfile update (changed))",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			if tc.previous != "" {
				s.Require().NoError(fs.WriteFile(path, []byte(tc.previous), 0o600))
			}
			files := managedfilemocks.NewMockProvider(ctrl)
			exec := hostmocks.NewMockExec(ctrl)
			tc.expect(files, exec)
			p, err := netplan.ExportNewDebian(host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			})
			s.Require().NoError(err)
			root := &behavior.Step{}

			_, err = behavior.Call(
				behavior.Record(context.Background(), root),
				netplan.Behavior,
				netplan.Update,
				p,
				netplan.UpdateInput{
					Path:    path,
					Content: "network:\n  version: 2\n  ethernets: {}\n",
				},
			)

			s.ErrorContains(err, tc.wantError)
			s.Equal(tc.wantSteps, client.FormatSteps(root.Wire()))
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
