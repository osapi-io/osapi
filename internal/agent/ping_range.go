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

package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// pingRangePath is the kernel setting that says which groups may send a
// ping on an ICMP datagram socket. Overridable in tests via export_test.go.
var pingRangePath = "/proc/sys/net/ipv4/ping_group_range"

// getgroups is os.Getgroups. Overridable in tests via export_test.go.
var getgroups = os.Getgroups

// processGroups are the groups of the agent: its group and its
// supplementary groups. Overridable in tests via export_test.go.
var processGroups = defaultProcessGroups

// defaultProcessGroups reads the groups of the agent from the process.
func defaultProcessGroups() ([]int, error) {
	groups, err := getgroups()
	if err != nil {
		return nil, err
	}

	return append(groups, os.Getgid()), nil
}

// PingProblem says why a ping from the agent fails, or "" when it can send
// one. On Linux the agent sends a ping on an ICMP datagram socket, also as
// root, and the kernel permits that only for the groups in
// net.ipv4.ping_group_range. Its default, "1 0", permits no group. A host
// without the setting, such as macOS, has no problem to report.
func PingProblem() string {
	data, err := os.ReadFile(pingRangePath)
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return fmt.Sprintf(
			"cannot read net.ipv4.ping_group_range %q",
			strings.TrimSpace(string(data)),
		)
	}
	low, errLow := strconv.Atoi(fields[0])
	high, errHigh := strconv.Atoi(fields[1])
	if errLow != nil || errHigh != nil {
		return fmt.Sprintf(
			"cannot read net.ipv4.ping_group_range %q",
			strings.TrimSpace(string(data)),
		)
	}

	groups, err := processGroups()
	if err != nil {
		return fmt.Sprintf("cannot read the groups of the agent: %v", err)
	}
	for _, g := range groups {
		if g >= low && g <= high {
			return ""
		}
	}

	return fmt.Sprintf(
		"no group of the agent %v is in net.ipv4.ping_group_range (%d %d); each ping fails",
		groups, low, high,
	)
}
