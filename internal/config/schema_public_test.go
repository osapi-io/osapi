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

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
)

type ConfigPublicTestSuite struct {
	suite.Suite
}

func (s *ConfigPublicTestSuite) TestValidate() {
	tests := []struct {
		name        string
		config      config.Config
		expectError bool
		errContains string
	}{
		{
			name: "valid config",
			config: config.Config{
				API: config.API{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: "test-signing-key",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing signing key",
			config: config.Config{
				API: config.API{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "test-bearer-token",
						},
					},
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: "",
						},
					},
				},
			},
			expectError: true,
			errContains: "SigningKey",
		},
		{
			name: "missing bearer token",
			config: config.Config{
				API: config.API{
					Client: config.Client{
						Security: config.ClientSecurity{
							BearerToken: "",
						},
					},
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: "test-signing-key",
						},
					},
				},
			},
			expectError: true,
			errContains: "BearerToken",
		},
		{
			name:        "missing both required fields",
			config:      config.Config{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := config.Validate(&tt.config)

			if tt.expectError {
				s.Error(err)
				if tt.errContains != "" {
					s.Contains(err.Error(), tt.errContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestConfigPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigPublicTestSuite))
}
