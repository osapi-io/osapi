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

package source_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/log/source"
)

type LinuxPublicTestSuite struct {
	suite.Suite
}

func (s *LinuxPublicTestSuite) TestSources() {
	exec := hostmocks.NewMockExec(gomock.NewController(s.T()))
	exec.EXPECT().RunCmd(gomock.Any(), "journalctl", []string{"--field=SYSLOG_IDENTIFIER"}).
		Return("", errors.New("exit 1"))
	p, err := source.ExportNewLinux(host.Host{Exec: exec})
	s.Require().NoError(err)

	_, err = p.Sources(context.Background())
	s.EqualError(err, "read the journal's sources: exit 1")
}

func (s *LinuxPublicTestSuite) TestContainer() {
	_, err := source.ExportNewLinux(host.Host{Container: true})
	s.ErrorIs(err, behavior.ErrUnsupported, "a container has no journald of its own")
}

func TestLinuxPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LinuxPublicTestSuite))
}
