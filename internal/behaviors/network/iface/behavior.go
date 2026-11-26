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

// Package iface is the interface behavior: the network interfaces of the
// host.
package iface

import (
	"regexp"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// macPattern is a MAC address as netplan's macaddress takes it: six octets
// in hex, split by colons.
const macPattern = `^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`

var macRE = regexp.MustCompile(macPattern)

// The host of the Linux examples: eth0 has a DHCP address and the default
// route, and lo is left out of a list.
var linuxHost = []behavior.HostOption{
	behavior.OnPlatform(behavior.PlatformDebian, behavior.PlatformLinux),
	behavior.Exists("/sys/class/net/eth0/ifindex", "2"),
	behavior.Exists("/sys/class/net/eth0/mtu", "1500"),
	behavior.Exists("/sys/class/net/eth0/address", "52:54:00:12:34:56"),
	behavior.Exists("/sys/class/net/eth0/flags", "0x1003"),
	behavior.Exists("/sys/class/net/eth0/type", "1"),
	behavior.Exists("/sys/class/net/lo/ifindex", "1"),
	behavior.Exists("/sys/class/net/lo/mtu", "65536"),
	behavior.Exists("/sys/class/net/lo/address", "00:00:00:00:00:00"),
	behavior.Exists("/sys/class/net/lo/flags", "0x9"),
	behavior.Exists("/sys/class/net/lo/type", "772"),
	behavior.Prints("ip -o addr show", "1: lo    inet 127.0.0.1/8 scope host lo\n"+
		"2: eth0    inet 10.0.0.5/24 brd 10.0.0.255 scope global dynamic eth0\n"+
		"2: eth0    inet6 2001:db8::5/64 scope global\n"+
		"2: eth0    inet6 fe80::5054:ff:fe12:3456/64 scope link\n"),
	behavior.Prints("ip -o -4 route show table main",
		"default via 10.0.0.1 dev eth0 proto dhcp src 10.0.0.5 metric 100\n"+
			"10.0.0.0/24 dev eth0 proto kernel scope link src 10.0.0.5\n"),
}

// eth0Entry is eth0 of linuxHost.
var eth0Entry = Entry{
	Name: "eth0", State: "up", MAC: "52:54:00:12:34:56", MTU: 1500,
	IPv4: "10.0.0.5", IPv6: "2001:db8::5",
	Addresses: []string{"10.0.0.5/24", "2001:db8::5/64", "fe80::5054:ff:fe12:3456/64"},
	DHCP4:     true, Primary: true,
}

// The host of the Mac examples.
var darwinHost = []behavior.HostOption{
	behavior.OnPlatform(behavior.PlatformDarwin),
	behavior.Prints("ifconfig -a -v",
		"lo0: flags=8049<UP,LOOPBACK,RUNNING,MULTICAST> mtu 16384 index 1\n"+
			"\tinet 127.0.0.1 netmask 0xff000000\n"+
			"en0: flags=8863<UP,BROADCAST,RUNNING,MULTICAST> mtu 1500 index 14\n"+
			"\tether 3c:22:fb:00:00:01\n"+
			"\tinet 192.168.0.40 netmask 0xffffff00 broadcast 192.168.0.255\n"),
	behavior.Prints("netstat -rn", "Routing tables\n\nInternet:\n"+
		"Destination        Gateway            Flags               Netif Expire\n"+
		"default            192.168.0.1        UGScg                 en0\n"),
}

// en0Entry is en0 of darwinHost.
var en0Entry = Entry{
	Name: "en0", State: "up", MAC: "3c:22:fb:00:00:01", MTU: 1500,
	IPv4: "192.168.0.40", Addresses: []string{"192.168.0.40/24"}, Primary: true,
}

// osapi0 is the dummy interface of the write examples, and the config osapi
// writes for it. The integration tests run the examples on a live host, so
// no example sets a gateway: a default route through osapi0 would take the
// host's traffic.
const (
	osapi0Status = `{"lo":{"type":"ethernet"}}`
	osapi0File   = netplan.Dir + "/osapi-osapi0.yaml"
	osapi0YAML   = "network:\n  version: 2\n  dummy-devices:\n    osapi0:\n" +
		"      dhcp4: false\n      dhcp6: false\n" +
		"      addresses:\n        - 10.0.1.5/24\n"
)

// osapi0Static is a dummy interface with a static address.
var osapi0Static = CreateInput{Name: "osapi0", Type: "dummy", Addresses: []string{"10.0.1.5/24"}}

// osapi0Host is a host with the link of osapi0, which netplan status does
// not list until netplan manages it.
var osapi0Host = []behavior.HostOption{
	behavior.Exists("/sys/class/net/osapi0/ifindex", "3"),
	behavior.Exists("/sys/class/net/osapi0/flags", "0x82"),
	behavior.Prints("netplan status --format json", osapi0Status),
}

// writes is what a write of osapi0's config runs, on a host that does not
// have osapi0 yet: netplan makes it.
var writes = []behavior.HostOption{
	behavior.Prints("netplan status --format json", osapi0Status),
	behavior.Ran("netplan", "generate"),
	behavior.Ran("netplan", "apply"),
}

// eth1Host is a host with eth1, an interface that netplan does not manage.
var eth1Host = []behavior.HostOption{
	behavior.Exists("/sys/class/net/eth1/ifindex", "3"),
	behavior.Exists("/sys/class/net/eth1/flags", "0x1003"),
	behavior.Prints("netplan status --format json", osapi0Status),
	behavior.Ran("netplan", "generate"),
	behavior.Ran("netplan", "apply"),
	behavior.NotLive("it configures an interface that the host has"),
}

// refuse adds to a write the refusals of each rule of Config, with as
// making the method's input of a Config.
func refuse(
	m *behavior.Method,
	as func(Config) any,
) *behavior.Method {
	for _, r := range []struct {
		doc   string
		in    Config
		field string
	}{
		{"An address without a prefix", Config{Addresses: []string{"10.0.1.5"}}, "Addresses"},
		{"More than 16 addresses", Config{Addresses: seventeen}, "Addresses"},
		{"An IPv4 gateway that is not IPv4", Config{Gateway4: "fe80::1"}, "Gateway4"},
		{"An IPv6 gateway that is not IPv6", Config{Gateway6: "10.0.1.1"}, "Gateway6"},
		{"An MTU below 68", Config{MTU: 10}, "MTU"},
		{"An MTU above 9000", Config{MTU: 9001}, "MTU"},
		{"A MAC address netplan does not take", Config{MACAddress: "0000.5e00.5301"}, "MACAddress"},
		{"A name that is not an interface name", Config{Name: "eth0;reboot"}, "Name"},
		{"A type osapi does not know", Config{Type: "vlan"}, "Type"},
	} {
		if r.in.Name == "" {
			r.in.Name = "osapi0"
		}
		m = m.Fails(r.doc, as(r.in), r.field)
	}

	return m
}

// seventeen is one address more than a config takes.
var seventeen = strings.Split(strings.Repeat("10.0.1.5/24 ", 17), " ")[:17]

// Behavior is the network interfaces of the host. gohai reads them on every
// platform; a netplan file of osapi's own configures one on the Debian
// family.
var Behavior = behavior.Register(behavior.Define(
	"interface", "The network interfaces of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Networking),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Requires(netplan.Behavior),
	behavior.RunsOnPlatform(behavior.PlatformLinux, "ip"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "ifconfig", "netstat"),
	behavior.RunsIfPresent("ethtool"),
	behavior.SudoOnPlatform(behavior.PlatformDebian, "ip"),
	behavior.Writes(netplan.Dir+"/osapi-*.yaml"),
	behavior.Validator("mac_address", macRE.MatchString,
		func(t rule.Target) { t.Pattern(macPattern) },
		"is not a MAC address of six octets split by colons"),
	behavior.Read("list", List).
		Describe("List the network interfaces of the host").
		Explain("Reads each interface but the loopback, in the order of its index: "+
			"/sys/class/net and ip on Linux, ifconfig and netstat on macOS. dhcp4 is true "+
			"when a route of the interface has proto dhcp, and primary when it has the "+
			"default route.").
		Example("A Linux host", ListInput{}, []Entry{eth0Entry}, linuxHost...).
		Example("A Mac", ListInput{}, []Entry{en0Entry}, darwinHost...),
	behavior.Read("get", Get).Key("name").
		Describe("Get a network interface of the host").
		Explain("Reads the interface as list does.").
		Example("An interface with a DHCP address", GetInput{Name: "eth0"}, eth0Entry, linuxHost...).
		Example("A Mac", GetInput{Name: "en0"}, en0Entry, darwinHost...).
		Fails("An interface the host does not have", GetInput{Name: "eth9"}, "not found",
			linuxHost...).
		Fails("A name that is not an interface name", GetInput{Name: "eth0;reboot"}, "Name").
		Fails("A fact the agent does not know", GetInput{Name: "@fact.nonexistent"}, "Name"),
	refuse(behavior.Write("create", Create).Key("name").
		On(behavior.PlatformDebian).
		Describe("Configure a network interface").
		Explain("Writes a netplan file of the interface's own, such as "+
			"/etc/netplan/osapi-osapi0.yaml for osapi0, under a netplan definition that "+
			"is the interface's own, and applies it. type is ethernet, the default, or "+
			"dummy. An ethernet is an interface the host has, in the netplan section that "+
			"netplan status gives for it: one the host does not have is not found. A dummy "+
			"goes under dummy-devices, and netplan apply makes it, so the host need not "+
			"have it. dummy-devices needs netplan 0.107 or later. dhcp4 and dhcp6 are "+
			"false unless given, so the interface has only the addresses given. An "+
			"interface that a match defines, such as cloud-init's all for en*, is refused, "+
			"because the config would change each interface the match takes, now and later. "+
			"An interface osapi already configures does not change: update replaces its "+
			"config. It runs on the Debian family, and elsewhere, and in a container, "+
			"reports skipped.").
		Example("A dummy interface with a static address", osapi0Static,
			Result{Name: "osapi0", Changed: true},
			append(
				append([]behavior.HostOption{behavior.Wrote(osapi0File, osapi0YAML)}, writes...),
				behavior.Reads(GetInput{Name: "osapi0"}, Entry{Name: "osapi0", Addresses: []string{"10.0.1.5/24"}}),
			)...,
		).
		Example("Each setting",
			CreateInput{
				Name: "osapi0", Type: "dummy", DHCP6: true, Addresses: []string{"2001:db8::5/64"},
				MTU: 9000, MACAddress: "52:54:00:aa:bb:cc",
			},
			Result{Name: "osapi0", Changed: true},
			append(
				append([]behavior.HostOption{
					behavior.Wrote(osapi0File, "network:\n  version: 2\n  dummy-devices:\n    osapi0:\n"+
						"      dhcp4: false\n      dhcp6: true\n"+
						"      addresses:\n        - 2001:db8::5/64\n"+
						"      mtu: 9000\n      macaddress: 52:54:00:aa:bb:cc\n"),
				}, writes...),
				behavior.Reads(GetInput{Name: "osapi0"}, Entry{Name: "osapi0", MTU: 9000, MAC: "52:54:00:aa:bb:cc"}),
			)...,
		).
		Example("An interface the host has",
			CreateInput{Name: "eth1", DHCP4: true, WakeOnLAN: true},
			Result{Name: "eth1", Changed: true},
			append([]behavior.HostOption{
				behavior.Wrote(netplan.Dir+"/osapi-eth1.yaml",
					"network:\n  version: 2\n  ethernets:\n    eth1:\n"+
						"      dhcp4: true\n      dhcp6: false\n      wakeonlan: true\n"),
			}, eth1Host...)...).
		Example("An interface osapi already configures", osapi0Static, Result{Name: "osapi0"},
			append(
				slices.Clone(osapi0Host),
				behavior.After(osapi0Static),
				behavior.Reads(GetInput{Name: "osapi0"}, Entry{Name: "osapi0", Addresses: []string{"10.0.1.5/24"}}),
			)...,
		).
		Fails("An interface in a match that takes other interfaces",
			CreateInput{Name: "enp0s1", DHCP4: true}, "also takes enp0s2",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json",
				`{"enp0s1":{"type":"ethernet","id":"all"},"enp0s2":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface in a match that takes it alone, for now",
			CreateInput{Name: "enp0s1", DHCP4: true}, "any interface it matches later",
			behavior.Exists("/sys/class/net/enp0s1/ifindex", "2"),
			behavior.Prints("netplan status --format json", `{"enp0s1":{"type":"ethernet","id":"all"}}`),
		).
		Fails("An interface the host does not have", CreateInput{Name: "eth9", DHCP4: true}, "not found",
			osapi0Host...),
		func(c Config) any { return CreateInput(c) }),
	refuse(behavior.Write("update", Update).Key("name").
		On(behavior.PlatformDebian).
		Describe("Change the config osapi has of a network interface").
		Explain("Writes the interface's netplan file over the one osapi has, as create "+
			"writes it, and applies it. Without type dummy, the config goes in the section "+
			"that netplan status gives, which is dummy-devices for a dummy, so a dummy "+
			"stays a dummy. An interface osapi does not configure is not found: create "+
			"configures it.").
		Example("Move an interface to DHCP", UpdateInput{Name: "osapi0", Type: "dummy", DHCP4: true},
			Result{Name: "osapi0", Changed: true},
			append(
				slices.Clone(writes),
				behavior.After(osapi0Static),
				behavior.Wrote(osapi0File, "network:\n  version: 2\n  dummy-devices:\n    osapi0:\n"+
					"      dhcp4: true\n      dhcp6: false\n"),
				behavior.Reads(GetInput{Name: "osapi0"}, Entry{Name: "osapi0"}, "name", "addresses"),
			)...,
		).
		Fails("An interface osapi does not configure", UpdateInput{Name: "osapi0", DHCP4: true},
			"not found"),
		func(c Config) any { return UpdateInput(c) }),
	behavior.Write("delete", Delete).Key("name").
		On(behavior.PlatformDebian).
		Describe("Drop the config osapi has of a network interface").
		Explain("Removes the interface's netplan file and applies what is left. netplan "+
			"apply leaves the link of a dummy, so delete then deletes the link with ip "+
			"link delete. An interface osapi does not configure does not change.").
		Example("A dummy interface osapi configures", DeleteInput{Name: "osapi0"},
			Result{Name: "osapi0", Changed: true},
			append(slices.Clone(osapi0Host),
				behavior.After(osapi0Static),
				behavior.Removed(osapi0File),
				behavior.Ran("netplan", "apply"),
				behavior.Ran("ip", "link", "delete", "osapi0"),
				behavior.ReadsNone(GetInput{Name: "osapi0"}, Entry{Name: "osapi0"}, "name"),
			)...,
		).
		Example("An interface osapi does not configure", DeleteInput{Name: "osapi0"},
			Result{Name: "osapi0"},
			behavior.ReadsNone(GetInput{Name: "osapi0"}, Entry{Name: "osapi0"}, "name"),
		).
		Fails("A name that is not an interface name", DeleteInput{Name: "eth0;reboot"}, "Name"),
))
