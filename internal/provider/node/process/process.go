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

// Package process provides process management operations.
package process

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/process"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/provider"
)

// allowedSignals maps signal names to their syscall equivalents.
// validatePid rejects a pid that kill(2) would not read as one process.
//
// kill(2) overloads the sign: 0 means every process in the caller's process
// group, -1 means every process the caller may signal, and anything below -1
// means a whole process group. The agent usually runs as root, so a value
// below 1 turns one signal into a host-wide one. The API validates this too,
// and this is the copy that matters, because this is the frame that calls
// kill(2).
func validatePid(
	pid int,
) error {
	if pid < 1 {
		return fmt.Errorf(
			"process: pid %d must be greater than zero, a lower value signals a process group rather than a process",
			pid,
		)
	}

	return nil
}

var allowedSignals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
}

// syscallSignaler wraps syscall.Kill to satisfy Signaler.
type syscallSignaler struct{}

// Kill sends a signal to a process.
func (s *syscallSignaler) Kill(
	pid int,
	sig syscall.Signal,
) error {
	return syscall.Kill(pid, sig)
}

// NewSyscallSignaler returns the real syscall-backed Signaler.
func NewSyscallSignaler() Signaler {
	return &syscallSignaler{}
}

var (
	_ Provider             = (*Gohai)(nil)
	_ provider.FactsSetter = (*Gohai)(nil)
)

// Gohai lists and reads processes with gohai's process collector, on
// every Linux family and on macOS, and signals them with kill(2).
type Gohai struct {
	provider.FactsAware

	logger   *slog.Logger
	signaler Signaler
	// Sources are what the collector reads. A zero field is the real
	// host's.
	Sources source.Sources
}

// New returns a provider that reads the real host and signals with
// signaler.
func New(
	logger *slog.Logger,
	signaler Signaler,
) *Gohai {
	return &Gohai{
		logger:   logger.With(slog.String("subsystem", "provider.process")),
		signaler: signaler,
	}
}

// List is each process of the host.
func (g *Gohai) List(
	ctx context.Context,
) ([]Info, error) {
	got, err := process.NewWith(g.Sources).Collect(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("process: list: %w", err)
	}
	procs := got.(*process.Info).Processes
	result := make([]Info, 0, len(procs))
	for _, p := range procs {
		result = append(result, toInfo(p))
	}

	return result, nil
}

// Get is the process pid. A process that gohai cannot read, because it
// is not there, is not found.
func (g *Gohai) Get(
	ctx context.Context,
	pid int,
) (*Info, error) {
	if err := validatePid(pid); err != nil {
		return nil, err
	}
	p, err := process.NewWith(g.Sources).Read(ctx, int32(pid))
	if err != nil {
		return nil, fmt.Errorf("process: get %d: %w: %w", pid, provider.ErrNotFound, err)
	}
	info := toInfo(*p)

	return &info, nil
}

// Signal sends a signal to a process by PID.
func (g *Gohai) Signal(
	_ context.Context,
	pid int,
	signal string,
) (*SignalResult, error) {
	if err := validatePid(pid); err != nil {
		return nil, err
	}

	sig, ok := allowedSignals[signal]
	if !ok {
		return nil, fmt.Errorf("process: signal: invalid signal %q", signal)
	}

	if err := g.signaler.Kill(pid, sig); err != nil {
		if err == syscall.ESRCH {
			return nil, fmt.Errorf("process signal: %w", provider.ErrNotFound)
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

// toInfo is the Info of a gohai process. The start time is RFC 3339, in
// UTC, as it was.
func toInfo(
	p process.Process,
) Info {
	info := Info{
		PID:        int(p.PID),
		Name:       p.Name,
		User:       p.Owner,
		State:      p.State,
		CPUPercent: p.CPUPercent,
		MemPercent: float32(p.MemPercent),
		MemRSS:     p.RSS,
		Command:    p.CmdLine,
	}
	if p.CreationTime > 0 {
		info.StartTime = time.Unix(p.CreationTime, 0).UTC().Format(time.RFC3339)
	}

	return info
}
