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

package iface

import (
	"context"
	"fmt"
	"strings"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
)

// debian reads through gohai and writes an interface's config with a
// netplan file of its own, through the netplan behavior.
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

// file is the netplan file of an interface's config.
func file(
	name string,
) string {
	return netplan.Dir + "/osapi-" + name + ".yaml"
}

// Create writes the interface's config. When osapi already has a config of
// the interface, it changes nothing: update replaces it.
func (d *debian) Create(
	ctx context.Context,
	in Config,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	if d.exists(in.Name) {
		return false, nil
	}

	return d.write(ctx, in)
}

// Update writes the interface's config over the one osapi has.
func (d *debian) Update(
	ctx context.Context,
	in Config,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	if !d.exists(in.Name) {
		return false, fmt.Errorf(
			"osapi has no config of interface %q: %w",
			in.Name,
			behavior.ErrNotFound,
		)
	}

	return d.write(ctx, in)
}

// Remove drops osapi's config of the interface.
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

// exists reports whether osapi has a config of the interface.
func (d *debian) exists(
	name string,
) bool {
	_, err := d.fs.Stat(file(name))

	return err == nil
}

// write writes the interface's config under its own netplan definition. An
// interface the host does not have is not found: netplan lists only the
// interfaces it manages, so gohai answers that. An interface that a match
// defines, such as cloud-init's all for en*, is refused: netplan applies that
// definition, so the config would change each interface it matches, now and
// later.
func (d *debian) write(
	ctx context.Context,
	in Config,
) (bool, error) {
	if _, err := d.Interface(ctx, in.Name); err != nil {
		return false, err
	}
	iface, err := behavior.Call(ctx, netplan.Behavior, netplan.Get, d.netplan,
		netplan.GetInput{Name: in.Name})
	if err != nil {
		return false, err
	}
	if iface.ID != in.Name {
		return false, fmt.Errorf(
			"interface %q is in netplan definition %q, a match that also takes %s: "+
				"give the interface a definition of its own first",
			in.Name, iface.ID, others(iface.Matches, in.Name),
		)
	}
	result, err := behavior.Call(
		ctx,
		netplan.Behavior,
		netplan.Update,
		d.netplan,
		netplan.UpdateInput{
			Path:     file(in.Name),
			Content:  yaml(iface.ID, iface.Section, in),
			Metadata: map[string]string{"domain": "interface", "interface": in.Name},
		},
	)

	return result.Changed, err
}

// others is the interfaces of names but name, or what the match takes
// when names has no other.
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

// yaml is the netplan config of an interface, under id, its netplan
// definition. dhcp4 and dhcp6 are always written, so a
// config without them is static, whatever another file of the definition
// says. The gateways are default routes: netplan deprecates gateway4 and
// gateway6.
func yaml(
	id string,
	section string,
	in Config,
) string {
	var b strings.Builder
	b.WriteString("network:\n  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n    %s:\n", section, id)
	fmt.Fprintf(&b, "      dhcp4: %t\n      dhcp6: %t\n", in.DHCP4, in.DHCP6)
	if len(in.Addresses) > 0 {
		b.WriteString("      addresses:\n")
		for _, a := range in.Addresses {
			fmt.Fprintf(&b, "        - %s\n", a)
		}
	}
	if in.Gateway4 != "" || in.Gateway6 != "" {
		b.WriteString("      routes:\n")
		for _, gw := range []string{in.Gateway4, in.Gateway6} {
			if gw != "" {
				fmt.Fprintf(&b, "        - to: default\n          via: %s\n", gw)
			}
		}
	}
	if in.MTU > 0 {
		fmt.Fprintf(&b, "      mtu: %d\n", in.MTU)
	}
	if in.MACAddress != "" {
		fmt.Fprintf(&b, "      macaddress: %s\n", in.MACAddress)
	}
	if in.WakeOnLAN {
		b.WriteString("      wakeonlan: true\n")
	}

	return b.String()
}
