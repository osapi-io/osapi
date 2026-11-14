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
	case "store":
		if _, err := kv.PutString(ctx, "a", "1"); err != nil {
			return "", err
		}
		if _, err := kv.Update(ctx, "a", []byte("2"), 0); err != nil {
			return "", err
		}
		if _, err := kv.Create(ctx, "a", []byte("3")); !errors.Is(err, jetstream.ErrKeyExists) {
			return "", fmt.Errorf("create over a value: %w", err)
		}
		if _, err := kv.Create(ctx, "b", []byte("4")); err != nil {
			return "", err
		}
		e, err := kv.Get(ctx, "a")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s=%s rev=%d zero=%t", e.Key(), e.Value(), e.Revision(),
			e.Created().IsZero()), nil
	case "remove":
		if _, err := kv.Put(ctx, "a", []byte("1")); err != nil {
			return "", err
		}
		if err := kv.Delete(ctx, "a"); err != nil {
			return "", err
		}
		if _, err := kv.Put(ctx, "b", []byte("1")); err != nil {
			return "", err
		}
		if err := kv.Purge(ctx, "b"); err != nil {
			return "", err
		}
		_, errA := kv.Get(ctx, "a")
		_, errB := kv.Get(ctx, "b")

		return fmt.Sprintf("a=%t b=%t", errors.Is(errA, jetstream.ErrKeyNotFound),
			errors.Is(errB, jetstream.ErrKeyNotFound)), nil
	default:
		_, err := kv.Watch(ctx, "a")

		return "", err
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
			name: "when values are put, updated, created and read",
			in:   probeIn{Op: "store"},
			want: "a=2 rev=3 zero=true",
		},
		{
			name: "when values are deleted and purged",
			in:   probeIn{Op: "remove"},
			want: "a=true b=true",
		},
		{
			name:    "when the provider makes a call the fake host has no answer for",
			in:      probeIn{Op: "watch"},
			wantErr: "the fake host does not support that call",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("probe", "Reaches the fake host", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (probe, error) {
					return probe{h: h}, nil
				}),
				behavior.Sudo("tee"),
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
