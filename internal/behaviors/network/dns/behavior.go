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
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
)

// The host of the examples: osapi0 has resolvectl's servers and domain. The
// integration test runs the write examples on a CI runner, which makes
// osapi0, a dummy interface, so they never change the runner's own network,
// whose interface is not always eth0.
var linuxHost = []behavior.HostOption{
	behavior.Exists("/sys/class/net/osapi0/ifindex", "2"),
	behavior.Prints("resolvectl dns", "Global:\nLink 2 (osapi0): 1.1.1.1 8.8.8.8\n"),
	behavior.Prints("resolvectl domain", "Global:\nLink 2 (osapi0): example.com\n"),
}

// osapi0File and osapi0YAML are the netplan file that sets osapi0's resolvers.
const (
	osapi0File = netplan.Dir + "/osapi-dns-osapi0.yaml"
	osapi0YAML = "network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n" +
		"        addresses:\n          - 9.9.9.9\n        search:\n          - example.com\n"
)

// Behavior is the DNS resolvers of each interface of the host. gohai reads
// them on every platform; a netplan file of their own sets them on the
// Debian family.
var Behavior = behavior.Register(behavior.Define(
	"dns", "The DNS resolvers of each interface of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaNetworking),
	behavior.CLI,
	behavior.Under("network"),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Requires(netplan.Behavior),
	behavior.RunsOnPlatform(behavior.PlatformLinux, "resolvectl", "ip"),
	behavior.RunsIfPresent("ethtool"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "scutil"),
	behavior.Writes(netplan.Dir+"/osapi-dns-*.yaml"),
	behavior.Read("get", Get).Key("interface_name").
		Describe("Get the DNS resolvers of an interface").
		Explain("Reads the servers and the search domains of the interface: resolvectl "+
			"on Linux, with resolv.conf on a host that has no resolvers per interface, "+
			"such as a container, and scutil on macOS.").
		Example("An interface with servers and a domain", GetInput{InterfaceName: "osapi0"},
			Entry{Servers: []string{"1.1.1.1", "8.8.8.8"}, SearchDomains: []string{"example.com"}},
			append(
				[]behavior.HostOption{
					behavior.OnPlatform(behavior.PlatformDebian, behavior.PlatformLinux),
				},
				linuxHost...)...,
		).
		Example("A container with no resolvers per interface", GetInput{InterfaceName: "osapi0"},
			Entry{Servers: []string{"10.0.0.2"}, SearchDomains: []string{"svc.local"}},
			behavior.OnPlatform(behavior.PlatformDebian, behavior.PlatformLinux),
			behavior.Exists("/etc/resolv.conf", "nameserver 10.0.0.2\nsearch svc.local\n"),
		).
		Example("A Mac", GetInput{InterfaceName: "en0"},
			Entry{Servers: []string{"1.1.1.1"}, SearchDomains: []string{}},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints(
				"scutil --dns",
				"DNS configuration\n\nresolver #1\n  nameserver[0] : 1.1.1.1\n  if_index : 14 (en0)\n",
			),
		).
		Fails("An interface the host does not have", GetInput{InterfaceName: "eth9"}, "not found",
			append([]behavior.HostOption{
				behavior.Prints(
					"scutil --dns",
					"resolver #1\n  nameserver[0] : 1.1.1.1\n  if_index : 14 (en0)\n",
				),
			}, linuxHost...)...).
		Fails("An interface name that is not alphanumeric", GetInput{InterfaceName: "osapi0;reboot"},
			"InterfaceName"),
	behavior.Write("update", Update).Key("interface_name").
		On(behavior.PlatformDebian).
		Describe("Set the DNS resolvers of an interface").
		Explain("Writes a netplan file of the interface's own, "+
			"such as /etc/netplan/osapi-dns-eth0.yaml for eth0, under the netplan definition that "+
			"matches the interface, and applies it. A definition that matches several "+
			"interfaces, such as cloud-init's all for en*, gives each of them the resolvers. An empty list keeps "+
			"the interface's servers or domains, and an interface with no resolvers yet starts from none. override_dhcp drops the servers DHCP "+
			"gives. It runs on the Debian family, and elsewhere, and in a container, "+
			"reports skipped.").
		Example("Set the servers of an interface", UpdateInput{InterfaceName: "osapi0", Servers: []string{"9.9.9.9"}},
			Result{InterfaceName: "osapi0", Changed: true},
			append([]behavior.HostOption{
				behavior.Prints("netplan status --format json", `{"osapi0":{"type":"ethernet","id":"osapi0"}}`),
				behavior.Wrote(osapi0File, osapi0YAML),
				behavior.Ran("netplan", "generate"),
				behavior.Ran("netplan", "apply"),
			}, linuxHost...)...,
		).
		Example("Set the search domains, and keep the servers",
			UpdateInput{
				InterfaceName: "osapi0",
				SearchDomains: []string{"corp.example"},
				OverrideDHCP:  true,
			},
			Result{InterfaceName: "osapi0", Changed: true},
			append([]behavior.HostOption{
				behavior.Prints(
					"netplan status --format json",
					`{"osapi0":{"type":"ethernet","id":"osapi0"}}`,
				),
				behavior.Wrote(
					osapi0File,
					"network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n"+
						"        addresses:\n          - 1.1.1.1\n          - 8.8.8.8\n        search:\n          - corp.example\n"+
						"      dhcp4-overrides:\n        use-dns: false\n      dhcp6-overrides:\n        use-dns: false\n",
				),
				behavior.Ran("netplan", "generate"),
				behavior.Ran("netplan", "apply"),
			}, linuxHost...)...,
		).
		Example("An interface with no resolvers yet",
			UpdateInput{InterfaceName: "osapi0", Servers: []string{"9.9.9.9"}},
			Result{InterfaceName: "osapi0", Changed: true},
			behavior.Exists("/sys/class/net/osapi0/ifindex", "2"),
			behavior.Prints("resolvectl dns", "Global:\nLink 3 (eth1): 1.1.1.1\n"),
			behavior.Prints("resolvectl domain", "Global:\n"),
			behavior.Prints("netplan status --format json", `{"osapi0":{"type":"ethernet","id":"osapi0"}}`),
			behavior.Wrote(osapi0File, "network:\n  version: 2\n  ethernets:\n    osapi0:\n      nameservers:\n"+
				"        addresses:\n          - 9.9.9.9\n"),
			behavior.Ran("netplan", "generate"),
			behavior.Ran("netplan", "apply"),
		).
		Fails("No servers and no search domains", UpdateInput{InterfaceName: "osapi0"},
			"give servers or search domains").
		Fails("An interface the host does not have, to keep the servers of",
			UpdateInput{
				InterfaceName: "eth9",
				SearchDomains: []string{"corp.example"},
			}, "not found",
			append([]behavior.HostOption{
				behavior.Prints(
					"netplan status --format json",
					`{"osapi0":{"type":"ethernet","id":"osapi0"}}`,
				),
			}, linuxHost...)...).
		Fails("A server that is not an IP address",
			UpdateInput{InterfaceName: "osapi0", Servers: []string{"one.one"}}, "ip").
		Fails("A search domain that is not a host name",
			UpdateInput{
				InterfaceName: "osapi0",
				SearchDomains: []string{"bad domain"},
			}, "hostname").
		Fails("More than 16 servers",
			UpdateInput{
				InterfaceName: "osapi0",
				Servers:       strings.Split(strings.Repeat("1.1.1.1 ", 17), " ")[:17],
			},
			"Servers").
		Fails("More than 16 search domains",
			UpdateInput{
				InterfaceName: "osapi0",
				SearchDomains: strings.Split(strings.Repeat("a.b ", 17), " ")[:17],
			},
			"SearchDomains").
		Fails("An interface name that is not alphanumeric",
			UpdateInput{
				InterfaceName: "osapi0;reboot",
				Servers:       []string{"9.9.9.9"},
			}, "InterfaceName"),
	behavior.Write("delete", Delete).Key("interface_name").
		On(behavior.PlatformDebian).
		Describe("Drop the DNS resolvers osapi set on an interface").
		Explain("Removes the interface's netplan file and applies what is left. An "+
			"interface osapi set nothing on does not change.").
		Example("An interface osapi set", DeleteInput{InterfaceName: "osapi0"},
			Result{InterfaceName: "osapi0", Changed: true},
			append(
				slices.Clone(linuxHost),
				behavior.After(UpdateInput{InterfaceName: "osapi0", Servers: []string{"9.9.9.9"}}),
				behavior.Prints("netplan status --format json", `{"osapi0":{"type":"ethernet","id":"osapi0"}}`),
				behavior.Removed(osapi0File),
				behavior.Ran("netplan", "apply"),
			)...,
		).
		Example("An interface osapi set nothing on", DeleteInput{InterfaceName: "osapi0"},
			Result{InterfaceName: "osapi0"}).
		Fails("An interface name that is not alphanumeric", DeleteInput{InterfaceName: "osapi0;reboot"},
			"InterfaceName"),
))
