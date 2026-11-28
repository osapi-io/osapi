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
	"strconv"
	"time"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/process"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// gohai reads the processes through gohai's process collector, and sends a
// signal with kill, on Linux and on macOS.
type gohai struct {
	src  source.Sources
	exec host.Exec
}

// maxPID is Linux's largest pid_max, the largest PID that GetInput and
// SignalInput allow.
const maxPID = 4194304

func newGohai(
	h host.Host,
) (Provider, error) {
	return &gohai{src: h.Sources(), exec: h.Exec}, nil
}

// List is each process that gohai reads.
func (g *gohai) List(
	ctx context.Context,
) ([]Entry, error) {
	got, err := process.NewWith(g.src).Collect(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list the processes: %w", err)
	}
	procs := got.(*process.Info).Processes
	out := make([]Entry, 0, len(procs))
	for _, p := range procs {
		out = append(out, entry(p))
	}

	return out, nil
}

// Get is the process with the PID, read through gohai.
func (g *gohai) Get(
	ctx context.Context,
	pid int,
) (Entry, error) {
	// The PID becomes an int32 and an argument of kill, so the provider
	// checks the bound of the validate rule again.
	if pid < 1 || pid > maxPID {
		return Entry{}, fmt.Errorf(
			"process %d: a PID is from 1 to %d: %w",
			pid,
			maxPID,
			behavior.ErrInvalid,
		)
	}
	p, err := process.NewWith(g.src).Read(ctx, int32(pid))
	if err != nil {
		return Entry{}, fmt.Errorf("process %d: %w", pid, behavior.ErrNotFound)
	}

	return entry(*p), nil
}

// Signal sends a signal to the process with the PID.
func (g *gohai) Signal(
	ctx context.Context,
	pid int,
	signal types.Signal,
) error {
	if _, err := g.Get(ctx, pid); err != nil {
		return err
	}
	if _, err := g.exec.RunPrivilegedCmd(
		ctx,
		"kill",
		[]string{"-" + string(signal), strconv.Itoa(pid)},
	); err != nil {
		return fmt.Errorf("signal %s to process %d: %w", signal, pid, err)
	}

	return nil
}

// entry is p as the behavior gives it.
func entry(
	p process.Process,
) Entry {
	e := Entry{
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
		e.StartTime = time.Unix(p.CreationTime, 0).UTC().Format(time.RFC3339)
	}

	return e
}
