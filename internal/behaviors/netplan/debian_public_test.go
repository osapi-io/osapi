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
	"bytes"
	"context"
	"errors"
	"io"
	"maps"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	managedfilemocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/pkg/sdk/types"
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

// TestOneChangeAtATime: changes from jobs that run at once, such as a dns
// write and a route write, reach netplan generate and apply one at a time.
// netplan reads each file in the directory, so two at once would apply one
// file half written, or roll back the other's.
func (s *DebianPublicTestSuite) TestOneChangeAtATime() {
	ctrl := gomock.NewController(s.T())
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
	files := managedfilemocks.NewMockProvider(ctrl)
	files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	var active, peak atomic.Int32
	exec := hostmocks.NewMockExec(ctrl)
	exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", gomock.Any()).
		DoAndReturn(func(context.Context, string, []string) (string, error) {
			n := active.Add(1)
			if n > peak.Load() {
				peak.Store(n)
			}
			time.Sleep(20 * time.Millisecond)
			active.Add(-1)

			return "", nil
		}).AnyTimes()
	p, err := netplan.ExportNewDebian(host.Host{
		FS:       fs,
		Exec:     exec,
		Required: map[string]any{managedfile.Behavior.Name: files},
	})
	s.Require().NoError(err)

	var wg sync.WaitGroup
	for _, name := range []string{"osapi-dns-eth0.yaml", "osapi-eth0-routes.yaml", "osapi-eth1.yaml"} {
		wg.Go(func() {
			_, err := p.Apply(context.Background(),
				netplan.UpdateInput{Path: netplan.Dir + "/" + name, Content: "network: {}\n"})
			s.NoError(err)
		})
	}
	wg.Wait()

	s.Equal(int32(1), peak.Load(), "one netplan command at a time")
}

// TestTry: with a way to reach NATS, a change runs through netplan try, and
// the agent confirms it only when it reaches NATS after the change. When it
// does not, netplan reverts the change by itself, and the record of the file
// goes back with it.
func (s *DebianPublicTestSuite) TestTry() {
	const (
		path     = netplan.Dir + "/osapi-eth1.yaml"
		previous = "network:\n  version: 2\n"
		content  = "network:\n  version: 2\n  ethernets: {}\n"
		ready    = "/run/netplan/netplan-try.ready"
	)
	try := []string{"try", "--timeout", "60"}
	cut := errors.New("nats: timeout")
	defer netplan.SetReachFor(1500 * time.Millisecond)()
	// netplan answers as netplan try does: it writes the ready stamp, waits
	// for the agent, and prints out.
	netplanTry := func(fs avfs.VFS, stamp, late bool, wrote *string, out string, err error) any {
		return func(
			ctx context.Context, _ string, _ []string,
			interact func(context.Context, io.Writer) error,
		) (string, error) {
			ictx := ctx
			if stamp {
				s.Require().NoError(fs.MkdirAll("/run/netplan", 0o700))
				s.Require().NoError(fs.WriteFile(ready, nil, 0o600))
			} else if late {
				// netplan takes a while to apply: the stamp comes after the
				// agent looked for it once.
				s.Require().NoError(fs.MkdirAll("/run/netplan", 0o700))
				go func() {
					time.Sleep(400 * time.Millisecond)
					_ = fs.WriteFile(ready, nil, 0o600)
				}()
			} else {
				// netplan try refused and exited before it applied.
				c, cancel := context.WithCancel(ctx)
				cancel()
				ictx = c
			}
			var stdin bytes.Buffer
			ierr := interact(ictx, &stdin)
			*wrote = stdin.String()

			return out, errors.Join(ierr, err)
		}
	}
	tests := []struct {
		name      string
		onDisk    bool
		reach     func(n int) error
		stamp     bool
		late      bool
		out       string
		runErr    error
		expect    func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec)
		wantWrote string
		wantError string
		changed   bool
	}{
		{
			name:      "when the agent reaches NATS, it confirms, and the change stays",
			reach:     func(int) error { return nil },
			stamp:     true,
			out:       "Do you want to keep these settings?\n\nConfiguration accepted.\n",
			wantWrote: "\n",
			changed:   true,
		},
		{
			name:      "when netplan takes a while to apply, the agent waits for it, then confirms",
			reach:     func(int) error { return nil },
			late:      true,
			out:       "Configuration accepted.\n",
			wantWrote: "\n",
			changed:   true,
		},
		{
			name:      "when the agent reaches NATS on a second try, after a reconnect, it confirms",
			reach:     func(n int) error { return map[bool]error{true: cut}[n == 0] },
			stamp:     true,
			out:       "Configuration accepted.\n",
			wantWrote: "\n",
			changed:   true,
		},
		{
			name:   "when the agent does not reach NATS, netplan reverts, and a new file's record goes",
			reach:  func(int) error { return cut },
			stamp:  true,
			out:    "Changes will revert in 60 seconds\nReverting.\n",
			runErr: nil,
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec) {
				files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
			},
			wantError: "netplan reverted the change",
		},
		{
			name:   "when the agent does not reach NATS, a file that was there gets its record back",
			onDisk: true,
			reach:  func(int) error { return cut },
			stamp:  true,
			out:    "Reverting.\n",
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
					return string(f.Content) == previous
				})).Return(true, nil)
			},
			wantError: "the agent did not reach NATS after the change: nats: timeout",
		},
		{
			name:   "a change that netplan try cannot revert is applied with netplan apply",
			reach:  func(int) error { return nil },
			out:    "bond0: reverting custom parameters for bridges and bonds is not supported\n\nPlease carefully review the configuration and use 'netplan apply' directly.\n",
			runErr: errors.New("exit status 78"),
			expect: func(_ *managedfilemocks.MockProvider, exec *hostmocks.MockExec) {
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"})
			},
			changed: true,
		},
		{
			name:   "netplan try that does not run is the error",
			reach:  func(int) error { return nil },
			runErr: errors.New("sudo: a password is required"),
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantError: "netplan try: sudo: a password is required",
		},
		{
			name:  "an answer that netplan try does not give is the error",
			reach: func(int) error { return nil },
			stamp: true,
			out:   "something else\n",
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantWrote: "\n",
			wantError: "netplan try: no answer it knows: something else",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			if tc.onDisk {
				s.Require().NoError(fs.WriteFile(path, []byte(previous), 0o600))
			}
			files := managedfilemocks.NewMockProvider(ctrl)
			files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
				return string(f.Content) == content
			})).Return(true, nil)
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
			var wrote string
			exec.EXPECT().RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
				DoAndReturn(netplanTry(fs, tc.stamp, tc.late, &wrote, tc.out, tc.runErr))
			if tc.expect != nil {
				tc.expect(files, exec)
			}
			calls := 0
			p, err := netplan.ExportNewDebian(host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
				Reach: func(context.Context) error {
					defer func() { calls++ }()

					return tc.reach(calls)
				},
			})
			s.Require().NoError(err)

			changed, err := p.Apply(context.Background(),
				netplan.UpdateInput{Path: path, Content: content})

			s.Equal(tc.changed, changed)
			s.Equal(tc.wantWrote, wrote)
			if tc.wantError != "" {
				s.ErrorContains(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

// TestTryOutlivesTheJob: netplan try runs to its end when the job's context
// ends, as it does when the change cuts NATS and the agent restarts its
// consumers. A netplan try that was killed then reverted nothing, and the
// change that cut the host stayed (found on the lab, osapi-io/osapi#1670).
func (s *DebianPublicTestSuite) TestTryOutlivesTheJob() {
	const path = netplan.Dir + "/osapi-eth1.yaml"
	ctrl := gomock.NewController(s.T())
	files := managedfilemocks.NewMockProvider(ctrl)
	files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil)
	exec := hostmocks.NewMockExec(ctrl)
	exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
	exec.EXPECT().RunPrivilegedCmdInteractive(gomock.Any(), "netplan", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, _ []string,
			_ func(context.Context, io.Writer) error,
		) (string, error) {
			s.NoError(ctx.Err(), "the job's context ended, and netplan try must not")

			return "Configuration accepted.\n", nil
		})
	p, err := netplan.ExportNewDebian(host.Host{
		FS:       memfs.New(),
		Exec:     exec,
		Required: map[string]any{managedfile.Behavior.Name: files},
		Reach:    func(context.Context) error { return nil },
	})
	s.Require().NoError(err)
	job, cancel := context.WithCancel(context.Background())
	cancel()

	changed, err := p.Apply(job, netplan.UpdateInput{Path: path, Content: "network: {}\n"})

	s.NoError(err)
	s.True(changed)
}

// TestTryRemove: a removal that netplan reverts gets its record back, as the
// file that netplan put back.
func (s *DebianPublicTestSuite) TestTryRemove() {
	const (
		path    = netplan.Dir + "/osapi-eth1.yaml"
		content = "network:\n  version: 2\n"
	)
	meta := map[string]string{"interface": "eth1"}
	defer netplan.SetReachFor(200 * time.Millisecond)()
	ctrl := gomock.NewController(s.T())
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
	s.Require().NoError(fs.WriteFile(path, []byte(content), 0o600))
	files := managedfilemocks.NewMockProvider(ctrl)
	files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
	files.EXPECT().
		Read(gomock.Any(), path).
		Return(managedfile.Record{Path: path, Metadata: meta}, nil)
	files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
		return string(f.Content) == content && maps.Equal(f.Metadata, meta) &&
			f.Metadata[managedfile.Unapplied] == ""
	})).Return(true, nil)
	exec := hostmocks.NewMockExec(ctrl)
	exec.EXPECT().RunPrivilegedCmdInteractive(gomock.Any(), "netplan", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, _ []string,
			interact func(context.Context, io.Writer) error,
		) (string, error) {
			s.Require().NoError(fs.MkdirAll("/run/netplan", 0o700))
			s.Require().NoError(fs.WriteFile("/run/netplan/netplan-try.ready", nil, 0o600))

			return "Reverting.\n", interact(ctx, io.Discard)
		})
	p, err := netplan.ExportNewDebian(host.Host{
		FS:       fs,
		Exec:     exec,
		Required: map[string]any{managedfile.Behavior.Name: files},
		Reach:    func(context.Context) error { return errors.New("nats: no servers") },
	})
	s.Require().NoError(err)

	changed, err := p.Remove(context.Background(), path)

	s.False(changed)
	s.ErrorContains(err, "netplan reverted the change")
	s.ErrorContains(err, "so "+path+" was put back")
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
			s.Equal(tc.wantSteps, types.FormatSteps(root.Wire()))
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
