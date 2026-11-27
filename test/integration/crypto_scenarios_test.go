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

// cryptoScenarios read the crypto report of each process: of the
// controller and the NATS server in health status, of an agent in agent
// get, and the approved flag of each key in pki list. A build of osapi that
// is not a FIPS build runs with FIPS 140-3 mode off, in CI and in the lab.
func cryptoScenarios() []scenario {
	return []scenario{
		{
			name: "each process reports its crypto",
			steps: []step{
				{
					name: "health status has the crypto of the controller and of NATS",
					cli:  []string{"client", "health", "status", "--json"},
					checks: []check{
						{path: "registry[type=controller].crypto.mode", op: "==", want: "off"},
						// Each listener negotiates only approved suites and
						// groups (#1719), so the mode is the only gap.
						{path: "registry[type=controller].fips", op: "==", want: "off"},
						{
							path: "registry[type=controller].crypto.listeners[*].approved",
							op:   "lacks", want: "false",
						},
						{
							path: "registry[type=nats].crypto.listeners[*].approved",
							op:   "lacks", want: "false",
						},
						{path: "registry[type=controller].crypto.module", op: "nonempty"},
						{
							path: "registry[type=controller].crypto.listeners[*].name",
							op:   "has",
							want: "api",
						},
						{
							path: "registry[type=controller].crypto.algorithms[*].name",
							op:   "has", want: "secrets sealed to agents",
						},
						{
							path: "registry[type=controller].crypto.algorithms[name=secrets sealed to agents].algorithm",
							op:   "==",
							want: "ECDH P-256, HKDF-SHA256, AES-256-GCM",
						},
						{
							path: "registry[type=controller].crypto.algorithms[name=API tokens].only_mode",
							op:   "==",
							want: "true",
						},
						{
							path: "registry[type=controller].crypto.gaps[*]", op: "has",
							want: "FIPS 140-3 mode is off: set GODEBUG=fips140=on",
						},
						{path: "registry[type=nats].crypto.mode", op: "==", want: "off"},
						{
							path: "registry[type=nats].crypto.listeners[*].name",
							op:   "has",
							want: "nats",
						},
						// NATS signs with nkeys, which FIPS 140-3 only mode
						// refuses until osapi#1527.
						{
							path: "registry[type=nats].crypto.algorithms[*].only_mode",
							op:   "has",
							want: "false",
						},
						{path: "registry[type=nats].crypto.enforceable", op: "==", want: "false"},
					},
				},
				{
					name:    "read a Ready agent",
					cli:     []string{"client", "agent", "list", "--json"},
					checks:  []check{{path: "items[state=Ready].hostname", op: "nonempty"}},
					capture: map[string]string{"host": "items[state=Ready].hostname"},
				},
				{
					name: "agent get has the crypto of the agent",
					cli:  []string{"client", "agent", "get", "--hostname", "{host}", "--json"},
					checks: []check{
						{path: "crypto.mode", op: "==", want: "off"},
						{path: "fips", op: "==", want: "off"},
						{path: "crypto.listeners[*].approved", op: "lacks", want: "false"},
						{
							path: "crypto.algorithms[*].name",
							op:   "has",
							want: "secrets sealed to agents",
						},
						{path: "crypto.algorithms[*].approved", op: "lacks", want: "false"},
					},
				},
			},
		},
		{
			name:             "pki list says that each key's algorithm is approved",
			fixture:          true,
			fixtureAutoAdmit: true,
			steps: []step{
				{
					name: "pki list has the approved flag of each key",
					on:   onFixture,
					cli:  []string{"pki", "list", "--json"},
					checks: []check{
						{path: "items[name=controller].approved", op: "==", want: "true"},
						{path: "items[name=api].approved", op: "==", want: "true"},
					},
				},
			},
		},
	}
}
