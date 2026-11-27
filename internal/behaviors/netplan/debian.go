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
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/avfs/avfs"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"

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
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		files: host.Require[managedfile.Provider](h, managedfile.Behavior),
		exec:  h.Exec,
		fs:    h.FS,
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
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"apply"}); err != nil {
		// The file stays as it was asked, and its record is marked: a retry
		// with the same input then changes and applies again.
		_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
			managedfile.MarkUnapplied(managedfile.UpdateInput{
				Path: in.Path, Content: in.Content, Mode: "0600", Metadata: in.Metadata,
			}))

		return false, fmt.Errorf("netplan apply: %w", err)
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
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"apply"}); err != nil {
		// The record of a file that osapi stopped managing keeps its
		// metadata, which the file gets back.
		rec, recErr := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, d.files,
			managedfile.GetInput{Path: path})
		if readErr == nil && recErr == nil {
			_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files,
				managedfile.MarkUnapplied(managedfile.UpdateInput{
					Path: path, Content: string(previous), Mode: "0600", Metadata: rec.Metadata,
				}))

			return false, fmt.Errorf("netplan apply: %w, so %s was put back", err, path)
		}

		return false, fmt.Errorf("netplan apply: %w", err)
	}

	return true, nil
}

// Interface reads netplan status, for the definition that matches the
// interface, and the type that netplan gives it. netplan lists only the
// interfaces it manages, and an older netplan has no status. For an interface
// that status does not list, the type is the interface's kind as gohai reads
// it, under its own name, and an interface of no kind is an ethernet, the
// section most interfaces are in. Whether the host has the interface is
// gohai's to read: the callers check it.
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
		t := s.Type
		if t == "" {
			t = d.kind(ctx, name)
		}

		return Interface{
			Name: name, Type: t, Section: Section(t), ID: name, Matches: []string{name},
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

// netplanType is the netplan type of each kind that gohai reads, where the two
// differ in name or netplan has a section for it. netplan defines a WireGuard
// link as a tunnel. Other kinds, such as a NIC, a veth or a loopback, are an
// ethernet.
var netplanType = map[string]string{
	"bridge":    "bridge",
	"bond":      "bond",
	"wifi":      "wifi",
	"vlan":      "vlan",
	"vxlan":     "vxlan",
	"wireguard": "tunnel",
}

// kind is the netplan type of the interface name from its kind as gohai reads
// it from sysfs, or "" when gohai does not know the interface or its kind
// has no section of its own. It reads sysfs only, with no commands, because
// a write asks it for each interface it changes.
func (d *debian) kind(
	ctx context.Context,
	name string,
) string {
	out, _ := (&network.Linux{FS: d.fs}).Collect(ctx, nil)
	if info, ok := out.(*network.Info); ok {
		for _, iface := range info.Interfaces {
			if iface.Name == name {
				return netplanType[iface.Kind]
			}
		}
	}

	return ""
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
	case "vlan":
		return "vlans"
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
