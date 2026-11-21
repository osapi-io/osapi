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

package updatepackages_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/updatepackages"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestRefresh() {
	for name, tc := range map[string]struct {
		out  string
		want bool
	}{
		"when apt-get gets new data first":  {out: "Get:1 http://archive.ubuntu.com/ubuntu noble InRelease [126 kB]\n", want: true},
		"when apt-get gets new data later":  {out: "Hit:1 http://a noble InRelease\nGet:2 http://a noble-updates InRelease [126 kB]\n", want: true},
		"when each repository is unchanged": {out: "Hit:1 http://archive.ubuntu.com/ubuntu noble InRelease\n"},
	} {
		s.Run(name, func() {
			exec := hostmocks.NewMockExec(gomock.NewController(s.T()))
			exec.EXPECT().
				RunPrivilegedCmd(gomock.Any(), "apt-get", []string{"update"}).
				Return(tc.out, nil)
			p, err := updatepackages.ExportNewDebian(host.Host{Exec: exec})
			s.Require().NoError(err)

			got, err := p.Refresh(context.Background())
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
