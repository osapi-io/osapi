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

package process_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"syscall"
	"testing"

	gopsutil "github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/node/process"
)

type DebianPublicTestSuite struct {
	suite.Suite

	logger   *slog.Logger
	provider *process.Debian
}

func (suite *DebianPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.provider = process.NewDebianProvider(suite.logger)
}

func (suite *DebianPublicTestSuite) TearDownSubTest() {
	process.ResetListProcesses()
	process.ResetGetProcess()
	process.ResetKillProcess()
	process.ResetGatherInfoFromP()
}

func (suite *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setupMock    func()
		wantErr      bool
		wantErrMsg   string
		validateFunc func(result []process.Info)
	}{
		{
			name: "when successful returns process list",
			setupMock: func() {
				process.SetListProcesses(func() ([]*gopsutil.Process, error) {
					return []*gopsutil.Process{
						{Pid: 1},
						{Pid: 2},
					}, nil
				})
				process.SetGatherInfoFromP(func(p *gopsutil.Process) (*process.Info, error) {
					return &process.Info{
						PID:        int(p.Pid),
						Name:       "test-proc",
						User:       "root",
						State:      "running",
						CPUPercent: 1.5,
						MemPercent: 2.3,
						MemRSS:     1024,
						Command:    "/usr/bin/test",
						StartTime:  "2026-01-01T00:00:00Z",
					}, nil
				})
			},
			validateFunc: func(result []process.Info) {
				suite.Len(result, 2)
				suite.Equal(1, result[0].PID)
				suite.Equal("test-proc", result[0].Name)
				suite.Equal("root", result[0].User)
				suite.Equal("running", result[0].State)
				suite.Equal(1.5, result[0].CPUPercent)
				suite.InDelta(2.3, result[0].MemPercent, 0.01)
				suite.Equal(int64(1024), result[0].MemRSS)
				suite.Equal("/usr/bin/test", result[0].Command)
				suite.Equal("2026-01-01T00:00:00Z", result[0].StartTime)
				suite.Equal(2, result[1].PID)
			},
		},
		{
			name: "when listProcesses errors returns error",
			setupMock: func() {
				process.SetListProcesses(func() ([]*gopsutil.Process, error) {
					return nil, errors.New("cannot read /proc")
				})
			},
			wantErr:    true,
			wantErrMsg: "process: list: cannot read /proc",
		},
		{
			name: "when gather info errors skips process",
			setupMock: func() {
				process.SetListProcesses(func() ([]*gopsutil.Process, error) {
					return []*gopsutil.Process{
						{Pid: 1},
						{Pid: 2},
						{Pid: 3},
					}, nil
				})
				callCount := 0
				process.SetGatherInfoFromP(func(p *gopsutil.Process) (*process.Info, error) {
					callCount++
					if p.Pid == 2 {
						return nil, errors.New("permission denied")
					}

					return &process.Info{
						PID:  int(p.Pid),
						Name: "ok-proc",
					}, nil
				})
			},
			validateFunc: func(result []process.Info) {
				suite.Len(result, 2)
				suite.Equal(1, result[0].PID)
				suite.Equal(3, result[1].PID)
			},
		},
		{
			name: "when all processes error returns empty list",
			setupMock: func() {
				process.SetListProcesses(func() ([]*gopsutil.Process, error) {
					return []*gopsutil.Process{
						{Pid: 1},
					}, nil
				})
				process.SetGatherInfoFromP(func(_ *gopsutil.Process) (*process.Info, error) {
					return nil, errors.New("permission denied")
				})
			},
			validateFunc: func(result []process.Info) {
				suite.Empty(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setupMock()

			got, err := suite.provider.List(context.Background())

			if tc.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.wantErrMsg)
				suite.Nil(got)

				return
			}

			suite.NoError(err)
			tc.validateFunc(got)
		})
	}
}

func (suite *DebianPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		pid          int
		setupMock    func()
		wantErr      bool
		wantErrMsg   string
		validateFunc func(result *process.Info)
	}{
		{
			name: "when successful returns process info",
			pid:  42,
			setupMock: func() {
				process.SetGetProcess(func(pid int32) (*gopsutil.Process, error) {
					return &gopsutil.Process{Pid: pid}, nil
				})
				process.SetGatherInfoFromP(func(p *gopsutil.Process) (*process.Info, error) {
					return &process.Info{
						PID:        int(p.Pid),
						Name:       "test-proc",
						User:       "root",
						State:      "sleeping",
						CPUPercent: 0.5,
						MemPercent: 1.2,
						MemRSS:     2048,
						Command:    "/usr/bin/test --flag",
						StartTime:  "2026-01-01T00:00:00Z",
					}, nil
				})
			},
			validateFunc: func(result *process.Info) {
				suite.Equal(42, result.PID)
				suite.Equal("test-proc", result.Name)
				suite.Equal("root", result.User)
				suite.Equal("sleeping", result.State)
				suite.Equal(0.5, result.CPUPercent)
				suite.InDelta(1.2, result.MemPercent, 0.01)
				suite.Equal(int64(2048), result.MemRSS)
				suite.Equal("/usr/bin/test --flag", result.Command)
				suite.Equal("2026-01-01T00:00:00Z", result.StartTime)
			},
		},
		{
			name: "when pid not found returns error",
			pid:  99999,
			setupMock: func() {
				process.SetGetProcess(func(_ int32) (*gopsutil.Process, error) {
					return nil, errors.New("process not found")
				})
			},
			wantErr:    true,
			wantErrMsg: "process: get: process not found",
		},
		{
			name: "when gather info errors returns error",
			pid:  42,
			setupMock: func() {
				process.SetGetProcess(func(pid int32) (*gopsutil.Process, error) {
					return &gopsutil.Process{Pid: pid}, nil
				})
				process.SetGatherInfoFromP(func(_ *gopsutil.Process) (*process.Info, error) {
					return nil, errors.New("permission denied")
				})
			},
			wantErr:    true,
			wantErrMsg: "process: get: permission denied",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setupMock()

			got, err := suite.provider.Get(context.Background(), tc.pid)

			if tc.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.wantErrMsg)
				suite.Nil(got)

				return
			}

			suite.NoError(err)
			suite.NotNil(got)
			tc.validateFunc(got)
		})
	}
}

func (suite *DebianPublicTestSuite) TestSignal() {
	tests := []struct {
		name         string
		pid          int
		signal       string
		setupMock    func()
		wantErr      bool
		wantErrMsg   string
		validateFunc func(result *process.SignalResult)
	}{
		{
			name:   "when TERM signal succeeds",
			pid:    42,
			signal: "TERM",
			setupMock: func() {
				process.SetKillProcess(func(_ int, _ syscall.Signal) error {
					return nil
				})
			},
			validateFunc: func(result *process.SignalResult) {
				suite.Equal(42, result.PID)
				suite.Equal("TERM", result.Signal)
				suite.True(result.Changed)
			},
		},
		{
			name:   "when KILL signal succeeds",
			pid:    42,
			signal: "KILL",
			setupMock: func() {
				process.SetKillProcess(func(_ int, sig syscall.Signal) error {
					suite.Equal(syscall.SIGKILL, sig)

					return nil
				})
			},
			validateFunc: func(result *process.SignalResult) {
				suite.Equal(42, result.PID)
				suite.Equal("KILL", result.Signal)
				suite.True(result.Changed)
			},
		},
		{
			name:       "when invalid signal returns error",
			pid:        42,
			signal:     "INVALID",
			setupMock:  func() {},
			wantErr:    true,
			wantErrMsg: `process: signal: invalid signal "INVALID"`,
		},
		{
			name:   "when process not found returns error",
			pid:    99999,
			signal: "TERM",
			setupMock: func() {
				process.SetKillProcess(func(_ int, _ syscall.Signal) error {
					return syscall.ESRCH
				})
			},
			wantErr:    true,
			wantErrMsg: "process: signal: process not found",
		},
		{
			name:   "when permission denied returns error",
			pid:    1,
			signal: "TERM",
			setupMock: func() {
				process.SetKillProcess(func(_ int, _ syscall.Signal) error {
					return syscall.EPERM
				})
			},
			wantErr:    true,
			wantErrMsg: "process: signal: permission denied",
		},
		{
			name:   "when other error returns wrapped error",
			pid:    42,
			signal: "HUP",
			setupMock: func() {
				process.SetKillProcess(func(_ int, _ syscall.Signal) error {
					return errors.New("unexpected error")
				})
			},
			wantErr:    true,
			wantErrMsg: "process: signal: unexpected error",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setupMock()

			got, err := suite.provider.Signal(context.Background(), tc.pid, tc.signal)

			if tc.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.wantErrMsg)
				suite.Nil(got)

				return
			}

			suite.NoError(err)
			suite.NotNil(got)
			tc.validateFunc(got)
		})
	}
}

// TestDefaultOSFunctions exercises the real OS-interaction wrappers
// against the current test process to ensure coverage of gopsutil calls.
func (suite *DebianPublicTestSuite) TestDefaultOSFunctions() {
	// Use real functions (not mocked).
	process.ResetListProcesses()
	process.ResetGetProcess()
	process.ResetGatherInfoFromP()

	pid := os.Getpid()

	tests := []struct {
		name         string
		fn           func() error
		validateFunc func()
	}{
		{
			name: "when listing with real OS returns current process",
			fn: func() error {
				provider := process.NewDebianProvider(slog.Default())
				results, err := provider.List(context.Background())
				if err != nil {
					return err
				}

				found := false
				for _, p := range results {
					if p.PID == pid {
						found = true

						break
					}
				}
				suite.True(found, "current process not found in list")
				suite.NotEmpty(results)

				return nil
			},
		},
		{
			name: "when getting with real OS returns current process info",
			fn: func() error {
				provider := process.NewDebianProvider(slog.Default())
				info, err := provider.Get(context.Background(), pid)
				if err != nil {
					return err
				}

				suite.Equal(pid, info.PID)
				suite.NotEmpty(info.Name)
				suite.NotEmpty(info.StartTime)

				return nil
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := tc.fn()
			suite.NoError(err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDebianPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianPublicTestSuite))
}
