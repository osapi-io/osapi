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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/internal/behaviors/network/iface"
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
	routesHead = "network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n"
	v4YAML     = "        - to: 198.51.100.0/24\n          via: 192.0.2.1\n"
	v6YAML     = "        - to: 2001:db8:1::/48\n          via: 2001:db8::1\n          metric: 200\n"
)

// v4 and v6 are two routes of osapi0.
var (
	v4 = CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.1"}
	v6 = CreateInput{
		Name:        "osapi0",
		Destination: "2001:db8:1::/48",
		Via:         "2001:db8::1",
		Metric:      200,
	}
)

// osapi0Host is a host with osapi0, which netplan does not manage, so
// netplan status does not list it.
var osapi0Host = []behavior.HostOption{
	behavior.Exists("/sys/class/net/osapi0/ifindex", "3"),
	behavior.Exists("/sys/class/net/osapi0/flags", "0x82"),
	behavior.Prints("netplan status --format json", `{"lo":{"type":"ethernet"}}`),
}

// writes is osapi0Host, what a write of osapi0's routes runs, and what the
// example adds.
func writes(
	more ...behavior.HostOption,
) []behavior.HostOption {
	return append(append(slices.Clone(osapi0Host),
		behavior.Ran("netplan", "generate"),
		behavior.Ran("netplan", "apply"),
	), more...)
}

// refuse adds to a write the refusals of the validate tags of Route, with as
// making the method's input of a Route.
func refuse(
	m *behavior.Method,
	as func(Route) any,
) *behavior.Method {
	for _, r := range []struct {
		doc   string
		edit  func(*Route)
		field string
	}{
		{"A destination that is not CIDR", func(r *Route) { r.Destination = "10.0.0.0" }, "Destination"},
		{"A gateway that is not an address", func(r *Route) { r.Via = "gateway" }, "Via"},
		{"A metric below 0", func(r *Route) { r.Metric = -1 }, "Metric"},
		{"A metric over 4294967295", func(r *Route) { r.Metric = 4294967296 }, "Metric"},
		{"A name that is not an interface name", func(r *Route) { r.Name = "eth0;reboot" }, "Name"},
	} {
		in := Route(v4)
		r.edit(&in)
		m = m.Fails(r.doc, as(in), r.field)
	}

	return m
}

// Behavior is the routes of an interface. gohai reads them on every platform;
// a netplan file of osapi's own writes them on the Debian family, one route a
// call, as ip route add and ip route del do.
var Behavior = behavior.Register(behavior.Define(
	"route", "The routes of an interface",
	behavior.Node,
	behavior.API,
	behavior.CLI,
	behavior.Of(iface.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Requires(netplan.Behavior),
	behavior.RunsOnPlatform(behavior.PlatformLinux, "ip"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "ifconfig", "netstat"),
	behavior.RunsIfPresent("ethtool"),
	behavior.Writes(netplan.Dir+"/osapi-*-routes.yaml"),
	behavior.Read("list", List).
		Describe("List the routes of an interface").
		Explain("Reads the routes of the main table that go out of the interface: ip route "+
			"on Linux, netstat on macOS. proto says where a route came from, such as dhcp, "+
			"kernel or static. An interface with no routes has none; one the host does not "+
			"have is not found.").
		Example("An interface with routes", ListInput{Name: "eth0"}, eth0Routes, linuxHost...).
		Example("An interface with no routes", ListInput{Name: "eth1"}, []Entry{}, linuxHost...).
		Example("A Mac", ListInput{Name: "en0"}, en0Routes, darwinHost...).
		Fails("An interface the host does not have", ListInput{Name: "eth9"}, "not found",
			linuxHost...).
		Fails("A name that is not an interface name", ListInput{Name: "eth0;reboot"}, "Name"),
	refuse(behavior.Write("create", Create).
		On(behavior.PlatformDebian).
		Describe("Add a static route to an interface").
		Explain("Adds the route to the interface's netplan file, such as "+
			"/etc/netplan/osapi-osapi0-routes.yaml for osapi0, under a netplan definition "+
			"that is the interface's own, and applies it. A route osapi already has, the "+
			"same, does not change; one to the same destination by another gateway or "+
			"metric is refused: update changes it. A default route is refused: the "+
			"interface behavior sets the gateway. An interface the host does not have is "+
			"not found. An interface that a match defines, such as cloud-init's all for "+
			"en*, is refused, because the route would go on each interface the match "+
			"takes. It runs on the Debian family, and elsewhere, and in a container, "+
			"reports skipped.").
		Example("A route", v4, Result{Destination: v4.Destination, Changed: true},
			writes(behavior.Wrote(osapi0File, routesHead+v4YAML))...).
		Example("A second route, with a metric", v6, Result{Destination: v6.Destination, Changed: true},
			writes(behavior.After(v4), behavior.Wrote(osapi0File, routesHead+v4YAML+v6YAML))...).
		Example("A route osapi already has", v4, Result{Destination: v4.Destination},
			append(slices.Clone(osapi0Host), behavior.After(v4))...).
		Fails("A default route",
			CreateInput{Name: "osapi0", Destination: "0.0.0.0/0", Via: "192.0.2.1"},
			"default route", osapi0Host...).
		Fails("A gateway of the other family",
			CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "2001:db8::1"},
			"not an address of the family", osapi0Host...).
		Fails("Another route to a destination osapi has",
			CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.2"},
			"update changes it", append(slices.Clone(osapi0Host), behavior.After(v4))...).
		Fails("An interface in a match that takes other interfaces",
			CreateInput{
				Name:        "enp0s1",
				Destination: v4.Destination,
				Via:         v4.Via,
			}, "also takes enp0s2",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json",
				`{"enp0s1":{"type":"ethernet","id":"all"},"enp0s2":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface in a match that takes it alone, for now",
			CreateInput{Name: "enp0s1", Destination: v4.Destination, Via: v4.Via},
			"any interface it matches later",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json", `{"enp0s1":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface the host does not have",
			CreateInput{Name: "eth9", Destination: v4.Destination, Via: v4.Via}, "not found",
			osapi0Host...),
		func(r Route) any { return CreateInput(r) }),
	refuse(behavior.Write("update", Update).
		On(behavior.PlatformDebian).
		Describe("Change the gateway or the metric of a route").
		Explain("Writes the route over the one osapi has to the same destination, and "+
			"applies it. A destination osapi has no route to is not found: create adds it.").
		Example("A new gateway and metric",
			UpdateInput{
				Name:        "osapi0",
				Destination: "198.51.100.0/24",
				Via:         "192.0.2.2",
				Metric:      10,
			},
			Result{Destination: v4.Destination, Changed: true},
			writes(behavior.After(v4), behavior.Wrote(osapi0File, routesHead+
				"        - to: 198.51.100.0/24\n          via: 192.0.2.2\n          metric: 10\n"))...).
		Fails("A destination osapi has no route to", UpdateInput(v4), "not found", osapi0Host...),
		func(r Route) any { return UpdateInput(r) }),
	behavior.Write("delete", Delete).
		On(behavior.PlatformDebian).
		Describe("Drop a route that osapi wrote").
		Explain("Drops the route from the interface's netplan file and applies what is "+
			"left; the last route removes the file. A route osapi does not have does not "+
			"change.").
		Example("One of two routes", DeleteInput{Name: "osapi0", Destination: v4.Destination},
			Result{Destination: v4.Destination, Changed: true},
			writes(behavior.After(v4), behavior.After(v6),
				behavior.Wrote(osapi0File, routesHead+v6YAML))...).
		Example("The last route", DeleteInput{Name: "osapi0", Destination: v4.Destination},
			Result{Destination: v4.Destination, Changed: true},
			append(slices.Clone(osapi0Host), behavior.After(v4),
				behavior.Removed(osapi0File), behavior.Ran("netplan", "apply"))...).
		Example("A route osapi does not have", DeleteInput{Name: "osapi0", Destination: v4.Destination},
			Result{Destination: v4.Destination}).
		Fails("A destination that is not CIDR", DeleteInput{Name: "osapi0", Destination: "10.0.0.0"},
			"Destination").
		Fails("A name that is not an interface name", DeleteInput{Name: "eth0;reboot", Destination: v4.Destination},
			"Name"),
))
