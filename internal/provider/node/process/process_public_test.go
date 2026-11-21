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

package process_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/provider"
	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/internal/provider/node/process"
	"github.com/osapi-io/osapi/internal/provider/node/process/mocks"
)

type ProcessPublicTestSuite struct {
	suite.Suite
}

func TestProcessPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessPublicTestSuite))
}

// host is a Linux host with one process, 7, started 5 s after boot.
func (s *ProcessPublicTestSuite) host(
	signaler process.Signaler,
) *process.Gohai {
	g := process.New(slog.Default(), signaler)
	g.Sources = gohaitest.Linux(s.T(), map[string]string{
		"/proc/stat":      "btime 1700000000\n",
		"/proc/meminfo":   "MemTotal: 4096 kB\n",
		"/etc/passwd":     "root:x:0:0::/root:/bin/sh\n",
		"/proc/7/stat":    "7 (nginx) S 1 7 7 0 -1 4194560 1 2 3 4 300 200 7 8 20 0 1 0 500 9 10\n",
		"/proc/7/status":  "Uid:\t0\t0\t0\t0\nVmRSS:\t1024 kB\n",
		"/proc/7/cmdline": "nginx\x00-g\x00daemon off;\x00",
	}, nil)
	g.Sources.Now = func() time.Time { return time.Unix(1700000015, 0) }

	return g
}

func (s *ProcessPublicTestSuite) want() process.Info {
	return process.Info{
		PID: 7, Name: "nginx", User: "root", State: "sleep",
		CPUPercent: 50, MemPercent: 25, MemRSS: 1024 * 1024,
		Command: "nginx -g daemon off;", StartTime: "2023-11-14T22:13:25Z",
	}
}

func (s *ProcessPublicTestSuite) TestList() {
	got, err := s.host(nil).List(context.Background())
	s.Require().NoError(err)
	s.Equal([]process.Info{s.want()}, got)

	g := process.New(slog.Default(), nil)
	g.Sources = gohaitest.Linux(s.T(), nil, nil)
	_, err = g.List(context.Background())
	s.ErrorContains(err, "process: list")
}

func (s *ProcessPublicTestSuite) TestGet() {
	got, err := s.host(nil).Get(context.Background(), 7)
	s.Require().NoError(err)
	s.Equal(s.want(), *got)

	_, err = s.host(nil).Get(context.Background(), 8)
	s.ErrorIs(err, provider.ErrNotFound)

	_, err = s.host(nil).Get(context.Background(), 0)
	s.ErrorContains(err, "must be greater than zero")
}

func (s *ProcessPublicTestSuite) TestSignal() {
	tests := []struct {
		name    string
		pid     int
		signal  string
		killErr error
		kill    bool
		wantErr string
		notFind bool
	}{
		{name: "sends TERM", pid: 7, signal: "TERM", kill: true},
		{name: "a pid below one", pid: 0, signal: "TERM", wantErr: "must be greater than zero"},
		{
			name:    "a signal that is not allowed",
			pid:     7,
			signal:  "STOP",
			wantErr: `invalid signal "STOP"`,
		},
		{
			name:    "no such process",
			pid:     7,
			signal:  "TERM",
			kill:    true,
			killErr: syscall.ESRCH,
			notFind: true,
		},
		{
			name:    "not permitted",
			pid:     7,
			signal:  "TERM",
			kill:    true,
			killErr: syscall.EPERM,
			wantErr: "permission denied",
		},
		{
			name:    "another error",
			pid:     7,
			signal:  "TERM",
			kill:    true,
			killErr: errors.New("boom"),
			wantErr: "boom",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			sig := mocks.NewMockSignaler(gomock.NewController(s.T()))
			if tt.kill {
				sig.EXPECT().Kill(tt.pid, syscall.SIGTERM).Return(tt.killErr)
			}
			got, err := s.host(sig).Signal(context.Background(), tt.pid, tt.signal)
			switch {
			case tt.notFind:
				s.ErrorIs(err, provider.ErrNotFound)
			case tt.wantErr != "":
				s.ErrorContains(err, tt.wantErr)
			default:
				s.Require().NoError(err)
				s.Equal(&process.SignalResult{PID: 7, Signal: "TERM", Changed: true}, got)
			}
		})
	}
}

func (s *ProcessPublicTestSuite) TestSyscallSignaler() {
	s.NoError(
		process.NewSyscallSignaler().Kill(os.Getpid(), 0),
		"signal 0 only checks the process is there",
	)
}
