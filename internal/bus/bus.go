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

// Package bus builds the NATS options, buckets and streams that osapi uses.
package bus

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// StorageType maps "memory"/"file" strings to jetstream.StorageType.
func StorageType(
	s string,
) jetstream.StorageType {
	if s == "memory" {
		return jetstream.MemoryStorage
	}

	return jetstream.FileStorage
}

// natsCloser is implemented by any NATS client that can be closed.
type natsCloser interface {
	Close()
}

// Close safely closes a NATS client connection.
func Close(
	nc natsCloser,
) {
	nc.Close()
}

// ConnectionOptions are the options of a connection to NATS from conn: its
// host, port, auth and TLS, with name. It is an error when conn's TLS does
// not load: the connection does not fall back to plain text.
func ConnectionOptions(
	conn config.NATSConnection,
	name string,
) (*natsclient.Options, error) {
	tlsConfig, err := conn.TLS.Dial()
	if err != nil {
		return nil, fmt.Errorf("nats tls: %w", err)
	}

	return &natsclient.Options{
		Host: conn.Host,
		Port: conn.Port,
		Auth: AuthOptions(conn.Auth),
		Name: name,
		TLS:  tlsConfig,
	}, nil
}

// AuthOptions converts a config NATSAuth to natsclient.AuthOptions.
func AuthOptions(
	auth config.NATSAuth,
) natsclient.AuthOptions {
	switch auth.Type {
	case "user_pass":
		return natsclient.AuthOptions{
			AuthType: natsclient.UserPassAuth,
			Username: auth.Username,
			Password: auth.Password,
		}
	case "nkey":
		return natsclient.AuthOptions{
			AuthType: natsclient.NKeyAuth,
			NKeyFile: auth.NKeyFile,
		}
	case "creds":
		return natsclient.AuthOptions{
			AuthType:  natsclient.CredsAuth,
			CredsFile: auth.CredsFile,
		}
	default:
		return natsclient.AuthOptions{
			AuthType: natsclient.NoAuth,
		}
	}
}

// MaxControlLine is the longest protocol line, in bytes, that the embedded
// server takes. A NATS server that osapi does not run needs at least this in
// its max_control_line, in operator mode.
const MaxControlLine = 32 * 1024

// ServerOptions is the embedded server's options: host, port and
// JetStream, and the authentication serverAuth describes. For operator auth the
// options start from serverAuth.ResolverFile, read by NATS's own config parser,
// so the trusted operator, the system account and the preloaded account JWTs
// are exactly what nsc wrote; the host, port and JetStream settings are laid
// over them.
func ServerOptions(
	host string,
	port int,
	storeDir string,
	serverAuth config.NATSServerAuth,
	serverTLS config.ServerTLS,
	keyDir string,
) (*natsserver.Options, error) {
	// With TLS on and no cert_file, the server serves its self-signed
	// nats.crt in keyDir, which it makes at its first start.
	serverTLS, err := serverTLS.WithSelfSigned(keyDir, "nats")
	if err != nil {
		return nil, fmt.Errorf("tls: %w", err)
	}
	tlsConfig, err := serverTLS.Config()
	if err != nil {
		return nil, fmt.Errorf("tls: %w", err)
	}
	opts := &natsserver.Options{}
	if serverAuth.Type == "operator" {
		if serverAuth.ResolverFile == "" {
			return nil, errors.New("operator auth needs resolver_file")
		}
		read, err := natsserver.ProcessConfigFile(serverAuth.ResolverFile)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", serverAuth.ResolverFile, err)
		}
		if len(read.TrustedOperators) == 0 {
			return nil, fmt.Errorf("%s names no operator", serverAuth.ResolverFile)
		}
		opts = read
	}

	opts.Host = host
	opts.Port = port
	opts.JetStream = true
	opts.StoreDir = storeDir
	// With a JetStream key, the store is encrypted with AES-256-GCM, which
	// FIPS 140-3 approves. The server's default is ChaCha20-Poly1305, which
	// it does not. The server does not open a stream with no messages that
	// ChaCha20 encrypted, so a store from before this is removed.
	opts.JetStreamCipher = natsserver.AES
	opts.NoSigs = true
	opts.NoLog = false
	// A user JWT carries the subjects of its role, and the controller's is
	// about 7 KB: more than the 4 KB of a CONNECT line that NATS takes by
	// default.
	opts.MaxControlLine = MaxControlLine
	opts.TLSConfig = tlsConfig

	switch serverAuth.Type {
	case "user_pass":
		users := make([]*natsserver.User, 0, len(serverAuth.Users))
		for _, u := range serverAuth.Users {
			users = append(users, &natsserver.User{
				Username: u.Username,
				Password: u.Password,
			})
		}
		opts.Users = users
	case "nkey":
		nkeys := make([]*natsserver.NkeyUser, 0, len(serverAuth.NKeys))
		for _, nk := range serverAuth.NKeys {
			nkeys = append(nkeys, &natsserver.NkeyUser{Nkey: nk})
		}
		opts.Nkeys = nkeys
	}

	return opts, nil
}

// JetStreamKey reads the key of the JetStream store from the file of cfg,
// or is "" when no file is set. A file with no key is an error, because an
// empty key turns the encryption off.
func JetStreamKey(
	cfg config.NATSServerJetStream,
) (string, error) {
	if cfg.KeyFile == "" {
		return "", nil
	}
	b, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		return "", fmt.Errorf("read the JetStream key: %w", err)
	}
	key := strings.TrimSpace(string(b))
	if key == "" {
		return "", fmt.Errorf("the JetStream key file %s is empty", cfg.KeyFile)
	}

	return key, nil
}

// ServerClientAuth is how the embedded server's own setup and
// heartbeat connect to it: as its first user for user_pass, with nkey_file for
// nkey, with creds_file for operator, and without auth otherwise.
func ServerClientAuth(
	serverAuth config.NATSServerAuth,
) natsclient.AuthOptions {
	switch serverAuth.Type {
	case "user_pass":
		if len(serverAuth.Users) > 0 {
			return natsclient.AuthOptions{
				AuthType: natsclient.UserPassAuth,
				Username: serverAuth.Users[0].Username,
				Password: serverAuth.Users[0].Password,
			}
		}
	case "nkey":
		return natsclient.AuthOptions{
			AuthType: natsclient.NKeyAuth,
			NKeyFile: serverAuth.NKeyFile,
		}
	case "operator":
		return natsclient.AuthOptions{
			AuthType:  natsclient.CredsAuth,
			CredsFile: serverAuth.CredsFile,
		}
	}

	return natsclient.AuthOptions{AuthType: natsclient.NoAuth}
}

// JobKV builds a jetstream.KeyValueConfig from job KV config values.
// The returned config includes TTL, MaxBytes, Storage, and Replicas from the
// NATSKV configuration.
func JobKV(
	namespace string,
	kvCfg config.NATSKV,
) jetstream.KeyValueConfig {
	bucket := job.ApplyNamespaceToInfraName(namespace, kvCfg.Bucket)
	ttl, _ := time.ParseDuration(kvCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   bucket,
		TTL:      ttl,
		MaxBytes: kvCfg.MaxBytes,
		Storage:  StorageType(kvCfg.Storage),
		Replicas: kvCfg.Replicas,
	}
}

// RegistryKV builds a jetstream.KeyValueConfig from registry config values.
func RegistryKV(
	namespace string,
	registryCfg config.NATSRegistry,
) jetstream.KeyValueConfig {
	registryBucket := job.ApplyNamespaceToInfraName(namespace, registryCfg.Bucket)
	registryTTL, _ := time.ParseDuration(registryCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   registryBucket,
		TTL:      registryTTL,
		Storage:  StorageType(registryCfg.Storage),
		Replicas: registryCfg.Replicas,
	}
}

// FactsKV builds a jetstream.KeyValueConfig from facts config values.
func FactsKV(
	namespace string,
	factsCfg config.NATSFacts,
) jetstream.KeyValueConfig {
	factsBucket := job.ApplyNamespaceToInfraName(namespace, factsCfg.Bucket)
	factsTTL, _ := time.ParseDuration(factsCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   factsBucket,
		TTL:      factsTTL,
		Storage:  StorageType(factsCfg.Storage),
		Replicas: factsCfg.Replicas,
	}
}

// StateKV builds a jetstream.KeyValueConfig from state config values.
// The state bucket has no TTL so drain flags and timeline events persist indefinitely.
func StateKV(
	namespace string,
	stateCfg config.NATSState,
) jetstream.KeyValueConfig {
	stateBucket := job.ApplyNamespaceToInfraName(namespace, stateCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   stateBucket,
		Storage:  StorageType(stateCfg.Storage),
		Replicas: stateCfg.Replicas,
	}
}

// AuditStream builds a jetstream.StreamConfig from audit
// config values.
func AuditStream(
	namespace string,
	auditCfg config.NATSAudit,
) jetstream.StreamConfig {
	streamName := job.ApplyNamespaceToInfraName(
		namespace,
		auditCfg.Stream,
	)
	subject := job.ApplyNamespaceToSubjects(
		namespace,
		auditCfg.Subject,
	)
	maxAge, _ := time.ParseDuration(auditCfg.MaxAge)

	return jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject + ".>"},
		MaxAge:   maxAge,
		MaxBytes: auditCfg.MaxBytes,
		Storage:  StorageType(auditCfg.Storage),
		Replicas: auditCfg.Replicas,
		Discard:  jetstream.DiscardOld,
	}
}

// ObjectStore builds a jetstream.ObjectStoreConfig from objects config values.
func ObjectStore(
	namespace string,
	objectsCfg config.NATSObjects,
) jetstream.ObjectStoreConfig {
	objectsBucket := job.ApplyNamespaceToInfraName(namespace, objectsCfg.Bucket)

	return jetstream.ObjectStoreConfig{
		Bucket:   objectsBucket,
		MaxBytes: objectsCfg.MaxBytes,
		Storage:  StorageType(objectsCfg.Storage),
		Replicas: objectsCfg.Replicas,
	}
}

// FileStateKV builds a jetstream.KeyValueConfig from file-state config values.
// The file-state bucket has no TTL so deployment SHA tracking persists indefinitely.
func FileStateKV(
	namespace string,
	fileStateCfg config.NATSFileState,
) jetstream.KeyValueConfig {
	fileStateBucket := job.ApplyNamespaceToInfraName(namespace, fileStateCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   fileStateBucket,
		Storage:  StorageType(fileStateCfg.Storage),
		Replicas: fileStateCfg.Replicas,
	}
}

// EnrollmentKV builds a jetstream.KeyValueConfig from enrollment config values.
// The enrollment bucket has no TTL so pending requests persist until accepted or rejected.
func EnrollmentKV(
	namespace string,
	enrollmentCfg config.NATSEnrollment,
) jetstream.KeyValueConfig {
	enrollmentBucket := job.ApplyNamespaceToInfraName(namespace, enrollmentCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   enrollmentBucket,
		Storage:  StorageType(enrollmentCfg.Storage),
		Replicas: enrollmentCfg.Replicas,
	}
}

// SecretsKV builds a jetstream.KeyValueConfig from the secrets config
// values. The bucket has no TTL, and keeps History versions of each key.
func SecretsKV(
	namespace string,
	secretsCfg config.NATSSecrets,
) jetstream.KeyValueConfig {
	return jetstream.KeyValueConfig{
		Bucket:   job.ApplyNamespaceToInfraName(namespace, secretsCfg.Bucket),
		History:  uint8(min(max(secretsCfg.History, 1), jetstream.KeyValueMaxHistory)),
		Storage:  StorageType(secretsCfg.Storage),
		Replicas: secretsCfg.Replicas,
	}
}
