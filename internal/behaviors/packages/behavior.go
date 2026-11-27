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

// Package packages is the package behavior: the packages installed on the
// host.
package packages

import (
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The queries gohai runs for the installed packages, as the fake host sees
// each command line.
const (
	dpkg = `dpkg-query -W -f=${Package}\t${Version}\t${Architecture}\t${db:Status-Status}\t${binary:Summary}\t${Installed-Size}\n`
	rpm  = `rpm -qa --qf %{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\t%{SUMMARY}\t%{SIZE}\n`
	brew = "brew list --versions"
)

// curl is the package of the examples.
var curl = Installed{
	Name: "curl", Version: "8.5.0-2ubuntu10", Architecture: "amd64", Manager: "dpkg",
	Description: "command line tool for transferring data with URL syntax", Size: 512000,
}

// tree is dpkg-query's line for tree, the package the write examples
// install and remove.
const tree = "tree\t2.1.1-2\tamd64\tinstalled\tdisplays an indented directory tree\t100\n"

// installed is dpkg-query's line for curl.
const installed = "curl\t8.5.0-2ubuntu10\tamd64\tinstalled\tcommand line tool for transferring data with URL syntax\t500\n"

// Behavior is the packages installed on the host. gohai reads them on every
// platform; apt-get installs and removes them on the Debian family.
var Behavior = behavior.Register(behavior.Define(
	"package", "The packages installed on the host",
	behavior.Node,
	behavior.API,
	behavior.Area(area.Software),
	behavior.CLI,
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Platform(behavior.PlatformDarwin, newReader),
	behavior.RunsOnPlatform(behavior.PlatformDebian, "dpkg-query"),
	behavior.RunsOnPlatform(behavior.PlatformDarwin, "brew"),
	behavior.RunsIfPresent("rpm"),
	behavior.Read("list", List).
		Describe("List the packages installed on the host").
		Explain("Lists each installed package with its version: dpkg on the Debian "+
			"family, rpm on the other Linux families, and Homebrew on macOS.").
		Example("A Debian host", ListInput{}, []Installed{curl},
			behavior.OnPlatform(behavior.PlatformDebian),
			behavior.Prints(dpkg, installed+"tree\t2.1.1-2\tamd64\tdeinstall\tdisplays an indented directory tree\t100\n"),
		).
		Example("A Fedora host", ListInput{}, []Installed{{
			Name: "curl", Version: "8.6.0-7.fc40", Architecture: "x86_64", Manager: "rpm",
			Description: "A utility for getting files from remote servers", Size: 790000,
		}},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Prints(rpm, "curl\t8.6.0-7.fc40\tx86_64\tA utility for getting files from remote servers\t790000\n"),
		).
		Example("A Mac", ListInput{}, []Installed{{Name: "jq", Version: "1.7.1", Manager: "brew"}},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints(brew, "jq 1.7 1.7.1\n"),
		),
	behavior.Read("get", Get).Key("name").
		Describe("Get one installed package").
		Explain("Reads one package from the same list. A package that is not "+
			"installed is not found.").
		Example("An installed package", GetInput{Name: "curl"}, curl,
			behavior.OnPlatform(behavior.PlatformDebian),
			behavior.Prints(dpkg, installed),
		).
		Example("An installed formula", GetInput{Name: "jq"},
			Installed{Name: "jq", Version: "1.7.1", Manager: "brew"},
			behavior.OnPlatform(behavior.PlatformDarwin),
			behavior.Prints(brew, "jq 1.7.1\n"),
		).
		Example("An installed RPM", GetInput{Name: "curl"},
			Installed{
				Name:         "curl",
				Version:      "8.6.0-7.fc40",
				Architecture: "x86_64",
				Manager:      "rpm",
			},
			behavior.OnPlatform(behavior.PlatformLinux),
			behavior.Prints(rpm, "curl\t8.6.0-7.fc40\tx86_64\n"),
		).
		Fails("A package that is not installed", GetInput{Name: "tree"}, "not found",
			behavior.NotLive("the suite installs tree on a live host, for the package examples")).
		Fails("A name that is an option", GetInput{Name: "-oAPT::Get=x"}, "not a valid package name").
		Fails("A name longer than 255 characters",
			GetInput{Name: strings.Repeat("a", 256)}, "not a valid package name"),
	behavior.Write("create", Create).
		On(behavior.PlatformDebian).
		Sudo("apt-get").
		Describe("Install a package").
		Explain("Installs the package with apt-get install, at the version when one is "+
			"given. A package that is installed, at that version when one is given, "+
			"does not change and reports changed: false. It runs on the Debian family "+
			"only, and elsewhere reports skipped.").
		// The integration test runs these on CI's runner as root, each from its
		// package removed first: tree is small, and nothing on the runner
		// needs it.
		Example("Install a package", CreateInput{Name: "tree"},
			Result{Name: "tree", Changed: true},
			behavior.PrintsAfter("apt-get install -y -o DPkg::Lock::Timeout=300 -- tree", dpkg, tree),
			behavior.Ran("apt-get", "install", "-y", "-o", "DPkg::Lock::Timeout=300", "--", "tree"),
			behavior.Reads(GetInput{Name: "tree"}, Installed{Name: "tree"}),
		).
		Fails("A name that is an option", CreateInput{Name: "-oAPT::Get=x"}, "not a valid package name").
		Fails("A name longer than 255 characters",
			CreateInput{Name: strings.Repeat("a", 256)}, "not a valid package name").
		Fails("A version that is an option", CreateInput{Name: "tree", Version: "-x"},
			"not a valid package version").
		Fails("A version longer than 255 characters",
			CreateInput{
				Name:    "tree",
				Version: strings.Repeat("1", 256),
			}, "not a valid package version"),
	behavior.Write("delete", Delete).Key("name").
		On(behavior.PlatformDebian).
		Sudo("apt-get").
		Describe("Remove a package").
		Explain("Removes the package with apt-get remove. A package that is not "+
			"installed does not change and reports changed: false. It runs on the "+
			"Debian family only, and elsewhere reports skipped.").
		Example("Remove a package", DeleteInput{Name: "tree"},
			Result{Name: "tree", Changed: true},
			behavior.After(CreateInput{Name: "tree"}),
			behavior.PrintsAfter("apt-get install -y -o DPkg::Lock::Timeout=300 -- tree", dpkg, tree),
			behavior.PrintsAfter("apt-get remove -y -o DPkg::Lock::Timeout=300 -- tree", dpkg, ""),
			behavior.Ran("apt-get", "remove", "-y", "-o", "DPkg::Lock::Timeout=300", "--", "tree"),
			behavior.ReadsNone(GetInput{Name: "tree"}, Installed{Name: "tree"}, "name"),
		).
		Example("A package that is not installed", DeleteInput{Name: "tree"},
			Result{Name: "tree"},
			behavior.ReadsNone(GetInput{Name: "tree"}, Installed{Name: "tree"}, "name"),
		).
		Fails("A name that is an option", DeleteInput{Name: "-oAPT::Get=x"}, "not a valid package name").
		Fails("A name longer than 255 characters",
			DeleteInput{Name: strings.Repeat("a", 256)}, "not a valid package name"),
))
