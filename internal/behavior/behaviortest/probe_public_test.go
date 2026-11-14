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
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// conf writes one key's line to its own file under /etc/conf.d and reloads
// it. safe says whether it checks its input first.
type conf struct {
	h    host.Host
	safe bool
}

type confIn struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type taggedIn struct {
	Key   string `json:"key"   validate:"alphanum"`
	Value string `json:"value" validate:"alphanum"`
}

type confOut struct {
	Changed bool `json:"changed"`
}

var confKey = regexp.MustCompile(`^[a-z]+$`)

func (c conf) set(
	ctx context.Context,
	key string,
	value string,
) (confOut, error) {
	if c.safe && (!confKey.MatchString(key) || strings.ContainsAny(value, "\r\n")) {
		return confOut{}, errors.New("refused")
	}
	path := "/etc/conf.d/" + key + ".conf"
	_ = c.h.FS.MkdirAll("/etc/conf.d", 0o755)
	_ = c.h.FS.WriteFile(path, []byte(key+" = "+value+"\n"), 0o644)
	if c.safe {
		_, _ = c.h.Exec.RunPrivilegedCmd(ctx, "reload", []string{"--", key})

		return confOut{Changed: true}, nil
	}
	_, _ = c.h.Exec.RunPrivilegedCmd(ctx, "reload", []string{key})
	_, _ = c.h.Exec.RunPrivilegedCmd(ctx, "sh", []string{"-c", "echo " + value})

	return confOut{Changed: true}, nil
}

func setConf(
	ctx context.Context,
	c conf,
	in confIn,
) (confOut, error) {
	return c.set(ctx, in.Key, in.Value)
}

func setTagged(
	ctx context.Context,
	c conf,
	in taggedIn,
) (confOut, error) {
	return c.set(ctx, in.Key, in.Value)
}

func setBare(
	ctx context.Context,
	c conf,
	key string,
) (confOut, error) {
	return c.set(ctx, key, "x")
}

func confBehavior(
	safe bool,
	opts ...behavior.Option,
) *behavior.Behavior {
	effects := []behavior.HostOption{
		behavior.Wrote("/etc/conf.d/a.conf", "a = b\n"),
		behavior.Ran("reload", "a"),
		behavior.Ran("sh", "-c", "echo b"),
	}
	if safe {
		effects = []behavior.HostOption{
			behavior.Wrote("/etc/conf.d/a.conf", "a = b\n"),
			behavior.Ran("reload", "--", "a"),
		}
	}
	opts = append(
		opts,
		behavior.Platform(
			"debian",
			func(h host.Host) (conf, error) { return conf{h: h, safe: safe}, nil },
		),
		behavior.Write("set", setConf).Describe("Set it").
			Example("A key", confIn{Key: "a", Value: "b"}, confOut{Changed: true}, effects...),
	)

	return behavior.Define(
		"conf",
		"Config lines",
		append([]behavior.Option{behavior.Node}, opts...)...)
}

// probed is the problems Check reports from the probes alone.
func probed(
	b *behavior.Behavior,
) []string {
	var out []string
	for _, p := range behaviortest.Check(b) {
		if strings.Contains(p, ": with ") && strings.Contains(p, " in ") {
			out = append(out, p)
		}
	}

	return out
}

type ProbePublicTestSuite struct {
	suite.Suite
}

func (s *ProbePublicTestSuite) TestProbes() {
	tests := []struct {
		name string
		b    *behavior.Behavior
		want []string
	}{
		{
			name: "when the provider trusts its input, each probe that gets through is reported",
			b:    confBehavior(false),
			want: []string{
				`with path traversal in Key, "../../../../osapi-probe": it wrote /osapi-probe.conf, outside the directories the example writes`,
				`with a new line in Key, "a\nosapi_probe = 1": its line reached`,
				`with an option in Key, "--osapi-probe": it reached a command as an option: reload --osapi-probe`,
				`with path traversal in Value, "../../../../osapi-probe": it reached a shell script`,
				`with a new line in Value, "b\nosapi_probe = 1": its line reached /etc/conf.d/a.conf`,
				`with an option in Value, "--osapi-probe": it reached a shell script`,
			},
		},
		{
			name: "when the provider refuses what it cannot use, nothing is reported",
			b:    confBehavior(true),
		},
		{
			name: "when the validate tags refuse every probe, nothing is reported",
			b: behavior.Define(
				"tagged",
				"Tagged",
				behavior.Node,
				behavior.Platform(
					"debian",
					func(h host.Host) (conf, error) { return conf{h: h}, nil },
				),
				behavior.Write("set", setTagged).Describe("Set it").
					Example("A key", taggedIn{Key: "a", Value: "b"}, confOut{Changed: true},
						behavior.Wrote("/etc/conf.d/a.conf", "a = b\n")),
			),
		},
		{
			name: "when the behavior is internal, its input is not probed",
			b:    confBehavior(false, behavior.Internal),
		},
		{
			name: "when the input is not a struct, there is nothing to probe",
			b: behavior.Define(
				"bare",
				"Bare",
				behavior.Node,
				behavior.Platform(
					"debian",
					func(h host.Host) (conf, error) { return conf{h: h}, nil },
				),
				behavior.Write("set", setBare).Describe("Set it").
					Example("A key", "a", confOut{Changed: true},
						behavior.Wrote("/etc/conf.d/a.conf", "a = x\n")),
			),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := probed(tc.b)

			s.Require().Len(got, len(tc.want), strings.Join(got, "\n"))
			for i, want := range tc.want {
				s.Contains(got[i], want)
			}
		})
	}
}

func TestProbePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProbePublicTestSuite))
}
