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

package updatepackages

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// debian reads the updates with apt and refreshes the index with apt-get.
// The index is not a fact about the host: it is what the repositories
// have, so it is not gohai's.
type debian struct {
	exec host.Exec
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{exec: h.Exec}, nil
}

// Updates runs apt list --upgradable.
func (d *debian) Updates(
	ctx context.Context,
) ([]Available, error) {
	out, err := d.exec.RunCmd(ctx, "apt", []string{"list", "--upgradable"})
	if err != nil {
		return nil, fmt.Errorf("list the updates: %w", err)
	}

	return parse(out), nil
}

// Refresh runs apt-get update. It got new data when apt-get fetched a
// file: a Get: line. A refresh with only Hit: lines changed nothing.
func (d *debian) Refresh(
	ctx context.Context,
) (bool, error) {
	out, err := d.exec.RunPrivilegedCmd(ctx, "apt-get", []string{"update"})
	if err != nil {
		return false, fmt.Errorf("refresh the package index: %w", err)
	}

	return strings.Contains(out, "\nGet:") || strings.HasPrefix(out, "Get:"), nil
}

// parse reads apt list --upgradable: a line per update, as
// name/suite version arch [upgradable from: current].
func parse(
	out string,
) []Available {
	updates := []Available{}
	for _, line := range strings.Split(out, "\n") {
		name, rest, ok := strings.Cut(line, "/")
		fields := strings.Fields(rest)
		if !ok || len(fields) < 6 {
			continue
		}
		updates = append(updates, Available{
			Name:           name,
			CurrentVersion: strings.TrimSuffix(fields[5], "]"),
			NewVersion:     fields[1],
		})
	}

	return updates
}
