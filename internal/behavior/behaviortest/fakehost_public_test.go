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

package behaviortest_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// probe reaches the fake host one way per op, and reports what it saw, so
// each part of the fake host is checked through Check as a provider uses it.
type probe struct{ h host.Host }

type probeIn struct {
	Op string `json:"op"`
}

type probeOut struct {
	Seen string `json:"seen"`
}

func (p probe) run(
	ctx context.Context,
	op string,
) (string, error) {
	kv := p.h.FileState
	switch op {
	case "stdin":
		return p.h.Exec.RunPrivilegedCmdWithStdin(ctx, "tee", []string{"/x"}, "secret")
	case "full":
		out, _, code, err := p.h.Exec.RunPrivilegedCmdFull(ctx, "tee", []string{"/y"})

		return fmt.Sprintf("%s code=%d", out, code), err
	case "fullcmd":
		out, err := p.h.Exec.RunCmdFull(ctx, "any-program", []string{"x"}, "/tmp", 5)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s exit=%d", out.Stdout, out.ExitCode), nil
	case "store":
		if _, err := kv.KVPut(ctx, "file-state", "a", []byte("1")); err != nil {
			return "", err
		}
		if _, err := kv.KVPut(ctx, "file-state", "a", []byte("2")); err != nil {
			return "", err
		}
		e, err := kv.KVGet(ctx, "file-state", "a")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s=%s rev=%d zero=%t", e.Key(), e.Value(), e.Revision(),
			e.Created().IsZero()), nil
	default:
		_, err := kv.KVGet(ctx, "file-state", "missing")

		return fmt.Sprintf("missing=%t", errors.Is(err, jetstream.ErrKeyNotFound)), nil
	}
}

func probeHost(
	ctx context.Context,
	p probe,
	in probeIn,
) (probeOut, error) {
	seen, err := p.run(ctx, in.Op)

	return probeOut{Seen: seen}, err
}

type FakeHostPublicTestSuite struct {
	suite.Suite
}

func (s *FakeHostPublicTestSuite) TestFakeHost() {
	tests := []struct {
		name    string
		in      probeIn
		want    string
		options []behavior.HostOption
		wantErr string
	}{
		{
			name:    "when a command is given input, it is recorded",
			in:      probeIn{Op: "stdin"},
			options: []behavior.HostOption{behavior.Ran("tee", "/x")},
		},
		{
			name:    "when a command's output is given apart, it is recorded",
			in:      probeIn{Op: "full"},
			want:    " code=0",
			options: []behavior.HostOption{behavior.Ran("tee", "/y")},
		},
		{
			name: "when a program of Runs(AnyProgram) runs in full, its output and exit code answer",
			in:   probeIn{Op: "fullcmd"},
			want: "out exit=2",
			options: []behavior.HostOption{
				behavior.Prints("any-program x", "out"),
				behavior.Exits("any-program x", 2),
				behavior.Ran("any-program", "x"),
			},
		},
		{
			name: "when a value is put, put again and read",
			in:   probeIn{Op: "store"},
			want: "a=2 rev=2 zero=true",
		},
		{
			name: "when a key is not there, the read says so",
			in:   probeIn{Op: "missing"},
			want: "missing=true",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("probe", "Reaches the fake host", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (probe, error) {
					return probe{h: h}, nil
				}),
				behavior.Sudo("tee"),
				behavior.Runs(behavior.AnyProgram),
				behavior.KeepsState,
				behavior.Read("get", probeHost).Describe("Probe").
					Example("A probe", tc.in, probeOut{Seen: tc.want}, tc.options...),
			)

			problems := behaviortest.Check(b)

			if tc.wantErr != "" {
				s.Require().NotEmpty(problems)
				s.Contains(problems[0], tc.wantErr)
				s.Contains(problems[0], "Watch")

				return
			}
			s.Empty(problems, strings.Join(problems, "\n"))
		})
	}
}

func TestFakeHostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FakeHostPublicTestSuite))
}

// TestClock holds the fake host's clock to Now, so an example that reads
// the time has one answer.
func (s *FakeHostPublicTestSuite) TestClock() {
	s.Equal(behaviortest.Now, behaviortest.ExportClock())
}

func (s *FakeHostPublicTestSuite) TestReport() {
	for _, fatal := range []bool{false, true} {
		s.PanicsWithValue(
			"the fake host has no answer for a call: Exec.Unknown",
			func() { behaviortest.ExportReport(fatal) },
		)
	}
}
