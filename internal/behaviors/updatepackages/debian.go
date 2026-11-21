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

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// debian reads the updates with apt and refreshes the index with apt-get.
// The index is not a fact about the host: it is what the repositories
// have, so it is not gohai's.
type debian struct {
	exec host.Exec
	fs   avfs.VFS
}

// lists is where apt keeps the index it got from each repository.
const lists = "/var/lib/apt/lists"

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{exec: h.Exec, fs: h.FS}, nil
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

// Refresh runs apt-get update. It got new data when a repository's
// Release or InRelease file under /var/lib/apt/lists changed: apt-get
// prints Get: lines on each run, so its output does not tell.
func (d *debian) Refresh(
	ctx context.Context,
) (bool, error) {
	before := d.releases()
	if _, err := d.exec.RunPrivilegedCmd(ctx, "apt-get", []string{"update"}); err != nil {
		return false, fmt.Errorf("refresh the package index: %w", err)
	}

	return d.releases() != before, nil
}

// releases is each Release and InRelease file of the index, by name, with
// its contents. A file it cannot read is left out.
func (d *debian) releases() string {
	var sb strings.Builder
	names, _ := d.fs.Glob(lists + "/*Release")
	for _, name := range names {
		data, _ := d.fs.ReadFile(name)
		sb.WriteString(name + "\n" + string(data) + "\n")
	}

	return sb.String()
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
