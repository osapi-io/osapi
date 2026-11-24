//go:build integration

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

import (
	"fmt"
	"os"
	"strings"
)

// The integration suite prepares its own target through osapi, the same on
// the CI runner and on a lab VM, so neither needs a setup step of its own.
// With OSAPI_INTEGRATION_WRITE_BEHAVIORS=1, TestMain runs prepare after the
// agent is up and before the tests, and unprepare after them.

// dummy is the dummy interface that the dns, interface and route write
// examples run on.
const dummy = "osapi0"

// pingRange is the net.ipv4.ping_group_range that admits every group, so
// the agent pings without root.
const (
	pingKey   = "net.ipv4.ping_group_range"
	pingRange = "0 2147483647"
)

// prepared is what prepare changed on the target, for unprepare to put
// back.
var prepared struct {
	// dummy is true while the suite has made dummy and not deleted it.
	dummy bool
	// ping is true when the suite set pingKey; pingWas and pingManaged are
	// its value and whether osapi managed it before.
	ping        bool
	pingWas     string
	pingManaged bool
}

// prepareWrites reports whether the suite prepares its target: only when the
// behavior write examples run, because preparing changes the target.
func prepareWrites() bool {
	return os.Getenv("OSAPI_INTEGRATION_WRITE_BEHAVIORS") == "1"
}

// cli runs a command of osapi client against the target and returns its
// result rows. It fails on an exit code other than 0, or a row that is not
// ok or skipped.
func cli(
	args ...string,
) ([]map[string]any, error) {
	args = append(args, "--target", target, "--json")
	stdout, stderr, code := runCLI(args...)
	if code != 0 {
		return nil, fmt.Errorf("%s: exit %d: %s", strings.Join(args, " "), code, stderr)
	}
	var body struct {
		Results []map[string]any `json:"results"`
	}
	if err := parseJSON(stdout, &body); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", strings.Join(args, " "), err, stdout)
	}
	for _, r := range body.Results {
		if r["status"] != "ok" && r["status"] != "skipped" {
			return nil, fmt.Errorf("%s: %v", strings.Join(args, " "), r)
		}
	}

	return body.Results, nil
}

// prepare gives the target what the tests need, through osapi: the ping
// parameter for the ping examples, chrony for the ntp examples, and dummy
// for the dns, interface and route examples. Each call changes nothing when
// the target has it. TestPrograms then waits for a heartbeat that finds
// chrony.
func prepare() error {
	if err := setPing(); err != nil {
		return err
	}
	if _, err := cli("client", "node", "package", "create", "--name", "chrony"); err != nil {
		return err
	}

	return makeDummy()
}

// setPing sets pingKey to pingRange, and keeps the value that it had.
func setPing() error {
	rows, err := cli("client", "node", "sysctl", "get", "--key", pingKey)
	if err != nil {
		return err
	}
	for _, r := range rows {
		value, _ := r["value"].(string)
		if r["status"] != "ok" || strings.Join(strings.Fields(value), " ") == pingRange {
			continue
		}
		prepared.pingWas = strings.Join(strings.Fields(value), " ")
		prepared.pingManaged, _ = r["managed"].(bool)
		if _, err := cli("client", "node", "sysctl", "update",
			"--key", pingKey, "--value", pingRange); err != nil {
			return err
		}
		prepared.ping = true
	}

	return nil
}

// makeDummy makes dummy on the target, and notes when the suite made it.
func makeDummy() error {
	rows, err := cli("client", "node", "interface", "create",
		"--name", dummy, "--type", "dummy")
	for _, r := range rows {
		if r["changed"] == true {
			prepared.dummy = true
		}
	}

	return err
}

// dropDummy deletes dummy, only when the suite made it.
func dropDummy() error {
	if !prepared.dummy {
		return nil
	}
	if _, err := cli("client", "node", "interface", "delete", "--name", dummy); err != nil {
		return err
	}
	prepared.dummy = false

	return nil
}

// unprepare puts back what prepare changed: no dummy, and the old value of
// pingKey, which osapi stops managing when it did not before. chrony stays.
func unprepare() error {
	if err := dropDummy(); err != nil {
		return err
	}
	if !prepared.ping {
		return nil
	}
	if _, err := cli("client", "node", "sysctl", "update",
		"--key", pingKey, "--value", prepared.pingWas); err != nil {
		return err
	}
	if !prepared.pingManaged {
		if _, err := cli("client", "node", "sysctl", "delete", "--key", pingKey); err != nil {
			return err
		}
	}
	prepared.ping = false

	return nil
}
