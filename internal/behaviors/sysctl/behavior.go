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

// Package sysctl is the sysctl behavior: kernel parameters osapi writes to
// /etc/sysctl.d and applies.
//
// The files follow the layout every behavior has (see docs/docs/sidebar/
// development/behaviors.md): behavior.go is the definition, types.go the
// input and output types, methods.go the methods, provider.go the interface
// each platform implements, and gen/ what osapi generates from them.
package sysctl

import (
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

// Behavior is the sysctl definition. Everything else, the routes, the agent's
// dispatch, the CLI commands, the spec and the docs, is built from it.
var Behavior = behavior.Register(behavior.Define(
	"sysctl", "Kernel parameters, written to sysctl.d files and applied",
	behavior.Node,
	behavior.Config(defaults),
	behavior.Requires(managedfile.Behavior),
	behavior.Platform("debian", newDebian),
	behavior.Sudo("sysctl"),
	behavior.API,
	behavior.CLI,
	behavior.Read("list", List).
		Describe("List the kernel's parameters and which osapi manages").
		Explain("Lists every parameter the kernel shows under /proc/sys, with its current "+
			"value, as sysctl -a does, and marks the ones osapi manages. A prefix, such as "+
			"net.ipv4, narrows it to that part of the tree. A parameter osapi manages that "+
			"the kernel does not have right now, such as one whose module is not loaded, "+
			"is listed with the value osapi wrote.").
		Example("Every parameter, and which osapi manages", ListInput{}, []Entry{
			{Key: "kernel.panic", Value: "0"},
			{Key: "net.ipv4.ip_forward", Value: "1", Managed: true},
			{Key: "net.ipv4.ping_group_range", Value: "0 2147483647"},
		}, kernel()...).
		Example("One part of the tree", ListInput{Prefix: "net.ipv4"}, []Entry{
			{Key: "net.ipv4.ip_forward", Value: "1", Managed: true},
			{Key: "net.ipv4.ping_group_range", Value: "0 2147483647"},
		}, kernel()...).
		Example("A managed parameter the kernel does not have right now",
			ListInput{Prefix: "net.bridge"}, []Entry{
				{Key: "net.bridge.bridge-nf-call-iptables", Value: "1", Managed: true},
			},
			behavior.After(CreateInput{Key: "net.bridge.bridge-nf-call-iptables", Value: "1"}),
		).
		Fails("A prefix that is not a parameter name", ListInput{Prefix: "../etc"},
			"invalid sysctl key"),
	behavior.Read("get", Get).Key("key").
		Describe("Get a kernel parameter's current value, and whether osapi manages it").
		Example("A parameter osapi manages", GetInput{Key: "vm.swappiness"},
			Entry{Key: "vm.swappiness", Value: "10", Managed: true},
			behavior.After(CreateInput{Key: "vm.swappiness", Value: "10"}),
			behavior.Exists(proc("vm.swappiness"), "10\n"),
		).
		Example("A parameter osapi does not manage", GetInput{Key: "vm.swappiness"},
			Entry{Key: "vm.swappiness", Value: "60"},
			behavior.Exists(proc("vm.swappiness"), "60\n"),
		).
		Example("A value of several fields", GetInput{Key: "net.ipv4.ping_group_range"},
			Entry{Key: "net.ipv4.ping_group_range", Value: "0 2147483647"},
			behavior.Exists(proc("net.ipv4.ping_group_range"), "0\t2147483647\n"),
		).
		Example("A managed parameter the kernel does not have right now",
			GetInput{Key: "net.bridge.bridge-nf-call-iptables"},
			Entry{Key: "net.bridge.bridge-nf-call-iptables", Value: "1", Managed: true},
			behavior.After(CreateInput{Key: "net.bridge.bridge-nf-call-iptables", Value: "1"}),
		).
		Fails("A parameter the kernel does not have", GetInput{Key: "vm.nosuch"}, "not found").
		Fails("No key", GetInput{}, "key must not be empty").
		Fails("A key that is not a parameter name", GetInput{Key: "vm.swappiness;reboot"},
			"invalid sysctl key").
		Fails("A key holding a value", GetInput{Key: "vm.swappiness=1"}, "invalid sysctl key"),
	behavior.Write("create", Create).
		Describe("Manage a kernel parameter and apply it").
		Explain("Writes osapi-<key>.conf to behaviors.sysctl.dir, /etc/sysctl.d by default, "+
			"and applies it with sysctl -p. "+
			"Running it again with the same value changes nothing and reports changed: false.").
		Example("Turn on IP forwarding",
			CreateInput{Key: "net.ipv4.ip_forward", Value: "1"},
			MutationResult{Key: "net.ipv4.ip_forward", Changed: true},
			behavior.Wrote(conf("net.ipv4.ip_forward"), "net.ipv4.ip_forward = 1\n"),
			behavior.Ran("sysctl", "-p", conf("net.ipv4.ip_forward")),
		).
		Fails("No key", CreateInput{Value: "1"}, "key must not be empty").
		Fails("A key that is not a parameter name",
			CreateInput{Key: "../etc/passwd", Value: "1"}, "invalid sysctl key").
		Fails("A key that would read as a flag",
			CreateInput{Key: "-w", Value: "1"}, "invalid sysctl key").
		Fails("A key that would be a hidden file",
			CreateInput{Key: ".hidden", Value: "1"}, "invalid sysctl key").
		Fails("No value", CreateInput{Key: "vm.swappiness"}, "value must not be empty").
		Fails("A value on two lines",
			CreateInput{Key: "vm.swappiness", Value: "10\nkernel.panic = 0"}, "line breaks"),
	behavior.Write("update", Update).Key("key").
		Describe("Set a kernel parameter, manage it, and apply it").
		Explain("The same as create, addressed by the parameter: osapi manages the parameter "+
			"from then on whether it did before or not. Setting the value it already has "+
			"reports changed: false.").
		Example("Lower swappiness",
			UpdateInput{Key: "vm.swappiness", Value: "10"},
			MutationResult{Key: "vm.swappiness", Changed: true},
			behavior.After(CreateInput{Key: "vm.swappiness", Value: "60"}),
			behavior.Wrote(conf("vm.swappiness"), "vm.swappiness = 10\n"),
			behavior.Ran("sysctl", "-p", conf("vm.swappiness")),
		).
		Example("A parameter osapi did not manage",
			UpdateInput{Key: "vm.swappiness", Value: "10"},
			MutationResult{Key: "vm.swappiness", Changed: true},
			behavior.Exists(proc("vm.swappiness"), "60\n"),
			behavior.Wrote(conf("vm.swappiness"), "vm.swappiness = 10\n"),
			behavior.Ran("sysctl", "-p", conf("vm.swappiness")),
		).
		Fails("No key", UpdateInput{Value: "10"}, "key must not be empty").
		Fails("A key that is not a parameter name",
			UpdateInput{Key: "vm.swappiness;reboot", Value: "1"}, "invalid sysctl key").
		Fails("A key holding a value",
			UpdateInput{Key: "vm.swappiness=1", Value: "1"}, "invalid sysctl key").
		Fails("No value", UpdateInput{Key: "vm.swappiness"}, "value must not be empty").
		Fails("A value on two lines",
			UpdateInput{Key: "vm.swappiness", Value: "10\nkernel.panic = 0"}, "line breaks"),
	behavior.Write("delete", Delete).Key("key").
		Describe("Stop managing a kernel parameter").
		Explain("Removes osapi's file for it and reloads with sysctl --system, so a setting "+
			"for the key in another file takes effect; otherwise the parameter keeps its "+
			"current value until the next reboot. Deleting one osapi does not manage "+
			"reports changed: false.").
		Example("Remove a parameter", DeleteInput{Key: "vm.swappiness"},
			MutationResult{Key: "vm.swappiness", Changed: true},
			behavior.After(CreateInput{Key: "vm.swappiness", Value: "10"}),
			behavior.Removed(conf("vm.swappiness")),
			behavior.Ran("sysctl", "--system"),
		).
		Example("A parameter osapi does not manage", DeleteInput{Key: "vm.swappiness"},
			MutationResult{Key: "vm.swappiness"}).
		Example("A parameter osapi already stopped managing", DeleteInput{Key: "vm.swappiness"},
			MutationResult{Key: "vm.swappiness"},
			behavior.After(CreateInput{Key: "vm.swappiness", Value: "10"}),
			behavior.After(DeleteInput{Key: "vm.swappiness"}),
		).
		Fails("No key", DeleteInput{}, "key must not be empty").
		Fails("A key that is not a parameter name", DeleteInput{Key: "vm.swappiness;reboot"},
			"invalid sysctl key").
		Fails("A key holding a value", DeleteInput{Key: "vm.swappiness=1"}, "invalid sysctl key"),
))

// conf is the file a parameter is written to under the default directory,
// for the examples.
func conf(
	key string,
) string {
	return DefaultDir + "/osapi-" + key + ".conf"
}

// proc is the file the kernel shows a parameter in, for the examples.
func proc(
	key string,
) string {
	return ProcDir + "/" + strings.ReplaceAll(key, ".", "/")
}

// kernel is a host with three parameters, one of which osapi manages, for
// the list examples. A file under the tree's root is a parameter too; a
// directory is not.
func kernel() []behavior.HostOption {
	return []behavior.HostOption{
		behavior.Exists(proc("kernel.panic"), "0\n"),
		behavior.Exists(proc("net.ipv4.ip_forward"), "1\n"),
		behavior.Exists(proc("net.ipv4.ping_group_range"), "0\t2147483647\n"),
		behavior.After(CreateInput{Key: "net.ipv4.ip_forward", Value: "1"}),
	}
}
