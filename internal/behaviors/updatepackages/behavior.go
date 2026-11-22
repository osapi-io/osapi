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

package updatepackages

import "github.com/osapi-io/osapi/internal/behavior"

// Behavior is the updates the package index has for the host. They are not
// facts about the host but what its repositories have, so apt reads them,
// not gohai. It runs on the Debian family.
var Behavior = behavior.Register(behavior.Define(
	"updatepackages", "The updates the package index has for the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaSoftware),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Read("list", List).
		NotAFact("the updates are what the repositories of the package index have, not a fact of the host").
		Runs("apt").
		Describe("List the packages the index has newer versions of").
		Explain("Runs apt list --upgradable: each package with its installed "+
			"version and the version the index has. Refresh the index first for "+
			"what the repositories have now.").
		Example("Two updates", ListInput{}, []Available{
			{Name: "curl", CurrentVersion: "8.5.0-2ubuntu10", NewVersion: "8.5.0-2ubuntu10.1"},
			{Name: "openssl", CurrentVersion: "3.0.13-0ubuntu3", NewVersion: "3.0.13-0ubuntu3.1"},
		},
			behavior.Prints("apt list --upgradable", "Listing...\n"+
				"curl/noble-updates 8.5.0-2ubuntu10.1 amd64 [upgradable from: 8.5.0-2ubuntu10]\n"+
				"openssl/noble-updates 3.0.13-0ubuntu3.1 amd64 [upgradable from: 3.0.13-0ubuntu3]\n"),
		),
	behavior.Write("update", Update).
		Sudo("apt-get").
		Describe("Refresh the package index").
		Explain("Runs apt-get update. It reports changed: true when the Release file "+
			"of a repository under /var/lib/apt/lists changed, and changed: false when "+
			"each repository was unchanged.").
		// The integration test runs this on CI's runner, whose index the
		// workflow has just refreshed. A refresh that gets new data is in
		// debian_public_test.go: what the mirrors have is not an example.
		// After refreshes first, as a live host can run it: the runner's index
		// can be out of date, and then its first refresh gets new data.
		Example("The repositories are unchanged since the last refresh", UpdateInput{}, Result{},
			behavior.After(UpdateInput{}),
			behavior.Prints("apt-get update", "Hit:1 http://archive.ubuntu.com/ubuntu noble-updates InRelease\n"),
		),
))
