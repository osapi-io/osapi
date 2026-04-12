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

package identity

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ioregFn is the function used to run ioreg and return its output.
// Injectable for testing via export_test.go.
var ioregFn = defaultIoregFn

// execCommandFn wraps exec.Command().Output() for testing.
// Injectable via export_test.go to simulate command failures.
var execCommandFn = defaultExecCommandFn

// uuidPattern matches the IOPlatformUUID value in ioreg output.
var uuidPattern = regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)

// defaultExecCommandFn runs the ioreg command via exec.Command.
func defaultExecCommandFn() ([]byte, error) {
	return exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
}

// defaultIoregFn runs the ioreg command to retrieve platform expert device info.
func defaultIoregFn() (string, error) {
	out, err := execCommandFn()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// GetDarwinMachineID retrieves the machine ID on macOS by parsing the
// IOPlatformUUID from ioreg output.
func GetDarwinMachineID() (string, error) {
	output, err := ioregFn()
	if err != nil {
		return "", fmt.Errorf("run ioreg: %w", err)
	}

	matches := uuidPattern.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("IOPlatformUUID not found in ioreg output")
	}

	return strings.TrimSpace(matches[1]), nil
}
