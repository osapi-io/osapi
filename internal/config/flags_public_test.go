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

package config_test

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type FlagsPublicTestSuite struct {
	suite.Suite
}

func (s *FlagsPublicTestSuite) TestFlagName() {
	s.Equal("agent-consumer-back-off", config.FlagName("agent.consumer.back_off"))
}

func (s *FlagsPublicTestSuite) TestBindFlags() {
	tests := []struct {
		name         string
		key          string
		validateFunc func(f *pflag.Flag)
	}{
		{
			name: "when the key is an int, the flag is one, with its default",
			key:  "agent.max_jobs",
			validateFunc: func(f *pflag.Flag) {
				s.Equal("int", f.Value.Type())
				s.Equal("10", f.DefValue)
				s.Equal("Maximum number of concurrent jobs to process (agent.max_jobs)", f.Usage)
			},
		},
		{
			name: "when the key is a string, the flag is one",
			key:  "agent.consumer.ack_wait",
			validateFunc: func(f *pflag.Flag) {
				s.Equal("string", f.Value.Type())
				s.Equal("2m", f.DefValue)
			},
		},
		{
			name: "when the key is a list, the flag is given once for each value",
			key:  "agent.consumer.back_off",
			validateFunc: func(f *pflag.Flag) {
				s.Equal("stringSlice", f.Value.Type())
				s.Equal("[30s,2m,5m,15m,30m]", f.DefValue)
			},
		},
		{
			name: "when the key is a bool, the flag is one",
			key:  "agent.command.enabled",
			validateFunc: func(f *pflag.Flag) {
				s.Equal("bool", f.Value.Type())
			},
		},
		{
			name: "when the key is a float, the flag is one",
			key:  "agent.conditions.high_load_multiplier",
			validateFunc: func(f *pflag.Flag) {
				s.Equal("float64", f.Value.Type())
				s.Equal("2", f.DefValue)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
			config.BindFlags(fs, tc.key)
			f := fs.Lookup(config.FlagName(tc.key))
			s.Require().NotNil(f)
			tc.validateFunc(f)
		})
	}
}

// TestBindFlagsSets holds that a flag a person gives sets its key.
func (s *FlagsPublicTestSuite) TestBindFlagAs() {
	fs := pflag.NewFlagSet("client", pflag.ContinueOnError)
	config.BindFlagAs(fs, "controller.client.url", "url")

	f := fs.Lookup("url")
	s.Require().NotNil(f)
	s.Equal("http://localhost:8080", f.DefValue, "the default of the owner, not a copy")
	s.Panics(func() { config.BindFlagAs(fs, "no.such.key", "nope") })
}

func (s *FlagsPublicTestSuite) TestBindFlagsSets() {
	fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
	config.BindFlags(fs, "agent.max_jobs")
	s.Require().NoError(fs.Parse([]string{"--agent-max-jobs", "7"}))
	s.Equal(7, viper.GetInt("agent.max_jobs"))
}

func (s *FlagsPublicTestSuite) TestBindFlagsUnknown() {
	s.PanicsWithValue(`config: no key "agent.nosuch" for a flag`, func() {
		config.BindFlags(pflag.NewFlagSet("t", pflag.ContinueOnError), "agent.nosuch")
	})
}

func TestFlagsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(FlagsPublicTestSuite))
}
