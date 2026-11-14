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
package behavior_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type SettingsPublicTestSuite struct {
	suite.Suite
}

type goodSettings struct {
	Dir   string `mapstructure:"dir"   doc:"Where."`
	Limit struct {
		Max int `mapstructure:"max" doc:"The most."`
	} `mapstructure:"limit"`
	hidden string
}

type badSettings struct {
	NoName string `doc:"No name."`
	NoDoc  string `               mapstructure:"no_doc"`
	Nested struct {
		Inner int `mapstructure:"inner"`
	} `               mapstructure:"nested"`
}

func (s *SettingsPublicTestSuite) TestSettings() {
	b := behavior.Define(
		"kv",
		"Key values",
		behavior.Node,
		behavior.Config(goodSettings{Dir: "/var/kv"}),
	)
	_ = goodSettings{}.hidden

	s.Equal("behaviors.kv", b.SettingsKey())
	s.Equal(reflect.TypeFor[goodSettings](), b.SettingsType())
	s.Equal("kv-limit-max", b.SettingFlag("limit.max"))
	s.Equal("kv-dry-run", b.SettingFlag("dry_run"))

	s.Equal("/var/kv", behavior.Settings[goodSettings](b).Dir, "before loading, the defaults")
	b.SetSettings(goodSettings{Dir: "/srv/kv"})
	s.Equal("/srv/kv", behavior.Settings[goodSettings](b).Dir, "after loading, the decoded value")

	plain := behavior.Define("plain", "No settings", behavior.Node)
	s.Nil(plain.SettingsType())
	s.Equal(goodSettings{}, behavior.Settings[goodSettings](plain))
}

func (s *SettingsPublicTestSuite) TestCurrentSettings() {
	tests := []struct {
		name    string
		b       *behavior.Behavior
		decoded any
		want    any
	}{
		{
			name: "when the config has not been read, the declared defaults",
			b:    behavior.Define("sysctl", "x", behavior.Config(fakeProvider{prefix: "default"})),
			want: fakeProvider{prefix: "default"},
		},
		{
			name: "when the config has been read, the decoded settings",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Config(fakeProvider{prefix: "default"}),
			),
			decoded: fakeProvider{prefix: "decoded"},
			want:    fakeProvider{prefix: "decoded"},
		},
		{
			name: "when no settings are declared",
			b:    behavior.Define("bare", "x"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.decoded != nil {
				tc.b.SetSettings(tc.decoded)
			}

			s.Equal(tc.want, tc.b.CurrentSettings())
		})
	}
}

func TestSettingsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SettingsPublicTestSuite))
}
