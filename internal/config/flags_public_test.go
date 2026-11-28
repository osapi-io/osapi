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
	"github.com/osapi-io/osapi/internal/wiring"
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
		validateFunc func(fs *pflag.FlagSet, wired error)
	}{
		{
			name: "when the key is an int, the flag is one, with its default",
			key:  "agent.max_jobs",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				f := fs.Lookup(config.FlagName("agent.max_jobs"))
				s.Require().NotNil(f)
				s.NoError(wired)
				s.Equal("int", f.Value.Type())
				s.Equal("10", f.DefValue)
				s.Equal("Maximum number of concurrent jobs to process (agent.max_jobs)", f.Usage)
			},
		},
		{
			name: "when the key is a string, the flag is one",
			key:  "agent.consumer.ack_wait",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				f := fs.Lookup(config.FlagName("agent.consumer.ack_wait"))
				s.Require().NotNil(f)
				s.NoError(wired)
				s.Equal("string", f.Value.Type())
				s.Equal("2m", f.DefValue)
			},
		},
		{
			name: "when the key is a list, the flag is given once for each value",
			key:  "agent.consumer.back_off",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				f := fs.Lookup(config.FlagName("agent.consumer.back_off"))
				s.Require().NotNil(f)
				s.NoError(wired)
				s.Equal("stringSlice", f.Value.Type())
				s.Equal("[30s,2m,5m,15m,30m]", f.DefValue)
			},
		},
		{
			name: "when the key is a bool, the flag is one",
			key:  "agent.command.enabled",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				f := fs.Lookup(config.FlagName("agent.command.enabled"))
				s.Require().NotNil(f)
				s.NoError(wired)
				s.Equal("bool", f.Value.Type())
			},
		},
		{
			name: "when the key is a float, the flag is one",
			key:  "agent.conditions.high_load_multiplier",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				f := fs.Lookup(config.FlagName("agent.conditions.high_load_multiplier"))
				s.Require().NotNil(f)
				s.NoError(wired)
				s.Equal("float64", f.Value.Type())
				s.Equal("2", f.DefValue)
			},
		},
		{
			name: "when a person gives the flag, it sets its key",
			key:  "agent.max_jobs",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				s.NoError(wired)
				s.Require().NoError(fs.Parse([]string{"--agent-max-jobs", "7"}))
				s.Equal(7, viper.GetInt("agent.max_jobs"))
			},
		},
		{
			name: "when the key does not exist, it is a wiring mistake and no flag",
			key:  "agent.nosuch",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				s.EqualError(wired, `config: no key "agent.nosuch" for a flag`)
				s.Nil(fs.Lookup(config.FlagName("agent.nosuch")))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_ = wiring.Take()
			fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
			config.BindFlags(fs, tc.key)
			tc.validateFunc(fs, wiring.Take())
		})
	}
}

func (s *FlagsPublicTestSuite) TestBindFlagAs() {
	tests := []struct {
		name         string
		key          string
		flag         string
		validateFunc func(fs *pflag.FlagSet, wired error)
	}{
		{
			name: "when the key exists, the flag has its name and the owner's default",
			key:  "controller.client.url",
			flag: "url",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				s.NoError(wired)
				f := fs.Lookup("url")
				s.Require().NotNil(f)
				s.Equal("https://localhost:8080", f.DefValue)
			},
		},
		{
			name: "when the key does not exist, it is a wiring mistake and no flag",
			key:  "no.such.key",
			flag: "nope",
			validateFunc: func(fs *pflag.FlagSet, wired error) {
				s.EqualError(wired, `config: no key "no.such.key" for a flag`)
				s.Nil(fs.Lookup("nope"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_ = wiring.Take()
			fs := pflag.NewFlagSet("client", pflag.ContinueOnError)
			config.BindFlagAs(fs, tc.key, tc.flag)
			tc.validateFunc(fs, wiring.Take())
		})
	}
}

func TestFlagsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(FlagsPublicTestSuite))
}
