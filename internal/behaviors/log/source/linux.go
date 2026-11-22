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

package source

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// linux reads the sources with journalctl. They are what journald holds
// now, not a fact of the host, so they are not gohai's.
type linux struct {
	exec host.Exec
}

// newLinux is the provider on each Linux host. A container has no journald
// of its own, so its methods report skipped there, as the hand-written log
// domain did.
func newLinux(
	h host.Host,
) (Provider, error) {
	if h.Container {
		return nil, behavior.ErrUnsupported
	}

	return &linux{exec: h.Exec}, nil
}

// Sources runs journalctl --field=SYSLOG_IDENTIFIER: each value once, one a
// line.
func (l *linux) Sources(
	ctx context.Context,
) ([]string, error) {
	out, err := l.exec.RunCmd(ctx, "journalctl", []string{"--field=SYSLOG_IDENTIFIER"})
	if err != nil {
		return nil, fmt.Errorf("read the journal's sources: %w", err)
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if name := strings.TrimSpace(line); name != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	return names, nil
}
