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
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// debian writes netplan files through managedfile, which keeps each one's
// record, and runs netplan.
type debian struct {
	files managedfile.Provider
	exec  host.Exec
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		files: host.Require[managedfile.Provider](h, managedfile.Behavior),
		exec:  h.Exec,
	}, nil
}

// Apply writes the file and, when it changed, validates and applies it.
// netplan reads the whole directory, so a file it refuses is removed again
// before it reaches a later apply.
func (d *debian) Apply(
	ctx context.Context,
	in UpdateInput,
) (bool, error) {
	if err := own(in.Path); err != nil {
		return false, err
	}
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
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"generate"}); err != nil {
		_, _ = behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
			managedfile.DeleteInput{Path: in.Path})

		return false, fmt.Errorf("netplan generate refused %s, so it was removed: %w", in.Path, err)
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"apply"}); err != nil {
		return false, fmt.Errorf("netplan apply: %w", err)
	}

	return true, nil
}

// Remove removes the file and, when it was there, applies what is left.
func (d *debian) Remove(
	ctx context.Context,
	path string,
) (bool, error) {
	if err := own(path); err != nil {
		return false, err
	}
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, d.files,
		managedfile.DeleteInput{Path: path})
	if err != nil || !result.Changed {
		return false, err
	}
	if _, err := d.exec.RunPrivilegedCmd(ctx, "netplan", []string{"apply"}); err != nil {
		return false, fmt.Errorf("netplan apply: %w", err)
	}

	return true, nil
}

// Interface reads netplan status. An interface that netplan does not list,
// or a status netplan cannot give, is an ethernet, the section most
// interfaces are in, under its own name.
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
		id = name
	}

	return Interface{Name: name, Type: s.Type, Section: Section(s.Type), ID: id}, nil
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
