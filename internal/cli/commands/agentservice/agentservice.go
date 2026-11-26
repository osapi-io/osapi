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

// Package agentservice installs the agent of this host as a service that
// starts at boot: a systemd unit on Linux, a launch daemon on macOS, with a
// config that has the join string.
package agentservice

import (
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is osapi agent service.
var Group = command.Register(command.Define(
	"service", "Run the agent of this host as a service that starts at boot",
	command.Under(command.ParentAgent),
	command.Uses[Deps](),
	command.Method[InstallInput, InstallOutput]("install").
		Short("Install the agent as a service, with a join string, and start it").
		Long("Run this as root on the host. It writes the config file with the join\n"+
			"string and agent.pki.key_dir, a systemd unit (Linux) or a launch daemon\n"+
			"(macOS) that runs this binary as osapi agent start, and starts it. Then\n"+
			"it waits up to a minute for the join. After the join, the host keeps its\n"+
			"keys and its NATS user in agent.pki.key_dir, and each start at boot\n"+
			"reads them: the join string is used one time.").
		Example("--join 'tls://nats.example.com:4222?creds=...&fp=SHA256%3A...&ns=osapi&token=osapi_jt_...'").
		Prints(sampleInstall).
		Fails("no join string", InstallInput{}).
		Fails("a join string of more than 8192 characters", InstallInput{
			Join: "tls://" + strings.Repeat("a", 8193),
		}).
		Handle(Install),
	command.Method[UninstallInput, UninstallOutput]("uninstall").
		Short("Stop the agent's service and remove it").
		Long("Run this as root on the host. It stops the service and removes its unit\n"+
			"or launch daemon. With --purge it also removes the config file and the\n"+
			"keys in agent.pki.key_dir: then the host is new to osapi, and joins again\n"+
			"only with a new join string.").
		Example("--purge").
		Prints(sampleUninstall).
		Handle(Uninstall),
))
