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
	"fmt"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
)

// debian reads through gohai and writes an interface's routes with a netplan
// file of their own, through the netplan behavior.
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

// Create writes the interface's routes. When osapi already has routes of the
// interface, it changes nothing: update replaces them.
func (d *debian) Create(
	ctx context.Context,
	in Config,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	specs, err := routes(in)
	if err != nil {
		return false, err
	}
	if d.exists(in.InterfaceName) {
		return false, nil
	}

	return d.write(ctx, in.InterfaceName, specs)
}

// Update writes the interface's routes over the ones osapi has.
func (d *debian) Update(
	ctx context.Context,
	in Config,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	specs, err := routes(in)
	if err != nil {
		return false, err
	}
	if !d.exists(in.InterfaceName) {
		return false, fmt.Errorf(
			"osapi has no routes of interface %q: %w", in.InterfaceName, behavior.ErrNotFound)
	}

	return d.write(ctx, in.InterfaceName, specs)
}

// Remove drops osapi's routes of the interface.
func (d *debian) Remove(
	ctx context.Context,
	name string,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	result, err := behavior.Call(ctx, netplan.Behavior, netplan.Delete, d.netplan,
		netplan.DeleteInput{Path: file(name)})

	return result.Changed, err
}

// exists reports whether osapi has routes of the interface.
func (d *debian) exists(
	name string,
) bool {
	_, err := d.fs.Stat(file(name))

	return err == nil
}

// routes is the routes of in. A default route is refused: a wrong one cuts
// the host off, and the interface behavior sets the gateway.
func routes(
	in Config,
) ([]spec, error) {
	specs := make([]spec, 0, len(in.Routes))
	for _, s := range in.Routes {
		// The route rule has checked each route before the job was sent.
		r, _ := parse(s)
		if r.isDefault() {
			return nil, fmt.Errorf(
				"route %q is a default route, which osapi does not write: "+
					"give the interface a gateway with osapi client node network interface", s)
		}
		specs = append(specs, r)
	}

	return specs, nil
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
	if _, err := d.InterfaceRoutes(ctx, name); err != nil {
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
			Content:  yaml(iface.ID, iface.Section, specs),
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

// yaml is the netplan config of an interface's routes, under id, its netplan
// definition.
func yaml(
	id string,
	section string,
	routes []spec,
) string {
	var b strings.Builder
	b.WriteString("network:\n  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n    %s:\n      routes:\n", section, id)
	for _, r := range routes {
		fmt.Fprintf(&b, "        - to: %s\n          via: %s\n", r.to, r.via)
		if r.metric > 0 {
			fmt.Fprintf(&b, "          metric: %d\n", r.metric)
		}
	}

	return b.String()
}
