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
	`"enp0s1":{"index":2,"type":"ethernet","id":"all"},"bond0":{"index":3,"type":"bond","id":"bond0"}}`

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
	behavior.Runs("netplan"),
	behavior.Writes(Dir+"/osapi-*.yaml"),
	behavior.Write("update", Update).
		Describe("Write a netplan file, validate it and apply it").
		Explain("Writes the file with mode 0600, validates the whole config with netplan "+
			"generate, and applies it with netplan apply. When netplan generate refuses "+
			"it, the file is removed again, because netplan reads every file in the "+
			"directory. A file that already holds the content changes nothing.").
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
		Describe("Get what netplan status reports for an interface").
		Explain("Reads netplan status: the interface's type, its section and the ID of the "+
			"definition that matches it, which a write goes under. An interface that it does "+
			"not list is an ethernet under its own name.").
		Example("A bond", GetInput{Name: "bond0"},
			Interface{Name: "bond0", Type: "bond", Section: "bonds", ID: "bond0"},
			behavior.Prints("netplan status --format json", status),
		).
		Example("An interface a match defines, as cloud-init's all for en*", GetInput{Name: "enp0s1"},
			Interface{Name: "enp0s1", Type: "ethernet", Section: "ethernets", ID: "all"},
			behavior.Prints("netplan status --format json", status),
		).
		Example("An interface netplan does not list", GetInput{Name: "eth9"},
			Interface{Name: "eth9", Section: "ethernets", ID: "eth9"},
			behavior.Prints("netplan status --format json", status),
		),
))
