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
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/machineid"
	"github.com/osapi-io/gohai/pkg/gohai/family"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/job"
)

// machineIDTimeout bounds the read of the machine ID. Identity resolution
// runs before the agent starts, with no deadline of its own. On macOS the
// read runs ioreg, and gohai's runner bounds each command as well.
var machineIDTimeout = 10 * time.Second

// getMachineIDFn dispatches to the platform-specific machine ID resolver.
// Overridable in tests via export_test.go.
var getMachineIDFn = defaultGetMachineID

// osPlatform holds the current platform string. Defaults to runtime.GOOS.
// Overridable in tests via export_test.go to cover all platform branches.
var osPlatform = runtime.GOOS

// machineIDRunner runs the commands of the read, such as ioreg on macOS. Nil
// is gohai's runner; a test replaces it.
var machineIDRunner source.Runner

// defaultGetMachineID reads the machine ID through gohai's machine_id
// collector: /etc/machine-id, then /var/lib/dbus/machine-id, then the DMI
// product UUID on Linux, and the IOPlatformUUID of ioreg on macOS. The agent
// must not start without one, so an empty ID is an error.
func defaultGetMachineID(
	fs avfs.VFS,
) (string, error) {
	src := source.Sources{FS: fs, Exec: machineIDRunner}
	switch osPlatform {
	case "linux":
	case "darwin":
		src.Family = family.Darwin
	default:
		return "", fmt.Errorf("unsupported platform: %s", osPlatform)
	}

	ctx, cancel := context.WithTimeout(context.Background(), machineIDTimeout)
	defer cancel()

	got, err := machineid.NewWith(src).Collect(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("read the machine ID: %w", err)
	}
	info, _ := got.(*machineid.Info)
	if info == nil || strings.TrimSpace(info.ID) == "" {
		return "", fmt.Errorf("read the machine ID: the host has none")
	}

	return strings.TrimSpace(info.ID), nil
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

	// GetAgentHostname never returns an error — it falls back to "unknown"
	// on any failure. The error return exists for interface consistency.
	hostname, _ := job.GetAgentHostname(configHostname)

	return &Identity{
		MachineID: machineID,
		Hostname:  hostname,
	}, nil
}
