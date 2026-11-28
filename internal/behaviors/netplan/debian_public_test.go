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
	"fmt"
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

func (s *DebianPublicTestSuite) TearDownSubTest() {
	netplan.ResetReachFor()
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

// tryAnswer answers as netplan try does: it writes its ready stamp now, later,
// or not at all, as one that refused and exited, then calls the agent's
// interact and prints out. wrote is what the agent wrote to its stdin.
func (s *DebianPublicTestSuite) tryAnswer(
	fs avfs.VFS,
	stamp string,
	wrote *string,
	out string,
	runErr error,
) func(context.Context, string, []string, func(context.Context, io.Writer) error) (string, error) {
	const ready = "/run/netplan/netplan-try.ready"

	return func(
		ctx context.Context,
		_ string,
		_ []string,
		interact func(context.Context, io.Writer) error,
	) (string, error) {
		s.NoError(ctx.Err(), "netplan try runs to its end when the job's context ends")
		s.Require().NoError(fs.MkdirAll("/run/netplan", 0o700))
		ictx := ctx
		switch stamp {
		case "now":
			s.Require().NoError(fs.WriteFile(ready, nil, 0o600))
		case "late":
			// netplan takes a while to apply: the stamp comes after the
			// agent looked for it once.
			go func() {
				time.Sleep(400 * time.Millisecond)
				_ = fs.WriteFile(ready, nil, 0o600)
			}()
		default:
			// netplan try refused and exited before it applied.
			c, cancel := context.WithCancel(ctx)
			cancel()
			ictx = c
		}
		var stdin bytes.Buffer
		ierr := interact(ictx, &stdin)
		*wrote = stdin.String()

		return out, errors.Join(ierr, runErr)
	}
}

// TestApply covers Apply past the validation of netplan generate.
//
// With no reach, the change runs through netplan apply. A file that netplan
// accepts and does not apply is applied by a retry with the same input: the
// record of the file is marked, so the retry's write changes; managedfile's
// own test shows that a marked record makes a write of the same content
// change. A call after an apply that worked changes nothing.
//
// With a reach, the change runs through netplan try, and the agent confirms it
// only when it reaches NATS after the change. When it does not, netplan
// reverts the change by itself, and the record of the file goes back with it.
// netplan try runs to its end when the job's context ends, as it does when the
// change cuts NATS and the agent restarts its consumers (found on the lab,
// osapi-io/osapi#1670).
func (s *DebianPublicTestSuite) TestApply() {
	const (
		path     = netplan.Dir + "/osapi-eth1.yaml"
		content  = "network:\n  version: 2\n"
		previous = "network:\n  version: 2\n  ethernets: {}\n"
	)
	meta := map[string]string{"interface": "eth1"}
	failed := errors.New("link is busy")
	cut := errors.New("nats: timeout")
	try := []string{"try", "--timeout", "60"}
	wrote := func(files *managedfilemocks.MockProvider) {
		files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
			return string(f.Content) == content && maps.Equal(f.Metadata, meta)
		})).Return(true, nil)
	}
	tests := []struct {
		name        string
		onDisk      bool
		reach       func(n int) error
		jobEnded    bool
		parallel    int
		expect      func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, wrote *string)
		wantWrote   string
		wantChanged bool
		wantError   string
	}{
		{
			name: "when netplan apply fails, the record is marked",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, _ avfs.VFS, _ *string) {
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
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, _ avfs.VFS, _ *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"})
			},
			wantChanged: true,
		},
		{
			name: "when the file was applied, it changes nothing",
			expect: func(files *managedfilemocks.MockProvider, _ *hostmocks.MockExec, _ avfs.VFS, _ *string) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name:   "when netplan generate refuses a file that had content, it gets the content back",
			onDisk: true,
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, _ avfs.VFS, _ *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("", errors.New("bad yaml"))
				files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
					return string(f.Content) == previous
				})).Return(true, nil)
			},
			wantError: "so its content was put back: bad yaml",
		},
		{
			name: "when netplan generate refuses a new file, it is removed",
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, _ avfs.VFS, _ *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"}).
					Return("", errors.New("bad yaml"))
				files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
			},
			wantError: "so it was removed: bad yaml",
		},
		{
			// dns, route and interface write their own files through netplan,
			// from jobs that run at once, and netplan reads each file in the
			// directory: two at once would apply one file half written, or
			// roll back the other's.
			name:     "when jobs that run at once apply changes, netplan runs one command at a time",
			parallel: 3,
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, _ avfs.VFS, _ *string) {
				files.EXPECT().Write(gomock.Any(), gomock.Any()).Return(true, nil).Times(3)
				var active, peak atomic.Int32
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", gomock.Any()).
					DoAndReturn(func(context.Context, string, []string) (string, error) {
						n := active.Add(1)
						if n > peak.Load() {
							peak.Store(n)
						}
						time.Sleep(20 * time.Millisecond)
						active.Add(-1)
						s.Equal(int32(1), peak.Load(), "one netplan command at a time")

						return "", nil
					}).Times(6)
			},
			wantChanged: true,
		},
		{
			name:  "when the agent reaches NATS, it confirms, and the change stays",
			reach: func(int) error { return nil },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "Configuration accepted.\n", nil))
			},
			wantWrote:   "\n",
			wantChanged: true,
		},
		{
			name:     "when the job's context ends, netplan try still runs to its end",
			reach:    func(int) error { return nil },
			jobEnded: true,
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "Configuration accepted.\n", nil))
			},
			wantWrote:   "\n",
			wantChanged: true,
		},
		{
			name:  "when netplan takes a while to apply, the agent waits for it, then confirms",
			reach: func(int) error { return nil },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "late", w, "Configuration accepted.\n", nil))
			},
			wantWrote:   "\n",
			wantChanged: true,
		},
		{
			name:  "when the agent reaches NATS on a second try, after a reconnect, it confirms",
			reach: func(n int) error { return map[bool]error{true: cut}[n == 0] },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "Configuration accepted.\n", nil))
			},
			wantWrote:   "\n",
			wantChanged: true,
		},
		{
			name:  "when the agent does not reach NATS, netplan reverts, and a new file's record goes",
			reach: func(int) error { return cut },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "Changes will revert in 60 seconds\nReverting.\n", nil))
				files.EXPECT().Remove(gomock.Any(), path).Return(true, nil)
			},
			wantError: "netplan reverted the change",
		},
		{
			name:   "when the agent does not reach NATS, a file that was there gets its record back",
			onDisk: true,
			reach:  func(int) error { return cut },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "Reverting.\n", nil))
				files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
					return string(f.Content) == previous
				})).Return(true, nil)
			},
			wantError: "the agent did not reach NATS after the change: nats: timeout",
		},
		{
			name:  "when netplan try cannot revert the change, it is applied with netplan apply",
			reach: func(int) error { return nil },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(
						fs,
						"",
						w,
						"bond0: reverting custom parameters for bridges and bonds is not supported\n\n"+
							"Please carefully review the configuration and use 'netplan apply' directly.\n",
						errors.New("exit status 78"),
					),
					)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"})
			},
			wantChanged: true,
		},
		{
			name:  "when netplan try does not run, it is the error, and the record is marked",
			reach: func(int) error { return nil },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "", w, "", errors.New("sudo: a password is required")))
				files.EXPECT().Write(gomock.Any(), marked(content, meta)).Return(true, nil)
			},
			wantError: "netplan try: sudo: a password is required",
		},
		{
			name:  "when netplan try gives an answer that it does not give, it is the error",
			reach: func(int) error { return nil },
			expect: func(files *managedfilemocks.MockProvider, exec *hostmocks.MockExec, fs avfs.VFS, w *string) {
				wrote(files)
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"generate"})
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", try, gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", w, "something else\n", nil))
				files.EXPECT().Write(gomock.Any(), marked(content, meta)).Return(true, nil)
			},
			wantWrote: "\n",
			wantError: "netplan try: no answer it knows: something else",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			netplan.SetReachFor(1500 * time.Millisecond)
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			if tc.onDisk {
				s.Require().NoError(fs.WriteFile(path, []byte(previous), 0o600))
			}
			files := managedfilemocks.NewMockProvider(ctrl)
			exec := hostmocks.NewMockExec(ctrl)
			var stdin string
			tc.expect(files, exec, fs, &stdin)
			h := host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			}
			if tc.reach != nil {
				calls := 0
				h.Reach = func(context.Context) error {
					defer func() { calls++ }()

					return tc.reach(calls)
				}
			}
			p, err := netplan.ExportNewDebian(h)
			s.Require().NoError(err)
			ctx := context.Background()
			if tc.jobEnded {
				c, cancel := context.WithCancel(ctx)
				cancel()
				ctx = c
			}

			if tc.parallel > 0 {
				var wg sync.WaitGroup
				for i := range tc.parallel {
					wg.Go(func() {
						changed, err := p.Apply(ctx, netplan.UpdateInput{
							Path: fmt.Sprintf(
								"%s/osapi-eth%d.yaml",
								netplan.Dir,
								i,
							),
							Content: content,
						})
						s.NoError(err)
						s.True(changed)
					})
				}
				wg.Wait()

				return
			}

			changed, err := p.Apply(ctx, netplan.UpdateInput{
				Path: path, Content: content, Metadata: meta,
			})

			s.Equal(tc.wantChanged, changed)
			s.Equal(tc.wantWrote, stdin)
			if tc.wantError != "" {
				s.ErrorContains(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

// TestRemove: when the apply after a remove fails, the file is put back with
// its record marked, so a retry removes it again. When netplan try reverts the
// remove, netplan put the file back, and its record comes back as it was.
func (s *DebianPublicTestSuite) TestRemove() {
	const (
		path    = netplan.Dir + "/osapi-eth1.yaml"
		content = "network:\n  version: 2\n"
	)
	meta := map[string]string{"interface": "eth1"}
	tests := []struct {
		name      string
		onDisk    bool
		recErr    error
		reach     func() error
		putBack   gomock.Matcher
		wantError string
	}{
		{
			name:      "when the file and its record are there, the file is put back",
			onDisk:    true,
			putBack:   marked(content, meta),
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
		{
			name:   "when netplan try reverts the remove, the record comes back, not marked",
			onDisk: true,
			reach:  func() error { return errors.New("nats: no servers") },
			putBack: gomock.Cond(func(f managedfile.File) bool {
				return string(f.Content) == content && maps.Equal(f.Metadata, meta)
			}),
			wantError: "so " + path + " was put back",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			netplan.SetReachFor(200 * time.Millisecond)
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
			var stdin string
			if tc.reach != nil {
				exec.EXPECT().
					RunPrivilegedCmdInteractive(gomock.Any(), "netplan", gomock.Any(), gomock.Any()).
					DoAndReturn(s.tryAnswer(fs, "now", &stdin, "Reverting.\n", nil))
			} else {
				exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "netplan", []string{"apply"}).
					Return("", errors.New("link is busy"))
			}
			if tc.putBack != nil {
				files.EXPECT().Write(gomock.Any(), tc.putBack).Return(true, nil)
			}
			h := host.Host{
				FS:       fs,
				Exec:     exec,
				Required: map[string]any{managedfile.Behavior.Name: files},
			}
			if tc.reach != nil {
				h.Reach = func(context.Context) error { return tc.reach() }
			}
			p, err := netplan.ExportNewDebian(h)
			s.Require().NoError(err)

			changed, err := p.Remove(context.Background(), path)

			s.False(changed)
			s.ErrorContains(err, tc.wantError)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
