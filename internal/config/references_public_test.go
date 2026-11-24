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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/vault"
)

const (
	// signingKey is a value that the rule signing_key allows.
	signingKey = "0123456789abcdef0123456789abcdef"
)

type ReferencesPublicTestSuite struct {
	suite.Suite
}

func TestReferencesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReferencesPublicTestSuite))
}

func (s *ReferencesPublicTestSuite) TearDownTest() {
	config.ResetRegisterValidatorsFn()
}

// referenceConfig is a config with no problem, and the vault default.
func referenceConfig() config.Config {
	var c config.Config
	c.Controller.API.Port = 8080
	c.Controller.API.Security.SigningKeys = []config.SigningKey{{ID: "k1", Key: signingKey}}
	c.Controller.Client.Security.BearerToken = "token"
	c.Controller.NATS.Port = 4222
	c.Agent.NATS.Port = 4222
	c.Agent.MaxJobs = 1
	c.Agent.Conditions = config.AgentConditions{
		MemoryPressureThreshold: 90, HighLoadMultiplier: 2, DiskPressureThreshold: 90,
	}
	c.NATS.Stream = config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}
	c.NATS.KV.Bucket = "job-queue"
	c.NATS.Registry.Bucket = "agent-registry"
	c.Controller.Secrets.Vaults = []vault.Spec{
		{Name: "default", Type: "nats", KeyFile: "/etc/osapi/secrets.key"},
	}

	return c
}

func (s *ReferencesPublicTestSuite) TestProblems() {
	tests := []struct {
		name  string
		setup func(c *config.Config)
		want  []config.Problem
	}{
		{
			name: "when a signing key is a reference of a vault there is no problem",
			want: []config.Problem{},
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://default/jwt"
				c.Controller.Client.Security.BearerToken = "secret://default/token"
			},
		},
		{
			name: "when a reference names a vault that is not in the config names it",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://other/jwt"
			},
			want: []config.Problem{
				{
					Rule:    config.RuleReference,
					Key:     "controller.api.security.signing_keys[0].key",
					Message: `is secret://other/jwt: the vault "other" is not in controller.secrets.vaults`,
				},
			},
		},
		{
			name: "when a reference does not parse refuses it",
			setup: func(c *config.Config) {
				c.Controller.Client.Security.BearerToken = "secret://default"
			},
			want: []config.Problem{
				{
					Rule:    config.RuleReference,
					Key:     "controller.client.security.bearer_token",
					Message: "secret://default: a reference is secret://<vault>/<key>, with a vault name of 1 to 64 letters, digits, - or _",
				},
			},
		},
		{
			name: "when a NATS password is a reference refuses it",
			setup: func(c *config.Config) {
				c.Controller.NATS.Auth.Password = "secret://default/nats"
				c.Agent.NATS.Auth.Password = "secret://default/nats"
				c.NATS.Server.Auth.Users = []config.NATSServerUser{
					{Username: "osapi", Password: "secret://default/nats"},
				}
			},
			want: []config.Problem{
				{
					Rule:    config.RuleReference,
					Key:     "controller.nats.auth.password",
					Message: "is a reference: a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault",
				},
				{
					Rule:    config.RuleReference,
					Key:     "agent.nats.auth.password",
					Message: "is a reference: a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault",
				},
				{
					Rule:    config.RuleReference,
					Key:     "nats.server.auth.users[0].password",
					Message: "is a reference: a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault",
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := referenceConfig()
			tc.setup(&c)

			problems := config.Problems(&c)
			got := make([]config.Problem, 0, len(problems))
			for _, p := range problems {
				p.Fix = ""
				got = append(got, p)
			}
			s.Equal(tc.want, got)
		})
	}
}

func (s *ReferencesPublicTestSuite) TestReferences() {
	c := referenceConfig()
	s.Empty(config.References(&c, "controller"))

	c.Controller.API.Security.SigningKeys = append(c.Controller.API.Security.SigningKeys,
		config.SigningKey{ID: "k2", Key: "secret://default/jwt"})
	c.Controller.Client.Security.BearerToken = "secret://default/token"

	s.Equal([]string{
		"controller.client.security.bearer_token",
		"controller.api.security.signing_keys[1].key",
	}, config.References(&c, "controller"))
	s.Equal(
		[]string{"controller.api.security.signing_keys[1].key"},
		config.References(&c, "controller.api"),
	)
}

func (s *ReferencesPublicTestSuite) TestResolveReferences() {
	values := map[string]string{
		"secret://default/jwt":   signingKey + "-from-the-vault",
		"secret://default/short": "short",
		"secret://default/token": "the-token",
	}
	resolve := func(_ context.Context, ref vault.Reference) ([]byte, error) {
		v, ok := values[ref.String()]
		if !ok {
			return nil, errors.New(ref.String() + ": no such key")
		}

		return []byte(v), nil
	}

	tests := []struct {
		name       string
		setup      func(c *config.Config)
		prefix     string
		validators func(*validator.Validate) error
		wantErr    string
		check      func(c config.Config)
	}{
		{
			name: "when each reference resolves puts each value in its place",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://default/jwt"
				c.Controller.Client.Security.BearerToken = "secret://default/token"
			},
			prefix: "controller.api",
			check: func(c config.Config) {
				s.Equal(signingKey+"-from-the-vault", c.Controller.API.Security.SigningKeys[0].Key)
				// A key outside the prefix keeps its reference.
				s.Equal("secret://default/token", c.Controller.Client.Security.BearerToken)
			},
		},
		{
			name: "when a key does not resolve names the key and the reference",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://default/nope"
			},
			prefix:  "controller.api",
			wantErr: "controller.api.security.signing_keys[0].key: secret://default/nope: no such key",
		},
		{
			name: "when a value breaks the rule of its key names the rule and not the value",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://default/short"
			},
			prefix:  "controller.api",
			wantErr: `controller.api.security.signing_keys[0].key: the value of secret://default/short breaks its rule "signing_key": signing key must be at least 32 characters — generate one with: openssl rand -hex 32`,
		},
		{
			name: "when a NATS password is a reference refuses it",
			setup: func(c *config.Config) {
				c.Controller.NATS.Auth.Password = "secret://default/nats"
			},
			prefix:  "controller",
			wantErr: "controller.nats.auth.password is a reference: a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault",
		},
		{
			name: "when a reference does not parse refuses it",
			setup: func(c *config.Config) {
				c.Controller.Client.Security.BearerToken = "secret://default"
			},
			prefix:  "controller.client",
			wantErr: "controller.client.security.bearer_token: secret://default: a reference is secret://<vault>/<key>, with a vault name of 1 to 64 letters, digits, - or _",
		},
		{
			name: "when the validators do not register gives the error",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys[0].Key = "secret://default/jwt"
			},
			prefix:     "controller.api",
			validators: func(*validator.Validate) error { return errors.New("no validators") },
			wantErr:    "no validators",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.validators != nil {
				config.SetRegisterValidatorsFn(tc.validators)
				defer config.ResetRegisterValidatorsFn()
			}
			c := referenceConfig()
			tc.setup(&c)

			err := config.ResolveReferences(context.Background(), &c, tc.prefix, resolve)
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				s.NotContains(err.Error(), "-from-the-vault")

				return
			}
			s.Require().NoError(err)
			tc.check(c)
		})
	}

	// A key with no rule keeps its value as the vault gives it.
	c := referenceConfig()
	c.Controller.Client.Security.BearerToken = "secret://default/token"
	s.Require().NoError(config.ResolveReferences(context.Background(), &c, "", resolve))
	s.Equal("the-token", c.Controller.Client.Security.BearerToken)
	s.False(strings.HasPrefix(c.Controller.Client.Security.BearerToken, secret.Scheme))
}
