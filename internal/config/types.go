// Copyright (c) 2024 John Dewey

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

package config

import (
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Config represents the root structure of the YAML configuration file.
// This struct is used to unmarshal configuration data from Viper.
type Config struct {
	Controller Controller  `mapstructure:"controller"`
	Agent      AgentConfig `mapstructure:"agent,omitempty"`
	NATS       NATS        `mapstructure:"nats"`
	Telemetry  Telemetry   `mapstructure:"telemetry"`
	// Debug enable or disable debug option set from CLI.
	Debug bool `mapstructure:"debug" doc:"Enable verbose logging."`
}

// NotificationsConfig holds settings for the pluggable condition notification
// system. When Enabled is true, a Watcher monitors the registry KV bucket and
// dispatches ConditionEvents via the configured Notifier.
type NotificationsConfig struct {
	// Enabled activates the condition watcher and notifier.
	Enabled bool `mapstructure:"enabled" doc:"Enable the condition watcher and notifier."`
	// Notifier selects the notification backend: "log" (default).
	Notifier string `mapstructure:"notifier" doc:"Notifier backend: 'log' (default)."`
	// RenotifyInterval is how often to re-fire active conditions.
	// Uses Go duration format (e.g., "1m", "5m", "1h"). Zero disables.
	RenotifyInterval string `mapstructure:"renotify_interval" doc:"How often to re-fire active conditions (Go duration). Zero disables re-notification." validate:"omitempty,go_duration"`
}

// Telemetry configuration settings.
type Telemetry struct {
	Tracing TracingConfig `mapstructure:"tracing,omitempty"`
}

// MetricsServer configures the per-component metrics HTTP server.
type MetricsServer struct {
	// Enabled activates the metrics server.
	Enabled bool `mapstructure:"enabled" doc:"Enable the metrics endpoint."`
	// Host the metrics server binds to.
	Host string `mapstructure:"host" doc:"Host the metrics server binds to."`
	// Port the metrics server listens on.
	Port int `mapstructure:"port" doc:"Port the metrics server listens on." validate:"omitempty,min=1,max=65535"`
	// URL is where a person opens the metrics endpoint. Health shows it
	// beside the bind address; osapi never requests it.
	URL string `mapstructure:"url" doc:"Where a person opens the metrics endpoint, shown by health beside the bind address." validate:"omitempty,http_url" choice:"where a person reaches it; empty shows the bind address"`
	// TLS makes the metrics server serve HTTPS.
	TLS ServerTLS `mapstructure:"tls"`
}

// TracingConfig configuration settings for distributed tracing.
type TracingConfig struct {
	// Enabled enables or disables tracing.
	Enabled bool `mapstructure:"enabled" doc:"Enable distributed tracing."`
	// Exporter selects the trace exporter: "stdout" or "otlp".
	Exporter string `mapstructure:"exporter" doc:"Trace exporter: 'stdout' or 'otlp'." choice:"empty sends no trace"`
	// OTLPEndpoint is the gRPC endpoint for the OTLP exporter (e.g., "localhost:4317").
	OTLPEndpoint string `mapstructure:"otlp_endpoint" doc:"gRPC endpoint for the OTLP exporter, such as 'localhost:4317'." choice:"the collector of the operator; set with exporter otlp"`
	// URL is where a person views the traces, such as a Jaeger UI. Health
	// links to it, and the controller reads a trace from its query API for
	// osapi client telemetry trace get. osapi does not run the backend.
	URL string `mapstructure:"url" doc:"Where a person views the traces, such as a Jaeger UI. Health links to it, and the controller reads traces from its query API. osapi does not run the backend." validate:"omitempty,http_url" choice:"the tracing backend of the operator; empty reads no trace"`
	// TLS is how osapi checks the tracing backend: the OTLP exporter, and
	// the controller's reads of its query API at url.
	TLS ClientTLS `mapstructure:"tls"`
}

// NATSAuth holds client-side authentication settings for connecting to NATS.
type NATSAuth struct {
	// Type is the auth method: "none", "user_pass", "nkey", or "creds".
	Type string `mapstructure:"type" doc:"Authentication type: 'none', 'user_pass', 'nkey', or 'creds'."`
	// Username for user_pass auth.
	Username string `mapstructure:"username" doc:"Username for user_pass auth." choice:"set for the auth type that uses it"`
	// Password for user_pass auth.
	Password string `mapstructure:"password" doc:"Password for user_pass auth." mask:"password" noref:"a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault" choice:"set for the auth type that uses it"`
	// NKeyFile path to the NKey seed file for nkey auth.
	NKeyFile string `mapstructure:"nkey_file" doc:"Path to the NKey seed file for nkey auth." choice:"set for the auth type that uses it"`
	// CredsFile path to a NATS credentials file, a user JWT and its seed, for
	// creds auth against a server in operator (JWT) mode.
	CredsFile string `mapstructure:"creds_file" doc:"Path to a NATS credentials file (user JWT and seed) for creds auth." choice:"set for the auth type that uses it"`
}

// NATSServerAuth holds server-side authentication settings for the embedded NATS server.
type NATSServerAuth struct {
	// Type is the auth method: "none", "user_pass", "nkey", or "operator".
	Type string `mapstructure:"type" doc:"Auth method: 'none', 'user_pass', 'nkey' or 'operator'."`
	// ResolverFile is, for operator auth, the server's half of an operator:
	// its JWT, the system account and each account's JWT, as
	// `nsc generate config --mem-resolver` writes it.
	ResolverFile string `mapstructure:"resolver_file" doc:"For operator auth: the operator, system account and account JWTs, as nsc generate config --mem-resolver writes them." choice:"set for the auth type that uses it"`
	// NKeyFile is, for nkey auth, a seed whose public key is in NKeys: the
	// identity the server's own setup and heartbeat connect with.
	NKeyFile string `mapstructure:"nkey_file" doc:"For nkey auth: a seed among nkeys, which the server's own setup and heartbeat connect with." choice:"set for the auth type that uses it"`
	// CredsFile is, for operator auth, the credentials the server's own setup
	// and heartbeat connect with: a user in an account with JetStream.
	CredsFile string `mapstructure:"creds_file" doc:"For operator auth: a credentials file the server's own setup and heartbeat connect with." choice:"set for the auth type that uses it"`
	// Users allowed to connect (for user_pass auth).
	Users []NATSServerUser `mapstructure:"users" doc:"Users allowed to connect (for user_pass auth)." choice:"set for the auth type that uses it"`
	// NKeys is a list of allowed public NKeys (for nkey auth).
	NKeys []string `mapstructure:"nkeys" doc:"Allowed public NKeys, for nkey auth." choice:"set for the auth type that uses it"`
	// NKeyRoles is the osapi role of each nkey of NKeys, by its public key.
	NKeyRoles map[string]string `mapstructure:"nkey_roles" doc:"The osapi role of each nkey, by its public key: controller, agent, setup or join. The nkey can then use only the subjects of that role. An nkey with no role can use each subject." choice:"empty: each nkey can use each subject"`
}

// NATSServerUser represents an allowed username/password pair for the NATS server.
type NATSServerUser struct {
	// Username for the user.
	Username string `mapstructure:"username"`
	// Password for the user.
	Password string `mapstructure:"password" mask:"password" noref:"a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault"`
	// Role limits the user to the subjects of one osapi process.
	Role string `mapstructure:"role" validate:"omitempty,oneof=controller agent setup join" doc:"The osapi process that the user is: controller, agent, setup, or join, the bootstrap user of a join over NATS. The user can then use only the subjects of that process. Empty lets the user use each subject."`
}

// NATS configuration settings.
type NATS struct {
	Server     NATSServer     `mapstructure:"server,omitempty"`
	Stream     NATSStream     `mapstructure:"stream,omitempty"`
	KV         NATSKV         `mapstructure:"kv,omitempty"`
	DLQ        NATSDLQ        `mapstructure:"dlq,omitempty"`
	Audit      NATSAudit      `mapstructure:"audit,omitempty"`
	Registry   NATSRegistry   `mapstructure:"registry,omitempty"`
	Facts      NATSFacts      `mapstructure:"facts,omitempty"`
	State      NATSState      `mapstructure:"state,omitempty"`
	Objects    NATSObjects    `mapstructure:"objects,omitempty"`
	FileState  NATSFileState  `mapstructure:"file_state,omitempty"`
	Enrollment NATSEnrollment `mapstructure:"enrollment,omitempty"`
	Secrets    NATSSecrets    `mapstructure:"secrets,omitempty"`
}

// NATSSecrets configuration for the KV bucket of the nats vaults. Each value
// in it is encrypted, under <vault>.<key>.
type NATSSecrets struct {
	// Bucket is the KV bucket name for the values of the nats vaults.
	Bucket string `mapstructure:"bucket" validate:"omitempty,nats_name" doc:"KV bucket for the encrypted values of each nats vault."`
	// History is how many versions of each key the bucket keeps.
	History  int    `mapstructure:"history"  validate:"omitempty,min=1,max=64" doc:"How many versions of each secret the bucket keeps, from 1 to 64."`
	Storage  string `mapstructure:"storage"                                    doc:"Storage backend: 'file' or 'memory'."`
	Replicas int    `mapstructure:"replicas"                                   doc:"Number of KV replicas."`
}

// NATSEnrollment configuration for the PKI enrollment KV bucket.
// No TTL — pending enrollment requests persist until accepted or rejected.
type NATSEnrollment struct {
	// Bucket is the KV bucket name for pending enrollment entries.
	Bucket   string `mapstructure:"bucket"   validate:"omitempty,nats_name" doc:"KV bucket for pending enrollment entries."`
	Storage  string `mapstructure:"storage"                                 doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                doc:"Number of replicas for the enrollment KV bucket."`
}

// NATSAudit configuration for the audit log stream.
type NATSAudit struct {
	// Stream is the JetStream stream name for audit log entries.
	Stream string `mapstructure:"stream" doc:"JetStream stream name for audit log entries."`
	// Subject is the base subject prefix for audit messages.
	Subject  string `mapstructure:"subject"   doc:"Base subject prefix for audit messages."`
	MaxAge   string `mapstructure:"max_age"   doc:"Maximum age of audit entries (Go duration). Default 30 days." validate:"omitempty,go_duration"` // e.g. "720h" (30 days)
	MaxBytes int64  `mapstructure:"max_bytes" doc:"Maximum total size of the audit stream in bytes."`
	Storage  string `mapstructure:"storage"   doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"  doc:"Number of stream replicas."`
}

// NATSRegistry configuration for the agent registry KV bucket.
type NATSRegistry struct {
	// Bucket is the KV bucket name for agent registration entries.
	Bucket   string `mapstructure:"bucket"   validate:"required,nats_name"    doc:"KV bucket for agent heartbeat registration."`
	TTL      string `mapstructure:"ttl"      validate:"omitempty,go_duration" doc:"TTL for registry entries (Go duration). Agents refresh every 10s; the TTL acts as a liveness timeout."` // e.g. "30s"
	Storage  string `mapstructure:"storage"                                   doc:"Storage backend: 'file' or 'memory'."`                                                                  // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                  doc:"Number of KV replicas."`
}

// NATSFacts configuration for the agent facts KV bucket.
type NATSFacts struct {
	// Bucket is the KV bucket name for agent facts entries.
	Bucket   string `mapstructure:"bucket"   validate:"omitempty,nats_name"   doc:"KV bucket for agent facts entries."`
	TTL      string `mapstructure:"ttl"      validate:"omitempty,go_duration" doc:"TTL for facts entries (Go duration)."` // e.g. "1h"
	Storage  string `mapstructure:"storage"                                   doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                  doc:"Number of KV replicas."`
}

// NATSState configuration for the agent state KV bucket (drain flags, timeline events).
type NATSState struct {
	// Bucket is the KV bucket name for persistent agent state.
	Bucket   string `mapstructure:"bucket"   validate:"omitempty,nats_name" doc:"KV bucket for persistent agent state (drain flags, timeline events). No TTL — operator actions persist indefinitely."`
	Storage  string `mapstructure:"storage"                                 doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                doc:"Number of KV replicas."`
}

// NATSObjects configuration for the NATS Object Store bucket.
type NATSObjects struct {
	// Bucket is the Object Store bucket name for file content.
	Bucket   string `mapstructure:"bucket"    validate:"omitempty,nats_name" doc:"Object Store bucket for uploaded file content."`
	MaxBytes int64  `mapstructure:"max_bytes"                                doc:"Maximum total size of the bucket in bytes."`
	Storage  string `mapstructure:"storage"                                  doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                 doc:"Number of Object Store replicas."`
}

// NATSFileState configuration for the file deployment state KV bucket.
// No TTL — deployed file state persists until explicitly removed.
type NATSFileState struct {
	// Bucket is the KV bucket name for file deployment SHA tracking.
	Bucket   string `mapstructure:"bucket"   validate:"omitempty,nats_name" doc:"KV bucket for file deploy state tracking. No TTL — state persists until explicitly removed."`
	Storage  string `mapstructure:"storage"                                 doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                doc:"Number of KV replicas."`
}

// NATSServer configuration settings for the embedded NATS server.
type NATSServer struct {
	// Host the server will bind to.
	Host string `mapstructure:"host" doc:"Host the embedded NATS server binds to."`
	// Port the server will bind to.
	Port int `mapstructure:"port" doc:"Port the embedded NATS server listens on."`
	// URL is where a person reaches the server, as nats:// or tls://. Health
	// shows it beside the bind address.
	URL string `mapstructure:"url" doc:"Where a person reaches the NATS server, as nats:// or tls://, shown by health beside the bind address." validate:"omitempty,url,startswith=nats://|startswith=tls://" choice:"where a person reaches it; empty shows the bind address"`
	// StoreDir the directory for JetStream file storage.
	StoreDir string `mapstructure:"store_dir" doc:"Directory for JetStream file storage."`
	// KeyDir is where the server keeps its self-signed certificates, of NATS
	// and of its metrics, when TLS is on with no cert_file.
	KeyDir string `mapstructure:"key_dir" doc:"Directory of the server's self-signed certificates, nats.crt and nats-metrics.crt, which it makes at its first start when its TLS is on with no cert_file."`
	// JetStream holds the encryption of the JetStream store.
	JetStream NATSServerJetStream `mapstructure:"jetstream"`
	// Namespace is a prefix for all NATS subjects and infrastructure names.
	Namespace string `mapstructure:"namespace" doc:"Prefix for all NATS subjects and infrastructure names." validate:"omitempty,nats_name"`
	// Auth holds server-side authentication configuration.
	Auth    NATSServerAuth `mapstructure:"auth,omitempty"`
	Metrics MetricsServer  `mapstructure:"metrics"`
	// TLS makes the server serve TLS: each client then connects with tls://.
	TLS ServerTLS `mapstructure:"tls"`
}

// NATSServerJetStream is the encryption of the JetStream store of the
// embedded server. With no KeyFile, the store is not encrypted.
type NATSServerJetStream struct {
	// KeyFile is the file of the key that encrypts the store.
	KeyFile string `mapstructure:"key_file" doc:"File of the key that encrypts each stream and bucket in store_dir. Unset, the store is not encrypted. osapi pki key create --name jetstream makes it." choice:"empty leaves the store unencrypted; osapi pki key create --name jetstream makes one"`
}

// NATSStream configuration for JetStream stream settings.
type NATSStream struct {
	// Name is the JetStream stream name.
	Name string `mapstructure:"name" validate:"required" doc:"JetStream stream name for job notifications."`
	// Subjects is the subject filter for the stream.
	Subjects string `mapstructure:"subjects" validate:"required"              doc:"Subject filter for the stream."`
	MaxAge   string `mapstructure:"max_age"  validate:"omitempty,go_duration" doc:"Maximum age of messages in the stream (Go duration)."` // e.g. "24h", "1h30m"
	MaxMsgs  int64  `mapstructure:"max_msgs"                                  doc:"Maximum number of messages retained."`
	Storage  string `mapstructure:"storage"                                   doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                  doc:"Number of stream replicas (1 for single-node)."`
	Discard  string `mapstructure:"discard"                                   doc:"Discard policy when limits are reached: 'old' or 'new'."` // "old" or "new"
}

// NATSKV configuration for KeyValue bucket settings.
type NATSKV struct {
	// Bucket is the KV bucket name for job definitions, status events and
	// agent results.
	Bucket   string `mapstructure:"bucket"    validate:"required,nats_name" doc:"KV bucket for job definitions, status events and each agent's result."`
	TTL      string `mapstructure:"ttl"                                     doc:"How long a job's definition, status events and results stay in the bucket (Go duration), as nats setup sets it. Empty keeps them until max_bytes."` // e.g. "24h"
	MaxBytes int64  `mapstructure:"max_bytes"                               doc:"Maximum total size of the bucket in bytes."`
	Storage  string `mapstructure:"storage"                                 doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                doc:"Number of KV replicas."`
}

// KVBucketInfo holds a KV bucket's human-readable name and its configured
// bucket name. It is returned by NATS.AllKVBuckets so callers can iterate
// all KV buckets without manually listing every sub-config field.
type KVBucketInfo struct {
	// Name is a human-readable label for the bucket (e.g. "job-queue").
	Name string
	// Key is the config key that sets the bucket (e.g. "nats.kv.bucket").
	Key string
	// Bucket is the bucket name from the config field.
	Bucket string
}

// ObjectStoreBucketInfo holds an Object Store bucket's human-readable name
// and its configured bucket name. It is returned by
// NATS.AllObjectStoreBuckets so callers can iterate all Object Store buckets
// without manually listing every sub-config field.
type ObjectStoreBucketInfo struct {
	// Name is a human-readable label for the bucket (e.g. "file-objects").
	Name string
	// Key is the config key that sets the bucket (e.g. "nats.objects.bucket").
	Key string
	// Bucket is the bucket name from the config field.
	Bucket string
}

// NATSDLQ configuration for Dead Letter Queue stream settings.
type NATSDLQ struct {
	MaxAge   string `mapstructure:"max_age"  validate:"omitempty,go_duration" doc:"Maximum age of messages in the DLQ."` // e.g. "7d", "24h"
	MaxMsgs  int64  `mapstructure:"max_msgs"                                  doc:"Maximum number of messages retained."`
	Storage  string `mapstructure:"storage"                                   doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                  doc:"Number of DLQ replicas."`
}

// NATSConnection is a reusable NATS connection configuration block.
type NATSConnection struct {
	// Host the NATS server hostname.
	Host string `mapstructure:"host" doc:"NATS server hostname to connect to."`
	// Port the NATS server port.
	Port int `mapstructure:"port" doc:"NATS server port to connect to." validate:"min=1,max=65535"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name" doc:"Client name sent to NATS for identification."`
	// Namespace is a prefix for all NATS subjects used by this client.
	Namespace string `mapstructure:"namespace" doc:"Subject namespace prefix. Must match nats.server.namespace." validate:"omitempty,nats_name"`
	// Auth holds client-side authentication configuration.
	Auth NATSAuth `mapstructure:"auth,omitempty"`
	// TLS is how the connection checks a NATS server that serves TLS.
	TLS ClientTLS `mapstructure:"tls"`
}

// Controller holds the control plane configuration.
type Controller struct {
	Client        Client              `mapstructure:"client"`
	API           APIServer           `mapstructure:"api"`
	NATS          NATSConnection      `mapstructure:"nats"`
	Metrics       MetricsServer       `mapstructure:"metrics"`
	Notifications NotificationsConfig `mapstructure:"notifications,omitempty"`
	// UI holds settings for the embedded management UI.
	UI UIConfig `mapstructure:"ui,omitempty"`
	// PKI holds PKI enrollment and signing settings.
	PKI ControllerPKI `mapstructure:"pki,omitempty"`
	// Secrets holds the vaults that the controller resolves secrets from.
	Secrets ControllerSecrets `mapstructure:"secrets,omitempty"`
	// Join is the join of an agent over NATS.
	Join ControllerJoin `mapstructure:"join,omitempty"`
}

// ControllerJoin is the join of an agent over NATS: where the agent reaches
// NATS, the bootstrap user that the join string carries, and the user that
// the join gives the agent. With the embedded server, the two users default
// to the users of nats.server.auth.users with the roles join and agent.
type ControllerJoin struct {
	// URL is where an agent reaches NATS, in the join string.
	URL string `mapstructure:"url" doc:"Where an agent reaches NATS, nats://host:port or tls://host:port, which the join string carries. Empty is nats.server.url." validate:"omitempty,url,startswith=nats://|startswith=tls://" choice:"where the agents reach NATS; empty is nats.server.url"`
	// User is the bootstrap user, which can only send the join request.
	User NATSUser `mapstructure:"user" doc:"The bootstrap NATS user that the join string carries: a user that can only send the join request. Empty is the user of nats.server.auth.users with the role join." choice:"the deployment's own user; empty takes the embedded server's"`
	// SigningKeyFile is, in operator mode, an account signing key seed, which
	// the controller mints the users of a join with.
	SigningKeyFile string `mapstructure:"signing_key_file" doc:"For a NATS server in operator mode: a file with a signing key seed (SA...) of osapi's account. The controller mints the join's bootstrap user, and each agent's own user, with it. Empty gives the static users." choice:"set for operator mode"`
	// AgentTTL is how long the JWT of an agent's NATS user lasts, in
	// operator mode. The agent renews it before then.
	AgentTTL string `mapstructure:"agent_ttl" doc:"For operator mode: how long the JWT of an agent's NATS user lasts, as a Go duration such as 24h. The agent renews it when a third is left, and the controller renews only an admitted machine: a revoked agent is cut from NATS when its JWT expires. Empty never expires." validate:"omitempty,go_duration" choice:"empty: the JWT never expires"`
	// JoinSigningKeyFile is the signing key of the join's bootstrap users,
	// when it is not SigningKeyFile.
	JoinSigningKeyFile string `mapstructure:"join_signing_key_file" doc:"For operator mode: a signing key seed (SA...) of osapi's account for the join's bootstrap users. Empty uses signing_key_file. With scoped, its template is the permissions of the join role." choice:"empty uses signing_key_file"`
	// Scoped is whether the signing keys are scoped: the template of each
	// sets the permissions, and the JWTs that they sign carry none.
	Scoped bool `mapstructure:"scoped" doc:"For operator mode: the signing keys are scoped, so the template of each key sets the permissions of the users that it signs: signing_key_file the agent role's, join_signing_key_file the join role's. osapi nats auth operator create makes them so." choice:"false: the JWTs carry the permissions of their role"`
	// Account is the public key of the account of SigningKeyFile.
	Account string `mapstructure:"account" doc:"For operator mode: the public key (A...) of osapi's account, which signing_key_file signs for." choice:"set for operator mode"`
	// Agent is the NATS user that the join gives the agent.
	Agent NATSUser `mapstructure:"agent" doc:"The NATS user that the join gives the agent. Empty is the user of nats.server.auth.users with the role agent." choice:"the deployment's own user; empty takes the embedded server's"`
	// Callout is osapi's auth callout service, for a NATS server whose
	// auth_callout block names osapi's issuer key.
	Callout JoinCallout `mapstructure:"callout"`
}

// JoinCallout is osapi's auth callout service: the NATS server sends it
// each connection that is not one of its auth_users, and the controller
// decides it, a join token or an admitted agent's key.
type JoinCallout struct {
	// IssuerKeyFile is the seed of the account key that the server's
	// auth_callout block names as its issuer.
	IssuerKeyFile string `mapstructure:"issuer_key_file" doc:"For a NATS server with auth callout: a file with the seed (SA...) of the account key that the server's auth_callout names as its issuer. The controller decides each connection of a join or an agent, and signs its user with it. Empty turns the callout off." choice:"set for auth callout"`
	// Account is the account that a join or an agent goes in.
	Account string `mapstructure:"account" doc:"For auth callout: the account that a join or an agent goes in, the account of osapi's streams, such as OSAPI." choice:"set for auth callout"`
	// Auth is how the controller connects as the user of the server's
	// auth_users that answers the server. Make it an nkey: a server with no
	// nkey user offers no nonce, and an agent signs the nonce with its key.
	Auth NATSAuth `mapstructure:"auth" doc:"For auth callout: how the controller connects as the user, among the server's auth_users, that answers the server's requests. Use type nkey: a NATS server offers each connection a nonce only when it has an nkey user, and an agent proves its key by signing it."`
}

// NATSUser is a NATS user and its password.
type NATSUser struct {
	Username string `mapstructure:"username" doc:"The username." choice:"the deployment's own user"`
	Password string `mapstructure:"password" doc:"The password." choice:"the deployment's own user" mask:"password" noref:"a NATS password cannot come from a vault, because osapi needs it to connect to NATS before it can read a vault"`
}

// JoinUsers are the bootstrap user and the agent user of the join over
// NATS: controller.join's, or else the users of the embedded server with
// the roles join and agent.
func (c Config) JoinUsers() (NATSUser, NATSUser) {
	join, agent := c.Controller.Join.User, c.Controller.Join.Agent
	for _, u := range c.NATS.Server.Auth.Users {
		switch {
		case u.Role == "join" && join.Username == "":
			join = NATSUser{Username: u.Username, Password: u.Password}
		case u.Role == "agent" && agent.Username == "":
			agent = NATSUser{Username: u.Username, Password: u.Password}
		}
	}

	return join, agent
}

// JoinURL is where an agent reaches NATS in a join: controller.join.url, or
// else nats.server.url.
func (c Config) JoinURL() string {
	if c.Controller.Join.URL != "" {
		return c.Controller.Join.URL
	}

	return c.NATS.Server.URL
}

// ControllerSecrets holds the vaults of the controller.
type ControllerSecrets struct {
	// Vaults are the vaults, each with its name, type and targets.
	Vaults []vault.Spec `mapstructure:"vaults" validate:"dive" doc:"The vaults that a reference such as secret://default/registry-token names, each with name, type, targets, key_file or key_env, and config."`
}

// UIConfig holds settings for the embedded management UI.
type UIConfig struct {
	// Enabled controls whether the embedded UI is served. Defaults to true.
	Enabled bool `mapstructure:"enabled" doc:"Serve the embedded UI."`
	// URL is where a person opens the UI. Health shows it beside the bind
	// address of the API, whose listener serves the UI.
	URL string `mapstructure:"url" doc:"Where a person opens the UI, shown by health beside the bind address." validate:"omitempty,http_url" choice:"where a person reaches it; empty shows the bind address"`
}

// APIServer holds the HTTP server config (port + security).
type APIServer struct {
	// Host the server binds to. Empty binds every interface. The UI uses the
	// same listener.
	Host string `mapstructure:"host" doc:"Host the REST API server, and the UI on its listener, binds to. Empty binds every interface." choice:"empty binds every interface"`
	// URL is where a person reaches the API. Health shows it beside the bind
	// address; osapi never requests it.
	URL string `mapstructure:"url" doc:"Where a person reaches the API, shown by health beside the bind address." validate:"omitempty,http_url" choice:"where a person reaches it; empty shows the bind address"`
	// Port the server will bind to.
	Port int `mapstructure:"port" doc:"Port the REST API server listens on." validate:"min=1,max=65535"`
	// Security contains security-related configuration for the server, such as CORS and tokens.
	Security ServerSecurity `mapstructure:"security"`
	// TLS makes the server serve HTTPS.
	TLS ServerTLS `mapstructure:"tls"`
	// JobTimeout is how long the controller waits for agent responses
	// before returning partial results. Uses Go duration format.
	// Defaults to 30s when empty.
	JobTimeout string `mapstructure:"job_timeout" doc:"How long to wait for agent responses before returning partial results. Go duration format." validate:"omitempty,go_duration"`
}

// Client configuration settings.
type Client struct {
	// URL the client will connect to
	URL string `mapstructure:"url" doc:"Base URL the CLI client connects to."`
	// Security contains security-related configuration for the client, such as access tokens.
	Security ClientSecurity `mapstructure:"security"`
	// TLS is what the client trusts on an https:// URL.
	TLS ClientTLS `mapstructure:"tls"`
}

// ClientTLS is how osapi checks a server that it connects to over TLS: the
// CLI to the controller, the OTLP exporter and the trace reader to the
// tracing backend, and the controller and each agent to NATS. ClientTLS.Config builds the TLS config of each.
type ClientTLS struct {
	// Enabled connects with TLS. A URL with https:// uses TLS without it.
	Enabled bool `mapstructure:"enabled" doc:"Connect with TLS. A URL that starts with https:// uses TLS without it."`
	// CAFile is a PEM certificate that osapi trusts, besides the system's
	// roots, such as a private CA or a self-signed certificate.
	CAFile string `mapstructure:"ca_file" doc:"A PEM certificate to trust besides the system's roots, such as a private CA or the self-signed api.crt in the controller.pki.key_dir of a controller." choice:"a certificate of the operator; empty trusts the system roots"`
	// CertFile and KeyFile are the client certificate, for a server that
	// asks for one (mutual TLS).
	CertFile string `mapstructure:"cert_file" doc:"The PEM client certificate, for a server that asks for one (mutual TLS). Set it with key_file." validate:"required_with=KeyFile" choice:"a certificate of the operator; empty sends none"`
	// KeyFile is the PEM private key of CertFile.
	KeyFile string `mapstructure:"key_file" doc:"The PEM private key of cert_file." validate:"required_with=CertFile" choice:"a certificate of the operator; empty sends none"`
	// ServerName is the name to check the certificate against, when it is
	// not the host that osapi connects to.
	ServerName string `mapstructure:"server_name" doc:"The name to check the server's certificate against, when it is not the host of the address." choice:"empty checks the host of the address"`
	// InsecureSkipVerify accepts any certificate. osapi logs a warning at
	// each start that has it on.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" doc:"Accept any certificate of the server, so anyone in the path can read and change the traffic. For a test only: osapi logs a warning at each start that has it on."`
}

// CustomRole defines a named set of permissions that can be assigned to tokens.
type CustomRole struct {
	// Permissions granted to this role.
	Permissions []types.Permission `mapstructure:"permissions" validate:"dive,permission"`
}

// ServerTLS makes a server of osapi serve TLS: the API, each metrics
// server, and the NATS server. With no certificate set, the API makes a
// self-signed one in controller.pki.key_dir at its first start, and uses it
// after that; a metrics server and the NATS server need one.
type ServerTLS struct {
	// Enabled serves TLS on the server's port, in place of plain text.
	Enabled bool `mapstructure:"enabled" doc:"Serve TLS on the port (HTTPS for the API and metrics), in place of plain text. Without cert_file and key_file, the API makes a self-signed certificate in controller.pki.key_dir; a metrics server and the NATS server need both."`
	// CertFile is the PEM certificate to serve, with its chain.
	CertFile string `mapstructure:"cert_file" doc:"The PEM certificate to serve, with its chain, such as one from tailscale cert or certbot. Set it with key_file, or neither for the self-signed certificate." validate:"required_with=KeyFile" choice:"a certificate of the operator; empty is the API's self-signed certificate"`
	// KeyFile is the PEM private key of CertFile.
	KeyFile string `mapstructure:"key_file" doc:"The PEM private key of cert_file." validate:"required_with=CertFile" choice:"a certificate of the operator; empty is the API's self-signed certificate"`
}

// ServerSecurity represents security-related settings for the server.
type ServerSecurity struct {
	// CORS Cross-Origin Resource Sharing (CORS) settings for the server.
	CORS CORS `mapstructure:"cors"`
	// SigningKeys are the keys that sign and verify tokens. The first key
	// signs a new token. Each key verifies a token whose kid is its ID.
	SigningKeys []SigningKey `mapstructure:"signing_keys" validate:"omitempty,unique=ID,dive" mask:"password" doc:"HS256 keys for JWTs, each an id and a key. Without one, the controller makes a key in controller.pki.key_dir at its first start, which controller token generate on its host also uses. The first key signs; each key verifies the tokens with its id as kid. The environment variable takes id=key pairs with commas between them. Generate a key with openssl rand -hex 32. A key can be a reference, secret://<vault>/<key>, which the controller resolves at startup." choice:"the controller makes one at its first start when there is none"`
	// Roles defines custom roles with fine-grained permissions.
	Roles map[types.Role]CustomRole `mapstructure:"roles" validate:"dive" doc:"Custom roles with fine-grained permissions." choice:"custom roles; the built-in roles need none"`
}

// SigningKey is one key of controller.api.security.signing_keys.
type SigningKey struct {
	// ID names the key in the kid header of each token that it signs.
	ID string `mapstructure:"id" validate:"required"`
	// Key is the HS256 key. Must be at least validation.MinSigningKeyLen
	// characters (see openssl rand -hex 32).
	Key string `mapstructure:"key" validate:"required,signing_key" mask:"password"`
}

// Keys are the signing keys as authtoken uses them, in the order of the
// config: the first signs.
func (s ServerSecurity) Keys() authtoken.Keys {
	keys := make(authtoken.Keys, 0, len(s.SigningKeys))
	for _, k := range s.SigningKeys {
		keys = append(keys, authtoken.Key{ID: k.ID, Secret: k.Key})
	}

	return keys
}

// ClientSecurity represents security-related settings for the client.
type ClientSecurity struct {
	// BearerToken is the JWT used for role-based access control.
	BearerToken string `mapstructure:"bearer_token" validate:"omitempty" mask:"password" doc:"JWT bearer token for client requests. Generate with osapi controller token generate. It can be a reference, secret://<vault>/<key>, which the command resolves from the vaults of its config." choice:"the token of the person or the program; osapi controller token generate makes one"`
}

// CORS represents the CORS (Cross-Origin Resource Sharing) settings.
type CORS struct {
	// List of origins allowed to access the server (e.g., "foo").
	AllowOrigins []string `mapstructure:"allow_origins,omitempty" doc:"Origins allowed to make cross-origin requests. An empty list disables CORS headers entirely." choice:"empty refuses cross-origin requests"`
}

// AgentConsumer configuration for the agent's JetStream consumer settings.
type AgentConsumer struct {
	// Name is the durable consumer name.
	Name string `mapstructure:"name" doc:"Durable consumer name."`
	// MaxDeliver is the maximum number of redelivery attempts before sending to DLQ.
	MaxDeliver int `mapstructure:"max_deliver" doc:"Maximum redelivery attempts before sending to DLQ."`
	// AckWait is the time to wait for an ACK before redelivering.
	AckWait string `mapstructure:"ack_wait" doc:"Time to wait for an ACK before redelivering." validate:"omitempty,go_duration"` // e.g. "30s", "1m"
	// MaxAckPending is the maximum outstanding unacknowledged messages.
	MaxAckPending int `mapstructure:"max_ack_pending" doc:"Maximum outstanding unacknowledged messages."`
	// ReplayPolicy is "instant" or "original".
	ReplayPolicy string `mapstructure:"replay_policy" doc:"Replay policy: 'instant' or 'original'."`
	// BackOff durations between redelivery attempts.
	BackOff []string `mapstructure:"back_off" doc:"Backoff durations between redelivery attempts."` // e.g. ["30s", "2m", "5m"]
	// InactiveThreshold is how long a consumer can be idle before NATS
	// removes it.
	InactiveThreshold string `mapstructure:"inactive_threshold" doc:"How long a job consumer can be idle before NATS removes it (Go duration). Empty keeps it." validate:"omitempty,go_duration"` // e.g. "24h"
}

// AgentFacts configuration for the agent's facts collection settings.
type AgentFacts struct {
	// Interval is how often the agent collects and publishes facts.
	Interval string `mapstructure:"interval" validate:"omitempty,go_duration" doc:"How often the agent collects and publishes facts."` // e.g. "5m", "1h"
}

// AgentConditions holds threshold configuration for node conditions.
type AgentConditions struct {
	MemoryPressureThreshold int     `mapstructure:"memory_pressure_threshold" validate:"min=1,max=100" doc:"Memory pressure threshold (percent used)."`
	HighLoadMultiplier      float64 `mapstructure:"high_load_multiplier"      validate:"gt=0"          doc:"High load multiplier (load1 / cpu_count)."`
	DiskPressureThreshold   int     `mapstructure:"disk_pressure_threshold"   validate:"min=1,max=100" doc:"Disk pressure threshold (percent used)."`
}

// ProcessConditions holds threshold configuration for process-level conditions.
type ProcessConditions struct {
	// MemoryPressureBytes is the RSS threshold in bytes (0 = disabled).
	MemoryPressureBytes int64 `mapstructure:"memory_pressure_bytes" doc:"Process RSS threshold in bytes (0 = disabled)." choice:"0 turns the condition off"`
	// HighCPUPercent is the CPU usage threshold as a percentage (0 = disabled).
	HighCPUPercent float64 `mapstructure:"high_cpu_percent" doc:"Process CPU threshold as percentage (0 = disabled)." choice:"0 turns the condition off"`
}

// PrivilegeEscalation configuration for least-privilege agent mode.
// When enabled, write commands use sudo and Linux capabilities are
// verified at startup.
type PrivilegeEscalation struct {
	// Enabled activates least-privilege mode: sudo for write commands
	// and capability verification at startup.
	Enabled bool `mapstructure:"enabled" doc:"Least-privilege mode: sudo for write commands and capability checks at startup."`
}

// AgentCommand configures the operations that run a command the caller
// supplies: command exec, command shell and docker exec.
type AgentCommand struct {
	// Enabled lets the agent run command exec, command shell and docker exec.
	// When it is false, the agent refuses those jobs. The agent makes the check,
	// so it also refuses a job that gets to it without the API.
	Enabled bool `mapstructure:"enabled" doc:"Run command exec, command shell and docker exec jobs. Off by default: each one runs a command the caller supplies, as the agent's user."`
}

// AgentPKI holds PKI configuration for the agent.
type AgentPKI struct {
	// KeyDir is the directory for agent keypair storage.
	KeyDir string `mapstructure:"key_dir" doc:"Directory of the agent's key pair."`
}

// ControllerPKI holds PKI configuration for the controller.
type ControllerPKI struct {
	// KeyDir is the directory for controller keypair storage.
	KeyDir string `mapstructure:"key_dir" doc:"Directory for controller keypair storage."`
	// AutoAdmit admits each machine that asks to join, with no operator.
	AutoAdmit bool `mapstructure:"auto_admit" doc:"Admit each machine that asks to join, with no operator (dev/test only)."`
	// RotationGracePeriod is how long both old and new keys are accepted
	// during key rotation. Uses Go duration format.
	RotationGracePeriod string `mapstructure:"rotation_grace_period" doc:"Grace period during key rotation where both old and new keys are accepted. Go duration format." validate:"omitempty,go_duration"`
	// DeleteAfter deletes an admitted machine that has not reported for
	// this long. Empty is never. Uses Go duration format.
	DeleteAfter string `mapstructure:"delete_after" doc:"Delete an admitted machine that has not reported for this long, such as 720h. Empty is never. Go duration format." validate:"omitempty,go_duration" choice:"empty keeps a machine that stops reporting"`
}

// AgentConfig configuration settings.
type AgentConfig struct {
	// Join is the join string: nats:// or tls:// for a join over NATS, or
	// osapi://<host>:<port>?token=...&fp=SHA256:... for one over HTTPS,
	// that osapi controller join token create prints. An agent that is not
	// admitted joins with it at its first start.
	Join string `mapstructure:"join" mask:"password" doc:"The join string that osapi controller join token create prints. An agent that is not admitted joins with it at its first start and is admitted at once: over NATS for nats:// and tls://, which also gives the agent its NATS user, or over HTTPS for osapi://. Without it, the agent asks to be admitted over NATS." choice:"the join string of the operator; empty asks over NATS"`
	// NATS connection settings for the agent.
	NATS NATSConnection `mapstructure:"nats"`
	// Consumer settings for the agent's JetStream consumer.
	Consumer AgentConsumer `mapstructure:"consumer,omitempty"`
	// Facts settings for the agent's facts collection.
	Facts AgentFacts `mapstructure:"facts,omitempty"`
	// Hostname identifies this agent instance for routing.
	Hostname string `mapstructure:"hostname" doc:"Agent hostname for direct routing. Defaults to the system hostname when empty." choice:"empty is the host's own hostname"`
	// MaxJobs maximum number of concurrent jobs to process.
	MaxJobs int `mapstructure:"max_jobs" doc:"Maximum number of concurrent jobs to process." validate:"min=1"`
	// Labels are key-value pairs for label-based routing (e.g., role: web, env: prod).
	// Maximum 5 labels per agent — each label creates multiple NATS consumers
	// for hierarchical prefix matching.
	Labels map[string]string `mapstructure:"labels" doc:"Key-value labels for label-based routing (max 5). Values can be hierarchical with dot separators. Each label creates NATS consumers for prefix matching. See Job System Architecture for details." validate:"max=5" choice:"the labels of the host, which the operator picks"`
	// Conditions holds threshold settings for node condition evaluation.
	Conditions AgentConditions `mapstructure:"conditions,omitempty"`
	// ProcessConditions holds threshold settings for process-level condition evaluation.
	ProcessConditions ProcessConditions `mapstructure:"process_conditions,omitempty"`
	// PrivilegeEscalation configures least-privilege agent mode.
	PrivilegeEscalation PrivilegeEscalation `mapstructure:"privilege_escalation,omitempty"`
	// Command enables the operations that run a command the caller supplies.
	Command AgentCommand `mapstructure:"command,omitempty"`
	// PKI holds PKI enrollment and signing settings.
	PKI     AgentPKI      `mapstructure:"pki,omitempty"`
	Metrics MetricsServer `mapstructure:"metrics"`
}
