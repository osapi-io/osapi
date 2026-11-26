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

// Package group is the group behavior: the groups of the host. It reads them
// through gohai on every platform, and adds, changes and removes them on the
// Debian family.
package group

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// etcGroup is the /etc/group of the examples: a host with the group
// developers, whose members are alice and bob.
const etcGroup = "root:x:0:\ndevelopers:x:1500:alice,bob\n"

var developers = Account{Name: "developers", GID: 1500, Members: []string{"alice", "bob"}}

// added is /etc/group once groupadd has added osapitest.
var added = behavior.ExistsAfter(
	"groupadd -- osapitest",
	"/etc/group",
	etcGroup+"osapitest:x:1501:\n",
)

// Behavior is the group behavior: the groups of the host. The integration test
// runs the write examples as root on CI's runner, each from its group removed
// first. osapitest is a name nothing on the runner has.
var Behavior = behavior.Register(behavior.Define(
	"group", "The groups of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Security),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsIfPresent("id", "dscl"),
	behavior.Read("list", List).
		Describe("List the groups of the host").
		Explain("Lists each group with its ID and its members, by name, from /etc/group "+
			"through gohai, on every platform. The members are the users that /etc/group "+
			"names; a user whose primary group it is is not listed again.").
		Example("A host with a group", ListInput{},
			[]Account{developers, {Name: "root", GID: 0}},
			behavior.Exists("/etc/group", etcGroup),
		),
	behavior.Read("get", Get).Key("name").
		Describe("Get one group").
		Explain("Reads one group from the same list. A group that the host does not "+
			"have is not found.").
		Example("A group", GetInput{Name: "developers"}, developers,
			behavior.Exists("/etc/group", etcGroup),
		).
		Fails("A group that the host does not have", GetInput{Name: "nobody"}, "not found").
		Fails("A name that is an option", GetInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			GetInput{Name: strings.Repeat("a", 33)}, "max"),
	behavior.Write("create", Create).
		On(behavior.PlatformDebian).
		Sudo("groupadd").
		Describe("Add a group").
		Explain("Adds the group with groupadd, with the ID when one is given. A group "+
			"that the host has already does not change and reports changed: false, "+
			"whatever its ID. It runs on the Debian family only, and elsewhere reports "+
			"skipped.").
		Example("Add a group", CreateInput{Name: "osapitest"},
			Result{Name: "osapitest", Changed: true},
			behavior.Exists("/etc/group", etcGroup),
			added,
			behavior.Ran("groupadd", "--", "osapitest"),
			behavior.Reads(GetInput{Name: "osapitest"}, Account{Name: "osapitest"}),
		).
		Example("Add a system group with its ID", CreateInput{Name: "osapitest", GID: 1999, System: true},
			Result{Name: "osapitest", Changed: true},
			behavior.Exists("/etc/group", etcGroup),
			behavior.ExistsAfter("groupadd -g 1999 -r -- osapitest", "/etc/group",
				etcGroup+"osapitest:x:1999:\n"),
			behavior.Ran("groupadd", "-g", "1999", "-r", "--", "osapitest"),
			behavior.Reads(GetInput{Name: "osapitest"}, Account{Name: "osapitest", GID: 1999}),
		).
		Example("A group that the host has", CreateInput{Name: "osapitest"},
			Result{Name: "osapitest"},
			behavior.Exists("/etc/group", etcGroup),
			added,
			behavior.After(CreateInput{Name: "osapitest"}),
			behavior.Reads(GetInput{Name: "osapitest"}, Account{Name: "osapitest"}),
		).
		Fails("A name that is an option", CreateInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			CreateInput{Name: strings.Repeat("a", 33)}, "max").
		Fails("A negative ID", CreateInput{Name: "osapitest", GID: -1}, "min"),
	behavior.Write("update", Update).Key("name").
		On(behavior.PlatformDebian).
		Sudo("gpasswd").
		Describe("Set the members of a group").
		Explain("Sets the members with gpasswd -M, in place of the members that the group "+
			"has. The same members, in any order, do not change it and report changed: "+
			"false. A group that the host does not have is not found. It runs on the "+
			"Debian family only, and elsewhere reports skipped.").
		Example("Set the members", UpdateInput{Name: "osapitest", Members: []string{"root"}},
			Result{Name: "osapitest", Changed: true},
			behavior.Exists("/etc/group", etcGroup),
			added,
			behavior.ExistsAfter("gpasswd -M root -- osapitest", "/etc/group",
				etcGroup+"osapitest:x:1501:root\n"),
			behavior.After(CreateInput{Name: "osapitest"}),
			behavior.Ran("gpasswd", "-M", "root", "--", "osapitest"),
			behavior.Reads(GetInput{Name: "osapitest"}, Account{Name: "osapitest", Members: []string{"root"}}),
		).
		Example("The members it has", UpdateInput{Name: "osapitest", Members: []string{"root"}},
			Result{Name: "osapitest"},
			behavior.Exists("/etc/group", etcGroup),
			added,
			behavior.ExistsAfter("gpasswd -M root -- osapitest", "/etc/group",
				etcGroup+"osapitest:x:1501:root\n"),
			behavior.After(CreateInput{Name: "osapitest"}),
			behavior.After(UpdateInput{Name: "osapitest", Members: []string{"root"}}),
			behavior.Reads(GetInput{Name: "osapitest"}, Account{Name: "osapitest", Members: []string{"root"}}),
		).
		Fails("A group that the host does not have", UpdateInput{Name: "nobody"}, "not found").
		Fails("A name that is an option", UpdateInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			UpdateInput{Name: strings.Repeat("a", 33)}, "max").
		Fails("A member that is an option",
			UpdateInput{Name: "developers", Members: []string{"-r"}}, "account_name").
		Fails("A member longer than 32 characters",
			UpdateInput{Name: "developers", Members: []string{strings.Repeat("a", 33)}}, "max"),
	behavior.Write("delete", Delete).Key("name").
		On(behavior.PlatformDebian).
		Sudo("groupdel").
		Describe("Remove a group").
		Explain("Removes the group with groupdel. A group that the host does not have "+
			"does not change and reports changed: false. It runs on the Debian family "+
			"only, and elsewhere reports skipped.").
		Example("Remove a group", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest", Changed: true},
			behavior.Exists("/etc/group", etcGroup),
			added,
			behavior.ExistsAfter("groupdel -- osapitest", "/etc/group", etcGroup),
			behavior.After(CreateInput{Name: "osapitest"}),
			behavior.Ran("groupdel", "--", "osapitest"),
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Account{Name: "osapitest"}, "name"),
		).
		Example("A group that the host does not have", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest"},
			behavior.Exists("/etc/group", etcGroup),
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Account{Name: "osapitest"}, "name"),
		).
		Fails("A name that is an option", DeleteInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			DeleteInput{Name: strings.Repeat("a", 33)}, "max"),
))
