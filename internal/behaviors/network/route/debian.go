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

package route

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/avfs/avfs"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
)

// debian reads through gohai and writes an interface's routes with a netplan
// file of their own, through the netplan behavior. Each write reads osapi's
// routes of the interface from that file, changes one, and writes the file
// again; the last route that goes removes it.
type debian struct {
	*reader
	netplan   netplan.Provider
	fs        avfs.VFS
	container bool
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		reader:    &reader{src: h.Sources()},
		netplan:   host.Require[netplan.Provider](h, netplan.Behavior),
		fs:        h.FS,
		container: h.Container,
	}, nil
}

// file is the netplan file of an interface's routes.
func file(
	name string,
) string {
	return netplan.Dir + "/osapi-" + name + "-routes.yaml"
}

// Set writes the route r of the interface.
func (d *debian) Set(
	ctx context.Context,
	in Route,
	create bool,
) (bool, error) {
	name, r := in.Name, toSpec(in)
	if d.container {
		return false, behavior.ErrUnsupported
	}
	if err := check(r); err != nil {
		return false, err
	}
	have, err := d.have(name)
	if err != nil {
		return false, err
	}
	i := slices.IndexFunc(have, func(s spec) bool { return s.To == r.To })
	switch {
	case i >= 0 && have[i] == r:
		return false, nil
	case i >= 0 && create:
		return false, fmt.Errorf("osapi has a route to %s of %s via %s: update changes it",
			r.To, name, have[i].Via)
	case i < 0 && !create:
		return false, fmt.Errorf(
			"osapi has no route to %s of %s: %w",
			r.To,
			name,
			behavior.ErrNotFound,
		)
	case i >= 0:
		have[i] = r
	default:
		have = append(have, r)
	}

	return d.write(ctx, name, have)
}

// Remove drops osapi's route to destination of the interface, and the file
// with its last route.
func (d *debian) Remove(
	ctx context.Context,
	name string,
	destination string,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	have, err := d.have(name)
	if err != nil {
		return false, err
	}
	i := slices.IndexFunc(have, func(s spec) bool { return s.To == canon(destination) })
	if i < 0 {
		return false, nil
	}
	have = slices.Delete(have, i, i+1)
	if len(have) > 0 {
		return d.write(ctx, name, have)
	}
	result, err := behavior.Call(ctx, netplan.Behavior, netplan.Delete, d.netplan,
		netplan.DeleteInput{Path: file(name)})

	return result.Changed, err
}

// have is osapi's routes of the interface, from the file that osapi wrote:
// none when there is no file.
func (d *debian) have(
	name string,
) ([]spec, error) {
	data, err := d.fs.ReadFile(file(name))
	if errors.Is(err, fs.ErrNotExist) {
		// No file is no routes.
		return nil, nil
	}
	if err != nil {
		// Any other error, such as permission denied, is not "no routes": a
		// write would drop the routes that the file has.
		return nil, fmt.Errorf("read %s: %w", file(name), err)
	}
	var doc struct {
		Network map[string]any `yaml:"network"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("read %s: %w", file(name), err)
	}
	for section, defs := range doc.Network {
		if section == "version" {
			continue
		}
		raw, _ := yaml.Marshal(defs)
		var byID map[string]struct {
			Routes []spec `yaml:"routes"`
		}
		if err := yaml.Unmarshal(raw, &byID); err != nil {
			return nil, fmt.Errorf("read %s: %w", file(name), err)
		}
		for _, def := range byID {
			// A file from before canon has a destination in the spelling it
			// was given: compare it in the one spelling.
			for i := range def.Routes {
				def.Routes[i].To = canon(def.Routes[i].To)
			}

			return def.Routes, nil
		}
	}

	return nil, nil
}

// write writes the interface's routes under its own netplan definition. An
// interface the host does not have is not found, as gohai reads it. An
// interface that a match defines, such as cloud-init's all for en*, is
// refused: netplan applies that definition, so the routes would go on each
// interface it matches, now and later.
func (d *debian) write(
	ctx context.Context,
	name string,
	specs []spec,
) (bool, error) {
	if _, err := d.Routes(ctx, name); err != nil {
		return false, err
	}
	iface, err := behavior.Call(ctx, netplan.Behavior, netplan.Get, d.netplan,
		netplan.GetInput{Name: name})
	if err != nil {
		return false, err
	}
	if iface.ID != name {
		return false, fmt.Errorf(
			"interface %q is in netplan definition %q, a match that also takes %s: "+
				"give the interface a definition of its own first",
			name, iface.ID, others(iface.Matches, name),
		)
	}
	result, err := behavior.Call(
		ctx,
		netplan.Behavior,
		netplan.Update,
		d.netplan,
		netplan.UpdateInput{
			Path:     file(name),
			Content:  render(iface.ID, iface.Section, specs),
			Metadata: map[string]string{"domain": "route", "interface": name},
		},
	)

	return result.Changed, err
}

// others is the interfaces of names but name, or what the match takes when
// names has no other.
func others(
	names []string,
	name string,
) string {
	out := []string{}
	for _, n := range names {
		if n != name {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return "any interface it matches later"
	}

	return strings.Join(out, ", ")
}

// render is the netplan config of an interface's routes, under id, its
// netplan definition.
func render(
	id string,
	section string,
	routes []spec,
) string {
	var b strings.Builder
	b.WriteString("network:\n  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n    %s:\n      routes:\n", section, id)
	for _, r := range routes {
		fmt.Fprintf(&b, "        - to: %s\n          via: %s\n", r.To, r.Via)
		if r.Metric > 0 {
			fmt.Fprintf(&b, "          metric: %d\n", r.Metric)
		}
	}

	return b.String()
}
