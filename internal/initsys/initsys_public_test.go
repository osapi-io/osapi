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
package initsys_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/initsys"
)

type InitsysPublicTestSuite struct {
	suite.Suite
}

func TestInitsysPublicTestSuite(t *testing.T) {
	suite.Run(t, new(InitsysPublicTestSuite))
}

var svc = initsys.Service{
	Name: "osapi-app", Label: "io.osapi.app", Description: "osapi app",
	Args:   []string{"/usr/local/bin/app", "--name", "a&b<c>"},
	Log:    "/var/log/app.log",
	Header: []string{"Written by a test."},
}

func (s *InitsysPublicTestSuite) TestSystemd() {
	s.Equal("/etc/systemd/system/osapi-app.service", initsys.SystemdPath("osapi-app"))
	s.Equal(
		[][]string{{"systemctl", "daemon-reload"}, {"systemctl", "enable", "--now", "osapi-app"}},
		initsys.SystemdEnable("osapi-app"),
	)
	s.Equal(
		[][]string{{"systemctl", "disable", "--now", "osapi-app"}, {"systemctl", "daemon-reload"}},
		initsys.SystemdDisable("osapi-app"),
	)
	s.Equal([]string{"systemctl", "daemon-reload"}, initsys.SystemdReload())
	s.Equal(
		[]string{"systemctl", "restart", "osapi-app.service"},
		initsys.SystemdAct("restart", "osapi-app"),
	)
}

func (s *InitsysPublicTestSuite) TestSystemdUnit() {
	unit := func(execStart string) string {
		return "# Written by a test.\n[Unit]\nDescription=osapi app\n" +
			"After=network-online.target\nWants=network-online.target\n\n" +
			"[Service]\nExecStart=" + execStart + "\nRestart=always\nRestartSec=5\n\n" +
			"[Install]\nWantedBy=multi-user.target\n"
	}
	tests := []struct {
		name         string
		args         []string
		validateFunc func(string)
	}{
		{
			name: "arguments that systemd does not split stay as they are",
			args: []string{"/usr/local/bin/app", "--name", "a&b<c>"},
			validateFunc: func(got string) {
				s.Equal(unit("/usr/local/bin/app --name a&b<c>"), got)
			},
		},
		{
			name: "an argument with a space is quoted",
			args: []string{"/usr/local/bin/osapi", "-f", "/etc/my osapi/agent.yaml"},
			validateFunc: func(got string) {
				s.Equal(unit(`/usr/local/bin/osapi -f "/etc/my osapi/agent.yaml"`), got)
			},
		},
		{
			name: "a specifier, a variable, a quote and a backslash are escaped",
			args: []string{"/bin/app", `100%`, `$HOME`, `say "hi"`, `a\b`},
			validateFunc: func(got string) {
				s.Equal(unit(`/bin/app "100%%" "$$HOME" "say \"hi\"" "a\\b"`), got)
			},
		},
		{
			name: "an empty argument stays an argument",
			args: []string{"/bin/app", ""},
			validateFunc: func(got string) {
				s.Equal(unit(`/bin/app ""`), got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			service := svc
			service.Args = tt.args

			tt.validateFunc(string(initsys.SystemdUnit(service)))
		})
	}
}

func (s *InitsysPublicTestSuite) TestLaunchd() {
	s.Equal("/Library/LaunchDaemons/io.osapi.app.plist", initsys.LaunchdPath("io.osapi.app"))
	plist := string(initsys.LaunchdPlist(svc))
	s.Contains(plist, "<!-- Written by a test. -->\n<plist")
	s.Contains(plist, "<string>io.osapi.app</string>")
	s.Contains(plist, "<string>a&amp;b&lt;c&gt;</string>", "an argument is XML-escaped")
	s.Contains(plist, "<key>StandardOutPath</key>\n  <string>/var/log/app.log</string>")
	s.Contains(plist, "<key>KeepAlive</key>\n  <true/>")

	noLog := svc
	noLog.Log = ""
	s.NotContains(string(initsys.LaunchdPlist(noLog)), "StandardOutPath")

	s.Equal(
		[][]string{
			{"launchctl", "bootstrap", "system", "/Library/LaunchDaemons/io.osapi.app.plist"},
		},
		initsys.LaunchdLoad("io.osapi.app"),
	)
	s.Equal(
		[][]string{{"launchctl", "bootout", "system/io.osapi.app"}},
		initsys.LaunchdUnload("io.osapi.app"),
	)
}
