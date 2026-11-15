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
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
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
							SigningKey: testSigningKey,
						},
					},
					NATS: config.NATSConnection{Port: 4222},
				},
				NATS: config.NATS{
					Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
					KV:       config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
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
							SigningKey: "",
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "SigningKey")
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
							SigningKey: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 31 characters
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "SigningKey")
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
							SigningKey: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 32 characters
						},
					},
					NATS: config.NATSConnection{Port: 4222},
				},
				NATS: config.NATS{
					Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
					KV:       config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
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
							SigningKey: "CHANGE_ME_BEFORE_DEPLOYING",
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "SigningKey")
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
							SigningKey: testSigningKey,
						},
					},
				},
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "BearerToken")
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

func (s *ConfigPublicTestSuite) TestValidateURL() {
	valid := func() config.Config {
		return config.Config{
			Controller: config.Controller{
				Client: config.Client{
					Security: config.ClientSecurity{BearerToken: "test-bearer-token"},
				},
				API: config.APIServer{
					Port:     8080,
					Security: config.ServerSecurity{SigningKey: testSigningKey},
				},
				NATS: config.NATSConnection{Port: 4222},
			},
			NATS: config.NATS{
				Stream:   config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
				KV:       config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
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
			wantErr: "URL",
		},
		{
			name: "tracing url with a scheme other than http",
			set: func(c *config.Config) {
				c.Telemetry.Tracing.URL = "grpc://hub.example.ts.net:4317"
			},
			wantErr: "URL",
		},
		{
			name: "ui url that is a path",
			set: func(c *config.Config) {
				c.Controller.UI.URL = "/ui"
			},
			wantErr: "URL",
		},
		{
			name: "http url without a host",
			set: func(c *config.Config) {
				c.Controller.Metrics.URL = "http://"
			},
			wantErr: "URL",
		},
		{
			name: "nats server url with http",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "http://hub.example.ts.net:4222"
			},
			wantErr: "URL",
		},
		{
			name: "nats server url with a made-up scheme",
			set: func(c *config.Config) {
				c.NATS.Server.URL = "hub:4222"
			},
			wantErr: "URL",
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
			wantErr: "URL",
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
							SigningKey: testSigningKey,
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
