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

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret/vault"
)

// validConfig is a config file holding the two required keys, then extra.
func validConfig(
	extra ...string,
) string {
	return strings.Join(append([]string{
		"controller:",
		"  api:",
		"    security:",
		"      signing_keys:",
		"        - id: test",
		"          key: " + testSigningKey,
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
		dir          bool
		wantErr      string
		wantRule     string
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
				s.Equal(config.SourceUnset, config.Source(v, "controller.api.url"),
					"a choice of the operator that nothing set")
				s.Equal(config.SourceUnset, config.Source(v, "no.such.key"))
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
		{
			name: "when the file has signing keys, the first is the signer",
			body: strings.Replace(validConfig(), "        - id: test",
				"        - id: new\n          key: "+testSigningKey+"-new\n        - id: test", 1),
			validateFunc: func(_ *viper.Viper, c config.Config) {
				keys := c.Controller.API.Security.Keys()
				s.Equal(authtoken.Keys{
					{ID: "new", Secret: testSigningKey + "-new"},
					{ID: "test", Secret: testSigningKey},
				}, keys)
			},
		},
		{
			name: "when the environment sets the signing keys, each id=key pair is a key",
			body: validConfig(),
			env: map[string]string{
				"OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS": "new=" + testSigningKey +
					"-new,old=" + testSigningKey,
			},
			validateFunc: func(v *viper.Viper, c config.Config) {
				s.Equal([]config.SigningKey{
					{ID: "new", Key: testSigningKey + "-new"},
					{ID: "old", Key: testSigningKey},
				}, c.Controller.API.Security.SigningKeys)
				s.Equal(config.SourceEnv, config.Source(v, "controller.api.security.signing_keys"))
			},
		},
		{
			name: "when an environment pair has no id, the key has none",
			body: validConfig(),
			env: map[string]string{
				"OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS": testSigningKey,
			},
			wantErr:  `controller.api.security.signing_keys[0].id breaks its rule "required"`,
			wantRule: config.RuleValue,
		},
		{
			name: "when the list of signing keys is empty, the config loads: the controller makes a key",
			body: strings.Replace(validConfig(),
				"      signing_keys:\n        - id: test\n          key: "+testSigningKey,
				`      signing_keys: ""`, 1),
			validateFunc: func(_ *viper.Viper, c config.Config) {
				s.Empty(c.Controller.API.Security.SigningKeys)
			},
		},
		{
			name: "when two signing keys have one id, the config does not load",
			body: validConfig(),
			env: map[string]string{
				"OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS": "a=" + testSigningKey +
					",a=" + testSigningKey,
			},
			wantErr:  `controller.api.security.signing_keys breaks its rule "unique=ID"`,
			wantRule: config.RuleValue,
		},
		{
			name: "when the environment sets a list, each comma starts an item",
			body: validConfig(),
			env: map[string]string{
				"OSAPI_CONTROLLER_API_SECURITY_CORS_ALLOW_ORIGINS": "https://a,https://b",
			},
			validateFunc: func(_ *viper.Viper, c config.Config) {
				s.Equal([]string{"https://a", "https://b"},
					c.Controller.API.Security.CORS.AllowOrigins)
			},
		},
		{
			name: "when a list is an empty string, it is empty",
			body: validConfig("agent:", "  consumer:", `    back_off: ""`),
			validateFunc: func(_ *viper.Viper, c config.Config) {
				s.Empty(c.Agent.Consumer.BackOff)
			},
		},
		{
			name:     "when the file is missing",
			missing:  true,
			wantErr:  "read ",
			wantRule: config.RuleFile,
		},
		{
			name:     "when the path cannot be read as a file",
			dir:      true,
			wantErr:  "read ",
			wantRule: config.RuleFile,
		},
		{
			name:     "when the file is not YAML",
			body:     "controller:\n  api:\n    port: [\n",
			wantErr:  "read ",
			wantRule: config.RuleParse,
		},
		{
			name:     "when a value has the wrong shape",
			body:     validConfig("agent: [1]"),
			wantErr:  "parse ",
			wantRule: config.RuleParse,
		},
		{
			name:     "when a value breaks a constraint",
			body:     strings.Replace(validConfig(), testSigningKey, "short", 1),
			wantErr:  `controller.api.security.signing_keys[0].key breaks its rule "signing_key": signing key`,
			wantRule: config.RuleValue,
		},
		{
			name: "when a custom role grants a permission that osapi does not know",
			body: strings.Replace(validConfig(), testSigningKey, testSigningKey+
				"\n      roles:\n        auditor:\n          permissions: [helth:read]", 1),
			wantErr:  `roles[auditor].permissions[0] breaks its rule "permission": "helth:read" is not a permission`,
			wantRule: config.RuleValue,
		},
		{
			name:     "when a behavior's setting breaks its rule",
			body:     validConfig("behaviors:", "  kv:", "    level: 9"),
			wantErr:  `behaviors.kv.level breaks its rule "max=`,
			wantRule: config.RuleSetting,
		},
		{
			name:     "when a behavior's setting has the wrong shape",
			body:     validConfig("behaviors:", "  kv:", "    level: many"),
			wantErr:  "behaviors.kv",
			wantRule: config.RuleParse,
		},
		{
			name: "when two keys name one bucket",
			body: validConfig(),
			env: map[string]string{
				"OSAPI_NATS_REGISTRY_BUCKET": "shared",
				"OSAPI_NATS_KV_BUCKET":       "shared",
			},
			wantErr:  `are both "shared"`,
			wantRule: config.RuleBucket,
		},
		{
			name:     "when the extra checks fail",
			body:     validConfig(),
			register: func(*validator.Validate) error { return errors.New("boom") },
			wantErr:  "boom",
			wantRule: config.RuleValue,
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
			switch {
			case tc.dir:
				path = s.T().TempDir()
			case !tc.missing:
				path = s.write(tc.body)
			}

			v := viper.New()
			var c config.Config
			err := config.Load(v, path, &c)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Equal(tc.wantRule, config.RuleOf(err))

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(v, c)
		})
	}
}

func (s *LoadPublicTestSuite) TestCheck() {
	tests := []struct {
		name    string
		body    string
		missing bool
		want    []config.Problem
	}{
		{name: "when the file is valid there are no problems", body: validConfig()},
		{
			name:    "when the file is missing it is one problem",
			missing: true,
			want: []config.Problem{{
				Rule: config.RuleFile,
				Fix:  config.Checks.Fix(config.RuleFile),
			}},
		},
		{
			name: "when values break rules each is a problem with its key",
			body: strings.Replace(
				validConfig("behaviors:", "  kv:", "    level: 9"), testSigningKey, "short", 1),
			want: []config.Problem{
				{
					Rule: config.RuleValue,
					Key:  "controller.api.security.signing_keys[0].key",
					Fix:  config.Checks.Fix(config.RuleValue),
				},
				{
					Rule: config.RuleSetting,
					Key:  "behaviors.kv.level",
					Fix:  config.Checks.Fix(config.RuleSetting),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			path := filepath.Join(s.T().TempDir(), "osapi.yaml")
			if !tc.missing {
				path = s.write(tc.body)
			}

			var c config.Config
			got := config.Check(viper.New(), path, &c)

			s.Require().Len(got, len(tc.want))
			for i, want := range tc.want {
				s.Equal(want.Rule, got[i].Rule)
				s.Equal(want.Key, got[i].Key)
				s.Equal(want.Fix, got[i].Fix)
				s.NotEmpty(got[i].Message)
			}
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

func (s *LoadPublicTestSuite) TestDefaultVault() {
	v := viper.New()
	var c config.Config
	s.Require().NoError(config.Load(v, s.write(validConfig()), &c))
	s.Equal([]vault.Spec{{Name: "default", Type: "nats", KeyFile: config.DefaultVaultKeyFile}},
		c.Controller.Secrets.Vaults, "a config with no vault has the default vault")

	v = viper.New()
	c = config.Config{}
	s.Require().NoError(config.Load(v, s.write(strings.Replace(
		validConfig(),
		"controller:\n",
		"controller:\n  secrets:\n    vaults:\n      - name: prod\n        type: nats\n        key_file: /tmp/prod.key\n",
		1,
	)), &c))
	s.Equal([]vault.Spec{{Name: "prod", Type: "nats", KeyFile: "/tmp/prod.key"}},
		c.Controller.Secrets.Vaults, "the vaults of the config replace the default")
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

// TestLoadDefaultFile is a host with no config file at the default path: it
// runs on the defaults, where a missing file that -f names is an error.
func (s *LoadPublicTestSuite) TestLoadDefaultFile() {
	if _, err := os.Stat(config.DefaultFile); err == nil {
		s.T().Skip(config.DefaultFile + " is there on this host")
	}
	var c config.Config
	s.Require().NoError(config.Load(viper.New(), config.DefaultFile, &c))
	s.Equal(8080, c.Controller.API.Port, "the default of controller start")
}
