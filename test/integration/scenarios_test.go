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

// This file has no build tag: the scenarios are data, which the unit gate
// of scenario_test.go checks against the command tree.

import "time"

// scenarios is each flow that the suite runs as a table of steps.
func scenarios() []scenario {
	out := joinScenarios()
	out = append(out, keyRotationScenarios()...)
	out = append(out, sealScenarios()...)

	out = append(out, agentServiceScenarios()...)
	out = append(out, cryptoScenarios()...)

	out = append(out, pkiScenarios()...)

	return append(out, jobScenarios()...)
}

// ready is a step that waits for a Ready agent in the fixture, and keeps
// its hostname and machine ID as host and mid.
func ready(
	name string,
) step {
	return step{
		name:   name,
		on:     onFixture,
		wait:   60 * time.Second,
		pause:  time.Second,
		cli:    []string{"client", "agent", "list", "--json"},
		checks: []check{{path: "items[state=Ready].hostname", op: "nonempty"}},
		capture: map[string]string{
			"host": "items[state=Ready].hostname",
			"mid":  "items[state=Ready].machine_id",
		},
	}
}

// admission is a step that waits until the fixture's admission of mid is
// in state.
func admission(
	state string,
) step {
	return step{
		name:   "the admission is " + state,
		on:     onFixture,
		wait:   30 * time.Second,
		pause:  time.Second,
		cli:    []string{"client", "admission", "get", "--machine-id", "{mid}", "--json"},
		checks: []check{{path: "state", op: "==", want: state}},
	}
}

// job is a step that runs a job on host, in the fixture. A job that may
// not reach host fails, and waits for that when wait is set.
func job(
	name string,
	fails bool,
	wait time.Duration,
) step {
	return step{
		name: name, on: onFixture, fails: fails, wait: wait, pause: time.Second,
		cli: []string{"client", "node", "hostname", "get", "--target", "{host}", "--json"},
	}
}

// joinToken is a step that makes a join token in the fixture, and keeps its
// join string as name and its ID as name+"ID".
func joinToken(
	name string,
	flags ...string,
) step {
	return step{
		name: "make the " + name + " token", on: onFixture,
		cli:     append([]string{"controller", "join", "token", "create", "--json"}, flags...),
		checks:  []check{{path: "join", op: "nonempty"}},
		capture: map[string]string{name: "join", name + "ID": "id"},
	}
}

// refused is a step that starts an agent with the join string of name in
// its own key directory: the controller must refuse it, so it stops.
func refused(
	name string,
	join string,
) step {
	return step{
		name: "a " + name + " token admits no agent", start: name, keyDir: "{dir:" + name + "}",
		join: join, exits: true, wait: 60 * time.Second,
	}
}

func joinScenarios() []scenario {
	return []scenario{{
		name: "join with a token, then each refusal and each admission change", writeOp: "JOIN",
		fixture: true,
		steps: []step{
			joinToken("join", "--uses", "1", "--expires", "10m", "--labels", "group=fixture"),
			{
				name:   "start the agent with the token",
				start:  "agent",
				keyDir: "{dir:token}",
				join:   "{join}",
			},
			ready("the agent joins"),
			{
				name: "the agent has the token's labels",
				on:   onFixture,
				cli:  []string{"client", "agent", "list", "--json"},
				checks: []check{
					{path: "items[hostname={host}].labels.group", op: "==", want: "fixture"},
				},
			},
			admission("admitted"),
			job("a job reaches the joined agent", false, 0),
			{
				name: "the join is in the audit log", on: onFixture,
				cli:    []string{"client", "audit", "list", "--limit", "100", "--json"},
				checks: []check{{path: "", op: "contains", want: `"agent.join"`}},
			},
			{name: "stop the agent", stop: "agent"},
			joinToken("revoked", "--uses", "1"),
			{
				name: "revoke the token",
				on:   onFixture,
				cli:  []string{"controller", "join", "token", "revoke", "--id", "{revokedID}"},
			},
			joinToken("expired", "--uses", "1", "--expires", "1s"),
			joinToken("valid", "--uses", "1"),
			{
				name:  "let the expired token expire",
				on:    onFixture,
				cli:   []string{"controller", "join", "token", "list", "--json"},
				pause: 2 * time.Second,
			},
			refused("used", "{join}"),
			refused("revoked", "{revoked}"),
			refused("expired", "{expired}"),
			refused("fingerprint", "{wrongfp:valid}"),
			admission("admitted"),
			{name: "start the agent again", start: "agent", keyDir: "{dir:token}"},
			ready("the agent is back"),
			{
				name: "revoke the admission",
				on:   onFixture,
				cli:  []string{"client", "admission", "revoke", "--machine-id", "{mid}"},
			},
			admission("revoked"),
			job("a job does not reach the revoked agent", true, 60*time.Second),
			{name: "stop the revoked agent", stop: "agent"},
			{
				name: "delete the admission",
				on:   onFixture,
				cli:  []string{"client", "admission", "delete", "--machine-id", "{mid}"},
			},
			{
				name:  "the admission is gone",
				on:    onFixture,
				fails: true,
				wait:  30 * time.Second,
				pause: time.Second,
				cli:   []string{"client", "admission", "get", "--machine-id", "{mid}", "--json"},
			},
			{name: "start an agent with no token", start: "pending", keyDir: "{dir:pending}"},
			admission("pending"),
			job("a job does not reach a pending agent", true, 0),
			{
				name:    "read the pending agent's fingerprint",
				on:      onFixture,
				keyDir:  "{dir:pending}",
				wait:    30 * time.Second,
				pause:   time.Second,
				cli:     []string{"pki", "key", "get", "--name", "agent", "--json"},
				checks:  []check{{path: "fingerprint", op: "prefix", want: "SHA256:"}},
				capture: map[string]string{"fp": "fingerprint"},
			},
			{
				name: "admit the agent",
				on:   onFixture,
				cli: []string{
					"client",
					"admission",
					"admit",
					"--machine-id",
					"{mid}",
					"--fingerprint",
					"{fp}",
				},
			},
			admission("admitted"),
			ready("the admitted agent is Ready"),
			job("a job reaches the admitted agent", false, 0),
			{name: "stop the admitted agent", stop: "pending"},
			{
				name: "delete the admission again",
				on:   onFixture,
				cli:  []string{"client", "admission", "delete", "--machine-id", "{mid}"},
			},
			{name: "start an agent to reject", start: "rejected", keyDir: "{dir:rejected}"},
			admission("pending"),
			{
				name: "reject the agent",
				on:   onFixture,
				cli: []string{
					"client",
					"admission",
					"reject",
					"--machine-id",
					"{mid}",
					"--reason",
					"integration test",
				},
			},
			admission("rejected"),
			job("a job does not reach a rejected agent", true, 0),
			{name: "stop the rejected agent", stop: "rejected"},
			{
				name: "the admission changes are in the audit log", on: onFixture,
				cli:    []string{"client", "audit", "list", "--limit", "100", "--json"},
				checks: []check{{path: "", op: "contains", want: "admission"}},
			},
		},
	}}
}

// agentKey is a step that reads the fingerprint of the agent's keys on its
// host, in keyDir, and keeps it as name.
func agentKey(
	name string,
	keyDir string,
	checks ...check,
) step {
	return step{
		name: "read the agent's fingerprint as " + name, on: onHost, keyDir: keyDir,
		cli:     []string{"pki", "key", "get", "--name", "agent", "--json"},
		checks:  append([]check{{path: "fingerprint", op: "prefix", want: "SHA256:"}}, checks...),
		capture: map[string]string{name: "fingerprint"},
	}
}

// admitted is a step whose checks read the fingerprint of each admitted
// machine.
func admitted(
	name string,
	checks ...check,
) step {
	return step{
		name:   name,
		cli:    []string{"client", "admission", "list", "--state", "admitted", "--json"},
		checks: checks,
	}
}

// controllerKey is a step that reads the controller's key and its rotation.
func controllerKey(
	name string,
	wait time.Duration,
	checks ...check,
) step {
	return step{
		name: name, wait: wait, pause: 2 * time.Second,
		cli: []string{"client", "pki", "key", "get", "--json"}, checks: checks,
	}
}

func keyRotationScenarios() []scenario {
	return []scenario{
		{
			name:      "rotate the agent's keys with no new admission",
			writeOp:   "KEY_ROTATE",
			ownServer: true,
			steps: []step{
				agentKey("before", ""),
				admitted(
					"the agent is admitted with its keys",
					check{path: "items[*].fingerprint", op: "has", want: "{before}"},
				),
				{
					name: "rotate the agent's keys", on: onHost,
					cli: []string{"pki", "key", "rotate", "--name", "agent", "--json"},
					checks: []check{
						{path: "key.name", op: "==", want: "agent"},
						{path: "key.fingerprint", op: "!=", want: "{before}"},
					},
					capture: map[string]string{"after": "key.fingerprint"},
				},
				agentKey("now", "", check{path: "fingerprint", op: "==", want: "{after}"}),
				admitted("the admission holds the new keys, not the old",
					check{path: "items[*].fingerprint", op: "has", want: "{after}"},
					check{path: "items[*].fingerprint", op: "lacks", want: "{before}"}),
				{
					name:    "the running agent takes the new keys at its heartbeat",
					wait:    30 * time.Second,
					pause:   time.Second,
					cli:     []string{"client", "agent", "list", "--json"},
					checks:  []check{{path: "items[fingerprint={after}].hostname", op: "nonempty"}},
					capture: map[string]string{"host": "items[fingerprint={after}].hostname"},
				},
				{
					name: "the controller takes the agent's result",
					cli: []string{
						"client",
						"node",
						"hostname",
						"get",
						"--target",
						"{host}",
						"--json",
					},
					checks: []check{
						{path: "results[*].error", op: "empty"},
						{path: "results[*].status", op: "!=", want: "refused"},
						{path: "results[*].name", op: "nonempty"},
					},
				},
			},
		},
		{
			name:      "keys that are not admitted do not rotate",
			writeOp:   "KEY_ROTATE",
			ownServer: true,
			steps: []step{
				{
					name:   "make keys that are not admitted",
					on:     onHost,
					keyDir: "{dir:agent}",
					cli:    []string{"pki", "key", "create", "--name", "agent"},
				},
				agentKey("before", "{dir:agent}"),
				{
					name:   "the controller refuses the rotation",
					on:     onHost,
					keyDir: "{dir:agent}",
					fails:  true,
					cli:    []string{"pki", "key", "rotate", "--name", "agent"},
					checks: []check{
						{path: "!all", op: "contains", want: "the controller refused the rotation"},
					},
				},
				agentKey("after", "{dir:agent}",
					check{path: "fingerprint", op: "==", want: "{before}"},
					check{op: "nofile", want: "{dir:agent}/agent.next.key"}),
			},
		},
		{
			name:    "rotate the controller's key with no new admission",
			writeOp: "CONTROLLER_KEY_ROTATE",
			steps: []step{
				{
					name:    "read the controller's key",
					cli:     []string{"client", "pki", "key", "get", "--json"},
					checks:  []check{{path: "current", op: "nonempty"}},
					capture: map[string]string{"before": "current"},
				},
				{
					name: "start a rotation",
					cli:  []string{"client", "pki", "key", "rotate", "--json"},
					checks: []check{
						{path: "next", op: "nonempty"},
						{path: "next", op: "!=", want: "{before}"},
						{path: "agents[*].hostname", op: "nonempty"},
					},
					capture: map[string]string{"next": "next"},
				},
				controllerKey(
					"the controller switches when each agent has the next key",
					90*time.Second,
					check{path: "current", op: "==", want: "{next}"},
					check{path: "next", op: "empty"},
				),
				{
					name:   "each agent trusts the new key",
					wait:   30 * time.Second,
					pause:  time.Second,
					cli:    []string{"client", "agent", "list", "--json"},
					checks: []check{{path: "items[*].controller_key", op: "==", want: "{next}"}},
				},
				{
					name: "an agent verifies a job that the new key signed",
					cli: []string{
						"client",
						"node",
						"hostname",
						"get",
						"--target",
						"{target}",
						"--json",
					},
					checks: []check{
						{path: "results[*].error", op: "empty"},
						{path: "results[*].name", op: "nonempty"},
					},
				},
			},
		},
		{
			name: "a rotation in progress is not started again", writeOp: "CONTROLLER_KEY_ROTATE",
			steps: []step{
				{
					name:    "start a rotation",
					cli:     []string{"client", "pki", "key", "rotate", "--json"},
					checks:  []check{{path: "next", op: "nonempty"}},
					capture: map[string]string{"first": "next"},
				},
				{
					name: "rotate again during it",
					cli:  []string{"client", "pki", "key", "rotate", "--json"},
				},
				controllerKey("the controller switches to the first rotation's key", 90*time.Second,
					check{path: "current", op: "==", want: "{first}"}),
			},
		},
	}
}

func sealScenarios() []scenario {
	name := "osapi-int-sealed-{stamp}"
	key := "int-sealed-{stamp}"
	value := "sealed-value-{stamp}"

	return []scenario{
		{
			name: "each admitted agent has a P-256 encryption key",
			steps: []step{admitted("read the admitted keys",
				check{path: "items[*].encryption_key", op: "p256"},
				check{path: "items[*].fingerprint", op: "prefix", want: "SHA256:"})},
		},
		{
			name:    "a job opens a sealed secret, and the stored job holds it sealed",
			writeOp: "SEALED_SECRET",
			steps: []step{
				{
					name:  "write the secret",
					stdin: value,
					cli: []string{
						"client",
						"secret",
						"update",
						"--vault",
						"default",
						"--key",
						key,
						"--value",
						"-",
					},
				},
				{
					name: "get shows the secret, not its value",
					cli: []string{
						"client",
						"secret",
						"get",
						"--vault",
						"default",
						"--key",
						key,
						"--json",
					},
					checks: []check{
						{path: "", op: "absent", want: value},
						{path: "version", op: "nonempty"},
					},
				},
				{
					name: "reveal shows the value",
					cli: []string{
						"client",
						"secret",
						"reveal",
						"--vault",
						"default",
						"--key",
						key,
						"--json",
					},
					checks: []check{{path: "value", op: "==", want: value}},
				},
				{
					name: "make a container",
					cli: []string{
						"client",
						"node",
						"docker",
						"create",
						"--target",
						"{target}",
						"--name",
						name,
						"--image",
						"alpine:3.20",
						"--command",
						"sleep",
						"--command",
						"300",
						"--auto-start",
						"--json",
					},
					skipIf:  &check{path: "results[0].status", op: "==", want: "skipped"},
					checks:  []check{{path: "results[0].status", op: "!=", want: "failed"}},
					capture: map[string]string{"host": "results[0].hostname"},
				},
				{
					name: "the command gets the secret's value",
					cli: []string{
						"client",
						"node",
						"docker",
						"exec",
						"--target",
						"{host}",
						"--name",
						name,
						"--env",
						"SEALED=secret://default/" + key,
						"--command",
						"printenv",
						"--command",
						"SEALED",
						"--json",
					},
					skipIf: &check{path: "", op: "contains", want: "command execution is disabled"},
					checks: []check{
						{path: "results[0].exit_code", op: "==", want: "0"},
						{path: "results[0].stdout", op: "==", want: value},
					},
					capture: map[string]string{"job": "job_id"},
				},
				{
					name: "the stored job holds the value sealed",
					cli:  []string{"client", "job", "get", "--id", "{job}", "--json"},
					checks: []check{
						{path: "operation.data.env.SEALED", op: "prefix", want: "sealed:"},
						{path: "operation.data.env.SEALED", op: "absent", want: value},
					},
				},
				{
					name: "the audit log holds no value",
					cli:  []string{"client", "audit", "list", "--limit", "100", "--json"},
					checks: []check{
						{path: "", op: "contains", want: key},
						{path: "", op: "absent", want: value},
					},
				},
			},
			cleanup: []step{
				{
					name: "stop the container",
					cli: []string{
						"client",
						"node",
						"docker",
						"stop",
						"--target",
						"{host}",
						"--name",
						name,
					},
				},
				{
					name: "delete the container",
					cli: []string{
						"client",
						"node",
						"docker",
						"delete",
						"--target",
						"{host}",
						"--name",
						name,
					},
				},
				{
					name: "delete the secret",
					cli: []string{
						"client",
						"secret",
						"delete",
						"--vault",
						"default",
						"--key",
						key,
					},
				},
			},
		},
	}
}

func agentServiceScenarios() []scenario {
	return []scenario{{
		name: "install and uninstall the agent's service", writeOp: "AGENT_SERVICE",
		fixture: true, fixtureAutoAdmit: true, ownServer: true, rootLinux: true,
		steps: []step{
			joinToken("join"),
			{
				name: "install the service with the join string", on: onFixtureHost,
				cli:     []string{"agent", "service", "install", "--join", "{join}", "--json"},
				capture: map[string]string{"service": "service", "config": "config_file"},
				checks: []check{
					{path: "joined", op: "==", want: "true"},
					{op: "file", want: "{service}"},
					{op: "file", want: "{config}"},
				},
			},
			{
				name:   "the service is enabled",
				exec:   []string{"systemctl", "is-enabled", "osapi-agent"},
				checks: []check{{path: "", op: "==", want: "enabled"}},
			},
			{
				name:   "the service's agent is Ready",
				on:     onFixture,
				wait:   60 * time.Second,
				pause:  time.Second,
				cli:    []string{"client", "agent", "list", "--json"},
				checks: []check{{path: "items[*].state", op: "==", want: "Ready"}},
			},
			{
				name: "uninstall the service", on: onFixtureHost,
				cli: []string{"agent", "service", "uninstall", "--purge", "--json"},
				checks: []check{
					{op: "nofile", want: "{service}"},
					{op: "nofile", want: "{config}"},
				},
			},
			{
				name:   "the service is not active",
				exec:   []string{"systemctl", "is-active", "osapi-agent"},
				checks: []check{{path: "", op: "!=", want: "active"}},
			},
		},
	}}
}

// pkiStep is a step that runs an osapi pki command against the fixture's
// config, so it reads the fixture's keys and certificates.
func pkiStep(
	name string,
	cli []string,
	checks ...check,
) step {
	return step{name: name, on: onFixture, cli: append([]string{"pki"}, cli...), checks: checks}
}

func pkiScenarios() []scenario {
	keyGet := pkiStep(
		"the controller's key, by name",
		[]string{"key", "get", "--name", "controller", "--json"},
		check{path: "fingerprint", op: "==", want: "{ctl}"},
		check{path: "algorithm", op: "==", want: "Ed25519"},
		check{path: "state", op: "==", want: "ok"},
		check{op: "file", want: "{ctlFile}"},
	)
	keyGet.capture = map[string]string{"ctlFile": "file"}
	certGet := pkiStep(
		"the API certificate, by name",
		[]string{"cert", "get", "--name", "api", "--json"},
		check{path: "fingerprint", op: "==", want: "{api}"},
		check{path: "expires", op: "nonempty"},
		check{path: "names[*]", op: "has", want: "localhost"},
		check{op: "file", want: "{apiFile}"},
	)
	certGet.capture = map[string]string{"apiFile": "file"}

	return []scenario{
		{
			name:             "the keys and certificates of a controller, from its host and from the API",
			writeOp:          "PKI",
			fixture:          true,
			fixtureAutoAdmit: true,
			steps: []step{
				{
					name: "pki list has the controller's key and its API certificate",
					on:   onFixture,
					cli:  []string{"pki", "list", "--json"},
					checks: []check{
						{path: "items[kind=key].name", op: "has", want: "controller"},
						{path: "items[kind=cert].name", op: "has", want: "api"},
						{path: "items[name=controller].fingerprint", op: "prefix", want: "SHA256:"},
					},
					capture: map[string]string{"ctl": "items[name=controller].fingerprint"},
				},
				{
					name: "pki key list agrees",
					on:   onFixture,
					cli:  []string{"pki", "key", "list", "--json"},
					checks: []check{
						{path: "items[name=controller].fingerprint", op: "==", want: "{ctl}"},
					},
				},
				keyGet,
				{
					name:  "an unknown key is refused",
					on:    onFixture,
					fails: true,
					cli:   []string{"pki", "key", "get", "--name", "nosuch"},
				},
				{
					name: "pki cert list has the API certificate",
					on:   onFixture,
					cli:  []string{"pki", "cert", "list", "--json"},
					checks: []check{
						{path: "items[name=api].fingerprint", op: "prefix", want: "SHA256:"},
						{path: "items[name=api].state", op: "==", want: "ok"},
					},
					capture: map[string]string{"api": "items[name=api].fingerprint"},
				},
				certGet,
				{
					name:  "an unknown certificate is refused",
					on:    onFixture,
					fails: true,
					cli:   []string{"pki", "cert", "get", "--name", "nosuch"},
				},
				{
					name: "the API shows the same certificate",
					on:   onFixture,
					cli:  []string{"client", "pki", "cert", "list", "--json"},
					checks: []check{
						{path: "items[name=api].fingerprint", op: "has", want: "{api}"},
					},
				},
				{
					name: "the API shows the same controller key", on: onFixture,
					cli:    []string{"client", "pki", "key", "get", "--json"},
					checks: []check{{path: "current", op: "==", want: "{ctl}"}},
				},
				{name: "start an agent", start: "agent", keyDir: "{dir:agent}"},
				ready("the agent is Ready"),
				{
					name:   "the agent trusts the controller's key",
					on:     onFixture,
					wait:   30 * time.Second,
					pause:  time.Second,
					cli:    []string{"client", "agent", "list", "--json"},
					checks: []check{{path: "items[*].controller_key", op: "==", want: "{ctl}"}},
				},
				{
					name:   "make an agent's keys in a new directory",
					on:     onFixture,
					keyDir: "{dir:new}",
					cli:    []string{"pki", "key", "create", "--name", "agent"},
				},
				{
					name: "read the new keys", on: onFixture, keyDir: "{dir:new}",
					cli:     []string{"pki", "key", "get", "--name", "agent", "--json"},
					checks:  []check{{path: "fingerprint", op: "prefix", want: "SHA256:"}},
					capture: map[string]string{"made": "fingerprint"},
				},
				{
					name:   "a second create refuses the key that is there",
					on:     onFixture,
					keyDir: "{dir:new}",
					fails:  true,
					cli:    []string{"pki", "key", "create", "--name", "agent"},
					checks: []check{{path: "!all", op: "contains", want: "exists"}},
				},
				{
					name: "the refused create left the keys", on: onFixture, keyDir: "{dir:new}",
					cli:    []string{"pki", "key", "get", "--name", "agent", "--json"},
					checks: []check{{path: "fingerprint", op: "==", want: "{made}"}},
				},
			},
		},
	}
}

// agentState is a step that waits until the fixture's agent host is in
// state.
func agentState(
	state string,
) step {
	return step{
		name: "the agent is " + state, on: onFixture, wait: 30 * time.Second, pause: time.Second,
		cli:    []string{"client", "agent", "list", "--json"},
		checks: []check{{path: "items[hostname={host}].state", op: "==", want: state}},
	}
}

func jobScenarios() []scenario {
	return []scenario{
		{
			name: "the controller's config of an agent reaches the agent", writeOp: "AGENT_CONFIG",
			fixture: true, fixtureAutoAdmit: true,
			steps: []step{
				{name: "start an agent", start: "agent", keyDir: "{dir:agent}"},
				ready("the agent is Ready"),
				{
					name: "set the agent's config", on: onFixture,
					cli: []string{
						"client",
						"agent",
						"config",
						"update",
						"--hostname",
						"{host}",
						"--labels",
						"group=cfg",
						"--max-jobs",
						"2",
						"--log-level",
						"debug",
						"--facts-interval",
						"2m",
					},
				},
				{
					name: "the controller holds the config",
					on:   onFixture,
					cli: []string{
						"client",
						"agent",
						"config",
						"get",
						"--hostname",
						"{host}",
						"--json",
					},
					checks: []check{
						{path: "labels.group", op: "==", want: "cfg"},
						{path: "max_jobs", op: "==", want: "2"},
						{path: "log_level", op: "==", want: "debug"},
						{path: "facts_interval", op: "==", want: "2m"},
					},
				},
				{
					name:  "the agent applies it at its heartbeat",
					on:    onFixture,
					wait:  30 * time.Second,
					pause: time.Second,
					cli:   []string{"client", "agent", "list", "--json"},
					checks: []check{
						{path: "items[hostname={host}].labels.group", op: "==", want: "cfg"},
						{path: "items[hostname={host}].config_behind", op: "empty"},
					},
				},
				{
					name: "a job to the label reaches the agent",
					on:   onFixture,
					cli: []string{
						"client",
						"node",
						"hostname",
						"get",
						"--target",
						"group:cfg",
						"--json",
					},
					checks: []check{{path: "results[*].hostname", op: "has", want: "{host}"}},
				},
				{
					name: "a job to _all reaches the agent",
					on:   onFixture,
					cli: []string{
						"client",
						"node",
						"hostname",
						"get",
						"--target",
						"_all",
						"--json",
					},
					checks: []check{{path: "results[*].hostname", op: "has", want: "{host}"}},
				},
			},
		},
		{
			name:             "a job's life: a failed job, its retry and delete, and a drained agent",
			writeOp:          "JOB_LIFE",
			fixture:          true,
			fixtureAutoAdmit: true,
			steps: []step{
				{name: "start an agent", start: "agent", keyDir: "{dir:agent}"},
				ready("the agent is Ready"),
				{
					name:  "a job that fails on the agent",
					on:    onFixture,
					fails: true,
					cli: []string{
						"client",
						"node",
						"user",
						"get",
						"--target",
						"{host}",
						"--name",
						"osapi-nosuch",
						"--json",
					},
					checks: []check{{path: "!all", op: "contains", want: "not found"}},
				},
				{
					name:    "the job list has the failed job",
					on:      onFixture,
					wait:    15 * time.Second,
					pause:   time.Second,
					cli:     []string{"client", "job", "list", "--status", "failed", "--json"},
					checks:  []check{{path: "items[0].id", op: "nonempty"}},
					capture: map[string]string{"failed": "items[0].id"},
				},
				{
					name: "the failed job is failed", on: onFixture,
					cli:    []string{"client", "job", "get", "--id", "{failed}", "--json"},
					checks: []check{{path: "status", op: "==", want: "failed"}},
				},
				{
					name: "retry the failed job", on: onFixture,
					cli:     []string{"client", "job", "retry", "--id", "{failed}", "--json"},
					checks:  []check{{path: "job_id", op: "!=", want: "{failed}"}},
					capture: map[string]string{"retried": "job_id"},
				},
				{
					name:   "the retry is a job of its own",
					on:     onFixture,
					wait:   15 * time.Second,
					pause:  time.Second,
					cli:    []string{"client", "job", "get", "--id", "{retried}", "--json"},
					checks: []check{{path: "status", op: "==", want: "failed"}},
				},
				{
					name: "delete the failed job",
					on:   onFixture,
					cli:  []string{"client", "job", "delete", "--id", "{failed}"},
				},
				{
					name:  "the deleted job is gone",
					on:    onFixture,
					fails: true,
					cli:   []string{"client", "job", "get", "--id", "{failed}", "--json"},
				},
				{
					name: "drain the agent",
					on:   onFixture,
					cli:  []string{"client", "agent", "drain", "--hostname", "{host}"},
				},
				agentState("Cordoned"),
				job("a job does not reach a drained agent", true, 0),
				{
					name: "undrain the agent",
					on:   onFixture,
					cli:  []string{"client", "agent", "undrain", "--hostname", "{host}"},
				},
				agentState("Ready"),
				job("a job reaches the undrained agent", false, 0),
			},
		},
	}
}
