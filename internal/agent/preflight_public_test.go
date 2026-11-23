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

package agent_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/behavior"
	execmocks "github.com/osapi-io/osapi/internal/exec/mocks"
)

type PreflightPublicTestSuite struct {
	suite.Suite

	mockCtrl    *gomock.Controller
	mockExecMgr *execmocks.MockManager
	logger      *slog.Logger
	tmpDir      string
}

func TestPreflightPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PreflightPublicTestSuite))
}

func (s *PreflightPublicTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	s.tmpDir = s.T().TempDir()
}

func (s *PreflightPublicTestSuite) SetupSubTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockExecMgr = execmocks.NewMockManager(s.mockCtrl)
}

func (s *PreflightPublicTestSuite) TearDownSubTest() {
	s.mockCtrl.Finish()
	agent.ResetProcStatusPath()
	agent.ResetLookPath()
	behavior.Reset()
}

func (s *PreflightPublicTestSuite) TestCheckCapabilities() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]agent.PreflightResult)
	}{
		{
			name: "when all caps present",
			setup: func() {
				path := filepath.Join(s.tmpDir, "status_all_caps")
				content := "Name:\tosapi\nCapEff:\t000000000000102f\n"
				err := os.WriteFile(path, []byte(content), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.True(r.Passed, "expected %s to pass", r.Name)
					s.Empty(r.Error)
				}
			},
		},
		{
			name: "when missing cap",
			setup: func() {
				path := filepath.Join(s.tmpDir, "status_no_caps")
				content := "Name:\tosapi\nCapEff:\t0000000000000000\n"
				err := os.WriteFile(path, []byte(content), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.False(r.Passed, "expected %s to fail", r.Name)
					s.Contains(r.Error, "not in effective set")
				}
			},
		},
		{
			name: "when file not found",
			setup: func() {
				agent.SetProcStatusPath(filepath.Join(s.tmpDir, "nonexistent"))
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.False(r.Passed, "expected %s to fail", r.Name)
					s.Contains(r.Error, "failed to read capabilities")
				}
			},
		},
		{
			name: "when CapEff line has invalid hex",
			setup: func() {
				path := filepath.Join(s.tmpDir, "status_bad_hex")
				content := "Name:\tosapi\nCapEff:\tNOTHEX\n"
				err := os.WriteFile(path, []byte(content), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.False(r.Passed, "expected %s to fail", r.Name)
					s.Contains(r.Error, "failed to read capabilities")
				}
			},
		},
		{
			name: "when scanner encounters read error",
			setup: func() {
				// Write a line longer than bufio.MaxScanTokenSize
				// (64 KiB) to trigger a scanner error before
				// reaching the CapEff line.
				path := filepath.Join(s.tmpDir, "status_long_line")
				longLine := make([]byte, 70000)
				for i := range longLine {
					longLine[i] = 'x'
				}
				content := string(longLine) + "\nCapEff:\t000000000000102f\n"
				err := os.WriteFile(path, []byte(content), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.False(r.Passed, "expected %s to fail", r.Name)
					s.Contains(r.Error, "failed to read capabilities")
				}
			},
		},
		{
			name: "when CapEff line not present",
			setup: func() {
				path := filepath.Join(s.tmpDir, "status_no_capeff")
				content := "Name:\tosapi\nCapInh:\t0000000000000000\n"
				err := os.WriteFile(path, []byte(content), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
			},
			validateFunc: func(results []agent.PreflightResult) {
				s.NotEmpty(results)
				for _, r := range results {
					s.False(r.Passed, "expected %s to fail", r.Name)
					s.Contains(r.Error, "failed to read capabilities")
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			results := agent.ExportCheckCapabilities(s.logger)
			tc.validateFunc(results)
		})
	}
}

const sudoListNarrow = `Matching Defaults entries for osapi on debian-1:
    env_reset, secure_path=/usr/sbin\\:/usr/bin, !requiretty

User osapi may run the following commands on debian-1:
    (root) NOPASSWD: /usr/sbin/sysctl -p /etc/sysctl.d/osapi-*.conf, /usr/sbin/sysctl --system
    (root) NOPASSWD: SETENV: /usr/bin/systemctl restart *
    (root) /usr/bin/timedatectl set-timezone *
`

// sudoListOdd has an empty rule line and a trailing comma, which the parser
// skips rather than reading as a program.
const sudoListOdd = `User osapi may run the following commands on debian-1:
    (root) NOPASSWD:
    (root) NOPASSWD: /usr/sbin/sysctl --system,
`

const sudoListAll = `User admin may run the following commands on debian-1:
    (ALL) NOPASSWD: ALL
`

func (s *PreflightPublicTestSuite) TestRunPreflight() {
	installed := func(names ...string) func(string) (string, error) {
		return func(p string) (string, error) {
			for _, n := range names {
				if n == p {
					return "/usr/bin/" + p, nil
				}
			}

			return "", fmt.Errorf("not found")
		}
	}

	tests := []struct {
		name      string
		sudoOut   string
		sudoErr   error
		capEff    string
		installed []string
		behaviors func()
		wantOK    bool
		wantFail  []string
		wantPass  []string
		wantNone  []string
	}{
		{
			name:      "when narrow rules cover the installed programs",
			sudoOut:   sudoListNarrow,
			capEff:    "000000000000102f",
			installed: []string{"sysctl", "systemctl", "timedatectl"},
			wantOK:    true,
			wantPass:  []string{"sudo", "sudo:sysctl", "sudo:systemctl", "sudo:timedatectl"},
		},
		{
			name:      "when a program is not installed, it is not checked",
			sudoOut:   sudoListNarrow,
			capEff:    "000000000000102f",
			installed: []string{"sysctl"},
			wantOK:    true,
			wantNone:  []string{"sudo:chronyc", "sudo:systemctl"},
		},
		{
			name:      "when an installed program has no rule, the check fails and the agent starts",
			sudoOut:   sudoListNarrow,
			capEff:    "000000000000102f",
			installed: []string{"sysctl", "useradd"},
			wantOK:    true,
			wantFail:  []string{"sudo:useradd"},
			wantPass:  []string{"sudo:sysctl"},
		},
		{
			name:      "when a rule grants ALL, every installed program passes",
			sudoOut:   sudoListAll,
			capEff:    "000000000000102f",
			installed: []string{"useradd", "chronyc"},
			wantOK:    true,
			wantPass:  []string{"sudo:useradd", "sudo:chronyc"},
		},
		{
			name:      "when a rule line is empty or ends in a comma, it is skipped",
			sudoOut:   sudoListOdd,
			capEff:    "000000000000102f",
			installed: []string{"sysctl", "systemctl"},
			wantOK:    true,
			wantPass:  []string{"sudo:sysctl"},
			wantFail:  []string{"sudo:systemctl"},
		},
		{
			name:      "when sudo needs a password, preflight fails",
			sudoErr:   fmt.Errorf("exit status 1"),
			capEff:    "000000000000102f",
			installed: []string{"sysctl"},
			wantOK:    false,
			wantFail:  []string{"sudo"},
			wantNone:  []string{"sudo:sysctl"},
		},
		{
			name:      "when a capability is missing, the check fails and the agent starts",
			sudoOut:   sudoListNarrow,
			capEff:    "0000000000000000",
			installed: []string{"sysctl"},
			wantOK:    true,
			wantFail:  []string{"cap:CAP_NET_ADMIN"},
		},
		{
			name:      "when a behavior declares a sudo program, it is checked once",
			sudoOut:   sudoListNarrow,
			capEff:    "000000000000102f",
			installed: []string{"motdctl", "sysctl"},
			behaviors: func() {
				behavior.Register(behavior.Define("motd", "Message of the day", behavior.Node,
					behavior.Sudo("motdctl", "sysctl"),
				))
			},
			wantOK:   true,
			wantFail: []string{"sudo:motdctl"},
			wantPass: []string{"sudo:sysctl"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.behaviors != nil {
				tc.behaviors()
			}
			agent.SetLookPath(installed(tc.installed...))
			s.mockExecMgr.EXPECT().
				RunCmd(gomock.Any(), "sudo", []string{"-n", "-l"}).
				Return(tc.sudoOut, tc.sudoErr)

			path := filepath.Join(s.tmpDir, "status")
			err := os.WriteFile(path, []byte("Name:\tosapi\nCapEff:\t"+tc.capEff+"\n"), 0o644)
			s.Require().NoError(err)
			agent.SetProcStatusPath(path)

			results, ok := agent.RunPreflight(
				context.Background(),
				s.logger,
				s.mockExecMgr,
				"debian",
			)

			s.Equal(tc.wantOK, ok)
			count := map[string]int{}
			passed := map[string]bool{}
			for _, r := range results {
				count[r.Name]++
				passed[r.Name] = r.Passed
				if !r.Passed {
					s.NotEmpty(r.Error, r.Name)
				}
			}
			for _, n := range tc.wantPass {
				s.Equal(1, count[n], n)
				s.True(passed[n], n)
			}
			for _, n := range tc.wantFail {
				s.Equal(1, count[n], n)
				s.False(passed[n], n)
			}
			for _, n := range tc.wantNone {
				s.Zero(count[n], n)
			}
		})
	}
}

func (s *PreflightPublicTestSuite) TestMissingPrograms() {
	tests := []struct {
		name  string
		setup func()
		want  []string
	}{
		{
			name: "when every declared program is installed",
			setup: func() {
				behavior.Register(behavior.Define("neighbors", "Neighbors", behavior.Node,
					behavior.Runs("ip"),
				))
				agent.SetLookPath(func(string) (string, error) { return "/usr/sbin/ip", nil })
			},
		},
		{
			name: "when a program two behaviors declare is missing, it is named once",
			setup: func() {
				behavior.Register(behavior.Define("neighbors", "Neighbors", behavior.Node,
					behavior.Runs("ip"),
				))
				behavior.Register(behavior.Define("routes", "Routes", behavior.Node,
					behavior.Runs("ip"),
				))
				calls := 0
				agent.SetLookPath(func(string) (string, error) {
					calls++
					s.Equal(1, calls, "a program is looked up once")

					return "", fmt.Errorf("not found")
				})
			},
			want: []string{"ip"},
		},
		{
			name: "when a sudo program is not on PATH but is in an sbin directory, it is not missing",
			setup: func() {
				behavior.Register(behavior.Define("users", "Users", behavior.Node,
					behavior.Sudo("useradd"),
				))
				agent.SetLookPath(func(p string) (string, error) {
					if p == "/usr/sbin/useradd" {
						return p, nil
					}

					return "", fmt.Errorf("not found")
				})
			},
		},
		{
			name: "when a sudo program is in no directory, it is missing",
			setup: func() {
				behavior.Register(behavior.Define("users", "Users", behavior.Node,
					behavior.Sudo("useradd"),
				))
				agent.SetLookPath(
					func(string) (string, error) { return "", fmt.Errorf("not found") },
				)
			},
			want: []string{"useradd"},
		},
		{
			name: "when a behavior runs any program, it needs none",
			setup: func() {
				behavior.Register(behavior.Define("command", "Command", behavior.Node,
					behavior.Runs(behavior.AnyProgram),
				))
				agent.SetLookPath(
					func(string) (string, error) { return "", fmt.Errorf("not found") },
				)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got := agent.MissingPrograms("debian")

			s.Equal(tc.want, got)
		})
	}
}
