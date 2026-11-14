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
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
)

// kvSettings is a behavior's declared settings, for these tests.
type kvSettings struct {
	Dir   string `mapstructure:"dir"   doc:"Where it writes."  validate:"required"`
	Level int    `mapstructure:"level" doc:"How much it says." validate:"max=5"`
	Limit struct {
		Max int `mapstructure:"max" doc:"The most it keeps."`
	} `mapstructure:"limit"`
}

// registerKV registers a behavior with settings and one without, and returns
// the first.
func registerKV() *behavior.Behavior {
	behavior.Reset()
	defaults := kvSettings{Dir: "/var/kv", Level: 1}
	defaults.Limit.Max = 10
	b := behavior.Register(
		behavior.Define("kv", "Key values", behavior.Node, behavior.Config(defaults)),
	)
	behavior.Register(behavior.Define("plain", "No settings", behavior.Node))

	return b
}

type KeysPublicTestSuite struct {
	suite.Suite
}

func (s *KeysPublicTestSuite) SetupTest() {
	registerKV()
}

func (s *KeysPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *KeysPublicTestSuite) TestEnvName() {
	s.Equal("OSAPI_CONTROLLER_API_PORT", config.EnvName("controller.api.port"))
}

func (s *KeysPublicTestSuite) TestKeys() {
	keys := config.Keys()
	byName := map[string]config.Key{}
	for _, k := range keys {
		byName[k.Name] = k
	}

	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "when every key has a description",
			validateFunc: func() {
				for _, k := range keys {
					s.NotEmpty(k.Doc, k.Name)
				}
			},
		},
		{
			name: "when every default names a real key",
			validateFunc: func() {
				for name := range config.Defaults {
					_, ok := byName[name]
					s.True(ok, "default for %s, which is not a key", name)
				}
			},
		},
		{
			name: "when a key carries its type, default, env and constraints",
			validateFunc: func() {
				s.Equal(config.Key{
					Name:     "controller.api.port",
					Env:      "OSAPI_CONTROLLER_API_PORT",
					Type:     "int",
					Default:  8080,
					Validate: "min=1,max=65535",
					Doc:      byName["controller.api.port"].Doc,
				}, byName["controller.api.port"])
			},
		},
		{
			name: "when lists, maps, floats, bools and pointers have their types",
			validateFunc: func() {
				s.Equal("list", byName["controller.api.security.cors.allow_origins"].Type)
				s.Equal("map", byName["agent.labels"].Type)
				s.Equal("float", byName["agent.conditions.high_load_multiplier"].Type)
				s.Equal("bool", byName["controller.ui.enabled"].Type)
			},
		},
		{
			name: "when passwords, keys and tokens are secret",
			validateFunc: func() {
				s.True(byName["controller.api.security.signing_key"].Secret)
				s.True(byName["controller.client.security.bearer_token"].Secret)
				s.True(byName["agent.nats.auth.password"].Secret)
				s.False(byName["controller.api.port"].Secret)
			},
		},
		{
			name: "when a field has no mapstructure name or is unexported, it is not a key",
			validateFunc: func() {
				type settings struct {
					Port     int `mapstructure:"port"`
					Internal string
					hidden   string `mapstructure:"hidden"`
				}
				_ = settings{}.hidden

				got := config.KeysOf(reflect.TypeFor[settings]())
				s.Len(got, 1)
				s.Equal("port", got[0].Name)
			},
		},
		{
			name: "when a behavior declares settings, they are keys under behaviors.<name>",
			validateFunc: func() {
				s.Equal(config.Key{
					Name:     "behaviors.kv.dir",
					Env:      "OSAPI_BEHAVIORS_KV_DIR",
					Type:     "string",
					Default:  "/var/kv",
					Validate: "required",
					Doc:      "Where it writes.",
					Behavior: "kv",
					Flag:     "kv-dir",
				}, byName["behaviors.kv.dir"])
				s.Equal("kv-limit-max", byName["behaviors.kv.limit.max"].Flag)
				s.Equal(10, byName["behaviors.kv.limit.max"].Default)
			},
		},
		{
			name: "when a behavior declares no settings, it adds no keys",
			validateFunc: func() {
				for name := range byName {
					s.False(strings.HasPrefix(name, "behaviors.plain"), name)
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, tc.validateFunc)
	}
}

func TestKeysPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeysPublicTestSuite))
}
