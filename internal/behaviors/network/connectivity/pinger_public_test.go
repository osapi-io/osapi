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

package connectivity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/network/connectivity"
)

type PingerPublicTestSuite struct {
	suite.Suite
}

func (s *PingerPublicTestSuite) TestPing() {
	tests := []struct {
		name     string
		platform behavior.PlatformName
		address  string
		command  string
		args     []string
		out      string
		err      error
		want     connectivity.PingResult
		wantErr  string
	}{
		{
			name:     "when an IPv6 address is pinged on macOS, it is ping6",
			platform: behavior.PlatformDarwin,
			address:  "::1",
			command:  "ping6",
			args:     []string{"-n", "-c", "3", "::1"},
			out:      "3 packets transmitted, 3 packets received, 0.0% packet loss\nround-trip min/avg/max/stddev = 0.1/0.2/0.3/0.1 ms\n",
			want: connectivity.PingResult{
				PacketsSent: 3, PacketsReceived: 3,
				MinRTT: "0.10ms", AvgRTT: "0.20ms", MaxRTT: "0.30ms",
			},
		},
		{
			name:     "when ping exits 1 with its summary, it is the loss",
			platform: behavior.PlatformLinux,
			address:  "192.0.2.1",
			command:  "ping",
			args:     []string{"-n", "-c", "3", "-W", "2", "-w", "4", "192.0.2.1"},
			out:      "3 packets transmitted, 0 received, 100% packet loss, time 2040ms\n",
			err:      errors.New("exit status 1"),
			want:     connectivity.PingResult{PacketsSent: 3, PacketLoss: 100},
		},
		{
			name:     "when ping fails with no summary, it is the error with the output",
			platform: behavior.PlatformLinux,
			address:  "10.255.255.1",
			command:  "ping",
			args:     []string{"-n", "-c", "3", "-W", "2", "-w", "4", "10.255.255.1"},
			out:      "connect: Network is unreachable\n",
			err:      errors.New("exit status 2"),
			wantErr:  "ping 10.255.255.1: exit status 2: connect: Network is unreachable",
		},
		{
			name:     "when ping prints no summary and no error, it is an error",
			platform: behavior.PlatformLinux,
			address:  "10.0.0.1",
			command:  "ping",
			args:     []string{"-n", "-c", "3", "-W", "2", "-w", "4", "10.0.0.1"},
			out:      "",
			wantErr:  "ping 10.0.0.1: no summary in its output",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunCmd(gomock.Any(), tc.command, tc.args).Return(tc.out, tc.err)
			p, err := connectivity.Behavior.Provider(tc.platform, host.Host{Exec: exec})
			s.Require().NoError(err)

			got, err := p.(connectivity.Provider).Ping(context.Background(), tc.address)

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestPingerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PingerPublicTestSuite))
}
