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

// Package user is the user behavior: the accounts of the host. It reads them
// through gohai on every platform, and adds, changes and removes them on the
// Debian family. A password is hashed on the controller, so the job data and
// NATS have only its hash. The password behavior, under the user, sets the
// password of a user that exists.
package user

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The files of the examples: a host with root and alice, and alice a member
// of developers.
const (
	etcPasswd = "root:x:0:0:root:/root:/bin/bash\n" +
		"alice:x:1000:1000:Alice:/home/alice:/bin/bash\n"
	etcGroup  = "root:x:0:\nalice:x:1000:\ndevelopers:x:1500:alice\n"
	etcShadow = "root:*:19000:0:99999:7:::\nalice:$6$s$h:19000:0:99999:7:::\n"
)

var alice = Account{
	Name: "alice", UID: 1000, GID: 1000, Home: "/home/alice", Shell: "/bin/bash",
	Groups: []string{"developers"},
}

// testHash is a sha512-crypt hash of the examples' password, as the
// controller's step makes it. An example is the input that the agent gets,
// so it has the hash; its password is what the caller sends on a live host.
const testHash = "$6$osapitest$ixskR1vdsZl2JBlRPn/2rqKZK4DltHGVEnTxDn.wN68cQnKpsMNWzH5Hxu6jZgEHWjcnzvFneyk/KHl6bEYtj1"

// newUser is /etc/passwd once useradd has added osapitest, and /etc/shadow
// with its password locked, as useradd leaves a user with none.
var newUser = behavior.ExistsAfter(
	"useradd --create-home -- osapitest",
	"/etc/passwd",
	etcPasswd+"osapitest:x:1001:1001::/home/osapitest:/bin/sh\n",
)

// The shadow lines of osapitest: with no password, as useradd leaves it,
// with the examples' password, and with that password locked.
const (
	shadowNone   = "osapitest:!:19000:0:99999:7:::\n"
	shadowSet    = "osapitest:" + testHash + ":19000:0:99999:7:::\n"
	shadowLocked = "osapitest:!" + testHash + ":19000:0:99999:7:::\n"
)

// withPassword is a host where osapitest has the examples' password: the
// files after useradd and chpasswd, and the call that makes them.
func withPassword() []behavior.HostOption {
	return []behavior.HostOption{
		newUser,
		behavior.ExistsAfter(
			"useradd --create-home -- osapitest",
			"/etc/shadow",
			etcShadow+shadowNone,
		),
		behavior.ExistsAfter("chpasswd -e", "/etc/shadow", etcShadow+shadowSet),
		behavior.After(
			CreateInput{Name: "osapitest", Password: "osapi-test-password", PasswordHash: testHash},
		),
	}
}

// etcFiles are the files of a host with root and alice.
var etcFiles = []behavior.HostOption{
	behavior.Exists("/etc/passwd", etcPasswd),
	behavior.Exists("/etc/group", etcGroup),
	behavior.Exists("/etc/shadow", etcShadow),
}

func withHost(
	opts ...behavior.HostOption,
) []behavior.HostOption {
	return append(append([]behavior.HostOption{}, etcFiles...), opts...)
}

// The integration test runs the write examples as root on CI's runner, each
// from its user removed first. osapitest is a name nothing on the runner
// has.
var Behavior = behavior.Register(behavior.Define(
	"user", "The users of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Security),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsIfPresent("id", "dscl"),
	behavior.Read("list", List).
		Describe("List the users of the host").
		Explain("Lists each user with its IDs, home, shell, groups and whether it is "+
			"locked, through gohai: from /etc/passwd, /etc/group and /etc/shadow on "+
			"Linux, and from the directory service with dscl on macOS. The groups are "+
			"those that list the user as a member, besides its primary group.").
		Example("A host with users", ListInput{},
			[]Account{alice, {Name: "root", Home: "/root", Shell: "/bin/bash", Locked: true}},
			withHost()...,
		),
	behavior.Read("get", Get).Key("name").
		Describe("Get one user").
		Explain("Reads one user from the same list. A user that the host does not have "+
			"is not found.").
		Example("A user", GetInput{Name: "alice"}, alice, withHost()...).
		Fails("A user that the host does not have", GetInput{Name: "nobody"}, "not found").
		Fails("A name that is an option", GetInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			GetInput{Name: strings.Repeat("a", 33)}, "max"),
	behavior.Write("create", Create).
		AdminOnly().
		On(behavior.PlatformDebian).
		Sudo("useradd", "chpasswd").
		Before(hashCreate).
		Describe("Add a user").
		Explain("Adds the user with useradd, with its home directory, and sets its "+
			"password with chpasswd when one is given. The controller hashes the "+
			"password, so the job and NATS have only the hash. A user that the host "+
			"has already does not change and reports changed: false. It runs on the "+
			"Debian family only, and elsewhere reports skipped.").
		Example("Add a user", CreateInput{Name: "osapitest"},
			Result{Name: "osapitest", Changed: true},
			withHost(newUser, behavior.Ran("useradd", "--create-home", "--", "osapitest"))...,
		).
		Example("Add a user with a shell, groups and a password",
			CreateInput{
				Name: "osapitest", Shell: "/bin/sh", Groups: []string{"root"},
				Password: "osapi-test-password", PasswordHash: testHash,
			},
			Result{Name: "osapitest", Changed: true},
			withHost(
				behavior.ExistsAfter(
					"useradd --create-home -s /bin/sh -G root -- osapitest", "/etc/passwd",
					etcPasswd+"osapitest:x:1001:1001::/home/osapitest:/bin/sh\n"),
				behavior.Ran("useradd", "--create-home", "-s", "/bin/sh", "-G", "root", "--", "osapitest"),
				behavior.Ran("chpasswd", "-e"),
			)...,
		).
		Example("Add a system user with its IDs and its home",
			CreateInput{
				Name:   "osapitest",
				UID:    1999,
				GID:    100,
				Home:   "/home/osapitest",
				System: true,
			},
			Result{Name: "osapitest", Changed: true},
			withHost(
				behavior.ExistsAfter(
					"useradd --create-home -u 1999 -g 100 -d /home/osapitest -r -- osapitest",
					"/etc/passwd", etcPasswd+"osapitest:x:1999:100::/home/osapitest:/bin/sh\n"),
				behavior.Ran("useradd", "--create-home", "-u", "1999", "-g", "100",
					"-d", "/home/osapitest", "-r", "--", "osapitest"),
			)...,
		).
		Example("A user that the host has", CreateInput{Name: "osapitest"},
			Result{Name: "osapitest"},
			withHost(newUser, behavior.After(CreateInput{Name: "osapitest"}))...,
		).
		Fails("A name that is an option", CreateInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			CreateInput{Name: strings.Repeat("a", 33)}, "max").
		Fails("A negative user ID", CreateInput{Name: "osapitest", UID: -1}, "min").
		Fails("A negative group ID", CreateInput{Name: "osapitest", GID: -1}, "min").
		Fails("A home that is not absolute", CreateInput{Name: "osapitest", Home: "home"}, "startswith").
		Fails("A home with a line break", CreateInput{Name: "osapitest", Home: "/home/a\nb"}, "no_linebreak").
		Fails("A shell that is not absolute", CreateInput{Name: "osapitest", Shell: "bash"}, "startswith").
		Fails("A shell with a line break", CreateInput{Name: "osapitest", Shell: "/bin/a\nb"}, "no_linebreak").
		Fails("A group that is an option",
			CreateInput{Name: "osapitest", Groups: []string{"-r"}}, "account_name").
		Fails("A group longer than 32 characters",
			CreateInput{Name: "osapitest", Groups: []string{strings.Repeat("a", 33)}}, "max").
		Fails("A password with a line break",
			CreateInput{Name: "osapitest", Password: "a\nb"}, "no_linebreak").
		Fails("A hash that is not a crypt hash",
			CreateInput{Name: "osapitest", PasswordHash: "plain"}, "crypt hash",
			withHost()...),
	behavior.Write("update", Update).Key("name").
		AdminOnly().
		On(behavior.PlatformDebian).
		Sudo("usermod").
		Describe("Change a user").
		Explain("Changes what the input gives and the user does not have already, with "+
			"usermod: its shell, its home, which usermod moves the files to, its groups, "+
			"which replace the groups it has, and whether it is locked. The same values "+
			"report changed: false. A user that the host does not have is not found. It "+
			"runs on the Debian family only, and elsewhere reports skipped.").
		Example("Set the shell", UpdateInput{Name: "osapitest", Shell: "/bin/bash"},
			Result{Name: "osapitest", Changed: true},
			withHost(newUser,
				behavior.ExistsAfter("usermod -s /bin/bash -- osapitest", "/etc/passwd",
					etcPasswd+"osapitest:x:1001:1001::/home/osapitest:/bin/bash\n"),
				behavior.After(CreateInput{Name: "osapitest"}),
				behavior.Ran("usermod", "-s", "/bin/bash", "--", "osapitest"),
			)...,
		).
		Example("The shell it has", UpdateInput{Name: "osapitest", Shell: "/bin/bash"},
			Result{Name: "osapitest"},
			withHost(newUser,
				behavior.ExistsAfter("usermod -s /bin/bash -- osapitest", "/etc/passwd",
					etcPasswd+"osapitest:x:1001:1001::/home/osapitest:/bin/bash\n"),
				behavior.After(CreateInput{Name: "osapitest"}),
				behavior.After(UpdateInput{Name: "osapitest", Shell: "/bin/bash"}),
			)...,
		).
		Example("Move the home, set the groups and lock it",
			UpdateInput{
				Name:   "osapitest",
				Home:   "/home/osapitest2",
				Groups: []string{"root"},
				Lock:   "locked",
			},
			Result{Name: "osapitest", Changed: true},
			withHost(append(
				withPassword(),
				behavior.ExistsAfter("usermod -d /home/osapitest2 -m -G root -L -- osapitest", "/etc/passwd",
					etcPasswd+"osapitest:x:1001:1001::/home/osapitest2:/bin/sh\n"),
				behavior.ExistsAfter("usermod -d /home/osapitest2 -m -G root -L -- osapitest", "/etc/group",
					"root:x:0:osapitest\nalice:x:1000:\ndevelopers:x:1500:alice\n"),
				behavior.ExistsAfter("usermod -d /home/osapitest2 -m -G root -L -- osapitest", "/etc/shadow",
					etcShadow+shadowLocked),
				behavior.Ran("usermod", "-d", "/home/osapitest2", "-m", "-G", "root", "-L", "--", "osapitest"),
			)...)...,
		).
		Example("Unlock it", UpdateInput{Name: "osapitest", Lock: "unlocked"},
			Result{Name: "osapitest", Changed: true},
			withHost(append(
				withPassword(),
				behavior.ExistsAfter("usermod -L -- osapitest", "/etc/shadow", etcShadow+shadowLocked),
				behavior.ExistsAfter("usermod -U -- osapitest", "/etc/shadow", etcShadow+shadowSet),
				behavior.After(UpdateInput{Name: "osapitest", Lock: "locked"}),
				behavior.Ran("usermod", "-U", "--", "osapitest"),
			)...)...,
		).
		Fails("A user that the host does not have", UpdateInput{Name: "nobody"}, "not found",
			withHost()...).
		Fails("A name that is an option", UpdateInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			UpdateInput{Name: strings.Repeat("a", 33)}, "max").
		Fails("A shell that is not absolute", UpdateInput{Name: "alice", Shell: "bash"}, "startswith").
		Fails("A shell with a line break", UpdateInput{Name: "alice", Shell: "/bin/a\nb"}, "no_linebreak").
		Fails("A home that is not absolute", UpdateInput{Name: "alice", Home: "home"}, "startswith").
		Fails("A home with a line break", UpdateInput{Name: "alice", Home: "/home/a\nb"}, "no_linebreak").
		Fails("A group that is an option",
			UpdateInput{Name: "alice", Groups: []string{"-r"}}, "account_name").
		Fails("A group longer than 32 characters",
			UpdateInput{Name: "alice", Groups: []string{strings.Repeat("a", 33)}}, "max").
		Fails("A lock that is not locked or unlocked",
			UpdateInput{Name: "alice", Lock: "maybe"}, "enum"),
	behavior.Write("delete", Delete).Key("name").
		AdminOnly().
		On(behavior.PlatformDebian).
		Sudo("userdel").
		Describe("Remove a user").
		Explain("Removes the user and its home directory with userdel -r. A user that "+
			"the host does not have does not change and reports changed: false. It runs "+
			"on the Debian family only, and elsewhere reports skipped.").
		Example("Remove a user", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest", Changed: true},
			withHost(newUser,
				behavior.ExistsAfter("userdel -r -- osapitest", "/etc/passwd", etcPasswd),
				behavior.After(CreateInput{Name: "osapitest"}),
				behavior.Ran("userdel", "-r", "--", "osapitest"),
			)...,
		).
		Example("A user that the host does not have", DeleteInput{Name: "osapitest"},
			Result{Name: "osapitest"},
			withHost()...,
		).
		Fails("A name that is an option", DeleteInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			DeleteInput{Name: strings.Repeat("a", 33)}, "max"),
))
