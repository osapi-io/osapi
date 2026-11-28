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

package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

// RunPrivilegedCmdInteractive runs name with args, through sudo when
// privilege escalation is on, and calls interact while it runs, with the
// command's standard input. interact decides what the command reads, and
// when: netplan try, for one, waits for a confirmation that the caller gives
// only after it checks the change. The standard input closes when interact
// returns. interact's context ends when the command exits. The result is the
// command's combined output once it exits, with interact's error first.
//
// The command has DefaultCommandTimeout, as each command does, and ctx ends
// it and the group of processes that it started.
func (e *Exec) RunPrivilegedCmdInteractive(
	ctx context.Context,
	name string,
	args []string,
	interact func(ctx context.Context, stdin io.Writer) error,
) (string, error) {
	if e.sudo {
		args = sudoArgs(name, args)
		name = sudoProgram
	}
	ctx, span := startSpan(ctx, name, args)
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, DefaultCommandTimeout)
	defer cancel()

	cmd := newCommand(ctx, name, args)
	// StdinPipe fails only for a command whose Stdin is set or that started,
	// and this one is neither.
	stdin, _ := cmd.StdinPipe()
	var out lockedBuffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Start(); err != nil {
		endSpan(span, cmd, err)

		return "", fmt.Errorf("command %s: %w", name, err)
	}

	// interact's context ends when the command exits, so interact does not
	// wait for a command that is gone, such as one that refused at once.
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	ictx, stop := context.WithCancel(ctx)
	go func() {
		select {
		case err := <-exited:
			exited <- err
			stop()
		case <-ictx.Done():
		}
	}()
	interactErr := interact(ictx, stdin)
	stop()
	_ = stdin.Close()
	err := <-exited

	output := out.String()
	e.logger.Debug(
		"exec",
		slog.String("command", strings.Join(cmd.Args, " ")),
		slog.Int("output_bytes", len(output)),
		slog.Any("error", err),
	)
	endSpan(span, cmd, err)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = fmt.Errorf("command %s: %w", name, ctxErr)
		}
	}

	return output, errors.Join(interactErr, err)
}

// lockedBuffer is a bytes.Buffer that the command's two output pipes write
// at once.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(
	p []byte,
) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}
