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
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// The host of the Linux examples: eth0 has a DHCP default route, the route
// of its link and a static route.
var linuxHost = []behavior.HostOption{
	behavior.OnPlatform(behavior.PlatformDebian, behavior.PlatformLinux),
	behavior.Exists("/sys/class/net/eth0/ifindex", "2"),
	behavior.Exists("/sys/class/net/eth0/flags", "0x1003"),
	behavior.Exists("/sys/class/net/eth1/ifindex", "3"),
	behavior.Exists("/sys/class/net/eth1/flags", "0x1003"),
	behavior.Prints("ip -o -4 route show table main",
		"default via 10.0.0.1 dev eth0 proto dhcp src 10.0.0.5 metric 100\n"+
			"10.0.0.0/24 dev eth0 proto kernel scope link src 10.0.0.5\n"+
			"10.20.0.0/16 via 10.0.0.254 dev eth0 proto static metric 50\n"),
}

// The routes of eth0 on linuxHost.
var eth0Routes = []Entry{
	{
		Destination: "default",
		Gateway:     "10.0.0.1",
		Interface:   "eth0",
		Family:      "inet",
		Metric:      100,
		Proto:       "dhcp",
	},
	{Destination: "10.0.0.0/24", Interface: "eth0", Family: "inet", Proto: "kernel", Scope: "link"},
	{
		Destination: "10.20.0.0/16",
		Gateway:     "10.0.0.254",
		Interface:   "eth0",
		Family:      "inet",
		Metric:      50,
		Proto:       "static",
	},
}

// The host of the Mac examples.
var darwinHost = []behavior.HostOption{
	behavior.OnPlatform(behavior.PlatformDarwin),
	behavior.Prints("ifconfig -a -v",
		"en0: flags=8863<UP,BROADCAST,RUNNING,MULTICAST> mtu 1500 index 14\n"+
			"\tinet 192.168.0.40 netmask 0xffffff00 broadcast 192.168.0.255\n"),
	behavior.Prints("netstat -rn", "Routing tables\n\nInternet:\n"+
		"Destination        Gateway            Flags               Netif Expire\n"+
		"default            192.168.0.1        UGScg                 en0\n"),
}

// en0Routes is the routes of en0 on darwinHost.
var en0Routes = []Entry{
	{Destination: "default", Gateway: "192.168.0.1", Interface: "en0", Family: "inet"},
}

// osapi0 is the interface of the write examples, and the routes osapi writes
// for it. The integration tests run the examples on a live host, with osapi0
// a dummy interface, so a route goes only to documentation addresses.
const (
	osapi0File = netplan.Dir + "/osapi-osapi0-routes.yaml"
	osapi0YAML = "network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n" +
		"        - to: 198.51.100.0/24\n          via: 192.0.2.1\n" +
		"        - to: 2001:db8:1::/48\n          via: 2001:db8::1\n          metric: 200\n"
)

// osapi0Static is two routes of osapi0.
var osapi0Static = CreateInput{
	InterfaceName: "osapi0",
	Routes: []string{
		"198.51.100.0/24 via 192.0.2.1",
		"2001:db8:1::/48 via 2001:db8::1 metric 200",
	},
}

// osapi0Host is a host with osapi0, which netplan does not manage, so
// netplan status does not list it.
var osapi0Host = []behavior.HostOption{
	behavior.Exists("/sys/class/net/osapi0/ifindex", "3"),
	behavior.Exists("/sys/class/net/osapi0/flags", "0x82"),
	behavior.Prints("netplan status --format json", `{"lo":{"type":"ethernet"}}`),
}

// writes is osapi0Host, and what a write of osapi0's routes runs.
var writes = append(slices.Clone(osapi0Host),
	behavior.Ran("netplan", "generate"),
	behavior.Ran("netplan", "apply"),
)

// refuse adds to a write the refusals of each rule of Config, with as making
// the method's input of a Config.
func refuse(
	m *behavior.Method,
	as func(Config) any,
) *behavior.Method {
	for _, r := range []struct {
		doc   string
		in    Config
		field string
	}{
		{"No route", Config{Routes: []string{}}, "Routes"},
		{"More than 64 routes", Config{Routes: sixtyFive}, "Routes"},
		{"A route without via", Config{Routes: []string{"10.0.0.0/8 192.168.1.1"}}, "Routes"},
		{"A destination that is not CIDR", Config{Routes: []string{"10.0.0.0 via 192.168.1.1"}}, "Routes"},
		{"A gateway of the other family", Config{Routes: []string{"10.0.0.0/8 via 2001:db8::1"}}, "Routes"},
		{"A metric that is not a number", Config{Routes: []string{"10.0.0.0/8 via 192.168.1.1 metric x"}}, "Routes"},
		{"A name that is not an interface name", Config{InterfaceName: "eth0;reboot"}, "InterfaceName"},
	} {
		if r.in.InterfaceName == "" {
			r.in.InterfaceName = "osapi0"
		}
		if r.in.Routes == nil {
			r.in.Routes = osapi0Static.Routes
		}
		m = m.Fails(r.doc, as(r.in), r.field)
	}

	return m
}

// sixtyFive is one route more than a config takes.
var sixtyFive = strings.Split(strings.Repeat("10.0.0.0/8 via 192.168.1.1,", 65), ",")[:65]

// Behavior is the routes of the host. gohai reads them on every platform; a
// netplan file of osapi's own writes an interface's routes on the Debian
// family.
var Behavior = behavior.Register(behavior.Define(
	"route", "The routes of the host",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	behavior.Under("network"),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Requires(netplan.Behavior),
	behavior.RunsOnPlatform(behavior.PlatformLinux, "ip"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "ifconfig", "netstat"),
	behavior.RunsIfPresent("ethtool"),
	behavior.Writes(netplan.Dir+"/osapi-*-routes.yaml"),
	behavior.Validator(
		"route",
		valid,
		func(t rule.Target) { t.Pattern(`^\S+ via \S+( metric [0-9]+)?$`) },
		"is not <destination> via <gateway> [metric <n>], with a CIDR destination and a gateway of its family",
	),
	behavior.Read("list", List).
		Describe("List the routes of the host").
		Explain("Reads each route of the main table: ip route on Linux, netstat on macOS. "+
			"proto says where a route came from, such as dhcp, kernel or static.").
		Example("A Linux host", ListInput{}, eth0Routes, linuxHost...).
		Example("A Mac", ListInput{}, en0Routes, darwinHost...),
	behavior.Read("get", Get).Key("interface_name").
		Describe("Get the routes of an interface").
		Explain("Reads the routes of the main table that go out of the interface, as list "+
			"does. An interface with no routes has none; one the host does not have is "+
			"not found.").
		Example("An interface with routes", GetInput{InterfaceName: "eth0"}, eth0Routes, linuxHost...).
		Example("An interface with no routes", GetInput{InterfaceName: "eth1"}, []Entry{}, linuxHost...).
		Example("A Mac", GetInput{InterfaceName: "en0"}, en0Routes, darwinHost...).
		Fails("An interface the host does not have", GetInput{InterfaceName: "eth9"}, "not found",
			linuxHost...).
		Fails("A name that is not an interface name", GetInput{InterfaceName: "eth0;reboot"},
			"InterfaceName"),
	refuse(behavior.Write("create", Create).Key("interface_name").
		On(behavior.PlatformDebian).
		Describe("Add static routes to an interface").
		Explain("Writes a netplan file of the interface's routes, such as "+
			"/etc/netplan/osapi-osapi0-routes.yaml for osapi0, under a netplan definition "+
			"that is the interface's own, and applies it. A default route is refused: the "+
			"interface behavior sets the gateway. An interface the host does not have is not "+
			"found. An interface that a match defines, such as cloud-init's all for en*, is "+
			"refused, because the routes would go on each interface the match takes, now and "+
			"later. An interface osapi already has routes of does not change: update "+
			"replaces them. It runs on the Debian family, and elsewhere, and in a "+
			"container, reports skipped.").
		Example("Two routes", osapi0Static, Result{InterfaceName: "osapi0", Changed: true},
			append(slices.Clone(writes), behavior.Wrote(osapi0File, osapi0YAML))...).
		Example("An interface osapi already has routes of", osapi0Static,
			Result{InterfaceName: "osapi0"},
			append(slices.Clone(osapi0Host), behavior.After(osapi0Static))...).
		Fails("A default route",
			CreateInput{InterfaceName: "osapi0", Routes: []string{"0.0.0.0/0 via 192.0.2.1"}},
			"default route", osapi0Host...).
		Fails("A default route on an interface osapi already has routes of",
			CreateInput{InterfaceName: "osapi0", Routes: []string{"::/0 via 2001:db8::1"}},
			"default route", append(slices.Clone(osapi0Host), behavior.After(osapi0Static))...).
		Fails("An interface in a match that takes other interfaces",
			CreateInput{InterfaceName: "enp0s1", Routes: osapi0Static.Routes}, "also takes enp0s2",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json",
				`{"enp0s1":{"type":"ethernet","id":"all"},"enp0s2":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface in a match that takes it alone, for now",
			CreateInput{
				InterfaceName: "enp0s1",
				Routes:        osapi0Static.Routes,
			}, "any interface it matches later",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json", `{"enp0s1":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface the host does not have",
			CreateInput{
				InterfaceName: "eth9",
				Routes:        osapi0Static.Routes,
			}, "not found", osapi0Host...),
		func(c Config) any { return CreateInput(c) }),
	refuse(behavior.Write("update", Update).Key("interface_name").
		On(behavior.PlatformDebian).
		Describe("Change the routes osapi has of an interface").
		Explain("Writes the interface's routes file over the one osapi has, as create writes "+
			"it, and applies it. An interface osapi has no routes of is not found: create "+
			"adds them.").
		Example("One route in place of two",
			UpdateInput{
				InterfaceName: "osapi0",
				Routes:        []string{"203.0.113.0/24 via 192.0.2.1 metric 10"},
			},
			Result{InterfaceName: "osapi0", Changed: true},
			append(
				slices.Clone(writes),
				behavior.After(osapi0Static),
				behavior.Wrote(
					osapi0File,
					"network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n"+
						"        - to: 203.0.113.0/24\n          via: 192.0.2.1\n          metric: 10\n",
				),
			)...).
		Fails("An interface osapi has no routes of",
			UpdateInput{InterfaceName: "osapi0", Routes: osapi0Static.Routes}, "not found").
		Fails("A default route",
			UpdateInput{InterfaceName: "osapi0", Routes: []string{"0.0.0.0/0 via 192.0.2.1"}},
			"default route"),
		func(c Config) any { return UpdateInput(c) }),
	behavior.Write("delete", Delete).Key("interface_name").
		On(behavior.PlatformDebian).
		Describe("Drop the routes osapi has of an interface").
		Explain("Removes the interface's routes file and applies what is left. An interface "+
			"osapi has no routes of does not change.").
		Example("An interface osapi has routes of", DeleteInput{InterfaceName: "osapi0"},
			Result{InterfaceName: "osapi0", Changed: true},
			append(slices.Clone(osapi0Host),
				behavior.After(osapi0Static),
				behavior.Removed(osapi0File),
				behavior.Ran("netplan", "apply"),
			)...,
		).
		Example("An interface osapi has no routes of", DeleteInput{InterfaceName: "osapi0"},
			Result{InterfaceName: "osapi0"}).
		Fails("A name that is not an interface name", DeleteInput{InterfaceName: "eth0;reboot"},
			"InterfaceName"),
))
