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
package agentservice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// The files of the service on each platform, and the name of the service.
const (
	systemdUnit  = "/etc/systemd/system/osapi-agent.service"
	launchdPlist = "/Library/LaunchDaemons/io.osapi.agent.plist"
	launchdLabel = "io.osapi.agent"
	serviceName  = "osapi-agent"
)

// enrolledKey is the file in agent.pki.key_dir that a join writes.
const enrolledKey = "enrolled-controller.pub"

// ErrNotRoot is install or uninstall run by a user that is not root.
var ErrNotRoot = errors.New("run this as root: the service and its config are root's")

// ErrPlatform is a platform with no service manager that osapi knows.
var ErrPlatform = errors.New("osapi installs a service on Linux (systemd) and macOS (launchd) only")

// ErrExists is a config file that exists, which install does not replace
// without --force.
var ErrExists = errors.New("the config file exists: --force replaces it")

// Install writes the config with the join string, the unit of the service,
// starts the service, and waits for the join.
func Install(
	ctx context.Context,
	d Deps,
	in InstallInput,
) (InstallOutput, error) {
	if d.EUID() != 0 {
		return InstallOutput{}, ErrNotRoot
	}
	if !validJoin(in.Join) {
		return InstallOutput{}, errors.New(
			"the join string starts with nats://, tls:// or osapi://",
		)
	}
	bin, err := d.Binary()
	if err != nil {
		return InstallOutput{}, fmt.Errorf("find this binary: %w", err)
	}
	unit, body, start, err := service(d.GOOS, bin, d.ConfigFile)
	if err != nil {
		return InstallOutput{}, err
	}
	if _, err := d.FS.Stat(d.ConfigFile); err == nil && !in.Force {
		return InstallOutput{}, fmt.Errorf("%s: %w", d.ConfigFile, ErrExists)
	}
	if err := writeFile(d, d.ConfigFile, configYAML(in, d.KeyDir), 0o600); err != nil {
		return InstallOutput{}, err
	}
	if err := writeFile(d, unit, body, 0o644); err != nil {
		return InstallOutput{}, err
	}
	for _, c := range start {
		if err := d.Run(ctx, c[0], c[1:]...); err != nil {
			return InstallOutput{}, fmt.Errorf("%s: %w", strings.Join(c, " "), err)
		}
	}
	out := InstallOutput{Service: unit, ConfigFile: d.ConfigFile}
	out.Joined = d.Wait(ctx, filepath.Join(d.KeyDir, enrolledKey))
	if out.Joined {
		out.Next = "osapi client admission list, on a host with the CLI"
	} else {
		out.Next = "the host has not joined yet: read the log of the service"
	}

	return out, nil
}

// Uninstall stops the service and removes its unit, and with Purge the
// config file and agent.pki.key_dir.
func Uninstall(
	ctx context.Context,
	d Deps,
	in UninstallInput,
) (UninstallOutput, error) {
	if d.EUID() != 0 {
		return UninstallOutput{}, ErrNotRoot
	}
	unit, _, _, err := service(d.GOOS, "", "")
	if err != nil {
		return UninstallOutput{}, err
	}
	// A service that is not running is not an error: uninstall ends with
	// none, whatever it found.
	for _, c := range stop(d.GOOS) {
		_ = d.Run(ctx, c[0], c[1:]...)
	}
	out := UninstallOutput{Removed: []string{}}
	files := []string{unit}
	if in.Purge {
		files = append(files, d.ConfigFile, d.KeyDir)
	}
	for _, f := range files {
		if _, err := d.FS.Stat(f); err != nil {
			continue
		}
		if err := d.FS.RemoveAll(f); err != nil {
			return out, fmt.Errorf("remove %s: %w", f, err)
		}
		out.Removed = append(out.Removed, f)
	}

	return out, nil
}

// validJoin is whether s has the scheme of a join string.
func validJoin(
	s string,
) bool {
	for _, p := range []string{"nats://", "tls://", "osapi://"} {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	return false
}

// configYAML is the config of a host that joins: the join string, the key
// directory, and the labels of the host.
func configYAML(
	in InstallInput,
	keyDir string,
) []byte {
	agent := map[string]any{"join": in.Join, "pki": map[string]any{"key_dir": keyDir}}
	if len(in.Labels) > 0 {
		agent["labels"] = in.Labels
	}
	// Maps of strings marshal with no error.
	out, _ := yaml.Marshal(map[string]any{"agent": agent})

	return append([]byte("# Written by osapi agent service install. The join string is\n"+
		"# used one time: after the join the agent reads agent.pki.key_dir.\n"), out...)
}

// service is the unit file of the platform, its body, and the commands that
// start it.
func service(
	goos string,
	bin string,
	configFile string,
) (string, []byte, [][]string, error) {
	switch goos {
	case "linux":
		body := fmt.Sprintf("# Written by osapi agent service install.\n[Unit]\n"+
			"Description=osapi agent\nAfter=network-online.target\nWants=network-online.target\n\n"+
			"[Service]\nExecStart=%s -f %s agent start\nRestart=always\nRestartSec=5\n\n"+
			"[Install]\nWantedBy=multi-user.target\n", bin, configFile)

		return systemdUnit, []byte(body), [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "enable", "--now", serviceName},
		}, nil
	case "darwin":
		body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<!-- Written by osapi agent service install. -->
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>-f</string>
    <string>%s</string>
    <string>agent</string>
    <string>start</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/var/log/osapi-agent.log</string>
  <key>StandardErrorPath</key>
  <string>/var/log/osapi-agent.log</string>
</dict>
</plist>
`, launchdLabel, bin, configFile)

		return launchdPlist, []byte(body), [][]string{
			{"launchctl", "bootstrap", "system", launchdPlist},
		}, nil
	}

	return "", nil, nil, fmt.Errorf("%s: %w", goos, ErrPlatform)
}

// stop is the commands that stop and disable the service.
func stop(
	goos string,
) [][]string {
	if goos == "darwin" {
		return [][]string{{"launchctl", "bootout", "system/" + launchdLabel}}
	}

	return [][]string{
		{"systemctl", "disable", "--now", serviceName},
		{"systemctl", "daemon-reload"},
	}
}

// writeFile writes data to file with perm, and makes its directory.
func writeFile(
	d Deps,
	file string,
	data []byte,
	perm os.FileMode,
) error {
	if err := d.FS.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return fmt.Errorf("make the directory of %s: %w", file, err)
	}
	if err := d.FS.WriteFile(file, data, perm); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}

	return nil
}
