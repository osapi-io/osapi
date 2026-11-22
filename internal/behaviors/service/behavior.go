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

package service

import (
	"context"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// The commands that gohai runs to read the services, and what they print on
// the examples' host.
const (
	units = "systemctl list-units --type=service --all --no-pager --plain"
	files = "systemctl list-unit-files --type=service --no-legend --no-pager"

	sshUnit  = "ssh.service loaded active running OpenBSD Secure Shell server\n"
	sshFile  = "ssh.service enabled enabled\n"
	appDead  = "osapi-app.service loaded inactive dead Example app\n"
	appUp    = "osapi-app.service loaded active running Example app\n"
	appOff   = "osapi-app.service disabled enabled\n"
	appOn    = "osapi-app.service enabled enabled\n"
	unitText = "[Unit]\nDescription=Example app\n\n[Service]\nExecStart=/bin/sleep infinity\n\n" +
		"[Install]\nWantedBy=multi-user.target\n"
	// staticText has no [Install] section, so systemd cannot enable it.
	staticText = "[Unit]\nDescription=Example static\n\n[Service]\nExecStart=/bin/sleep infinity\n"
	appPath    = "/etc/systemd/system/osapi-app.service"
)

// ssh is the service of the examples that osapi did not deploy.
var ssh = Entry{
	Name: "ssh", State: "running", Enabled: true,
	Description: "OpenBSD Secure Shell server",
}

// appEntry is the service that the examples deploy, as a read gives it: by
// the name it was deployed with, and managed.
var appEntry = Entry{Name: "app", State: "dead", Description: "Example app", Managed: true}

// stored and deployed are the unit file of the examples: uploaded, then
// deployed with create, so osapi keeps a record of it.
var (
	stored   = behavior.Stored("app.service", unitText, client.ContentRaw)
	deployed = behavior.After(CreateInput{Name: "app", Object: "app.service"})
)

// agentUnit is a unit file with the prefix that osapi did not deploy: the
// agent's own.
var agentUnit = behavior.Exists("/etc/systemd/system/osapi-agent.service", unitText)

// Behavior is the systemd services of the host. gohai reads every service.
// osapi writes and acts on the services it deploys only: each has a unit file
// /etc/systemd/system/osapi-<name>.service that managedfile has a record of,
// and whose content is an object that a person uploads once with osapi client
// file upload. A unit file with the prefix that osapi did not write, such as
// the agent's own osapi-agent.service, is not osapi's to change.
var Behavior = behavior.Register(behavior.Define(
	"service", "The systemd services of the host",
	behavior.Node,
	behavior.API,
	behavior.Area(behavior.AreaServices),
	behavior.CLI,
	behavior.Requires(object.Behavior),
	behavior.Requires(managedfile.Behavior),
	behavior.Platform(behavior.PlatformDebian, newDebian),
	behavior.Platform(behavior.PlatformLinux, newReader),
	behavior.Writes("/etc/systemd/system/osapi-*.service"),
	behavior.Runs("systemctl"),
	behavior.Read("list", List).
		Describe("List the services of the host").
		Explain("Lists each service that systemd has loaded, with its state, whether it "+
			"starts at boot and its description, through gohai. A service that osapi "+
			"deployed shows by the name it was deployed with, not its unit's osapi- "+
			"name, and is managed. A Mac has no systemd, and reports skipped.").
		Example("A service that osapi deployed reads by its name", ListInput{},
			[]Entry{ssh, appEntry},
			// Only the Debian writer deploys a service.
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, deployed,
			behavior.Prints(units, sshUnit+appDead),
			behavior.Prints(files, sshFile+appOff),
		).
		Example("A host with two services", ListInput{},
			[]Entry{ssh, {Name: "osapi-app", State: "dead", Description: "Example app"}},
			behavior.Prints(units, sshUnit+appDead),
			behavior.Prints(files, sshFile+appOff),
		),
	behavior.Read("get", Get).Key("name").
		Describe("Get one service").
		Explain("Reads one service from the same list, by its name with .service or "+
			"without it. A service that osapi deployed reads by the name it was deployed "+
			"with, such as app, as its writes take it, or by its unit's name, osapi-app. "+
			"A service that systemd has not loaded is not found.").
		Example("A running service", GetInput{Name: "ssh"}, ssh,
			behavior.Prints(units, sshUnit),
			behavior.Prints(files, sshFile),
		).
		Example("A name with .service", GetInput{Name: "ssh.service"}, ssh,
			behavior.Prints(units, sshUnit),
			behavior.Prints(files, sshFile),
		).
		Example("A service that osapi deployed, by the name it was deployed with",
			GetInput{Name: "app"}, appEntry,
			// Only the Debian writer deploys a service.
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, deployed,
			behavior.Prints(units, appDead),
			behavior.Prints(files, appOff),
		).
		Example("A service that osapi deployed, by its unit's name",
			GetInput{Name: "osapi-app"}, appEntry,
			// Only the Debian writer deploys a service.
			behavior.OnPlatform(behavior.PlatformDebian),
			stored, deployed,
			behavior.Prints(units, appDead),
			behavior.Prints(files, appOff),
		).
		Fails("A service that systemd has not loaded", GetInput{Name: "nosuch"}, "not found").
		Fails("A name with a space", GetInput{Name: "a b"}, "Name"),
	behavior.Write("create", Create).
		On(behavior.PlatformDebian).
		Sudo("systemctl").
		Describe("Deploy a service").
		Explain("Gets the object from the object store and writes it as the unit file "+
			"of the service, such as /etc/systemd/system/osapi-app.service for app, then reloads systemd. "+
			"A unit file at that path that osapi did not write, such as the agent's own "+
			"osapi-agent.service, is refused. A unit file "+
			"that already has the content does not change. The service does not start: "+
			"start and enable are their own actions.").
		Example("Deploy a unit file", CreateInput{Name: "app", Object: "app.service"},
			Result{Name: "app", Changed: true},
			behavior.Stored("app.service", unitText, client.ContentRaw),
			behavior.Wrote(appPath, unitText),
			behavior.Ran("systemctl", "daemon-reload"),
		).
		Example("A unit file that has the content", CreateInput{Name: "app", Object: "app.service"},
			Result{Name: "app"},
			behavior.Stored("app.service", unitText, client.ContentRaw),
			behavior.After(CreateInput{Name: "app", Object: "app.service"}),
		).
		Fails("An object the store does not have", CreateInput{Name: "app", Object: "nosuch"}, "not found").
		Fails("A unit file that osapi did not deploy", CreateInput{Name: "agent", Object: "app.service"},
			"osapi did not deploy it", stored, agentUnit).
		Fails("No object", CreateInput{Name: "app"}, "Object").
		Fails("An object name over 255 characters",
			CreateInput{Name: "app", Object: strings.Repeat("a", 256)}, "Object").
		Fails("A name with a space", CreateInput{Name: "a b", Object: "app.service"}, "Name").
		Fails("A name that is an option", CreateInput{Name: "--now", Object: "app.service"},
			"not a valid service name"),
	behavior.Write("update", Update).Key("name").
		On(behavior.PlatformDebian).
		Sudo("systemctl").
		Describe("Deploy a new unit file for a service").
		Explain("Writes the object as the unit file of a service that osapi deployed, "+
			"then reloads systemd when the file changed. A service that osapi did not "+
			"deploy is not found.").
		Example("A new unit file", UpdateInput{Name: "app", Object: "app-v2.service"},
			Result{Name: "app", Changed: true},
			stored,
			deployed,
			behavior.Stored("app-v2.service", unitText+"Restart=always\n", client.ContentRaw),
			behavior.Wrote(appPath, unitText+"Restart=always\n"),
			behavior.Ran("systemctl", "daemon-reload"),
		).
		Fails("A service that osapi did not deploy", UpdateInput{Name: "ssh", Object: "app.service"}, "not found").
		Fails("A unit file that osapi did not deploy", UpdateInput{Name: "agent", Object: "app.service"},
			"not found", stored, agentUnit).
		Fails("No object", UpdateInput{Name: "app"}, "Object").
		Fails("An object name over 255 characters",
			UpdateInput{Name: "app", Object: strings.Repeat("a", 256)}, "Object").
		Fails("A name with a space", UpdateInput{Name: "a b", Object: "app.service"}, "Name").
		Fails("A name that is an option", UpdateInput{Name: "--now", Object: "app.service"},
			"not a valid service name"),
	behavior.Write("delete", Delete).Key("name").
		On(behavior.PlatformDebian).
		Sudo("systemctl").
		Describe("Remove a service that osapi deployed").
		Explain("Stops and disables the service, removes its unit file and reloads "+
			"systemd. A service with no unit file does not change.").
		Example("Remove a service", DeleteInput{Name: "app"},
			Result{Name: "app", Changed: true},
			behavior.Stored("app.service", unitText, client.ContentRaw),
			behavior.After(CreateInput{Name: "app", Object: "app.service"}),
			behavior.Removed(appPath),
			behavior.Ran("systemctl", "disable", "--now", "osapi-app.service"),
			behavior.Ran("systemctl", "daemon-reload"),
		).
		Example("A service that osapi did not deploy", DeleteInput{Name: "app"},
			Result{Name: "app"},
		).
		Example("A unit file that osapi did not deploy", DeleteInput{Name: "agent"},
			Result{Name: "agent"},
			agentUnit,
		).
		Fails("A name with a space", DeleteInput{Name: "a b"}, "Name").
		Fails("A name that is an option", DeleteInput{Name: "--now"}, "not a valid service name"),
	action(
		"start",
		Start,
		"Start a service that osapi deployed",
		"Starts the service. A service that runs does not change.",
		StartInput{
			Name: "app",
		},
		appDead,
		appOff,
		appUp,
		appOff,
		StartInput{Name: "ssh"},
		StartInput{Name: "agent"},
		StartInput{Name: "a b"},
		StartInput{Name: "--now"},
	),
	action(
		"stop",
		Stop,
		"Stop a service that osapi deployed",
		"Stops the service. A service that does not run does not change.",
		StopInput{
			Name: "app",
		},
		appUp,
		appOff,
		appDead,
		appOff,
		StopInput{Name: "ssh"},
		StopInput{Name: "agent"},
		StopInput{Name: "a b"},
		StopInput{Name: "--now"},
	),
	action(
		"restart",
		Restart,
		"Restart a service that osapi deployed",
		"Restarts the service. It always changes the host.",
		RestartInput{
			Name: "app",
		},
		appUp,
		appOn,
		appUp,
		appOn,
		RestartInput{Name: "ssh"},
		RestartInput{Name: "agent"},
		RestartInput{Name: "a b"},
		RestartInput{Name: "--now"},
	),
	action(
		"enable",
		Enable,
		"Make a service that osapi deployed start at boot",
		"Enables the service. A service that is enabled does not change, and nor does a unit "+
			"with no [Install] section, which systemd calls static: it cannot start at boot.",
		EnableInput{
			Name: "app",
		},
		appDead,
		appOff,
		appDead,
		appOn,
		EnableInput{Name: "ssh"},
		EnableInput{Name: "agent"},
		EnableInput{Name: "a b"},
		EnableInput{Name: "--now"},
	).
		Example("A unit with no [Install] section", EnableInput{Name: "static"},
			Result{Name: "static"},
			behavior.Stored("static.service", staticText, client.ContentRaw),
			behavior.After(CreateInput{Name: "static", Object: "static.service"}),
			behavior.Prints(units, "osapi-static.service loaded inactive dead Example static\n"),
			behavior.Prints(files, "osapi-static.service static -\n"),
			behavior.Ran("systemctl", "enable", "osapi-static.service"),
		),
	action(
		"disable",
		Disable,
		"Make a service that osapi deployed not start at boot",
		"Disables the service. A service that is not enabled does not change.",
		DisableInput{
			Name: "app",
		},
		appDead,
		appOn,
		appDead,
		appOff,
		DisableInput{Name: "ssh"},
		DisableInput{Name: "agent"},
		DisableInput{Name: "a b"},
		DisableInput{Name: "--now"},
	),
))

// action is one action on a service that osapi deployed. Its example runs on
// a host where systemd prints unit and file, and unitAfter and fileAfter once
// the action ran.
func action[In any](
	name string,
	fn func(context.Context, Provider, In) (Result, error),
	describe string,
	explain string,
	in In,
	unit string,
	file string,
	unitAfter string,
	fileAfter string,
	missing In,
	agent In,
	slash In,
	option In,
) *behavior.Method {
	ran := "systemctl " + name + " osapi-app.service"
	// A live host starts each example from a unit that create deployed: dead
	// and disabled. stop and disable need it running or enabled first.
	host := []behavior.HostOption{stored, deployed}
	switch name {
	case "stop":
		host = append(host, behavior.After(StartInput{Name: "app"}))
	case "disable":
		host = append(host, behavior.After(EnableInput{Name: "app"}))
	}
	host = append(host,
		behavior.Prints(units, unit),
		behavior.Prints(files, file),
		behavior.PrintsAfter(ran, units, unitAfter),
		behavior.PrintsAfter(ran, files, fileAfter),
		behavior.Ran("systemctl", name, "osapi-app.service"),
	)
	m := behavior.Action(name, fn).Key("name").
		On(behavior.PlatformDebian).
		Sudo("systemctl").
		Describe(describe).
		Explain(explain+" A service that osapi did not deploy is not found.").
		Example(strings.ToUpper(name[:1])+name[1:]+" the service", in,
			Result{Name: "app", Changed: true},
			host...,
		).
		Fails("A service that osapi did not deploy", missing, "not found").
		Fails("A unit file that osapi did not deploy", agent, "not found", agentUnit).
		Fails("A name with a space", slash, "Name").
		Fails("A name that is an option", option, "not a valid service name")
	if name == "restart" {
		m = m.Always()
	}

	return m
}
