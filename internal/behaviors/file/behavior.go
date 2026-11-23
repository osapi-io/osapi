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

// Package file is the file behavior: it deploys the objects of the object
// store to a host, through the object and managedfile building blocks, and
// reads how a deployed file is against what osapi wrote. The controller's
// file endpoints upload, list and delete the objects.
package file

import (
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// The objects and the path of the examples. The integration test runs the
// write examples as root on CI's runner; nothing there has this path.
const (
	path     = "/etc/osapi/file-example.conf"
	raw      = "listen 8080\n"
	tmpl     = "host {{ .Hostname }} port {{ .Vars.port }}\n"
	rendered = "host test-host port 8080\n"
	rawSHA   = "0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60"
	tmplSHA  = "28ef41e1ff26302a25da25eaa7b2762f2d55d0fdc0631a4f6a8c380d2b8640d4"
)

var (
	stored   = behavior.Stored("example.conf", raw, client.ContentRaw)
	template = behavior.Stored("example.conf.tmpl", tmpl, client.ContentTemplate)
	deployed = DeployInput{ObjectName: "example.conf", Path: path}
)

// Behavior is the file definition.
var Behavior = behavior.Register(behavior.Define(
	"file", "Files of the object store, deployed to the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaConfig),
	behavior.CLI,
	behavior.Requires(object.Behavior),
	behavior.Requires(managedfile.Behavior),
	behavior.Platform(behavior.PlatformLinux, newLinux),
	// It writes the path its caller names.
	behavior.Writes("*"),
	behavior.Read("status", GetStatus).
		Describe("Show whether a deployed file holds what osapi wrote").
		Explain("Compares the file on the host with the record that osapi keeps of what it "+
			"deployed: in-sync when they agree, drifted when something changed the file, "+
			"and missing when osapi did not deploy the path, undeployed it, or the file is gone.").
		Example("A deployed file", StatusInput{Path: path},
			Status{Path: path, State: InSync, SHA256: rawSHA},
			stored,
			behavior.After(deployed),
		).
		Example("A path that osapi did not deploy", StatusInput{Path: path},
			Status{Path: path, State: Missing},
		).
		Fails("A relative path", StatusInput{Path: "etc/example.conf"}, "startswith"),
	behavior.Action("deploy", Deploy).
		Describe("Deploy an object of the object store to a path").
		Explain("Writes the object at the path with its mode and ownership, and keeps a record "+
			"of it. A template object, or content_type template, is rendered first with "+
			".Facts, .Vars and .Hostname. A file that holds the content, mode and owner "+
			"already does not change and reports changed: false. osapi client file stale "+
			"list shows the deployments whose object changed since.").
		Example("Deploy a file", deployed,
			DeployResult{Path: path, SHA256: rawSHA, Changed: true},
			stored,
			behavior.Wrote(path, raw),
		).
		Example("Render a template",
			DeployInput{
				ObjectName: "example.conf.tmpl",
				Path:       path,
				Vars:       map[string]string{"port": "8080"},
			},
			DeployResult{Path: path, SHA256: tmplSHA, Changed: true},
			template,
			behavior.Wrote(path, rendered),
		).
		Example("A file that holds the object", deployed,
			DeployResult{Path: path, SHA256: rawSHA},
			stored,
			behavior.After(deployed),
		).
		Fails("A relative path", DeployInput{ObjectName: "example.conf", Path: "etc/x"}, "startswith").
		Fails("No object", DeployInput{Path: path}, "required").
		Fails("A mode that is not octal",
			DeployInput{ObjectName: "example.conf", Path: path, Mode: "rw-r--r--"}, "file_mode").
		Fails("An object name longer than 255 characters",
			DeployInput{ObjectName: strings.Repeat("a", 256), Path: path}, "max").
		Fails("A group that is an option",
			DeployInput{ObjectName: "example.conf", Path: path, Group: "-r"}, "account_name").
		Fails("An owner that is an option",
			DeployInput{ObjectName: "example.conf", Path: path, Owner: "-r"}, "account_name").
		Fails("A content type that is neither",
			DeployInput{ObjectName: "example.conf", Path: path, ContentType: "binary"}, "oneof"),
	behavior.Action("undeploy", Undeploy).
		Describe("Remove a file that osapi deployed").
		Explain("Removes the file and keeps its record, marked undeployed, so status reports "+
			"missing. A path that osapi did not deploy, or undeployed already, does not "+
			"change and reports changed: false. The object stays in the object store.").
		Example("Remove a deployed file", UndeployInput{Path: path},
			UndeployResult{Path: path, Changed: true},
			stored,
			behavior.After(deployed),
			behavior.Removed(path),
		).
		Example("A path that osapi did not deploy", UndeployInput{Path: path},
			UndeployResult{Path: path},
		).
		Fails("A relative path", UndeployInput{Path: "etc/x"}, "startswith"),
))
