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

// Package netplan is the netplan behavior: the netplan files that osapi
// writes, validated and applied.
package netplan

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// The file of the examples, and its YAML.
const (
	dnsFile = Dir + "/osapi-dns.yaml"
	dnsYAML = "network:\n  version: 2\n  ethernets:\n    eth0:\n      nameservers:\n        addresses: [1.1.1.1]\n"
)

// status is netplan status for a host with a bond and an ethernet.
const status = `{"netplan-global-state":{"online":true},` +
	`"enp0s1":{"index":2,"type":"ethernet","id":"all"},"enp0s2":{"index":3,"type":"ethernet","id":"all"},` +
	`"bond0":{"index":4,"type":"bond","id":"bond0"},"lo":{"index":1,"type":"ethernet"}}`

// Behavior writes, validates and applies the netplan files of the dns,
// interface and route behaviors, which require it. It keeps each file's
// record through managedfile.
var Behavior = behavior.Register(behavior.Define(
	"netplan", "The netplan files osapi writes, validated and applied",
	behavior.Node,
	behavior.Internal,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Requires(managedfile.Behavior),
	behavior.Sudo("netplan"),
	behavior.Runs("netplan", "ip"),
	behavior.RunsIfPresent("ethtool"),
	behavior.Writes(Dir+"/osapi-*.yaml"),
	behavior.Write("update", Update).
		Describe("Write a netplan file, validate it and apply it").
		Explain("Writes the file with mode 0600, validates the whole config with netplan "+
			"generate, and applies it with netplan apply. When netplan generate refuses "+
			"it, the file gets back the content it had, or a new file is removed, because "+
			"netplan reads every file in the directory. A file that already holds the "+
			"content changes nothing.").
		Example("A new file", UpdateInput{Path: dnsFile, Content: dnsYAML},
			Result{Path: dnsFile, Changed: true},
			behavior.Wrote(dnsFile, dnsYAML),
			behavior.Ran("netplan", "generate"),
			behavior.Ran("netplan", "apply"),
		).
		Fails("A file outside osapi's own", UpdateInput{Path: Dir + "/50-cloud-init.yaml", Content: dnsYAML},
			"'startswith' tag").
		Fails("A name that leaves the directory",
			UpdateInput{Path: Dir + "/osapi-../../shadow.yaml", Content: dnsYAML},
			"not one of osapi's netplan files").
		Fails("A name that is not YAML", UpdateInput{Path: Dir + "/osapi-dns.conf", Content: dnsYAML},
			"not one of osapi's netplan files"),
	behavior.Write("delete", Delete).
		Describe("Remove a netplan file and apply what is left").
		Example("A file osapi wrote", DeleteInput{Path: dnsFile},
			Result{Path: dnsFile, Changed: true},
			behavior.After(UpdateInput{Path: dnsFile, Content: dnsYAML}),
			behavior.Removed(dnsFile),
			behavior.Ran("netplan", "apply"),
		).
		Example("A file that is not there", DeleteInput{Path: dnsFile}, Result{Path: dnsFile}).
		Fails("A file outside osapi's own", DeleteInput{Path: Dir + "/50-cloud-init.yaml"},
			"'startswith' tag").
		Fails("A name that leaves the directory", DeleteInput{Path: Dir + "/osapi-../../shadow.yaml"},
			"not one of osapi's netplan files"),
	behavior.Read("get", Get).Key("name").
		NotAFact("the definition that netplan has of an interface is the configuration of netplan, not a fact of the host").
		Describe("Get what netplan status reports for an interface").
		Explain("Reads netplan status: the interface's type, its section and the ID of the "+
			"definition that matches it, which a write goes under, and each interface that "+
			"definition matches. netplan lists only the interfaces it manages, and an older "+
			"netplan has no status. An interface that status does not list goes under its "+
			"own name, with the type of its kind as gohai reads it from sysfs: a bridge, a "+
			"bond, a Wi-Fi card, a VLAN or a tunnel. One of no other kind is an ethernet.").
		Example("A bond", GetInput{Name: "bond0"},
			Interface{
				Name:    "bond0",
				Type:    "bond",
				Section: "bonds",
				ID:      "bond0",
				Matches: []string{"bond0"},
			},
			behavior.Prints("netplan status --format json", status),
		).
		Example("A dummy", GetInput{Name: "osapi0"},
			Interface{
				Name:    "osapi0",
				Type:    "dummy-device",
				Section: "dummy-devices",
				ID:      "osapi0",
				Matches: []string{"osapi0"},
			},
			behavior.Prints("netplan status --format json",
				`{"osapi0":{"type":"dummy-device","id":"osapi0"}}`),
		).
		Example("An interface a match defines, as cloud-init's all for en*", GetInput{Name: "enp0s1"},
			Interface{
				Name: "enp0s1", Type: "ethernet", Section: "ethernets", ID: "all",
				Matches: []string{"enp0s1", "enp0s2"},
			},
			behavior.Prints("netplan status --format json", status),
		).
		Example("An interface netplan does not define", GetInput{Name: "lo"},
			Interface{
				Name:    "lo",
				Type:    "ethernet",
				Section: "ethernets",
				ID:      "lo",
				Matches: []string{"lo"},
			},
			behavior.Prints("netplan status --format json", status),
		).
		Example("A netplan without status", GetInput{Name: "eth0"},
			Interface{Name: "eth0", Section: "ethernets", ID: "eth0", Matches: []string{"eth0"}},
			behavior.Prints("netplan status --format json", "Invalid command: status\n"),
		).
		Example("An interface netplan does not manage", GetInput{Name: "eth9"},
			Interface{Name: "eth9", Section: "ethernets", ID: "eth9", Matches: []string{"eth9"}},
			behavior.Prints("netplan status --format json", status),
			behavior.Exists("/sys/class/net/eth0/type", "1\n"),
		).
		Example("A bridge on a netplan without status", GetInput{Name: "br0"},
			Interface{
				Name:    "br0",
				Type:    "bridge",
				Section: "bridges",
				ID:      "br0",
				Matches: []string{"br0"},
			},
			behavior.Prints("netplan status --format json", "Invalid command: status\n"),
			behavior.Exists("/sys/class/net/br0/type", "1\n"),
			behavior.Exists("/sys/class/net/br0/bridge/bridge_id", "8000.000000000000\n"),
		).
		Example("A Wi-Fi card that netplan does not list", GetInput{Name: "wlan0"},
			Interface{
				Name:    "wlan0",
				Type:    "wifi",
				Section: "wifis",
				ID:      "wlan0",
				Matches: []string{"wlan0"},
			},
			behavior.Prints("netplan status --format json", status),
			behavior.Exists("/sys/class/net/wlan0/type", "1\n"),
			behavior.Exists("/sys/class/net/wlan0/wireless/.keep", ""),
		).
		Example("A VLAN on a netplan without status", GetInput{Name: "eth0.10"},
			Interface{
				Name: "eth0.10", Type: "vlan", Section: "vlans", ID: "eth0.10",
				Matches: []string{"eth0.10"},
			},
			behavior.Prints("netplan status --format json", "Invalid command: status\n"),
			behavior.Exists("/sys/class/net/eth0.10/type", "1\n"),
			behavior.Exists("/sys/class/net/eth0.10/uevent", "DEVTYPE=vlan\nINTERFACE=eth0.10\n"),
		),
))
