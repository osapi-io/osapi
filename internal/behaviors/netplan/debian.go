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

package netplan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// debian writes netplan files through managedfile, which keeps each one's
// record, and runs netplan.
// applying holds one netplan change at a time on the host. dns, route and
// interface each write their own file, through this behavior, from jobs that
// can run at once, and netplan generate and apply read each file: two at
// once would apply one's file half written, or roll back the other's.
var applying sync.Mutex

type debian struct {
	files managedfile.Provider
	exec  host.Exec
	fs    avfs.VFS
	// reach is the agent's round trip to NATS. With one, a change runs
	// through netplan try and stays only when reach answers after it.
	reach func(ctx context.Context) error
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		files: host.Require[managedfile.Provider](h, managedfile.Behavior),
		exec:  h.Exec,
		fs:    h.FS,
		reach: h.Reach,
	}, nil
}

// Apply writes the file and, when it changed, validates and applies it.
// netplan reads the whole directory, so a file it refuses does not stay to
// reach a later apply: the content it had is put back, or a new file is
// removed. A file that netplan accepts and does not apply stays, with its
// record marked, so a retry with the same input applies it.
func (d *debian) Apply(
	ctx context.Context,
	in UpdateInput,
) (bool, error) {
	if err := own(in.Path); err != nil {
		return false, err
	}
	applying.Lock()
	defer applying.Unlock()
	previous, readErr := d.fs.ReadFile(in.Path)
	result, err := behavior.Call(
		ctx,
		managedfile.Behavior,
		managedfile.Update,
		d.files,
		managedfile.UpdateInput{
			Path:     in.Path,
			Content:  in.Content,
			Mode:     "0600",
			Metadata: in.Metadata,
		},
	)
	if err != nil || !result.Changed {
		return false, err
	}
	if out, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"generate"}); err != nil {
		// netplan generate tells in its output which key it refused.
		if msg := strings.TrimSpace(out); msg != "" {
			err = fmt.Errorf("%w: %s", err, msg)
		}
		if readErr == nil {
			_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
				managedfile.UpdateInput{
					Path: in.Path, Content: string(previous), Mode: "0600", Metadata: in.Metadata,
				})

			return false, fmt.Errorf(
				"netplan generate refused %s, so its content was put back: %w", in.Path, err)
		}
		_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
			managedfile.DeleteInput{Path: in.Path})

		return false, fmt.Errorf("netplan generate refused %s, so it was removed: %w", in.Path, err)
	}
	err = d.apply(ctx)
	if errors.Is(err, errReverted) {
		// netplan put the directory back as it was, so the record goes back
		// with it: a retry with the same input then tries again.
		d.putBack(ctx, in.Path, previous, readErr, in.Metadata)

		return false, err
	}
	if err != nil {
		// The file stays as it was asked, and its record is marked: a retry
		// with the same input then changes and applies again.
		_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
			managedfile.MarkUnapplied(managedfile.UpdateInput{
				Path: in.Path, Content: in.Content, Mode: "0600", Metadata: in.Metadata,
			}))

		return false, err
	}

	return true, nil
}

// Remove removes the file and, when it was there, applies what is left. When
// netplan apply fails, the file is put back with its record marked, so a
// retry removes it and applies again.
func (d *debian) Remove(
	ctx context.Context,
	path string,
) (bool, error) {
	if err := own(path); err != nil {
		return false, err
	}
	applying.Lock()
	defer applying.Unlock()
	previous, readErr := d.fs.ReadFile(path)
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path})
	if err != nil || !result.Changed {
		return false, err
	}
	if err := d.apply(ctx); err != nil {
		// The record of a file that osapi stopped managing keeps its
		// metadata, which the file gets back.
		rec, recErr := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
			managedfile.GetInput{Path: path})
		if errors.Is(err, errReverted) && readErr == nil && recErr == nil {
			// netplan put the file back itself: its record comes back as
			// it was, so a retry removes it and tries again.
			d.putBack(ctx, path, previous, nil, rec.Metadata)

			return false, fmt.Errorf("%w, so %s was put back", err, path)
		}
		if readErr == nil && recErr == nil {
			_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
				managedfile.MarkUnapplied(managedfile.UpdateInput{
					Path: path, Content: string(previous), Mode: "0600", Metadata: rec.Metadata,
				}))

			return false, fmt.Errorf("%w, so %s was put back", err, path)
		}

		return false, err
	}

	return true, nil
}

// tryTimeout is how long netplan try waits for the agent to confirm a change
// before it reverts it, in seconds. The agent confirms when it reaches NATS,
// within reachFor of the change.
const tryTimeout = 60

// defaultReachFor is how long the agent tries to reach NATS after a change
// before it lets netplan revert it, and reachFor is that time, which a test
// sets shorter.
const defaultReachFor = 45 * time.Second

var reachFor = defaultReachFor

// tryReady is the file that netplan try writes when it has applied the
// change and waits for the confirmation.
const tryReady = "/run/netplan/netplan-try.ready"

// errReverted is a change that netplan try reverted, because the agent did
// not reach NATS after it.
var errReverted = errors.New("netplan reverted the change")

// apply applies the files of Dir. With reach, it runs netplan try: netplan
// applies the change, and keeps it only when the agent confirms, which it
// does when it still reaches NATS. Otherwise netplan reverts the files and
// the network by itself at the timeout, also when the job ends first. A
// change that netplan try cannot revert, such as a bond with parameters of
// its own, is applied with netplan apply, as with no reach.
func (d *debian) apply(
	ctx context.Context,
) error {
	if d.reach == nil {
		return d.applyDirect(ctx)
	}
	// netplan try runs to its end whatever happens to the job: when the
	// change cuts NATS, the agent restarts its consumers, which ends the
	// job's context, and a netplan try killed then reverts nothing. It ends
	// on its own, at its timeout at the latest.
	out, err := d.exec.RunPrivilegedCmdInteractive(context.WithoutCancel(ctx), "netplan",
		[]string{"try", "--timeout", strconv.Itoa(tryTimeout)},
		func(ctx context.Context, stdin io.Writer) error {
			if !d.waitFor(ctx, tryReady) {
				return nil
			}
			if err := d.reachWithin(ctx, reachFor); err != nil {
				// No confirmation: netplan reverts at its timeout.
				return fmt.Errorf("the agent did not reach NATS after the change: %w", err)
			}
			_, err := io.WriteString(stdin, "\n")

			return err
		})
	switch {
	case strings.Contains(out, "Configuration accepted"):
		return nil
	case strings.Contains(out, "use 'netplan apply' directly"):
		return d.applyDirect(ctx)
	case strings.Contains(out, "Reverting"):
		return fmt.Errorf(
			"%w: %w",
			errReverted,
			errors.Join(err, errors.New(strings.TrimSpace(out))),
		)
	case err != nil:
		return fmt.Errorf("netplan try: %w: %s", err, strings.TrimSpace(out))
	default:
		return fmt.Errorf("netplan try: no answer it knows: %s", strings.TrimSpace(out))
	}
}

// applyDirect runs netplan apply.
func (d *debian) applyDirect(
	ctx context.Context,
) error {
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"apply"}); err != nil {
		return fmt.Errorf("netplan apply: %w", err)
	}

	return nil
}

// waitFor reports whether path is there before ctx ends. ctx ends when
// netplan try exits, such as when it refuses a change that it cannot
// revert.
func (d *debian) waitFor(
	ctx context.Context,
	path string,
) bool {
	for {
		if _, err := d.fs.Stat(path); err == nil {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(250 * time.Millisecond):
		}
	}
}

// reachWithin calls reach until it answers, for at most within: a change
// that moves the host's address breaks the connection, and the client
// reconnects on its own.
func (d *debian) reachWithin(
	ctx context.Context,
	within time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, within)
	defer cancel()
	for {
		attempt, stop := context.WithTimeout(ctx, 5*time.Second)
		err := d.reach(attempt)
		stop()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(time.Second):
		}
	}
}

// putBack puts the record of path back as it was before a change that
// netplan reverted: its previous content, or no record for a file that was
// new. netplan put the file itself back.
func (d *debian) putBack(
	ctx context.Context,
	path string,
	previous []byte,
	readErr error,
	metadata map[string]string,
) {
	if readErr == nil {
		_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
			managedfile.UpdateInput{
				Path: path, Content: string(previous), Mode: "0600", Metadata: metadata,
			})

		return
	}
	_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path})
}

// Interface reads netplan status. netplan lists only the interfaces it
// manages, so an interface that it does not list, or a status netplan cannot
// give, is an ethernet, the section most interfaces are in, under its own
// name. Whether the host has the interface is gohai's to read: the callers
// check it.
func (d *debian) Interface(
	ctx context.Context,
	name string,
) (Interface, error) {
	out, _ := d.exec.RunCmd(ctx, "netplan", []string{"status", "--format", "json"})
	var status map[string]struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}
	_ = json.Unmarshal([]byte(out), &status)
	s := status[name]
	// A definition that matches several interfaces, such as cloud-init's
	// all for en*, is the one netplan applies: a second definition under
	// the interface's name is ignored, so a write goes under the ID.
	id := s.ID
	if id == "" {
		return Interface{
			Name: name, Type: s.Type, Section: Section(s.Type), ID: name, Matches: []string{name},
		}, nil
	}
	matches := []string{}
	for other, o := range status {
		if o.ID == id {
			matches = append(matches, other)
		}
	}
	sort.Strings(matches)

	return Interface{
		Name:    name,
		Type:    s.Type,
		Section: Section(s.Type),
		ID:      id,
		Matches: matches,
	}, nil
}

// Section is the netplan section of an interface of type t.
func Section(
	t string,
) string {
	switch t {
	case "wifi":
		return "wifis"
	case "bridge":
		return "bridges"
	case "bond":
		return "bonds"
	case "tunnel", "vxlan":
		return "tunnels"
	case "dummy-device":
		return "dummy-devices"
	default:
		return "ethernets"
	}
}

// own refuses a path that is not one of osapi's netplan files: a file
// directly in Dir, named osapi-<something>.yaml. A caller's name cannot
// reach another directory, or a file that another tool owns.
func own(
	path string,
) error {
	name := filepath.Base(path)
	if filepath.Clean(path) != path || filepath.Dir(path) != Dir ||
		!strings.HasPrefix(name, "osapi-") || !strings.HasSuffix(name, ".yaml") {
		return fmt.Errorf("path %q is not one of osapi's netplan files, %s/osapi-*.yaml", path, Dir)
	}

	return nil
}
