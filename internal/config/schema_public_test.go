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
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/validation"
)

// testSigningKey is a fixture signing key long enough to satisfy the
// minimum-length validation, used wherever a test only needs a valid value.
const testSigningKey = "test-signing-key-0123456789abcdef"

type ConfigPublicTestSuite struct {
	suite.Suite
}

func (s *ConfigPublicTestSuite) TestValidate() {
	tests := []struct {
		name         string
		config       config.Config
		validateFunc func(error)
	}{
		{
			name: "valid config",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					API: config.APIServer{
						Port: 8080,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: testSigningKey}},
						},
					},
					NATS: config.NATSConnection{Port: 4222},
				},
				NATS: config.NATS{
					Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
					KV:       config.NATSKV{Bucket: "job-queue"},
					Registry: config.NATSRegistry{Bucket: "agent-registry"},
				},
				Agent: config.AgentConfig{
					NATS:    config.NATSConnection{Port: 4222},
					MaxJobs: 10,
					Conditions: config.AgentConditions{
						MemoryPressureThreshold: 90,
						HighLoadMultiplier:      2.0,
						DiskPressureThreshold:   90,
					},
				},
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "missing signing key",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: ""}},
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(
					err.Error(),
					"controller.api.security.signing_keys[0].key breaks its rule",
				)
			},
		},
		{
			name: "signing key one character under the minimum",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{
								{ID: "test", Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
							}, // 31 characters
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(
					err.Error(),
					"controller.api.security.signing_keys[0].key breaks its rule",
				)
			},
		},
		{
			name: "signing key at the minimum length",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					API: config.APIServer{
						Port: 8080,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{
								{ID: "test", Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
							}, // 32 characters
						},
					},
					NATS: config.NATSConnection{Port: 4222},
				},
				NATS: config.NATS{
					Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
					KV:       config.NATSKV{Bucket: "job-queue"},
					Registry: config.NATSRegistry{Bucket: "agent-registry"},
				},
				Agent: config.AgentConfig{
					NATS:    config.NATSConnection{Port: 4222},
					MaxJobs: 10,
					Conditions: config.AgentConditions{
						MemoryPressureThreshold: 90,
						HighLoadMultiplier:      2.0,
						DiskPressureThreshold:   90,
					},
				},
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "committed example placeholder signing key",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{
								{ID: "test", Key: "CHANGE_ME_BEFORE_DEPLOYING"},
							},
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(
					err.Error(),
					"controller.api.security.signing_keys[0].key breaks its rule",
				)
			},
		},
		{
			name: "missing bearer token",
			config: config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "",
						},
					},
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: testSigningKey}},
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "controller.client.security.bearer_token breaks its rule")
			},
		},
		{
			name:   "missing both required fields",
			config: config.Config{},
			validateFunc: func(err error) {
				s.Error(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(config.Validate(&tt.config))
		})
	}
}

// validConfig returns a config that passes Validate, with nats replaced.
func (s *ConfigPublicTestSuite) validConfig(
	nats config.NATS,
) config.Config {
	return config.Config{
		Controller: config.Controller{
			Client: config.Client{
				Security: config.ClientSecurity{BearerToken: "test-bearer-token"},
			},
			API: config.APIServer{
				Port: 8080,
				Security: config.ServerSecurity{
					SigningKeys: []config.SigningKey{{ID: "test", Key: testSigningKey}},
				},
			},
			NATS: config.NATSConnection{Port: 4222},
		},
		NATS: nats,
		Agent: config.AgentConfig{
			NATS:    config.NATSConnection{Port: 4222},
			MaxJobs: 10,
			Conditions: config.AgentConditions{
				MemoryPressureThreshold: 90,
				HighLoadMultiplier:      2.0,
				DiskPressureThreshold:   90,
			},
		},
	}
}

func (s *ConfigPublicTestSuite) TestValidateBuckets() {
	base := func() config.NATS {
		return config.NATS{
			Stream:    config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
			KV:        config.NATSKV{Bucket: "job-queue"},
			Registry:  config.NATSRegistry{Bucket: "agent-registry"},
			Facts:     config.NATSFacts{Bucket: "agent-facts"},
			State:     config.NATSState{Bucket: "agent-state"},
			FileState: config.NATSFileState{Bucket: "file-state"},
			Objects:   config.NATSObjects{Bucket: "file-objects"},
		}
	}

	tests := []struct {
		name    string
		change  func(*config.NATS)
		wantErr string
	}{
		{
			name:   "when each bucket has its own name passes",
			change: func(*config.NATS) {},
		},
		{
			name:   "when an optional bucket is unset passes",
			change: func(n *config.NATS) { n.Facts.Bucket, n.State.Bucket = "", "" },
		},
		{
			name: "when an Object Store bucket has the name of a KV bucket passes",
			change: func(n *config.NATS) {
				n.Objects.Bucket = "file-state"
			},
		},
		{
			name:    "when a bucket name is not valid in NATS fails",
			change:  func(n *config.NATS) { n.Registry.Bucket = "agent.registry" },
			wantErr: "not a valid NATS bucket name",
		},
		{
			name:    "when an optional bucket name is not valid in NATS fails",
			change:  func(n *config.NATS) { n.State.Bucket = "agent state" },
			wantErr: "not a valid NATS bucket name",
		},
		{
			name:    "when the namespace is not valid in NATS fails",
			change:  func(n *config.NATS) { n.Server.Namespace = "osapi.prod" },
			wantErr: "not a valid NATS bucket name",
		},
		{
			name:    "when two KV settings name the same bucket fails",
			change:  func(n *config.NATS) { n.Registry.Bucket = "job-queue" },
			wantErr: `nats.kv.bucket and nats.registry.bucket are both "job-queue": each needs its own bucket`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			n := base()
			tt.change(&n)
			c := s.validConfig(n)

			err := config.Validate(&c)
			if tt.wantErr == "" {
				s.NoError(err)
				return
			}
			s.Require().Error(err)
			msg := err.Error()
			if m, ok := validation.Struct(context.Background(), c); !ok {
				msg = m
			}
			s.Contains(msg, tt.wantErr)
		})
	}
}

func (s *ConfigPublicTestSuite) TestValidateObjectStoreBuckets() {
	s.Run("when two Object Store settings name the same bucket fails", func() {
		// There is one Object Store setting today, so the check runs on a
		// list that has two.
		err := config.ExportSharedObjectStoreBucket([]config.ObjectStoreBucketInfo{
			{Key: "nats.objects.bucket", Bucket: "files"},
			{Key: "nats.other.bucket", Bucket: "files"},
		})
		s.EqualError(
			err,
			`nats.objects.bucket and nats.other.bucket are both "files": each needs its own bucket`,
		)
	})
}

func (s *ConfigPublicTestSuite) TestValidateURL() {
	valid := func() config.Config {
		return config.Config{
			Controller: config.Controller{
				Client: config.Client{
					Security: config.ClientSecurity{BearerToken: "test-bearer-token"},
				},
				API: config.APIServer{
					Port: 8080,
					Security: config.ServerSecurity{
						SigningKeys: []config.SigningKey{{ID: "test", Key: testSigningKey}},
					},
				},
				NATS: config.NATSConnection{Port: 4222},
			},
			NATS: config.NATS{
				Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
				KV:       config.NATSKV{Bucket: "job-queue"},
				Registry: config.NATSRegistry{Bucket: "agent-registry"},
			},
			Agent: config.AgentConfig{
				NATS:    config.NATSConnection{Port: 4222},
				MaxJobs: 10,
				Conditions: config.AgentConditions{
					MemoryPressureThreshold: 90,
					HighLoadMultiplier:      2.0,
					DiskPressureThreshold:   90,
				},
			},
		}
	}

	tests := []struct {
		name    string
		set     func(c *config.Config)
		wantErr string
	}{
		{
			name: "every url set to an http or https URL",
			set: func(c *config.Config) {
				c.Controller.API.Host = "127.0.0.1"
				c.Controller.API.URL = "https://hub.example.ts.net/"
				c.Controller.UI.URL = "https://hub.example.ts.net/"
				c.Controller.Metrics.URL = "http://hub.example.ts.net:9090/metrics"
				c.Telemetry.Tracing.URL = "http://hub.example.ts.net:16686/"
				c.Agent.Metrics.URL = "http://vm.example:9091/metrics"
				c.NATS.Server.URL = "nats://hub.example.ts.net:4222"
				c.NATS.Server.Metrics.URL = "http://hub.example.ts.net:9092/"
			},
		},
		{
			name: "api url without a scheme",
			set: func(c *config.Config) {
				c.Controller.API.URL = "hub.example.ts.net:8080"
			},
			wantErr: "controller.api.url breaks its rule",
		},
		{
			name: "tracing url with a scheme other than http",
			set: func(c *config.Config) {
				c.Telemetry.Tracing.URL = "grpc://hub.example.ts.net:4317"
			},
			wantErr: "telemetry.tracing.url breaks its rule",
		},
		{
			name: "ui url that is a path",
			set: func(c *config.Config) {
				c.Controller.UI.URL = "/ui"
			},
			wantErr: "controller.ui.url breaks its rule",
		},
		{
			name: "http url without a host",
			set: func(c *config.Config) {
				c.Controller.Metrics.URL = "http://"
			},
			wantErr: "controller.metrics.url breaks its rule",
		},
		{
			name: "nats server url with http",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "http://hub.example.ts.net:4222"
			},
			wantErr: "nats.server.url breaks its rule \"startswith=nats://|startswith=tls://\"",
		},
		{
			name: "nats server url with a made-up scheme",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "hub:4222"
			},
			wantErr: "nats.server.url breaks its rule \"startswith=nats://|startswith=tls://\"",
		},
		{
			name: "nats server url with tls",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "tls://hub.example.ts.net:4222"
			},
		},
		{
			name: "nats server url that is a bare hostname",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "hub.example.ts.net"
			},
			wantErr: "nats.server.url breaks its rule \"url\"",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := valid()
			tt.set(&c)

			err := config.Validate(&c)

			if tt.wantErr == "" {
				s.NoError(err)

				return
			}
			s.Error(err)
			s.Contains(err.Error(), tt.wantErr)
		})
	}
}

func (s *ConfigPublicTestSuite) TestValidateRegisterValidatorsError() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(err error)
	}{
		{
			name: "when validator registration fails returns error",
			setupMock: func() {
				config.SetRegisterValidatorsFn(func(_ *validator.Validate) error {
					return errors.New("validator registration failed")
				})
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "validator registration failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			defer config.ResetRegisterValidatorsFn()
			c := &config.Config{
				Controller: config.Controller{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-token",
						},
					},
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: testSigningKey}},
						},
					},
				},
			}

			err := config.Validate(c)
			tt.validateFunc(err)
		})
	}
}

func TestConfigPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConfigPublicTestSuite))
}
