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

package process

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	gopsutil "github.com/shirou/gopsutil/v4/process"

	"github.com/retr0h/osapi/internal/provider"
)

// allowedSignals maps signal names to their syscall equivalents.
var allowedSignals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
}

// Injectable functions for testing.
var (
	listProcesses   = defaultListProcesses
	getProcess      = defaultGetProcess
	killProcess     = defaultKillProcess
	gatherInfoFromP = defaultGatherInfoFromP
)

func defaultListProcesses() ([]*gopsutil.Process, error) {
	return gopsutil.Processes()
}

func defaultGetProcess(
	pid int32,
) (*gopsutil.Process, error) {
	return gopsutil.NewProcess(pid)
}

func defaultKillProcess(
	pid int,
	sig syscall.Signal,
) error {
	return syscall.Kill(pid, sig)
}

func defaultGatherInfoFromP(
	p *gopsutil.Process,
) (*Info, error) {
	name, err := p.Name()
	if err != nil {
		return nil, err
	}

	user, err := p.Username()
	if err != nil {
		return nil, err
	}

	statuses, err := p.Status()
	if err != nil {
		return nil, err
	}

	state := ""
	if len(statuses) > 0 {
		state = statuses[0]
	}

	cpuPercent, err := p.CPUPercent()
	if err != nil {
		return nil, err
	}

	memPercent, err := p.MemoryPercent()
	if err != nil {
		return nil, err
	}

	var memRSS int64

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}

	if memInfo != nil {
		memRSS = int64(memInfo.RSS)
	}

	cmdline, err := p.Cmdline()
	if err != nil {
		return nil, err
	}

	createTime, err := p.CreateTime()
	if err != nil {
		return nil, err
	}

	startTime := time.UnixMilli(createTime).UTC().Format(time.RFC3339)

	return &Info{
		PID:        int(p.Pid),
		Name:       name,
		User:       user,
		State:      state,
		CPUPercent: cpuPercent,
		MemPercent: memPercent,
		MemRSS:     memRSS,
		Command:    cmdline,
		StartTime:  startTime,
	}, nil
}

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
type Debian struct {
	provider.FactsAware
	logger *slog.Logger
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
) *Debian {
	return &Debian{
		logger: logger.With(slog.String("subsystem", "provider.process")),
	}
}

// List returns all running processes.
func (d *Debian) List(
	_ context.Context,
) ([]Info, error) {
	procs, err := listProcesses()
	if err != nil {
		return nil, fmt.Errorf("process: list: %w", err)
	}

	var result []Info

	for _, p := range procs {
		info, err := gatherInfoFromP(p)
		if err != nil {
			d.logger.Debug(
				"skipping process",
				slog.Int("pid", int(p.Pid)),
				slog.String("error", err.Error()),
			)

			continue
		}

		result = append(result, *info)
	}

	return result, nil
}

// Get returns details for a specific process by PID.
func (d *Debian) Get(
	_ context.Context,
	pid int,
) (*Info, error) {
	p, err := getProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("process: get: %w", err)
	}

	info, err := gatherInfoFromP(p)
	if err != nil {
		return nil, fmt.Errorf("process: get: %w", err)
	}

	return info, nil
}

// Signal sends a signal to a process by PID.
func (d *Debian) Signal(
	_ context.Context,
	pid int,
	signal string,
) (*SignalResult, error) {
	sig, ok := allowedSignals[signal]
	if !ok {
		return nil, fmt.Errorf("process: signal: invalid signal %q", signal)
	}

	if err := killProcess(pid, sig); err != nil {
		if err == syscall.ESRCH {
			return nil, fmt.Errorf("process: signal: process not found")
		}

		if err == syscall.EPERM {
			return nil, fmt.Errorf("process: signal: permission denied")
		}

		return nil, fmt.Errorf("process: signal: %w", err)
	}

	return &SignalResult{
		PID:     pid,
		Signal:  signal,
		Changed: true,
	}, nil
}
