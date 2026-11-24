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
	"maps"
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
// netplan generate accepted, so a failed netplan apply leaves the new file,
// and a second write marks its record so a retry applies it.
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
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("", nil)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"}).
					Return("", errors.New("link down"))
			},
			wantError: "netplan apply: link down",
			wantSteps: "netplan update (failed) → (managedfile update (changed), managedfile update (changed))",
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
		{
			name:     "when netplan generate refuses the file, the error has what it said",
			previous: "network:\n  version: 2\n",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("Error in network definition: unknown key 'dummy-devices'\n",
						errors.New("exit status 1"))
			},
			wantError: "exit status 1: Error in network definition: unknown key 'dummy-devices'",
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

// marked matches a write of content whose record carries the mark of a step
// that did not finish, with metadata beside it.
func marked(
	content string,
	metadata map[string]string,
) gomock.Matcher {
	return gomock.Cond(func(f managedfile.File) bool {
		want := map[string]string{managedfile.Unapplied: "true"}
		for k, v := range metadata {
			want[k] = v
		}

		return string(f.Content) == content && f.Mode == 0o600 && maps.Equal(f.Metadata, want)
	})
}

// TestRetryAfterAFailedApply: a file that netplan accepts and does not apply
// is applied by a retry with the same input. The record of the file is
// marked, so the retry's write changes; managedfile's own test shows that a
// marked record makes a write of the same content change. A call after an
// apply that worked changes nothing.
func (s *DebianPublicTestSuite) TestRetryAfterAFailedApply() {
	const (
		path    = netplan.Dir + "/osapi-eth1.yaml"
		content = "network:\n  version: 2\n"
	)
	meta := map[string]string{"interface": "eth1"}
	failed := errors.New("link is busy")
	tests := []struct {
		name        string
		expect      func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec)
		wantChanged bool
		wantError   string
	}{
		{
			name: "when netplan apply fails, the record is marked",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"}).
					Return("", failed)
				files.EXPECT().Write(gomock.Any(), marked(content, meta)).Return(true, nil)
			},
			wantError: "netplan apply: link is busy",
		},
		{
			name: "when the retry's write changes the marked record, it applies",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
					return maps.Equal(f.Metadata, meta)
				})).Return(true, nil)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"})
			},
			wantChanged: true,
		},
		{
			name: "when the file was applied, it changes nothing",
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			files := managedfilemocks.NewMockProvider(ctrl)
			exec := hostmocks.NewMockExec(ctrl)
			tc.expect(files, exec)
			p, err := netplan.ExportNewDebian(host.Host{
				FS:       memfs.New(),
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			})
			s.Require().NoError(err)

			changed, err := p.Apply(context.Background(), netplan.UpdateInput{
				Path: path, Content: content, Metadata: meta,
			})

			s.Equal(tc.wantChanged, changed)
			if tc.wantError != "" {
				s.EqualError(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

// TestRetryAfterAFailedRemove: when netplan apply fails after a remove, the
// file is put back with its record marked, so a retry removes it and applies
// again. Without the file or its record, nothing can be put back.
func (s *DebianPublicTestSuite) TestRetryAfterAFailedRemove() {
	const (
		path    = netplan.Dir + "/osapi-eth1.yaml"
		content = "network:\n  version: 2\n"
	)
	meta := map[string]string{"interface": "eth1"}
	tests := []struct {
		name      string
		onDisk    bool
		recErr    error
		putBack   bool
		wantError string
	}{
		{
			name:      "when the file and its record are there, the file is put back",
			onDisk:    true,
			putBack:   true,
			wantError: "netplan apply: link is busy, so " + path + " was put back",
		},
		{
			name:      "when the file cannot be read, it is not put back",
			wantError: "netplan apply: link is busy",
		},
		{
			name:      "when the record cannot be read, the file is not put back",
			onDisk:    true,
			recErr:    managedfile.ErrUnreadable,
			wantError: "netplan apply: link is busy",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			if tc.onDisk {
				s.Require().NoError(fs.WriteFile(path, []byte(content), 0o600))
			}
			files := managedfilemocks.NewMockProvider(ctrl)
			files.EXPECT().Read(gomock.Any(), path).
				Return(managedfile.Record{Path: path, Metadata: meta}, tc.recErr)
			files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"}).
				Return("", errors.New("link is busy"))
			if tc.putBack {
				files.EXPECT().Write(gomock.Any(), marked(content, meta)).Return(true, nil)
			}
			p, err := netplan.ExportNewDebian(host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			})
			s.Require().NoError(err)

			changed, err := p.Remove(context.Background(), path)

			s.False(changed)
			s.EqualError(err, tc.wantError)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
