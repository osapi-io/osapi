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

package integration_test

import "time"

// liveWriteScenarios run the write methods that no example runs on a live
// host: a signal to a process that the scenario starts, a command on an
// agent that refuses commands, and a reboot of a lab VM.
func liveWriteScenarios() []scenario {
	return []scenario{
		{
			name: "a signal stops a process that a command started", writeOp: "PROCESS",
			steps: []step{
				{
					name: "the target is Linux",
					cli: []string{
						"client",
						"node",
						"os",
						"get",
						"--target",
						"{target}",
						"--json",
					},
					skipIf: &check{
						path: "results[0].os_info.distribution",
						op:   "==",
						want: "darwin",
					},
				},
				{
					name: "start a sleep in its own session",
					cli: []string{
						"client", "node", "command", "shell", "--target", "{target}", "--json",
						"--command", "setsid sleep 300 </dev/null >/dev/null 2>&1 & printf %s $!",
					},
					checks:  []check{{path: "results[0].stdout", op: "nonempty"}},
					capture: map[string]string{"pid": "results[0].stdout"},
				},
				{
					name: "the process runs",
					cli: []string{
						"client",
						"node",
						"process",
						"get",
						"--target",
						"{target}",
						"--pid",
						"{pid}",
						"--json",
					},
					checks: []check{{path: "results[0].name", op: "==", want: "sleep"}},
				},
				{
					name: "send it TERM",
					cli: []string{
						"client", "node", "process", "signal", "--target", "{target}",
						"--pid", "{pid}", "--signal", "TERM", "--json",
					},
					checks: []check{
						{path: "results[0].signal", op: "==", want: "TERM"},
						{path: "results[0].changed", op: "==", want: "true"},
					},
				},
				{
					name:  "the process is gone",
					fails: true,
					wait:  15 * time.Second,
					pause: time.Second,
					cli: []string{
						"client",
						"node",
						"process",
						"get",
						"--target",
						"{target}",
						"--pid",
						"{pid}",
						"--json",
					},
					checks: []check{{path: "!all", op: "contains", want: "not found"}},
				},
			},
			cleanup: []step{{
				name: "stop the sleep",
				cli: []string{
					"client", "node", "process", "signal", "--target", "{target}",
					"--pid", "{pid}", "--signal", "KILL", "--json",
				},
			}},
		},
		{
			name:             "an agent with commands off refuses a command, with the reason",
			writeOp:          "COMMAND",
			fixture:          true,
			fixtureAutoAdmit: true,
			fixtureAgent:     map[string]any{"command": map[string]any{"enabled": false}},
			steps: []step{
				{name: "start an agent", start: "agent", keyDir: "{dir:agent}"},
				ready("the agent is Ready"),
				{
					// A host's refusal is a row of the result, not the exit code.
					name: "command exec is refused",
					on:   onFixture,
					cli: []string{
						"client", "node", "command", "exec", "--target", "{host}",
						"--command", "true", "--json",
					},
					checks: []check{
						{path: "results[0].status", op: "==", want: "failed"},
						{
							path: "results[0].error",
							op:   "contains",
							want: "command execution is disabled",
						},
					},
				},
				{
					name: "command shell is refused",
					on:   onFixture,
					cli: []string{
						"client", "node", "command", "shell", "--target", "{host}",
						"--command", "true", "--json",
					},
					checks: []check{
						{path: "results[0].status", op: "==", want: "failed"},
						{
							path: "results[0].error",
							op:   "contains",
							want: "command execution is disabled",
						},
					},
				},
			},
		},
		{
			name: "a reboot of a lab VM brings its agent back on a new boot", writeOp: "POWER",
			labOnly: true,
			steps: []step{
				{
					name: "the target is a Linux VM",
					cli: []string{
						"client",
						"node",
						"os",
						"get",
						"--target",
						"{target}",
						"--json",
					},
					skipIf: &check{
						path: "results[0].os_info.distribution",
						op:   "==",
						want: "darwin",
					},
				},
				{
					name: "read the boot ID",
					cli: []string{
						"client", "node", "command", "shell", "--target", "{target}", "--json",
						"--command", "tr -d '\\n' < /proc/sys/kernel/random/boot_id",
					},
					checks:  []check{{path: "results[0].stdout", op: "nonempty"}},
					capture: map[string]string{"boot": "results[0].stdout"},
				},
				{
					name: "reboot in one minute",
					cli: []string{
						"client",
						"node",
						"power",
						"reboot",
						"--target",
						"{target}",
						"--json",
					},
					checks: []check{{path: "results[0].action", op: "==", want: "reboot"}},
				},
				{
					name:  "the agent answers from a new boot",
					wait:  6 * time.Minute,
					pause: 10 * time.Second,
					cli: []string{
						"client", "node", "command", "shell", "--target", "{target}", "--json",
						"--command", "tr -d '\\n' < /proc/sys/kernel/random/boot_id",
					},
					checks: []check{
						{path: "results[0].stdout", op: "nonempty"},
						{path: "results[0].stdout", op: "!=", want: "{boot}"},
					},
				},
			},
		},
	}
}
