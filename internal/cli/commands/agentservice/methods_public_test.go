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
package agentservice_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/agentservice"
)

type MethodsPublicTestSuite struct {
	suite.Suite
	fs   avfs.VFS
	ran  []string
	fail string
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.fs = memfs.New()
	s.ran = nil
	s.fail = ""
}

// deps are root's on goos, with a join that arrives when joined.
func (s *MethodsPublicTestSuite) deps(
	goos string,
	joined bool,
) agentservice.Deps {
	return agentservice.Deps{
		FS:     s.fs,
		GOOS:   goos,
		EUID:   func() int { return 0 },
		Binary: func() (string, error) { return "/usr/local/bin/osapi", nil },
		Run: func(_ context.Context, name string, args ...string) error {
			c := strings.Join(append([]string{name}, args...), " ")
			s.ran = append(s.ran, c)
			if s.fail != "" && strings.Contains(c, s.fail) {
				return errors.New("failed")
			}

			return nil
		},
		ConfigFile: "/etc/osapi/osapi.yaml",
		KeyDir:     "/etc/osapi/pki",
		Wait:       func(context.Context, string) bool { return joined },
	}
}

// noRemove is a file system that cannot remove.
type noRemove struct {
	avfs.VFS
}

func (noRemove) RemoveAll(string) error { return errors.New("read-only") }

const join = "tls://nats.example.com:4222?fp=SHA256%3Aab&ns=osapi&token=osapi_jt_t1.secret"

func (s *MethodsPublicTestSuite) TestInstall() {
	ctx := context.Background()

	s.Run("linux: the config, the unit, systemd started, and the join", func() {
		got, err := agentservice.Install(ctx, s.deps("linux", true), agentservice.InstallInput{
			Join: join, Labels: map[string]string{"group": "web"},
		})
		s.Require().NoError(err)
		s.True(got.Joined)
		s.Equal("/etc/systemd/system/osapi-agent.service", got.Service)
		cfg, _ := avfs.ReadFile(s.fs, "/etc/osapi/osapi.yaml")
		s.Contains(string(cfg), "join: "+join)
		s.Contains(string(cfg), "key_dir: /etc/osapi/pki")
		s.Contains(string(cfg), "group: web")
		info, _ := s.fs.Stat("/etc/osapi/osapi.yaml")
		s.Equal("-rw-------", info.Mode().Perm().String(), "the join string is a secret")
		unit, _ := avfs.ReadFile(s.fs, got.Service)
		s.Contains(
			string(unit),
			"ExecStart=/usr/local/bin/osapi -f /etc/osapi/osapi.yaml agent start",
		)
		s.Equal([]string{"systemctl daemon-reload", "systemctl enable --now osapi-agent"}, s.ran)
	})
	s.Run("macOS: the launch daemon, bootstrapped", func() {
		got, err := agentservice.Install(
			ctx,
			s.deps("darwin", true),
			agentservice.InstallInput{Join: join},
		)
		s.Require().NoError(err)
		s.Equal("/Library/LaunchDaemons/io.osapi.agent.plist", got.Service)
		plist, _ := avfs.ReadFile(s.fs, got.Service)
		s.Contains(string(plist), "<string>io.osapi.agent</string>")
		s.Equal(
			[]string{"launchctl bootstrap system /Library/LaunchDaemons/io.osapi.agent.plist"},
			s.ran,
		)
	})
	s.Run("a host that has not joined in a minute says where to look", func() {
		got, err := agentservice.Install(
			ctx,
			s.deps("linux", false),
			agentservice.InstallInput{Join: join},
		)
		s.Require().NoError(err)
		s.False(got.Joined)
		s.Contains(got.Next, "read the log")
	})
	s.Run("a config that exists is kept, unless --force", func() {
		s.Require().NoError(s.fs.MkdirAll("/etc/osapi", 0o700))
		s.Require().NoError(avfs.WriteFile(s.fs, "/etc/osapi/osapi.yaml", []byte("x"), 0o600))
		_, err := agentservice.Install(
			ctx,
			s.deps("linux", true),
			agentservice.InstallInput{Join: join},
		)
		s.ErrorIs(err, agentservice.ErrExists)
		_, err = agentservice.Install(
			ctx,
			s.deps("linux", true),
			agentservice.InstallInput{Join: join, Force: true},
		)
		s.NoError(err)
	})
	s.Run("not root", func() {
		d := s.deps("linux", true)
		d.EUID = func() int { return 501 }
		_, err := agentservice.Install(ctx, d, agentservice.InstallInput{Join: join})
		s.ErrorIs(err, agentservice.ErrNotRoot)
	})
	s.Run("a string that is not a join string", func() {
		_, err := agentservice.Install(
			ctx,
			s.deps("linux", true),
			agentservice.InstallInput{Join: "http://x"},
		)
		s.ErrorContains(err, "nats://, tls:// or osapi://")
	})
	s.Run("a platform with no service manager that osapi knows", func() {
		_, err := agentservice.Install(
			ctx,
			s.deps("windows", true),
			agentservice.InstallInput{Join: join},
		)
		s.ErrorIs(err, agentservice.ErrPlatform)
	})
	s.Run("a binary that cannot be found", func() {
		d := s.deps("linux", true)
		d.Binary = func() (string, error) { return "", errors.New("gone") }
		_, err := agentservice.Install(ctx, d, agentservice.InstallInput{Join: join})
		s.ErrorContains(err, "find this binary: gone")
	})
	s.Run("a service that does not start", func() {
		s.fail = "enable"
		_, err := agentservice.Install(
			ctx,
			s.deps("linux", true),
			agentservice.InstallInput{Join: join},
		)
		s.ErrorContains(err, "systemctl enable --now osapi-agent: failed")
	})
	s.Run("a config that cannot be written", func() {
		d := s.deps("linux", true)
		s.Require().NoError(avfs.WriteFile(s.fs, "/blocked", []byte("x"), 0o600))
		d.ConfigFile = "/blocked/osapi.yaml"
		_, err := agentservice.Install(ctx, d, agentservice.InstallInput{Join: join})
		s.ErrorContains(err, "make the directory of /blocked/osapi.yaml")
	})
	s.Run("a config file that is a directory cannot be written", func() {
		s.Require().NoError(s.fs.MkdirAll("/etc/osapi/osapi.yaml", 0o700))
		_, err := agentservice.Install(ctx, s.deps("linux", true),
			agentservice.InstallInput{Join: join, Force: true})
		s.ErrorContains(err, "write /etc/osapi/osapi.yaml")
	})
	s.Run("a unit that cannot be written", func() {
		s.Require().NoError(s.fs.MkdirAll("/etc", 0o755))
		s.Require().NoError(avfs.WriteFile(s.fs, "/etc/systemd", []byte("x"), 0o600))
		_, err := agentservice.Install(
			ctx,
			s.deps("linux", true),
			agentservice.InstallInput{Join: join},
		)
		s.ErrorContains(err, "osapi-agent.service")
	})
}

func (s *MethodsPublicTestSuite) TestUninstall() {
	ctx := context.Background()
	s.Run("stops the service and removes the unit, and with --purge the config and keys", func() {
		s.Require().NoError(s.fs.MkdirAll("/etc/systemd/system", 0o755))
		s.Require().
			NoError(avfs.WriteFile(s.fs, "/etc/systemd/system/osapi-agent.service", []byte("x"), 0o644))
		s.Require().NoError(s.fs.MkdirAll("/etc/osapi/pki", 0o700))
		s.Require().NoError(avfs.WriteFile(s.fs, "/etc/osapi/osapi.yaml", []byte("x"), 0o600))
		got, err := agentservice.Uninstall(
			ctx,
			s.deps("linux", true),
			agentservice.UninstallInput{Purge: true},
		)
		s.Require().NoError(err)
		s.Equal([]string{
			"/etc/systemd/system/osapi-agent.service", "/etc/osapi/osapi.yaml", "/etc/osapi/pki",
		}, got.Removed)
		s.Equal([]string{"systemctl disable --now osapi-agent", "systemctl daemon-reload"}, s.ran)
	})
	s.Run("macOS, with nothing installed, removes nothing and is not an error", func() {
		got, err := agentservice.Uninstall(
			ctx,
			s.deps("darwin", true),
			agentservice.UninstallInput{},
		)
		s.Require().NoError(err)
		s.Empty(got.Removed)
		s.Equal([]string{"launchctl bootout system/io.osapi.agent"}, s.ran)
	})
	s.Run("a unit that cannot be removed is the error", func() {
		s.Require().NoError(s.fs.MkdirAll("/etc/systemd/system", 0o755))
		s.Require().
			NoError(avfs.WriteFile(s.fs, "/etc/systemd/system/osapi-agent.service", []byte("x"), 0o644))
		d := s.deps("linux", true)
		d.FS = noRemove{s.fs}
		_, err := agentservice.Uninstall(ctx, d, agentservice.UninstallInput{})
		s.ErrorContains(err, "remove /etc/systemd/system/osapi-agent.service: read-only")
	})
	s.Run("not root, or no service manager", func() {
		d := s.deps("linux", true)
		d.EUID = func() int { return 501 }
		_, err := agentservice.Uninstall(ctx, d, agentservice.UninstallInput{})
		s.ErrorIs(err, agentservice.ErrNotRoot)
		_, err = agentservice.Uninstall(ctx, s.deps("windows", true), agentservice.UninstallInput{})
		s.ErrorIs(err, agentservice.ErrPlatform)
	})
}
