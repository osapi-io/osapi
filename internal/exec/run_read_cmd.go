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
	"context"
	"log/slog"
	"strings"

	"github.com/osapi-io/gohai/pkg/gohai/source"
)

// RunReadCmd runs name with args to read the host, as gohai runs a command:
// its standard output is the data, its standard error is in the error, it
// runs with LC_ALL=C so that its output is the same in each locale, it
// stops after source.CommandTimeout, and it keeps at most source.MaxOutput
// of output. gohai's runner of a host is this, so that each read of a fact
// runs one way.
func (e *Exec) RunReadCmd(
	ctx context.Context,
	name string,
	args []string,
) ([]byte, error) {
	ctx, span := startSpan(ctx, name, args)
	defer span.End()

	out, err := source.Commands().Execute(ctx, name, args...)
	e.logger.Debug(
		"exec read",
		slog.String("command", strings.Join(append([]string{name}, args...), " ")),
		slog.Int("output_bytes", len(out)),
		slog.Any("error", err),
	)
	endReadSpan(span, err)

	return out, err
}
