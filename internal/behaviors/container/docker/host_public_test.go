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

package docker_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/container/docker"
	"github.com/osapi-io/osapi/internal/secret"
)

const (
	cid     = "abc123"
	psLine  = `{"ID":"abc123","Names":"web","Image":"nginx","State":"%s","Status":"x"}`
	inspect = `[{"Id":"abc123","NetworkSettings":{"Ports":{"80/tcp":[],"443/tcp":[{"HostIp":"0.0.0.0","HostPort":"8443"}]}}}]`
)

type HostPublicTestSuite struct {
	suite.Suite
}

// docker answers as a host whose one container, web, is in state; an empty
// state is a host with no container. calls are the privileged calls the
// test expects, with what each answers.
func (s *HostPublicTestSuite) provider(
	state string,
	enabled bool,
	expect func(m *hostmocks.MockExec),
) docker.Provider {
	m := hostmocks.NewMockExec(gomock.NewController(s.T()))
	m.EXPECT().RunCmd(gomock.Any(), "docker", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, args []string) (string, error) {
			switch strings.Join(args, " ") {
			case "version --format {{.Server.Version}}":
				if state == "down" {
					return "", errors.New("no daemon")
				}
				return "29.1.3\n", nil
			case "ps -a --no-trunc --format {{json .}}":
				if state == "" {
					return "", nil
				}
				if state == "two" {
					return `{"ID":"def456","Names":"web","State":"running"}` + "\n" +
						`{"ID":"abc123","Names":"api","State":"exited"}` + "\n", nil
				}
				return strings.Replace(psLine, "%s", state, 1) + "\n", nil
			case "inspect " + cid:
				return inspect, nil
			}
			return "", nil
		}).AnyTimes()
	if expect != nil {
		expect(m)
	}
	p, err := docker.NewCLI(host.Host{Exec: m, CommandEnabled: enabled})
	s.Require().NoError(err)

	return p
}

func (s *HostPublicTestSuite) TestReads() {
	s.Run("a host whose docker does not answer is skipped", func() {
		_, err := s.provider("down", true, nil).List(context.Background(), "")
		s.ErrorIs(err, behavior.ErrUnsupported)
		_, err = s.provider("down", true, nil).Get(context.Background(), "web")
		s.ErrorIs(err, behavior.ErrUnsupported)
	})
	s.Run(
		"running leaves out a stopped container, and ports without a host port show alone",
		func() {
			got, err := s.provider("exited", true, nil).List(context.Background(), "running")
			s.Require().NoError(err)
			s.Empty(got)
			got, err = s.provider("running", true, nil).List(context.Background(), "")
			s.Require().NoError(err)
			s.Equal([]string{"0.0.0.0:8443->443/tcp", "80/tcp"}, got[0].Ports)
			got, err = s.provider("two", true, nil).List(context.Background(), "")
			s.Require().NoError(err)
			s.Equal([]string{"api", "web"}, []string{got[0].Name, got[1].Name}, "by name")
		},
	)
}

func (s *HostPublicTestSuite) TestCreate() {
	ctx := context.Background()
	boom := errors.New("boom")
	full := docker.CreateInput{
		Name: "web", Image: "nginx", HostName: "box", DNS: []string{"1.1.1.1"},
		Ports: []string{"8080:80"}, Volumes: []string{"/a:/b"},
		Env: map[string]secret.Value{"B": "2", "A": "1"}, Command: []string{"run"}, AutoStart: true,
	}

	s.Run("each option goes to docker create, the environment on standard input", func() {
		p := s.provider("", true, func(m *hostmocks.MockExec) {
			m.EXPECT().RunPrivilegedCmdWithStdin(gomock.Any(), "docker", []string{
				"create", "--name", "web", "--hostname", "box", "--dns", "1.1.1.1",
				"-p", "8080:80", "-v", "/a:/b", "--env-file", "/dev/stdin", "nginx", "run",
			}, "A=1\nB=2\n").Return(cid+"\n", nil)
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"start", "web"}).
				Return("", nil)
		})
		got, err := p.Create(ctx, full)
		s.Require().NoError(err)
		s.Equal(docker.Result{Name: "web", ID: cid, Changed: true}, got)
	})
	s.Run("a create that fails names the container", func() {
		p := s.provider("", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", gomock.Any()).
				Return("no such image", boom)
		})
		_, err := p.Create(ctx, docker.CreateInput{Name: "web", Image: "nginx"})
		s.ErrorIs(err, boom)
		s.ErrorContains(err, "no such image")
	})
	s.Run("a start after the create that fails is the error", func() {
		p := s.provider("", true, func(m *hostmocks.MockExec) {
			m.EXPECT().RunPrivilegedCmd(gomock.Any(), "docker", gomock.Any()).Return(cid, nil)
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"start", "web"}).
				Return("", boom)
		})
		_, err := p.Create(ctx, docker.CreateInput{Name: "web", Image: "nginx", AutoStart: true})
		s.ErrorIs(err, boom)
	})
	s.Run("a container that exists and is stopped is started when auto_start is set", func() {
		p := s.provider("exited", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"start", "web"}).
				Return("", nil)
		})
		got, err := p.Create(ctx, docker.CreateInput{Name: "web", Image: "nginx", AutoStart: true})
		s.Require().NoError(err)
		s.Equal(docker.Result{Name: "web", ID: cid, Changed: true}, got)
	})
	s.Run("a start of a container that exists that fails is the error", func() {
		p := s.provider("exited", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"start", "web"}).
				Return("", boom)
		})
		_, err := p.Create(ctx, docker.CreateInput{Name: "web", Image: "nginx", AutoStart: true})
		s.ErrorIs(err, boom)
	})
	s.Run("a host whose docker does not answer is skipped", func() {
		_, err := s.provider("down", true, nil).
			Create(ctx, docker.CreateInput{Name: "web", Image: "nginx"})
		s.ErrorIs(err, behavior.ErrUnsupported)
	})
}

func (s *HostPublicTestSuite) TestChanges() {
	ctx := context.Background()
	boom := errors.New("boom")

	s.Run("delete runs docker rm -f, and its failure is the error", func() {
		p := s.provider("running", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"rm", "-f", "web"}).
				Return("busy", boom)
		})
		_, err := p.Delete(ctx, "web")
		s.ErrorIs(err, boom)
	})
	s.Run("delete on a host whose docker does not answer is skipped", func() {
		_, err := s.provider("down", true, nil).Delete(ctx, "web")
		s.ErrorIs(err, behavior.ErrUnsupported)
	})
	s.Run("start and stop that fail are the error", func() {
		p := s.provider("exited", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"start", "web"}).
				Return("", boom)
		})
		_, err := p.Start(ctx, "web")
		s.ErrorIs(err, boom)
		p = s.provider("running", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "docker", []string{"stop", "web"}).
				Return("", boom)
		})
		_, err = p.Stop(ctx, "web", 0)
		s.ErrorIs(err, boom)
	})
}

func (s *HostPublicTestSuite) TestExec() {
	ctx := context.Background()

	s.Run("an agent that runs no caller's command refuses it", func() {
		_, err := s.provider("running", false, nil).
			Exec(ctx, docker.ExecInput{Name: "web", Command: []string{"id"}})
		s.ErrorContains(err, "agent.command.enabled")
	})
	s.Run(
		"the environment and the directory go to docker exec, and a code other than 0 is an answer",
		func() {
			p := s.provider("running", true, func(m *hostmocks.MockExec) {
				m.EXPECT().RunPrivilegedCmdFull(gomock.Any(), "docker", []string{
					"exec", "-e", "A=1", "-e", "B=2", "-w", "/tmp", "web", "false",
				}).Return("out", "err", 1, nil)
			})
			got, err := p.Exec(ctx, docker.ExecInput{
				Name: "web", Command: []string{"false"}, WorkingDir: "/tmp",
				Env: map[string]secret.Value{"B": "2", "A": "1"},
			})
			s.Require().NoError(err)
			s.Equal(
				docker.ExecResult{Stdout: "out", Stderr: "err", ExitCode: 1, Changed: true},
				got,
			)
		},
	)
	s.Run("a command that docker cannot run is the error", func() {
		boom := errors.New("boom")
		p := s.provider("running", true, func(m *hostmocks.MockExec) {
			m.EXPECT().
				RunPrivilegedCmdFull(gomock.Any(), "docker", gomock.Any()).
				Return("", "", -1, boom)
		})
		_, err := p.Exec(ctx, docker.ExecInput{Name: "web", Command: []string{"x"}})
		s.ErrorIs(err, boom)
	})
}

func TestHostPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HostPublicTestSuite))
}
