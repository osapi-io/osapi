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

// gopsutilLister wraps gopsutil calls to satisfy ProcessLister.
// These are thin OS wrappers — error paths are not coverable in
// unit tests because gopsutil.Processes() always succeeds on a
// running system. The provider's List/Get/Signal methods are fully
// tested via the mocked ProcessLister/ProcessSignaler interfaces.
type gopsutilLister struct{}

// Processes returns all running processes as ProcessItem instances.
func (g *gopsutilLister) Processes() ([]ProcessItem, error) {
	procs, err := gopsutil.Processes()
	if err != nil {
		return nil, err // not coverable: gopsutil always succeeds on a running system
	}

	result := make([]ProcessItem, len(procs))
	for i, p := range procs {
		result[i] = ProcessItem{
			PID:     p.Pid,
			Querier: p,
		}
	}

	return result, nil
}

// NewProcess returns a ProcessQuerier for the given PID.
func (g *gopsutilLister) NewProcess(
	pid int32,
) (ProcessQuerier, error) {
	return gopsutil.NewProcess(pid)
}

// syscallSignaler wraps syscall.Kill to satisfy ProcessSignaler.
type syscallSignaler struct{}

// Kill sends a signal to a process.
func (s *syscallSignaler) Kill(
	pid int,
	sig syscall.Signal,
) error {
	return syscall.Kill(pid, sig)
}

// NewGopsutilLister returns the real gopsutil-backed ProcessLister.
func NewGopsutilLister() ProcessLister {
	return &gopsutilLister{}
}

// NewSyscallSignaler returns the real syscall-backed ProcessSignaler.
func NewSyscallSignaler() ProcessSignaler {
	return &syscallSignaler{}
}

// Compile-time checks.
var (
	_ Provider             = (*Debian)(nil)
	_ provider.FactsSetter = (*Debian)(nil)
)

// Debian implements the Provider interface for Debian-family systems.
type Debian struct {
	provider.FactsAware
	logger   *slog.Logger
	lister   ProcessLister
	signaler ProcessSignaler
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	logger *slog.Logger,
	lister ProcessLister,
	signaler ProcessSignaler,
) *Debian {
	return &Debian{
		logger:   logger.With(slog.String("subsystem", "provider.process")),
		lister:   lister,
		signaler: signaler,
	}
}

// List returns all running processes.
func (d *Debian) List(
	_ context.Context,
) ([]Info, error) {
	procs, err := d.lister.Processes()
	if err != nil {
		return nil, fmt.Errorf("process: list: %w", err)
	}

	var result []Info

	for _, item := range procs {
		info, err := gatherInfo(item.PID, item.Querier)
		if err != nil {
			d.logger.Debug(
				"skipping process",
				slog.Int("pid", int(item.PID)),
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
	p, err := d.lister.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("process: get: %w", err)
	}

	info, err := gatherInfo(int32(pid), p)
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

	if err := d.signaler.Kill(pid, sig); err != nil {
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

// gatherInfo extracts process information from a ProcessQuerier.
// PID is passed separately because gopsutil exposes it as a struct
// field rather than a method.
func gatherInfo(
	pid int32,
	p ProcessQuerier,
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
		PID:        int(pid),
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
