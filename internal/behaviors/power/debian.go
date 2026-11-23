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

package power

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// minDelayMinutes is the least delay before the host goes down: the agent
// writes the job's result, and the controller answers, before it does.
const minDelayMinutes = 1

// debian runs shutdown on the Debian family.
type debian struct {
	exec host.Exec
}

// newDebian is the provider on the Debian family.
func newDebian(
	h host.Host,
) (Provider, error) {
	// A container shares its host's kernel, and stopping it is not the
	// agent's to do.
	if h.Container {
		return nil, behavior.ErrUnsupported
	}

	return &debian{exec: h.Exec}, nil
}

// Reboot runs shutdown -r +N.
func (d *debian) Reboot(
	ctx context.Context,
	in Input,
) (Result, error) {
	return d.run(ctx, "reboot", "-r", in)
}

// Shutdown runs shutdown -h +N.
func (d *debian) Shutdown(
	ctx context.Context,
	in Input,
) (Result, error) {
	return d.run(ctx, "shutdown", "-h", in)
}

// run runs shutdown with flag after the delay in whole minutes, at least
// one, and the message.
func (d *debian) run(
	ctx context.Context,
	action string,
	flag string,
	in Input,
) (Result, error) {
	minutes := max(in.Delay/60, minDelayMinutes)
	// -- ends the options, so a message that starts with - is the message.
	args := []string{flag, "--", fmt.Sprintf("+%d", minutes)}
	if in.Message != "" {
		args = append(args, in.Message)
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "shutdown", args); err != nil {
		return Result{}, fmt.Errorf("power: %s: %w", action, err)
	}

	return Result{Action: action, Delay: minutes * 60, Changed: true}, nil
}
