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

package sshkey

import (
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

// The examples' keys, made with ssh-keygen -t ed25519, and their
// fingerprints as ssh-keygen -l prints them.
const (
	key1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z osapitest@example"
	fp1  = "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk"
	key2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGLPeEZQDbhOrSIx6gISjxXIcREV3rwL517pX6m9G5jc other@example"
	fp2  = "SHA256:a5CeY6Lh3ZXkJOqGyuwb24kQeVjgag1gc9ujzv++nhE"
)

// authorizedKeys is the file of the examples' user.
const authorizedKeys = "/home/osapitest/.ssh/authorized_keys"

// osapitest is a host that has the user osapitest, and a live host where
// the user behavior adds it first.
var osapitest = []behavior.HostOption{
	behavior.Exists("/etc/passwd", "root:x:0:0:root:/root:/bin/bash\n"+
		"osapitest:x:1001:1001::/home/osapitest:/bin/sh\n"),
	behavior.Exists("/etc/group", "root:x:0:\nosapitest:x:1001:\n"),
	behavior.Exists("/etc/shadow", "root:*:19000:0:99999:7:::\nosapitest:!:19000:0:99999:7:::\n"),
}

// with is the host of osapitest and more, for a write: the user behavior
// adds osapitest first, as a live host needs. This host has it already.
func with(
	more ...behavior.HostOption,
) []behavior.HostOption {
	return append(
		append(slices.Clone(osapitest), behavior.After(user.CreateInput{Name: "osapitest"})),
		more...)
}

// chown is the call that gives the user its directory again.
var chown = behavior.Ran("chown", "-R", "1001:1001", "/home/osapitest/.ssh")

// Behavior is the authorized SSH keys of a user.
var Behavior = behavior.Register(behavior.Define(
	"sshkey", "The authorized SSH keys of a user",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaSecurity),
	behavior.CLI,
	behavior.Of(user.Behavior),
	behavior.Requires(user.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.Writes("/home/*/.ssh/authorized_keys"),
	behavior.Read("list", List).
		Describe("List the authorized SSH keys of a user").
		Explain("Reads ~/.ssh/authorized_keys of the user through gohai, with the type, the "+
			"SHA256 fingerprint, the comment and the options of each key. A user with no "+
			"file has no keys. A user that the host does not have is not found.").
		Example("A user with keys", ListInput{Name: "osapitest"},
			[]Key{
				{Type: "ssh-ed25519", Fingerprint: fp1, Comment: "osapitest@example"},
				{
					Type:        "ssh-ed25519",
					Fingerprint: fp2,
					Comment:     "other@example",
					Options:     `from="10.0.0.0/8"`,
				},
			},
			append(slices.Clone(osapitest), behavior.Exists(
				authorizedKeys,
				key1+"\n# a comment\n"+`from="10.0.0.0/8" `+key2+"\n",
			))...).
		Example("A user with no keys", ListInput{Name: "osapitest"}, []Key{}, osapitest...).
		Fails("A user that the host does not have", ListInput{Name: "nobody"}, "not found",
			osapitest...).
		Fails("A host key that gohai cannot read", ListInput{Name: "osapitest"}, "parse",
			append(slices.Clone(osapitest),
				behavior.Exists("/etc/ssh/ssh_host_ed25519_key.pub", "ssh-ed25519"))...).
		Fails("A name that is not a user name", ListInput{Name: "Bad;name"}, "Name").
		Fails("A name over 32 characters", ListInput{Name: strings.Repeat("a", 33)}, "max"),
	behavior.Write("create", Create).
		AdminOnly().
		On(behavior.PlatformDebian).
		Describe("Add an authorized SSH key to a user").
		Explain("Adds the key line to the end of ~/.ssh/authorized_keys of the user, and makes "+
			"the directory and the file if they are missing, with modes 700 and 600, owned by "+
			"the user. A key that the file has already, by its fingerprint, does not change "+
			"it. A line with no public key is refused. A user that the host does not have is "+
			"not found. It runs on the Debian family only, and elsewhere reports skipped.").
		Sudo("chown").
		Example("Add a key", CreateInput{Name: "osapitest", Key: key1},
			Result{Fingerprint: fp1, Changed: true},
			with(behavior.Wrote(authorizedKeys, key1+"\n"), chown)...).
		Example("Add a second key", CreateInput{Name: "osapitest", Key: key2},
			Result{Fingerprint: fp2, Changed: true},
			with(
				behavior.After(CreateInput{Name: "osapitest", Key: key1}),
				behavior.Wrote(authorizedKeys, key1+"\n"+key2+"\n"),
				chown,
			)...).
		Example("A key after a last line with no line break", CreateInput{Name: "osapitest", Key: key2},
			Result{Fingerprint: fp2, Changed: true},
			with(
				behavior.Exists(authorizedKeys, key1),
				behavior.Wrote(authorizedKeys, key1+"\n"+key2+"\n"),
				chown,
			)...).
		Example("A key the user has", CreateInput{Name: "osapitest", Key: key1},
			Result{Fingerprint: fp1},
			with(behavior.After(CreateInput{Name: "osapitest", Key: key1}))...).
		Fails("A line with no public key", CreateInput{Name: "osapitest", Key: "not a key"},
			"not an SSH public key", with()...).
		Fails("A key that is not base64", CreateInput{Name: "osapitest", Key: "ssh-ed25519 !!!"},
			"not an SSH public key", with()...).
		Fails("A user that the host does not have", CreateInput{Name: "nobody", Key: key1},
			"not found", with()...).
		Fails("A name that is not a user name", CreateInput{Name: "Bad;name", Key: key1}, "Name").
		Fails("A name over 32 characters", CreateInput{Name: strings.Repeat("a", 33), Key: key1}, "max").
		Fails("A key with a line break", CreateInput{Name: "osapitest", Key: key1 + "\n" + key2},
			"no_linebreak"),
	behavior.Write("delete", Delete).
		AdminOnly().
		On(behavior.PlatformDebian).
		Describe("Remove an authorized SSH key of a user").
		Explain("Removes each line of the key with the fingerprint from ~/.ssh/authorized_keys "+
			"of the user, and keeps the other lines as they are. A key that the file does not "+
			"have, a user with no file, or a user that the host does not have, changes "+
			"nothing. It runs on the Debian family only, and elsewhere reports "+
			"skipped.").
		Example("Remove a key", DeleteInput{Name: "osapitest", Fingerprint: fp1},
			Result{Fingerprint: fp1, Changed: true},
			with(
				behavior.After(CreateInput{Name: "osapitest", Key: key1}),
				behavior.After(CreateInput{Name: "osapitest", Key: key2}),
				behavior.Wrote(authorizedKeys, key2+"\n"),
			)...).
		Example("A key the user does not have", DeleteInput{Name: "osapitest", Fingerprint: fp2},
			Result{Fingerprint: fp2},
			with(behavior.After(CreateInput{Name: "osapitest", Key: key1}))...).
		Example("A user that the host does not have", DeleteInput{Name: "nobody", Fingerprint: fp1},
			Result{Fingerprint: fp1}, with()...).
		Fails("A fingerprint that is not SHA256", DeleteInput{Name: "osapitest", Fingerprint: "MD5:aa"},
			"Fingerprint").
		Fails("A fingerprint with a line break",
			DeleteInput{Name: "osapitest", Fingerprint: fp1 + "\n" + fp2}, "no_linebreak").
		Fails("A name that is not a user name", DeleteInput{Name: "Bad;name", Fingerprint: fp1}, "Name").
		Fails("A name over 32 characters",
			DeleteInput{Name: strings.Repeat("a", 33), Fingerprint: fp1}, "max"),
))
