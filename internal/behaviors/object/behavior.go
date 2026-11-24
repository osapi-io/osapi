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

package object

import (
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/behavior"
)

// unit is the object of the examples: a systemd unit, uploaded raw.
const unit = "[Unit]\nDescription=Example\n\n[Service]\nExecStart=/usr/bin/example\n"

// Behavior reads the objects of the object store of the fleet: the files that
// a person uploads once with osapi client file upload, and that behaviors
// deploy by name. It is a building block: a behavior that deploys a file
// requires it, gets the content with it, and writes the content with
// managedfile.
var Behavior = behavior.Register(behavior.Define(
	"object",
	"The objects of the object store of the fleet, such as a file uploaded with osapi client file upload",
	behavior.Node,
	behavior.Internal,
	behavior.Platform(behavior.PlatformDebian, newStore),
	behavior.Platform(behavior.PlatformLinux, newStore),
	behavior.Platform(behavior.PlatformDarwin, newStore),
	behavior.Read("get", Get).Key("name").
		Describe("Get an object of the object store").
		Explain("Reads the object from the object store of the fleet, with the content type that "+
			"its upload gave it, raw or template, and the SHA-256 of its content. A template is "+
			"not rendered: the behavior that deploys it renders it. An agent with no object store "+
			"reports skipped.").
		Example("A unit file", GetInput{Name: "example.service"},
			Object{
				Name: "example.service", Content: unit, ContentType: "raw",
				SHA256: "4d38310b40344ae465b2f1f88758809f31ef1aa8287c756a1d65c76ca27dc243",
			},
			behavior.Stored("example.service", unit, types.ContentRaw),
		).
		Example("A template", GetInput{Name: "motd.tmpl"},
			Object{
				Name: "motd.tmpl", Content: "Welcome to {{ .Hostname }}\n", ContentType: "template",
				SHA256: "08c60853fd0933d9903a112a64bb4c742c03961a4e56af1940173403a09746e2",
			},
			behavior.Stored("motd.tmpl", "Welcome to {{ .Hostname }}\n", types.ContentTemplate),
		).
		Fails("An object the store does not have", GetInput{Name: "nosuch"}, "not found").
		Fails("No name", GetInput{}, "Name").
		Fails("A name over 255 characters", GetInput{Name: strings.Repeat("a", 256)}, "Name"),
))
