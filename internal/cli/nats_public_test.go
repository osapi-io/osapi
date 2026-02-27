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

package cli_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/cli"
	climocks "github.com/retr0h/osapi/internal/cli/mocks"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
)

type NATSTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
}

func TestNATSTestSuite(t *testing.T) {
	suite.Run(t, new(NATSTestSuite))
}

func (suite *NATSTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
}

func (suite *NATSTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NATSTestSuite) TestCloseNATSClient() {
	tests := []struct {
		name    string
		setupFn func() func()
	}{
		{
			name: "when mock client does not panic",
			setupFn: func() func() {
				mock := mocks.NewMockNATSClient(suite.ctrl)

				return func() {
					cli.CloseNATSClient(mock)
				}
			},
		},
		{
			name: "when real client with nil NC does not panic",
			setupFn: func() func() {
				client := &natsclient.Client{}

				return func() {
					cli.CloseNATSClient(client)
				}
			},
		},
		{
			name: "when real client with non-nil NC closes connection",
			setupFn: func() func() {
				mockConn := climocks.NewMockNATSConnector(suite.ctrl)
				mockConn.EXPECT().Close()
				client := &natsclient.Client{NC: mockConn}

				return func() {
					cli.CloseNATSClient(client)
				}
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			closeFn := tc.setupFn()

			assert.NotPanics(suite.T(), closeFn)
		})
	}
}

func (suite *NATSTestSuite) TestBuildNATSAuthOptions() {
	tests := []struct {
		name string
		auth config.NATSAuth
		want natsclient.AuthOptions
	}{
		{
			name: "when user_pass returns user pass auth",
			auth: config.NATSAuth{
				Type:     "user_pass",
				Username: "osapi",
				Password: "secret",
			},
			want: natsclient.AuthOptions{
				AuthType: natsclient.UserPassAuth,
				Username: "osapi",
				Password: "secret",
			},
		},
		{
			name: "when nkey returns nkey auth",
			auth: config.NATSAuth{
				Type:     "nkey",
				NKeyFile: "/path/to/nkey",
			},
			want: natsclient.AuthOptions{
				AuthType: natsclient.NKeyAuth,
				NKeyFile: "/path/to/nkey",
			},
		},
		{
			name: "when none returns no auth",
			auth: config.NATSAuth{
				Type: "none",
			},
			want: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
		},
		{
			name: "when empty type defaults to no auth",
			auth: config.NATSAuth{},
			want: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.BuildNATSAuthOptions(tc.auth)

			assert.Equal(suite.T(), tc.want, got)
		})
	}
}

func (suite *NATSTestSuite) TestBuildRegistryKVConfig() {
	tests := []struct {
		name        string
		namespace   string
		registryCfg config.NATSRegistry
		validateFn  func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			registryCfg: config.NATSRegistry{
				Bucket:   "worker-registry",
				TTL:      "30s",
				Storage:  "file",
				Replicas: 1,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-worker-registry", cfg.Bucket)
				assert.Equal(suite.T(), 30*time.Second, cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			registryCfg: config.NATSRegistry{
				Bucket:   "worker-registry",
				TTL:      "1m",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "worker-registry", cfg.Bucket)
				assert.Equal(suite.T(), 1*time.Minute, cfg.TTL)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when TTL is invalid defaults to zero",
			namespace: "",
			registryCfg: config.NATSRegistry{
				Bucket:   "worker-registry",
				TTL:      "invalid",
				Storage:  "file",
				Replicas: 1,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.BuildRegistryKVConfig(tc.namespace, tc.registryCfg)

			tc.validateFn(got)
		})
	}
}

func (suite *NATSTestSuite) TestBuildAuditKVConfig() {
	tests := []struct {
		name       string
		namespace  string
		auditCfg   config.NATSAudit
		validateFn func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			auditCfg: config.NATSAudit{
				Bucket:   "audit-log",
				TTL:      "720h",
				MaxBytes: 52428800,
				Storage:  "file",
				Replicas: 1,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-audit-log", cfg.Bucket)
				assert.Equal(suite.T(), 720*time.Hour, cfg.TTL)
				assert.Equal(suite.T(), int64(52428800), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			auditCfg: config.NATSAudit{
				Bucket:   "audit-log",
				TTL:      "24h",
				MaxBytes: 1048576,
				Storage:  "memory",
				Replicas: 3,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "audit-log", cfg.Bucket)
				assert.Equal(suite.T(), 24*time.Hour, cfg.TTL)
				assert.Equal(suite.T(), int64(1048576), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when TTL is invalid defaults to zero",
			namespace: "",
			auditCfg: config.NATSAudit{
				Bucket:   "audit-log",
				TTL:      "invalid",
				MaxBytes: 0,
				Storage:  "file",
				Replicas: 1,
			},
			validateFn: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := cli.BuildAuditKVConfig(tc.namespace, tc.auditCfg)

			tc.validateFn(got)
		})
	}
}
