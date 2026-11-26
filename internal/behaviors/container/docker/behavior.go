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

// Package docker is the docker behavior: the Docker containers of the host.
// It reads them through gohai's docker collector, and changes them with the
// docker CLI, on each platform where docker answers.
package docker

import (
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The examples' container. The integration test runs the write examples on
// a live host, each from its container removed first: osapitest is a name
// that nothing else on the host has, and alpine is small.
const (
	name  = "osapitest"
	image = "alpine:3.20"
	id    = "3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7"
)

// The commands that gohai's docker collector runs, and the docker CLI calls
// of the writes.
const (
	versionCmd = "docker version --format {{.Server.Version}}"
	psCmd      = "docker ps -a --no-trunc --format {{json .}}"
	inspectCmd = "docker inspect " + id
	createCmd  = "docker create --name osapitest alpine:3.20 sleep 300"
)

// ps is what docker ps prints for the container in state.
func ps(
	state string,
	summary string,
) string {
	return `{"ID":"` + id + `","Names":"osapitest","Image":"alpine:3.20","State":"` + state +
		`","Status":"` + summary + `"}` + "\n"
}

// inspected is what docker inspect prints for the container.
const inspected = `[{"Id":"` + id + `","Created":"2026-11-22T10:00:00Z","State":{},` +
	`"Config":{"Env":["PATH=/usr/local/bin:/usr/bin","APP_TOKEN=x"]},` +
	`"Mounts":[{"Type":"bind","Source":"/srv/data","Destination":"/data","RW":false}],` +
	`"NetworkSettings":{"Ports":{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"8080"}]},` +
	`"Networks":{"bridge":{"IPAddress":"172.17.0.2","Gateway":"172.17.0.1"}}}}]`

// The host's docker, and the container's state on it.
var (
	answers = behavior.Prints(versionCmd, "29.1.3\n")
	details = behavior.Prints(inspectCmd, inspected)
	running = behavior.Prints(psCmd, ps("running", "Up 2 minutes"))
	made    = behavior.Prints(createCmd, id+"\n")
	created = behavior.PrintsAfter(createCmd, psCmd, ps("created", "Created"))
	started = behavior.PrintsAfter("docker start osapitest", psCmd, ps("running", "Up 1 second"))
	stopped = behavior.PrintsAfter(
		"docker stop -t 1 osapitest",
		psCmd,
		ps("exited", "Exited (143) 1 second ago"),
	)
	removed  = behavior.PrintsAfter("docker rm -f osapitest", psCmd, "")
	toCreate = CreateInput{Name: name, Image: image, Command: []string{"sleep", "300"}}
	toRun    = CreateInput{
		Name:      name,
		Image:     image,
		Command:   []string{"sleep", "300"},
		AutoStart: true,
	}
)

// web is the container that the read examples list.
var web = Entry{
	ID: id, Name: name, Image: image, State: "running", Summary: "Up 2 minutes",
	Created: "2026-11-22T10:00:00Z", Ports: []string{"0.0.0.0:8080->80/tcp"},
	Mounts: []string{"/srv/data:/data:ro"}, Networks: []string{"bridge 172.17.0.2 via 172.17.0.1"},
	EnvNames: []string{"PATH", "APP_TOKEN"},
}

// Behavior is the docker definition.
var Behavior = behavior.Register(behavior.Define(
	"docker", "The Docker containers of the host",
	behavior.Title("Container"),
	behavior.Node,
	behavior.API,
	behavior.Area(area.Containers),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newCLI),
	behavior.Platform(behavior.PlatformDarwin, newCLI),
	behavior.Sudo("docker"),
	behavior.Runs("docker"),
	behavior.Read("list", List).
		Describe("List the containers of the host").
		Explain("Reads each container through gohai's docker collector, which runs docker ps "+
			"and docker inspect. state limits the list to the running or the stopped ones. "+
			"A host whose docker does not answer reports skipped. The environment of a "+
			"container shows only the names of its variables: osapi never reads their values.").
		Example("A host with a container", ListInput{}, []Entry{web},
			answers, running, details,
		).
		Example("Only the stopped ones", ListInput{State: "stopped"}, []Entry{},
			answers, running, details,
		).
		Fails("A state that is not one", ListInput{State: "paused"}, "enum"),
	behavior.Read("get", Get).Key("name").
		Describe("Get one container").
		Explain("Reads one container by its name, its ID or the start of its ID. A container "+
			"that the host does not have is not found.").
		Example("A container", GetInput{Name: name}, web, answers, running, details).
		Fails("A container that the host does not have", GetInput{Name: "nothere"}, "not found").
		Fails("A name that is an option", GetInput{Name: "-f"}, "container_id"),
	behavior.Write("create", Create).
		Describe("Create a container").
		Explain("Runs docker create, which pulls the image when the host does not have it. "+
			"The environment goes on docker's standard input, not in its arguments, which "+
			"each user of the host can read. auto_start starts the container once it is made. "+
			"A container of the name that the host has already is not made again, and "+
			"reports changed: false.").
		Example("Create a container", toCreate, Result{Name: name, ID: id, Changed: true},
			answers, made, created, details,
			behavior.Ran("docker", "create", "--name", name, image, "sleep", "300"),
			behavior.Reads(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", Image: "alpine:3.20"}),
		).
		Example("A container that the host has", toCreate, Result{Name: name, ID: id},
			answers, made, created, details,
			behavior.After(toCreate),
			behavior.Reads(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", Image: "alpine:3.20"}),
		).
		Fails("A name that is an option", CreateInput{Name: "-f", Image: image}, "container_id").
		Fails("No image", CreateInput{Name: name}, "required").
		Fails("An image that is an option", CreateInput{Name: name, Image: "--privileged"}, "image_ref").
		Fails("A hostname that is an option",
			CreateInput{Name: name, Image: image, HostName: "-x"}, "dns_name").
		Fails("A hostname longer than 253 characters",
			CreateInput{Name: name, Image: image, HostName: strings.Repeat("a", 254)}, "max").
		Fails("A DNS server that is not an address",
			CreateInput{Name: name, Image: image, DNS: []string{"nameserver"}}, "ip").
		Fails("A port that is not host:container",
			CreateInput{Name: name, Image: image, Ports: []string{"80"}}, "port_mapping").
		Fails("A bind mount that is not two absolute paths",
			CreateInput{
				Name:    name,
				Image:   image,
				Volumes: []string{"data:/data"},
			}, "volume_mapping").
		Fails("No name", CreateInput{Image: image}, "required"),
	behavior.Write("delete", Delete).Key("name").
		Describe("Remove a container").
		Explain("Runs docker rm -f, which also removes a container that runs: a delete "+
			"leaves no container of the name. A container that the host does not have "+
			"reports changed: false.").
		Example("Remove a container", DeleteInput{Name: name}, Result{Name: name, Changed: true},
			answers, made, created, removed, details,
			behavior.After(toCreate),
			behavior.Ran("docker", "rm", "-f", name),
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Entry{Name: "osapitest"}, "name"),
		).
		Example("A container that the host does not have", DeleteInput{Name: name},
			Result{Name: name}, answers,
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Entry{Name: "osapitest"}, "name"),
		).
		Fails("A name that is an option", DeleteInput{Name: "-f"}, "container_id"),
	behavior.Action("start", Start).Key("name").
		Describe("Start a container").
		Explain("Runs docker start. A container that runs already reports changed: false, and "+
			"one that the host does not have is not found.").
		Example("Start a container", StartInput{Name: name}, Result{Name: name, Changed: true},
			answers, made, created, started, details,
			behavior.After(toCreate),
			behavior.Ran("docker", "start", name),
			behavior.Reads(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", State: "running"}),
		).
		Example("A container that runs already", StartInput{Name: name}, Result{Name: name},
			answers, made, created, started, details,
			behavior.After(toRun),
			behavior.Reads(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", State: "running"}),
		).
		Fails("A container that the host does not have", StartInput{Name: "nothere"}, "not found").
		Fails("A name that is an option", StartInput{Name: "-f"}, "container_id"),
	behavior.Action("stop", Stop).Key("name").
		Describe("Stop a container").
		Explain("Runs docker stop, which waits timeout seconds before docker kills the "+
			"container, or docker's default of 10 when it is 0. A container that does not "+
			"run reports changed: false.").
		Example("Stop a container", StopInput{Name: name, Timeout: 1}, Result{Name: name, Changed: true},
			answers, made, created, started, stopped, details,
			behavior.After(toRun),
			behavior.Ran("docker", "stop", "-t", "1", name),
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", State: "running"}, "name", "state"),
		).
		Example("A container that does not run", StopInput{Name: name, Timeout: 1}, Result{Name: name},
			answers, made, created, details,
			behavior.After(toCreate),
			behavior.ReadsNone(GetInput{Name: "osapitest"}, Entry{Name: "osapitest", State: "running"}, "name", "state"),
		).
		Fails("A container that the host does not have", StopInput{Name: "nothere"}, "not found").
		Fails("A name that is an option", StopInput{Name: "-f"}, "container_id").
		Fails("A timeout longer than 300 seconds", StopInput{Name: name, Timeout: 301}, "max").
		Fails("A negative timeout", StopInput{Name: name, Timeout: -1}, "min"),
	behavior.Action("exec", Exec).Key("name").
		ReadsNothing("the output of the command in the container is its result, which the call answers").
		PermitAs(types.VerbExec).
		Always().
		Describe("Run a command in a container").
		Explain("Runs docker exec, and answers what the command wrote to standard output and "+
			"to standard error, and its exit code. A command that exits with a code other "+
			"than 0 is an answer, not a failure. It changes what the command changes, which "+
			"osapi cannot tell, so it reports changed: true. Its permission is docker:exec, "+
			"which only the admin role gets, apart from the docker:execute of start and stop. An agent with agent.command.enabled false refuses it, as it "+
			"refuses a command.").
		Example("Run a command", ExecInput{Name: name, Command: []string{"echo", "hello"}},
			ExecResult{Stdout: "hello\n", Changed: true},
			answers, made, created, started, details,
			behavior.Prints("docker exec osapitest echo hello", "hello\n"),
			behavior.After(toRun),
			behavior.Ran("docker", "exec", name, "echo", "hello"),
		).
		Fails("An empty command", ExecInput{Name: name, Command: []string{}}, "min").
		Fails("No command", ExecInput{Name: name}, "required").
		Fails("A name that is an option", ExecInput{Name: "-f", Command: []string{"true"}}, "container_id").
		Fails("A working directory that is not absolute",
			ExecInput{Name: name, Command: []string{"pwd"}, WorkingDir: "tmp"}, "startswith"),
))
