// Copyright (c) 2024 John Dewey

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
	"context"
	"os/exec"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// startSpan opens a span for one command, named for the program that runs:
// exec sysctl, and exec sysctl for sudo sysctl too, with osapi.exec.sudo.
// The arguments stay out of the span, because they can hold a secret.
func startSpan(
	ctx context.Context,
	name string,
	args []string,
) (context.Context, trace.Span) {
	program := name
	sudo := name == "sudo" && len(args) > 0
	if sudo {
		program = args[0]
	}

	return otel.Tracer("osapi-exec").Start(ctx, "exec "+program,
		trace.WithAttributes(
			attribute.String("process.executable.name", program),
			attribute.Bool("osapi.exec.sudo", sudo),
		),
	)
}

// endSpan records the exit code of the command, and marks the span as failed
// when the command failed. A command that did not start has no exit code.
func endSpan(
	span trace.Span,
	cmd *exec.Cmd,
	err error,
) {
	if cmd.ProcessState != nil {
		span.SetAttributes(attribute.Int("process.exit.code", cmd.ProcessState.ExitCode()))
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}
