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
	"runtime"

	"github.com/avfs/avfs"

	"github.com/retr0h/osapi/internal/job"
)

// getMachineIDFn dispatches to the platform-specific machine ID resolver.
// Overridable in tests via export_test.go.
var getMachineIDFn = defaultGetMachineID

// defaultGetMachineID selects the machine ID reader based on runtime.GOOS.
func defaultGetMachineID(
	fs avfs.VFS,
) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return GetMachineIDFromFS(fs)
	case "darwin":
		return GetDarwinMachineID()
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// GetIdentity resolves the machine identity by reading the machine ID and
// hostname. The agent must refuse to start without a machine ID; this
// function returns an error rather than a fallback value.
func GetIdentity(
	fs avfs.VFS,
	configHostname string,
) (*Identity, error) {
	machineID, err := getMachineIDFn(fs)
	if err != nil {
		return nil, fmt.Errorf("resolve machine-id: %w", err)
	}

	hostname, err := job.GetAgentHostname(configHostname)
	if err != nil {
		return nil, fmt.Errorf("resolve hostname: %w", err)
	}

	return &Identity{
		MachineID: machineID,
		Hostname:  hostname,
	}, nil
}
