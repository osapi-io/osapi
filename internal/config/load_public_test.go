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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
)

// validConfig is a config file holding the two required keys, then extra.
func validConfig(
	extra ...string,
) string {
	return strings.Join(append([]string{
		"controller:",
		"  api:",
		"    security:",
		"      signing_key: " + testSigningKey,
		"  client:",
		"    security:",
		"      bearer_token: token",
	}, extra...), "\n")
}

type LoadPublicTestSuite struct {
	suite.Suite

	kv *behavior.Behavior
}

func (s *LoadPublicTestSuite) SetupTest() {
	s.kv = registerKV()
}

func (s *LoadPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *LoadPublicTestSuite) TearDownSubTest() {
	config.ResetRegisterValidatorsFn()
}

// write puts body in a config file and returns its path.
func (s *LoadPublicTestSuite) write(
	body string,
) string {
	path := filepath.Join(s.T().TempDir(), "osapi.yaml")
	s.Require().NoError(os.WriteFile(path, []byte(body), 0o600))

	return path
}

func (s *LoadPublicTestSuite) TestLoad() {
	tests := []struct {
		name         string
		body         string
		missing      bool
		env          map[string]string
		register     func(*validator.Validate) error
		wantErr      string
		validateFunc func(v *viper.Viper, c config.Config)
	}{
		{
			name: "when the file is valid, the defaults apply",
			body: validConfig(),
			validateFunc: func(v *viper.Viper, c config.Config) {
				s.Equal(8080, c.Controller.API.Port)
				s.Equal("localhost", c.Controller.NATS.Host, "the controller's NATS default")
				s.Equal("/var/kv", behavior.Settings[kvSettings](s.kv).Dir,
					"a behavior's setting default")
				s.Equal(config.SourceDefault, config.Source(v, "behaviors.kv.dir"))
			},
		},
		{
			name: "when the file and the environment set a behavior's settings",
			body: validConfig("behaviors:", "  kv:", "    dir: /srv/kv"),
			env:  map[string]string{"OSAPI_BEHAVIORS_KV_LEVEL": "3"},
			validateFunc: func(v *viper.Viper, _ config.Config) {
				got := behavior.Settings[kvSettings](s.kv)
				s.Equal("/srv/kv", got.Dir)
				s.Equal(3, got.Level)
				s.Equal(10, got.Limit.Max)
				s.Equal(config.SourceFile, config.Source(v, "behaviors.kv.dir"))
				s.Equal(config.SourceEnv, config.Source(v, "behaviors.kv.level"))
			},
		},
		{name: "when the file is missing", missing: true, wantErr: "read "},
		{
			name:    "when a value has the wrong shape",
			body:    validConfig("agent: [1]"),
			wantErr: "parse ",
		},
		{
			name:    "when a value breaks a constraint",
			body:    strings.Replace(validConfig(), testSigningKey, "short", 1),
			wantErr: "signing",
		},
		{
			name:    "when a behavior's setting breaks its rule",
			body:    validConfig("behaviors:", "  kv:", "    level: 9"),
			wantErr: "behaviors.kv",
		},
		{
			name:    "when a behavior's setting has the wrong shape",
			body:    validConfig("behaviors:", "  kv:", "    level: many"),
			wantErr: "behaviors.kv",
		},
		{
			name:     "when the extra checks fail",
			body:     validConfig(),
			register: func(*validator.Validate) error { return errors.New("boom") },
			wantErr:  "boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.register != nil {
				config.SetRegisterValidatorsFn(tc.register)
			}
			for k, v := range tc.env {
				s.T().Setenv(k, v)
			}
			path := filepath.Join(s.T().TempDir(), "osapi.yaml")
			if !tc.missing {
				path = s.write(tc.body)
			}

			v := viper.New()
			var c config.Config
			err := config.Load(v, path, &c)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(v, c)
		})
	}
}

func (s *LoadPublicTestSuite) TestRead() {
	tests := []struct {
		name      string
		body      string
		missing   bool
		wantFound bool
		wantErr   string
	}{
		{name: "when the file is missing the defaults still apply", missing: true},
		{name: "when the file is not YAML", body: "controller: [", wantErr: "read "},
		{
			name:      "when the file is there it is read, unchecked",
			body:      "agent:\n  max_jobs: 0",
			wantFound: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			path := filepath.Join(s.T().TempDir(), "osapi.yaml")
			if !tc.missing {
				path = s.write(tc.body)
			}

			var c config.Config
			found, err := config.Read(viper.New(), path, &c)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantFound, found)
			s.Equal(8080, c.Controller.API.Port)
		})
	}
}

func (s *LoadPublicTestSuite) TestSource() {
	s.T().Setenv("OSAPI_AGENT_HOSTNAME", "web-01")
	v := viper.New()
	var c config.Config
	s.Require().NoError(config.Load(v, s.write(strings.Replace(validConfig(),
		"  api:\n", "  api:\n    port: 9000\n", 1)), &c))

	tests := []struct {
		key  string
		want string
	}{
		{key: "agent.hostname", want: config.SourceEnv},
		{key: "controller.api.port", want: config.SourceFile},
		{key: "agent.max_jobs", want: config.SourceDefault},
		{key: "agent.labels", want: config.SourceUnset},
	}

	for _, tc := range tests {
		s.Run(tc.key, func() {
			s.Equal(tc.want, config.Source(v, tc.key))
		})
	}
	s.Equal("web-01", c.Agent.Hostname, "an environment variable applies with no key in the file")
}

func (s *LoadPublicTestSuite) TestValue() {
	s.T().Setenv("OSAPI_CONTROLLER_API_PORT", "9000")
	s.T().Setenv("OSAPI_AGENT_CONDITIONS_HIGH_LOAD_MULTIPLIER", "1.5")
	s.T().Setenv("OSAPI_DEBUG", "true")
	v := viper.New()
	var c config.Config
	s.Require().NoError(config.Load(v, s.write(validConfig()), &c))

	byName := map[string]config.Key{}
	for _, k := range config.Keys() {
		byName[k.Name] = k
	}

	tests := []struct {
		key  string
		want any
	}{
		{key: "controller.api.port", want: int64(9000)},
		{key: "agent.conditions.high_load_multiplier", want: 1.5},
		{key: "debug", want: true},
		{key: "controller.api.job_timeout", want: "30s"},
		{key: "agent.labels", want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.key, func() {
			s.Equal(tc.want, config.Value(v, byName[tc.key]))
		})
	}
}

func TestLoadPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LoadPublicTestSuite))
}
