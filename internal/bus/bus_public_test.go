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

package bus_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/mocks"
	"github.com/osapi-io/osapi/internal/config"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
)

// testFS is the host's filesystem, through avfs. The certificates, keys and NATS configs here are read by path, by crypto/tls and the NATS server as well as by the code under test.
var testFS avfs.VFS = osfs.NewWithNoIdm()

type BusPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
}

func TestBusPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BusPublicTestSuite))
}

func (suite *BusPublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
}

func (suite *BusPublicTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *BusPublicTestSuite) TestCloseNATSClient() {
	tests := []struct {
		name         string
		setupFn      func() func()
		validateFunc func(assert.PanicTestFunc)
	}{
		{
			name: "when real client with nil NC does not panic",
			setupFn: func() func() {
				client := &natsclient.Client{}

				return func() {
					bus.Close(client)
				}
			},
			validateFunc: func(closeFn assert.PanicTestFunc) {
				assert.NotPanics(suite.T(), closeFn)
			},
		},
		{
			name: "when real client with non-nil NC closes connection",
			setupFn: func() func() {
				mockConn := mocks.NewMockNATSConnector(suite.ctrl)
				mockConn.EXPECT().Close()
				client := &natsclient.Client{NC: mockConn}

				return func() {
					bus.Close(client)
				}
			},
			validateFunc: func(closeFn assert.PanicTestFunc) {
				assert.NotPanics(suite.T(), closeFn)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(tc.setupFn())
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildNATSAuthOptions() {
	tests := []struct {
		name         string
		auth         config.NATSAuth
		validateFunc func(natsclient.AuthOptions)
	}{
		{
			name: "when user_pass returns user pass auth",
			auth: config.NATSAuth{
				Type:     "user_pass",
				Username: "osapi",
				Password: "secret",
			},
			validateFunc: func(got natsclient.AuthOptions) {
				assert.Equal(suite.T(), natsclient.AuthOptions{
					AuthType: natsclient.UserPassAuth,
					Username: "osapi",
					Password: "secret",
				}, got)
			},
		},
		{
			name: "when nkey returns nkey auth",
			auth: config.NATSAuth{
				Type:     "nkey",
				NKeyFile: "/path/to/nkey",
			},
			validateFunc: func(got natsclient.AuthOptions) {
				assert.Equal(suite.T(), natsclient.AuthOptions{
					AuthType: natsclient.NKeyAuth,
					NKeyFile: "/path/to/nkey",
				}, got)
			},
		},
		{
			name: "when creds returns creds auth",
			auth: config.NATSAuth{
				Type:      "creds",
				CredsFile: "/path/to/user.creds",
			},
			validateFunc: func(got natsclient.AuthOptions) {
				assert.Equal(suite.T(), natsclient.AuthOptions{
					AuthType:  natsclient.CredsAuth,
					CredsFile: "/path/to/user.creds",
				}, got)
			},
		},
		{
			name: "when none returns no auth",
			auth: config.NATSAuth{
				Type: "none",
			},
			validateFunc: func(got natsclient.AuthOptions) {
				assert.Equal(suite.T(), natsclient.AuthOptions{
					AuthType: natsclient.NoAuth,
				}, got)
			},
		},
		{
			name: "when empty type defaults to no auth",
			auth: config.NATSAuth{},
			validateFunc: func(got natsclient.AuthOptions) {
				assert.Equal(suite.T(), natsclient.AuthOptions{
					AuthType: natsclient.NoAuth,
				}, got)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.AuthOptions(tc.auth)

			tc.validateFunc(got)
		})
	}
}

// resolverFile writes what `nsc generate config --mem-resolver` writes for an
// operator and one system account, and returns its path and the account's
// public key.
func (suite *BusPublicTestSuite) resolverFile() (string, string) {
	op, err := nkeys.CreateOperator()
	suite.Require().NoError(err)
	opPub, _ := op.PublicKey()
	opJWT, err := jwt.NewOperatorClaims(opPub).Encode(op)
	suite.Require().NoError(err)

	acct, err := nkeys.CreateAccount()
	suite.Require().NoError(err)
	acctPub, _ := acct.PublicKey()
	acctJWT, err := jwt.NewAccountClaims(acctPub).Encode(op)
	suite.Require().NoError(err)

	path := filepath.Join(suite.T().TempDir(), "resolver.conf")
	content := fmt.Sprintf("operator: %s\nsystem_account: %s\nresolver: MEMORY\n"+
		"resolver_preload: {\n  %s: %s\n}\n", opJWT, acctPub, acctPub, acctJWT)
	suite.Require().NoError(avfs.WriteFile(testFS, path, []byte(content), 0o600))

	return path, acctPub
}

func (suite *BusPublicTestSuite) TestBuildNATSServerOptions() {
	resolver, sysAccount := suite.resolverFile()
	noOperator := filepath.Join(suite.T().TempDir(), "empty.conf")
	suite.Require().NoError(avfs.WriteFile(testFS, noOperator, []byte("port: 4222\n"), 0o600))

	tests := []struct {
		name         string
		auth         config.NATSServerAuth
		wantErr      string
		validateFunc func(*natsserver.Options)
	}{
		{
			name: "when operator, the resolver file's operator and accounts are trusted",
			auth: config.NATSServerAuth{Type: "operator", ResolverFile: resolver},
			validateFunc: func(o *natsserver.Options) {
				suite.Len(o.TrustedOperators, 1)
				suite.Equal(sysAccount, o.SystemAccount)
				suite.NotNil(o.AccountResolver)
			},
		},
		{
			name:    "when operator without a resolver file",
			auth:    config.NATSServerAuth{Type: "operator"},
			wantErr: "operator auth needs resolver_file",
		},
		{
			name:    "when the resolver file cannot be read",
			auth:    config.NATSServerAuth{Type: "operator", ResolverFile: "/nonexistent.conf"},
			wantErr: "read /nonexistent.conf",
		},
		{
			name:    "when the resolver file names no operator",
			auth:    config.NATSServerAuth{Type: "operator", ResolverFile: noOperator},
			wantErr: "names no operator",
		},
		{
			name: "when user_pass, its users",
			auth: config.NATSServerAuth{
				Type:  "user_pass",
				Users: []config.NATSServerUser{{Username: "osapi", Password: "secret"}},
			},
			validateFunc: func(o *natsserver.Options) {
				suite.Require().Len(o.Users, 1)
				suite.Equal("osapi", o.Users[0].Username)
			},
		},
		{
			name: "when nkey, its public keys",
			auth: config.NATSServerAuth{Type: "nkey", NKeys: []string{"UABC"}},
			validateFunc: func(o *natsserver.Options) {
				suite.Require().Len(o.Nkeys, 1)
				suite.Equal("UABC", o.Nkeys[0].Nkey)
			},
		},
		{
			name: "when none, no auth",
			auth: config.NATSServerAuth{Type: "none"},
			validateFunc: func(o *natsserver.Options) {
				suite.Empty(o.Users)
				suite.Empty(o.TrustedOperators)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got, err := bus.ServerOptions(
				"0.0.0.0",
				4222,
				"/var/lib/nats",
				tc.auth,
				config.ServerTLS{},
				"",
			)

			if tc.wantErr != "" {
				suite.ErrorContains(err, tc.wantErr)

				return
			}
			suite.Require().NoError(err)
			suite.Equal("0.0.0.0", got.Host)
			suite.Equal(4222, got.Port)
			suite.True(got.JetStream)
			suite.Equal("/var/lib/nats", got.StoreDir)
			suite.Equal(natsserver.AES, got.JetStreamCipher)
			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestJetStreamKey() {
	dir := suite.T().TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		suite.Require().NoError(avfs.WriteFile(testFS, path, []byte(content), 0o600))

		return path
	}
	key := write("js.key", "  thekey\n")
	empty := write("empty.key", " \n")

	tests := []struct {
		name    string
		cfg     config.NATSServerJetStream
		want    string
		wantErr string
	}{
		{
			name: "when no file is set, no key",
		},
		{
			name: "when key_file is set, its key without whitespace",
			cfg:  config.NATSServerJetStream{KeyFile: key},
			want: "thekey",
		},
		{
			name:    "when key_file does not exist",
			cfg:     config.NATSServerJetStream{KeyFile: filepath.Join(dir, "missing.key")},
			wantErr: "read the JetStream key",
		},
		{
			name:    "when key_file is empty",
			cfg:     config.NATSServerJetStream{KeyFile: empty},
			wantErr: "is empty",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got, err := bus.JetStreamKey(tc.cfg)

			if tc.wantErr != "" {
				suite.ErrorContains(err, tc.wantErr)
				suite.Empty(got)

				return
			}
			suite.Require().NoError(err)
			suite.Equal(tc.want, got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildNATSServerClientAuth() {
	tests := []struct {
		name string
		auth config.NATSServerAuth
		want natsclient.AuthOptions
	}{
		{
			name: "when user_pass, as its first user",
			auth: config.NATSServerAuth{
				Type:  "user_pass",
				Users: []config.NATSServerUser{{Username: "osapi", Password: "secret"}},
			},
			want: natsclient.AuthOptions{
				AuthType: natsclient.UserPassAuth,
				Username: "osapi",
				Password: "secret",
			},
		},
		{
			name: "when user_pass with no users, without auth",
			auth: config.NATSServerAuth{Type: "user_pass"},
			want: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		},
		{
			name: "when nkey, with its seed",
			auth: config.NATSServerAuth{Type: "nkey", NKeyFile: "/etc/osapi/server.nk"},
			want: natsclient.AuthOptions{
				AuthType: natsclient.NKeyAuth,
				NKeyFile: "/etc/osapi/server.nk",
			},
		},
		{
			name: "when operator, with its credentials",
			auth: config.NATSServerAuth{Type: "operator", CredsFile: "/etc/osapi/server.creds"},
			want: natsclient.AuthOptions{
				AuthType:  natsclient.CredsAuth,
				CredsFile: "/etc/osapi/server.creds",
			},
		},
		{
			name: "when none, without auth",
			auth: config.NATSServerAuth{Type: "none"},
			want: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Equal(tc.want, bus.ServerClientAuth(tc.auth))
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildJobKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		kvCfg        config.NATSKV
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			kvCfg: config.NATSKV{
				Bucket:   "job-queue",
				TTL:      "1h",
				MaxBytes: 104857600,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-job-queue", cfg.Bucket)
				assert.Equal(suite.T(), 1*time.Hour, cfg.TTL)
				assert.Equal(suite.T(), int64(104857600), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			kvCfg: config.NATSKV{
				Bucket:   "job-queue",
				TTL:      "30m",
				MaxBytes: 52428800,
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "job-queue", cfg.Bucket)
				assert.Equal(suite.T(), 30*time.Minute, cfg.TTL)
				assert.Equal(suite.T(), int64(52428800), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when TTL is invalid defaults to zero",
			namespace: "",
			kvCfg: config.NATSKV{
				Bucket:   "job-queue",
				TTL:      "invalid",
				MaxBytes: 0,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.JobKV(tc.namespace, tc.kvCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildRegistryKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		registryCfg  config.NATSRegistry
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			registryCfg: config.NATSRegistry{
				Bucket:   "agent-registry",
				TTL:      "30s",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-agent-registry", cfg.Bucket)
				assert.Equal(suite.T(), 30*time.Second, cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			registryCfg: config.NATSRegistry{
				Bucket:   "agent-registry",
				TTL:      "1m",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "agent-registry", cfg.Bucket)
				assert.Equal(suite.T(), 1*time.Minute, cfg.TTL)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when TTL is invalid defaults to zero",
			namespace: "",
			registryCfg: config.NATSRegistry{
				Bucket:   "agent-registry",
				TTL:      "invalid",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.RegistryKV(tc.namespace, tc.registryCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildFactsKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		factsCfg     config.NATSFacts
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			factsCfg: config.NATSFacts{
				Bucket:   "agent-facts",
				TTL:      "1h",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-agent-facts", cfg.Bucket)
				assert.Equal(suite.T(), 1*time.Hour, cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			factsCfg: config.NATSFacts{
				Bucket:   "agent-facts",
				TTL:      "30m",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "agent-facts", cfg.Bucket)
				assert.Equal(suite.T(), 30*time.Minute, cfg.TTL)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when TTL is invalid defaults to zero",
			namespace: "",
			factsCfg: config.NATSFacts{
				Bucket:   "agent-facts",
				TTL:      "invalid",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.FactsKV(tc.namespace, tc.factsCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildStateKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		stateCfg     config.NATSState
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			stateCfg: config.NATSState{
				Bucket:   "agent-state",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-agent-state", cfg.Bucket)
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			stateCfg: config.NATSState{
				Bucket:   "agent-state",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "agent-state", cfg.Bucket)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.StateKV(tc.namespace, tc.stateCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildAuditStreamConfig() {
	tests := []struct {
		name         string
		namespace    string
		auditCfg     config.NATSAudit
		validateFunc func(jetstream.StreamConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			auditCfg: config.NATSAudit{
				Stream:   "AUDIT",
				Subject:  "audit",
				MaxAge:   "720h",
				MaxBytes: 52428800,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.StreamConfig) {
				assert.Equal(suite.T(), "osapi-AUDIT", cfg.Name)
				assert.Equal(suite.T(), []string{"osapi.audit.>"}, cfg.Subjects)
				assert.Equal(suite.T(), 720*time.Hour, cfg.MaxAge)
				assert.Equal(suite.T(), int64(52428800), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
				assert.Equal(suite.T(), jetstream.DiscardOld, cfg.Discard)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			auditCfg: config.NATSAudit{
				Stream:   "AUDIT",
				Subject:  "audit",
				MaxAge:   "24h",
				MaxBytes: 1048576,
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.StreamConfig) {
				assert.Equal(suite.T(), "AUDIT", cfg.Name)
				assert.Equal(suite.T(), []string{"audit.>"}, cfg.Subjects)
				assert.Equal(suite.T(), 24*time.Hour, cfg.MaxAge)
				assert.Equal(suite.T(), int64(1048576), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
				assert.Equal(suite.T(), jetstream.DiscardOld, cfg.Discard)
			},
		},
		{
			name:      "when MaxAge is invalid defaults to zero",
			namespace: "",
			auditCfg: config.NATSAudit{
				Stream:   "AUDIT",
				Subject:  "audit",
				MaxAge:   "invalid",
				MaxBytes: 0,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.StreamConfig) {
				assert.Equal(suite.T(), time.Duration(0), cfg.MaxAge)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.AuditStream(tc.namespace, tc.auditCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildObjectStoreConfig() {
	tests := []struct {
		name         string
		namespace    string
		objectsCfg   config.NATSObjects
		validateFunc func(jetstream.ObjectStoreConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			objectsCfg: config.NATSObjects{
				Bucket:   "file-objects",
				MaxBytes: 104857600,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.ObjectStoreConfig) {
				assert.Equal(suite.T(), "osapi-file-objects", cfg.Bucket)
				assert.Equal(suite.T(), int64(104857600), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			objectsCfg: config.NATSObjects{
				Bucket:   "file-objects",
				MaxBytes: 52428800,
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.ObjectStoreConfig) {
				assert.Equal(suite.T(), "file-objects", cfg.Bucket)
				assert.Equal(suite.T(), int64(52428800), cfg.MaxBytes)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
		{
			name:      "when max_bytes is zero",
			namespace: "osapi",
			objectsCfg: config.NATSObjects{
				Bucket:   "file-objects",
				MaxBytes: 0,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.ObjectStoreConfig) {
				assert.Equal(suite.T(), "osapi-file-objects", cfg.Bucket)
				assert.Equal(suite.T(), int64(0), cfg.MaxBytes)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.ObjectStore(tc.namespace, tc.objectsCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildEnrollmentKVConfig() {
	tests := []struct {
		name          string
		namespace     string
		enrollmentCfg config.NATSEnrollment
		validateFunc  func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			enrollmentCfg: config.NATSEnrollment{
				Bucket:   "agent-enrollment",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-agent-enrollment", cfg.Bucket)
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			enrollmentCfg: config.NATSEnrollment{
				Bucket:   "agent-enrollment",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "agent-enrollment", cfg.Bucket)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.EnrollmentKV(tc.namespace, tc.enrollmentCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildFileStateKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		fileStateCfg config.NATSFileState
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set",
			namespace: "osapi",
			fileStateCfg: config.NATSFileState{
				Bucket:   "file-state",
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-file-state", cfg.Bucket)
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:      "when namespace is empty",
			namespace: "",
			fileStateCfg: config.NATSFileState{
				Bucket:   "file-state",
				Storage:  "memory",
				Replicas: 3,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "file-state", cfg.Bucket)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
				assert.Equal(suite.T(), 3, cfg.Replicas)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.FileStateKV(tc.namespace, tc.fileStateCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestBuildSecretsKVConfig() {
	tests := []struct {
		name         string
		namespace    string
		secretsCfg   config.NATSSecrets
		validateFunc func(jetstream.KeyValueConfig)
	}{
		{
			name:      "when namespace is set keeps the history of each key",
			namespace: "osapi",
			secretsCfg: config.NATSSecrets{
				Bucket:   "secrets",
				History:  10,
				Storage:  "file",
				Replicas: 1,
			},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "osapi-secrets", cfg.Bucket)
				assert.Equal(suite.T(), uint8(10), cfg.History)
				assert.Equal(suite.T(), time.Duration(0), cfg.TTL)
				assert.Equal(suite.T(), jetstream.FileStorage, cfg.Storage)
				assert.Equal(suite.T(), 1, cfg.Replicas)
			},
		},
		{
			name:       "when no history is set keeps one version",
			secretsCfg: config.NATSSecrets{Bucket: "secrets", Storage: "memory"},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), "secrets", cfg.Bucket)
				assert.Equal(suite.T(), uint8(1), cfg.History)
				assert.Equal(suite.T(), jetstream.MemoryStorage, cfg.Storage)
			},
		},
		{
			name:       "when the history is above the limit of NATS keeps the limit",
			secretsCfg: config.NATSSecrets{Bucket: "secrets", History: 100},
			validateFunc: func(cfg jetstream.KeyValueConfig) {
				assert.Equal(suite.T(), uint8(jetstream.KeyValueMaxHistory), cfg.History)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got := bus.SecretsKV(tc.namespace, tc.secretsCfg)

			tc.validateFunc(got)
		})
	}
}

func (suite *BusPublicTestSuite) TestServerOptionsTLS() {
	cert, key := writeCert(suite.T())
	got, err := bus.ServerOptions("0.0.0.0", 4222, "/var/lib/nats",
		config.NATSServerAuth{Type: "none"},
		config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}, suite.T().TempDir())
	suite.Require().NoError(err)
	suite.NotNil(got.TLSConfig)

	// With no cert_file, the server makes its self-signed nats.crt.
	dir := suite.T().TempDir()
	got, err = bus.ServerOptions("0.0.0.0", 4222, "/var/lib/nats",
		config.NATSServerAuth{Type: "none"}, config.ServerTLS{Enabled: true}, dir)
	suite.Require().NoError(err)
	suite.NotNil(got.TLSConfig)
	suite.FileExists(filepath.Join(dir, "nats.crt"))

	blocked := filepath.Join(suite.T().TempDir(), "file")
	suite.Require().NoError(avfs.WriteFile(testFS, blocked, nil, 0o600))
	_, err = bus.ServerOptions("0.0.0.0", 4222, "/var/lib/nats",
		config.NATSServerAuth{Type: "none"}, config.ServerTLS{Enabled: true},
		filepath.Join(blocked, "pki"))
	suite.ErrorContains(err, "tls: make the key directory")

	// A cert_file that does not load is the error.
	_, err = bus.ServerOptions("0.0.0.0", 4222, "/var/lib/nats",
		config.NATSServerAuth{Type: "none"},
		config.ServerTLS{Enabled: true, CertFile: "/no/such.crt", KeyFile: "/no/such.key"}, "")
	suite.ErrorContains(err, "tls: load the server certificate")
}

func (suite *BusPublicTestSuite) TestConnectionOptions() {
	conn := config.NATSConnection{
		Host: "nats.lab",
		Port: 4222,
		Auth: config.NATSAuth{Type: "none"},
	}
	got, err := bus.ConnectionOptions(conn, "osapi-agent")
	suite.Require().NoError(err)
	suite.Equal("nats.lab", got.Host)
	suite.Equal("osapi-agent", got.Name)
	suite.Nil(got.TLS, "plain text when TLS is off")

	conn.TLS = config.ClientTLS{Enabled: true, ServerName: "hub.lab"}
	got, err = bus.ConnectionOptions(conn, "osapi-agent")
	suite.Require().NoError(err)
	suite.Equal("hub.lab", got.TLS.ServerName)

	conn.TLS = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.crt"}
	_, err = bus.ConnectionOptions(conn, "osapi-agent")
	suite.ErrorContains(err, "nats tls: ")
}
