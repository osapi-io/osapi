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

package image_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/container/image"
)

type HostPublicTestSuite struct {
	suite.Suite
}

func (s *HostPublicTestSuite) TestProvider() {
	ctx := context.Background()
	boom := errors.New("boom")
	inspect := []string{"image", "inspect", "--format", "{{.Id}} {{.Size}}", "alpine"}

	s.Run("a host whose docker does not answer is skipped", func() {
		m := hostmocks.NewMockExec(gomock.NewController(s.T()))
		behaviortest.ReadsThroughRunCmd(m)
		m.EXPECT().RunCmd(gomock.Any(), "docker", gomock.Any()).Return("", boom).AnyTimes()
		p, _ := image.NewCLI(host.Host{Exec: m})
		_, err := p.List(ctx)
		s.ErrorIs(err, behavior.ErrUnsupported)
	})
	s.Run("the images come by repository and tag", func() {
		m := hostmocks.NewMockExec(gomock.NewController(s.T()))
		behaviortest.ReadsThroughRunCmd(m)
		m.EXPECT().
			RunCmd(gomock.Any(), "docker", []string{"version", "--format", "{{.Server.Version}}"}).
			Return("29\n", nil)
		m.EXPECT().RunCmd(gomock.Any(), "docker", []string{"images", "--format", "{{json .}}"}).
			Return(`{"Repository":"nginx","Tag":"1"}`+"\n"+`{"Repository":"alpine","Tag":"3"}`+"\n", nil)
		m.EXPECT().RunCmd(gomock.Any(), "docker", gomock.Any()).Return("", nil).AnyTimes()
		p, _ := image.NewCLI(host.Host{Exec: m})
		got, err := p.List(ctx)
		s.Require().NoError(err)
		s.Equal([]string{"alpine", "nginx"}, []string{got[0].Repository, got[1].Repository})
	})
	s.Run("a pull that fails is the error", func() {
		m := hostmocks.NewMockExec(gomock.NewController(s.T()))
		behaviortest.ReadsThroughRunCmd(m)
		m.EXPECT().RunPrivilegedCmd(gomock.Any(), "docker", inspect).Return("", boom)
		m.EXPECT().
			RunPrivilegedCmd(gomock.Any(), "docker", []string{"pull", "alpine"}).
			Return("denied", boom)
		p, _ := image.NewCLI(host.Host{Exec: m})
		_, err := p.Pull(ctx, "alpine")
		s.ErrorIs(err, boom)
	})
	s.Run("delete with force runs docker rmi -f, and its failure is the error", func() {
		m := hostmocks.NewMockExec(gomock.NewController(s.T()))
		behaviortest.ReadsThroughRunCmd(m)
		m.EXPECT().RunPrivilegedCmd(gomock.Any(), "docker", inspect).Return("sha256:x 1\n", nil)
		m.EXPECT().
			RunPrivilegedCmd(gomock.Any(), "docker", []string{"rmi", "-f", "alpine"}).
			Return("in use", boom)
		p, _ := image.NewCLI(host.Host{Exec: m})
		_, err := p.Delete(ctx, "alpine", true)
		s.ErrorIs(err, boom)
	})
}

func TestHostPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HostPublicTestSuite))
}
