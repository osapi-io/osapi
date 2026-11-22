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

package entry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/log/entry"
)

type LinuxPublicTestSuite struct {
	suite.Suite
}

// TestEntries holds the parsing that the hand-written log provider's tests
// held: a malformed line is left out, and a timestamp or a priority that is
// not a number is kept as journald gave it.
func (s *LinuxPublicTestSuite) TestEntries() {
	tests := []struct {
		name    string
		out     string
		err     error
		want    []entry.Entry
		wantErr string
	}{
		{
			name: "when a line is not JSON it is left out",
			out:  "not json\n" + `{"MESSAGE":"ok","PRIORITY":"3"}` + "\n\n",
			want: []entry.Entry{{Message: "ok", Priority: "err"}},
		},
		{
			name: "when the time and the priority are not numbers they stay as given",
			out:  `{"__REALTIME_TIMESTAMP":"yesterday","PRIORITY":"weird","MESSAGE":"m","_PID":"x"}`,
			want: []entry.Entry{{Timestamp: "yesterday", Priority: "weird", Message: "m"}},
		},
		{
			name: "when the journal has no entries",
			want: []entry.Entry{},
		},
		{
			name:    "when journalctl fails",
			err:     errors.New("exit 1"),
			wantErr: "read the journal: exit 1",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			exec := hostmocks.NewMockExec(gomock.NewController(s.T()))
			exec.EXPECT().
				RunCmd(gomock.Any(), "journalctl", []string{"--output=json", "-n", "100"}).
				Return(tc.out, tc.err)
			p, err := entry.ExportNewLinux(host.Host{Exec: exec})
			s.Require().NoError(err)

			got, err := p.Entries(context.Background(), entry.ListInput{})
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *LinuxPublicTestSuite) TestContainer() {
	_, err := entry.ExportNewLinux(host.Host{Container: true})
	s.ErrorIs(err, behavior.ErrUnsupported, "a container has no journald of its own")
}

func TestLinuxPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LinuxPublicTestSuite))
}
