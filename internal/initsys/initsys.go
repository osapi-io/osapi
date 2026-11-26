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

// Package initsys is how osapi runs a program as a service under the service
// manager of a host: systemd on Linux and launchd on macOS. It makes the file
// of the service and the commands that load, start, stop and remove it. The
// caller writes the file and runs the commands with its own file system and
// runner: osapi agent service install as root, or the service behavior through
// its host, with sudo.
package initsys

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"
)

// The directories of the service files.
const (
	SystemdDir = "/etc/systemd/system"
	LaunchdDir = "/Library/LaunchDaemons"
)

// Service is a program that a service manager runs: its name, such as
// osapi-agent, its launchd label, such as io.osapi.agent, its description,
// the program and its arguments, and the log of launchd.
type Service struct {
	Name        string
	Label       string
	Description string
	Args        []string
	Log         string
	// Header is the comment at the top of the file, one line each, such as
	// who wrote it.
	Header []string
}

// SystemdPath is the unit file of the service name, such as
// /etc/systemd/system/osapi-agent.service.
func SystemdPath(
	name string,
) string {
	return filepath.Join(SystemdDir, name+".service")
}

// SystemdUnit is the unit file of s: it starts after the network is up, and
// systemd starts it again 5 seconds after it stops.
func SystemdUnit(
	s Service,
) []byte {
	var b strings.Builder
	for _, l := range s.Header {
		b.WriteString("# " + l + "\n")
	}
	fmt.Fprintf(
		&b,
		"[Unit]\nDescription=%s\nAfter=network-online.target\nWants=network-online.target\n\n"+
			"[Service]\nExecStart=%s\nRestart=always\nRestartSec=5\n\n"+
			"[Install]\nWantedBy=multi-user.target\n",
		s.Description,
		strings.Join(s.Args, " "),
	)

	return []byte(b.String())
}

// SystemdEnable are the commands that load the unit of name and start it,
// now and at each boot.
func SystemdEnable(
	name string,
) [][]string {
	return [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "--now", name},
	}
}

// SystemdReload is the command that makes systemd read the unit files
// again, after one of them changed.
func SystemdReload() []string {
	return []string{"systemctl", "daemon-reload"}
}

// SystemdDisable are the commands that stop the service name and keep it from
// starting at boot, then forget its unit. Remove the unit file between the
// two, to remove the service.
func SystemdDisable(
	name string,
) [][]string {
	return [][]string{
		{"systemctl", "disable", "--now", name},
		{"systemctl", "daemon-reload"},
	}
}

// SystemdAct is the command of action, such as start or restart, on the
// service name.
func SystemdAct(
	action string,
	name string,
) []string {
	return []string{"systemctl", action, name + ".service"}
}

// LaunchdPath is the plist of the label, such as
// /Library/LaunchDaemons/io.osapi.agent.plist.
func LaunchdPath(
	label string,
) string {
	return filepath.Join(LaunchdDir, label+".plist")
}

// LaunchdPlist is the plist of s, a daemon that launchd starts at boot and
// starts again when it stops.
func LaunchdPlist(
	s Service,
) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
`)
	for _, l := range s.Header {
		b.WriteString("<!-- " + l + " -->\n")
	}
	b.WriteString(
		"<plist version=\"1.0\">\n<dict>\n  <key>Label</key>\n  <string>" + esc(
			s.Label,
		) + "</string>\n",
	)
	b.WriteString("  <key>ProgramArguments</key>\n  <array>\n")
	for _, a := range s.Args {
		b.WriteString("    <string>" + esc(a) + "</string>\n")
	}
	b.WriteString(
		"  </array>\n  <key>RunAtLoad</key>\n  <true/>\n  <key>KeepAlive</key>\n  <true/>\n",
	)
	if s.Log != "" {
		b.WriteString("  <key>StandardOutPath</key>\n  <string>" + esc(s.Log) + "</string>\n")
		b.WriteString("  <key>StandardErrorPath</key>\n  <string>" + esc(s.Log) + "</string>\n")
	}
	b.WriteString("</dict>\n</plist>\n")

	return []byte(b.String())
}

// LaunchdLoad are the commands that load the plist of label into the system
// domain, which starts it.
func LaunchdLoad(
	label string,
) [][]string {
	return [][]string{{"launchctl", "bootstrap", "system", LaunchdPath(label)}}
}

// LaunchdUnload are the commands that stop the daemon label and unload it.
// Remove its plist after them, to remove the service.
func LaunchdUnload(
	label string,
) [][]string {
	return [][]string{{"launchctl", "bootout", "system/" + label}}
}

// esc is s with the characters of XML escaped.
func esc(
	s string,
) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))

	return b.String()
}
