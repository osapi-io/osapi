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

// Package password is the password behavior: the password of a user, under
// the user. It sets the password on the Debian family. The controller
// hashes the password, so the job data and NATS have only its hash.
package password

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

// The files of the examples: a host with root and alice.
const (
	etcPasswd = "root:x:0:0:root:/root:/bin/bash\n" +
		"alice:x:1000:1000:Alice:/home/alice:/bin/bash\n"
	etcGroup  = "root:x:0:\nalice:x:1000:\n"
	etcShadow = "root:*:19000:0:99999:7:::\nalice:$6$s$h:19000:0:99999:7:::\n"
)

// testHash is a sha512-crypt hash of the examples' password, as the
// controller's step makes it. An example is the input that the agent gets,
// so it has the hash; its password is what the caller sends on a live host.
const testHash = "$6$osapitest$ixskR1vdsZl2JBlRPn/2rqKZK4DltHGVEnTxDn.wN68cQnKpsMNWzH5Hxu6jZgEHWjcnzvFneyk/KHl6bEYtj1"

// newUser is /etc/passwd once useradd has added osapitest.
var newUser = behavior.ExistsAfter(
	"useradd --create-home -- osapitest",
	"/etc/passwd",
	etcPasswd+"osapitest:x:1001:1001::/home/osapitest:/bin/sh\n",
)

// withHost is a host with root and alice, and opts.
func withHost(
	opts ...behavior.HostOption,
) []behavior.HostOption {
	return append([]behavior.HostOption{
		behavior.Exists("/etc/passwd", etcPasswd),
		behavior.Exists("/etc/group", etcGroup),
		behavior.Exists("/etc/shadow", etcShadow),
	}, opts...)
}

// Behavior is the password of a user, under the user:
// /api/node/{hostname}/user/{name}/password. Its permission is the user's
// user:write, which only the admin role gets, because an account gets to
// root.
var Behavior = behavior.Register(behavior.Define(
	"password", "The password of a user",
	behavior.Title("User password"),
	behavior.Node,
	behavior.API,
	behavior.Area(area.Security),
	behavior.CLI,
	behavior.Of(user.Behavior),
	behavior.Requires(user.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Write("update", Update).
		AdminOnly().
		Sudo("chpasswd").
		Always().
		Before(hashPassword).
		Describe("Set the password of a user").
		Explain("Sets the password with chpasswd -e, from its hash on standard input. "+
			"The controller hashes the password, so the job and NATS have only the "+
			"hash. Each call makes a new hash, so it always reports changed: true. A "+
			"user that the host does not have is not found. It runs on the Debian "+
			"family only, and elsewhere reports skipped.").
		Example("Set a password",
			UpdateInput{
				Name:         "osapitest",
				Password:     "osapi-test-password",
				PasswordHash: testHash,
			},
			Result{Name: "osapitest", Changed: true},
			withHost(newUser,
				behavior.After(user.CreateInput{Name: "osapitest"}),
				behavior.Ran("chpasswd", "-e"),
			)...,
		).
		Fails("No password", UpdateInput{Name: "alice"}, "give the password", withHost()...).
		Fails("A user that the host does not have",
			UpdateInput{Name: "nobody", PasswordHash: testHash}, "not found", withHost()...).
		Fails("A hash that is not a crypt hash",
			UpdateInput{Name: "alice", PasswordHash: "plain"}, "crypt hash", withHost()...).
		Fails("A name that is an option", UpdateInput{Name: "-r"}, "account_name").
		Fails("A name longer than 32 characters",
			UpdateInput{Name: strings.Repeat("a", 33)}, "max").
		Fails("A password with a line break",
			UpdateInput{Name: "alice", Password: "a\nb"}, "no_linebreak"),
))
