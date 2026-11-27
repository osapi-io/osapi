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

package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
)

// debian reads through gohai and sets the resolvers with a netplan file of
// their own per interface, through the netplan behavior.
type debian struct {
	*reader
	netplan   netplan.Provider
	container bool
}

func newDebian(
	h host.Host,
) (Provider, error) {
	return &debian{
		reader:    &reader{src: h.Sources()},
		netplan:   host.Require[netplan.Provider](h, netplan.Behavior),
		container: h.Container,
	}, nil
}

// file is the netplan file of an interface's resolvers.
func file(
	name string,
) string {
	return netplan.Dir + "/osapi-dns-" + name + ".yaml"
}

// Set writes the netplan file of the interface's resolvers. An empty list
// writes nothing for that list, so the interface keeps the servers or domains
// that DHCP gives it, as they change. With overrideDHCP, DHCP gives none, so
// an empty list keeps the interface's current servers or domains, written in
// the file. In a container the runtime owns resolv.conf, so it reports
// skipped.
func (d *debian) Set(
	ctx context.Context,
	in UpdateInput,
) (bool, error) {
	if d.container {
		return false, behavior.ErrUnsupported
	}
	if len(in.Servers) == 0 && len(in.SearchDomains) == 0 {
		return false, fmt.Errorf("give servers or search domains: nothing to set")
	}
	if err := d.has(ctx, in.InterfaceName); err != nil {
		return false, err
	}
	servers, domains := in.Servers, in.SearchDomains
	if in.OverrideDHCP && (len(servers) == 0 || len(domains) == 0) {
		// Resolvers fails only with not found: an interface with no
		// resolvers yet, such as one with a static address. It starts from
		// none.
		current, _ := d.Resolvers(ctx, in.InterfaceName)
		if len(servers) == 0 {
			servers = current.Servers
		}
		if len(domains) == 0 {
			domains = current.SearchDomains
		}
	}
	iface, err := behavior.Call(ctx, netplan.Behavior, netplan.Get, d.netplan,
		netplan.GetInput{Name: in.InterfaceName})
	if err != nil {
		return false, err
	}
	result, err := behavior.Call(
		ctx,
		netplan.Behavior,
		netplan.Update,
		d.netplan,
		netplan.UpdateInput{
			Path:     file(in.InterfaceName),
			Content:  yaml(iface.ID, iface.Section, servers, domains, in.OverrideDHCP),
			Metadata: map[string]string{"domain": "dns", "interface": in.InterfaceName},
		},
	)

	return result.Changed, err
}

// has is nil when the host has the interface, read through gohai: netplan
// lists only the interfaces it manages.
func (d *debian) has(
	ctx context.Context,
	name string,
) error {
	out, err := network.NewWith(d.src).Collect(ctx, nil)
	if err != nil {
		return fmt.Errorf("read the interfaces: %w", err)
	}
	for _, i := range out.(*network.Info).Interfaces {
		if i.Name == name {
			return nil
		}
	}

	return fmt.Errorf("interface %q: %w", name, behavior.ErrNotFound)
}

// Remove drops the netplan file of the interface's resolvers.
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

// yaml is the netplan config of an interface's resolvers, under id, the
// netplan definition that matches it: netplan merges it into that
// definition, so each interface the definition matches gets them. "." is the root
// domain that resolvectl shows for an interface with none, not a search
// domain, so it is left out. With overrideDHCP, DHCPv4 and DHCPv6 give no
// servers; IPv6 router advertisements still can, because turning them off
// would drop the host's IPv6 route.
func yaml(
	id string,
	section string,
	servers []string,
	domains []string,
	overrideDHCP bool,
) string {
	var b strings.Builder
	b.WriteString("network:\n  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n    %s:\n      nameservers:\n", section, id)
	if len(servers) > 0 {
		b.WriteString("        addresses:\n")
		for _, s := range servers {
			fmt.Fprintf(&b, "          - %s\n", s)
		}
	}
	search := make([]string, 0, len(domains))
	for _, d := range domains {
		if d != "." {
			search = append(search, d)
		}
	}
	if len(search) > 0 {
		b.WriteString("        search:\n")
		for _, d := range search {
			fmt.Fprintf(&b, "          - %s\n", d)
		}
	}
	if overrideDHCP {
		b.WriteString("      dhcp4-overrides:\n        use-dns: false\n")
		b.WriteString("      dhcp6-overrides:\n        use-dns: false\n")
	}

	return b.String()
}
