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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret/vault"
)

// storeConfig is the config block of the type store, which the tests
// register.
type storeConfig struct {
	Address string `mapstructure:"address" validate:"required" doc:"The address of the store."`
}

type VaultsPublicTestSuite struct {
	suite.Suite
}

func TestVaultsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(VaultsPublicTestSuite))
}

func (s *VaultsPublicTestSuite) SetupTest() {
	vault.Register(vault.Define("store", "A store for the tests", nil,
		func(context.Context, vault.Env, vault.Spec, storeConfig) (vault.Vault, error) {
			return nil, nil
		},
	))
}

func (s *VaultsPublicTestSuite) TearDownTest() {
	vault.Unregister("store")
}

// vaultProblems are the problems of the rule vault in a valid config with
// specs as its vaults.
func (s *VaultsPublicTestSuite) vaultProblems(
	specs []vault.Spec,
) []config.Problem {
	var c config.Config
	c.Controller.API.Port = 8080
	c.Controller.API.Security.SigningKeys = []config.SigningKey{
		{ID: "k1", Key: "0123456789abcdef0123456789abcdef"},
	}
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
	c.Controller.Secrets.Vaults = specs

	var out []config.Problem
	for _, p := range config.Problems(&c) {
		if p.Rule == config.RuleVault || strings.HasPrefix(p.Key, "controller.secrets") {
			p.Fix = ""
			out = append(out, p)
		}
	}

	return out
}

func (s *VaultsPublicTestSuite) TestVaultProblems() {
	tests := []struct {
		name  string
		specs []vault.Spec
		want  []config.Problem
	}{
		{
			name: "when each vault is valid there is no problem",
			specs: []vault.Spec{
				{Name: "default", Type: "nats", KeyFile: "/etc/osapi/secrets.key"},
				{
					Name:    "prod",
					Type:    "store",
					Targets: []string{"group:prod"},
					Config:  map[string]any{"address": "a"},
				},
			},
		},
		{
			name: "when two vaults have one name and one key file names both",
			specs: []vault.Spec{
				{Name: "default", Type: "nats", KeyFile: "/k"},
				{Name: "default", Type: "nats", KeyFile: "/k"},
			},
			want: []config.Problem{
				{
					Rule:    config.RuleVault,
					Key:     "controller.secrets.vaults[1].name",
					Message: `is "default", as vault 0 is: each vault needs its own name`,
				},
				{
					Rule:    config.RuleVault,
					Key:     "controller.secrets.vaults[1].key_file",
					Message: `is "/k", as vault 0 is: a rotation of one key would drop the key of the other vault`,
				},
			},
		},
		{
			name: "when the key file is a reference refuses it",
			specs: []vault.Spec{
				{Name: "default", Type: "nats", KeyFile: "secret://other/vault-key"},
			},
			want: []config.Problem{
				{
					Rule:    config.RuleVault,
					Key:     "controller.secrets.vaults[0].key_file",
					Message: "is a reference: the vault key cannot come from a vault, because the controller needs the key to read the vault: give the path of the key file",
				},
			},
		},
		{
			name:  "when a type is not registered names the types",
			specs: []vault.Spec{{Name: "default", Type: "nope"}},
			want: []config.Problem{
				{
					Rule:    config.RuleVault,
					Key:     "controller.secrets.vaults[0].type",
					Message: `vault "default": type "nope" is not a type of vault: the types are nats, store`,
				},
			},
		},
		{
			name: "when the config block has a key that the type does not take names it",
			specs: []vault.Spec{
				{Name: "prod", Type: "store", Config: map[string]any{"adress": "a"}},
			},
			want: []config.Problem{{
				Rule:    config.RuleVault,
				Key:     "controller.secrets.vaults[0].config",
				Message: `vault "prod": config: the type store has no setting adress`,
			}},
		},
		{
			name:  "when a value of the config block breaks its rule names its key",
			specs: []vault.Spec{{Name: "prod", Type: "store"}},
			want: []config.Problem{{
				Rule:    config.RuleVault,
				Key:     "controller.secrets.vaults[0].config.address",
				Message: `breaks its rule "required"`,
			}},
		},
		{
			name:  "when the check of the type fails names the vault",
			specs: []vault.Spec{{Name: "default", Type: "nats"}},
			want: []config.Problem{{
				Rule:    config.RuleVault,
				Key:     "controller.secrets.vaults[0]",
				Message: `vault "default": a nats vault needs its key: set key_file or key_env`,
			}},
		},
		{
			name:  "when a vault has no valid name it is a value problem",
			specs: []vault.Spec{{Name: "a/b", Type: "nats", KeyFile: "/k"}},
			want: []config.Problem{
				{
					Rule:    config.RuleValue,
					Key:     "controller.secrets.vaults[0].name",
					Message: `breaks its rule "resource_name": "a/b" is not a valid name (expected: ^[a-zA-Z0-9_-]+$)`,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, s.vaultProblems(tc.specs))
		})
	}
}

func (s *VaultsPublicTestSuite) TestVaultKeys() {
	keys := config.VaultKeys()
	names := make([]string, 0, len(keys))
	for _, k := range keys {
		names = append(names, k.Name)
		s.Empty(k.Env, k.Name)
	}

	s.Equal([]string{
		"controller.secrets.vaults[].name",
		"controller.secrets.vaults[].type",
		"controller.secrets.vaults[].targets",
		"controller.secrets.vaults[].key_file",
		"controller.secrets.vaults[].key_env",
		"controller.secrets.vaults[].config",
		"controller.secrets.vaults[].config.address",
	}, names, "a type with no settings, nats, adds none")
	s.Equal("Type store: The address of the store.", config.VaultKeys()[6].Doc)
}
