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

// Config represents the root structure of the YAML configuration file.
// This struct is used to unmarshal configuration data from Viper.
type Config struct {
	Controller Controller  `mapstructure:"controller"      mask:"struct"`
	Agent      AgentConfig `mapstructure:"agent,omitempty"`
	NATS       NATS        `mapstructure:"nats"`
	Telemetry  Telemetry   `mapstructure:"telemetry"`
	// Debug enable or disable debug option set from CLI.
	Debug bool `mapstructure:"debug"                         doc:"Enable verbose logging."`
}

// NotificationsConfig holds settings for the pluggable condition notification
// system. When Enabled is true, a Watcher monitors the registry KV bucket and
// dispatches ConditionEvents via the configured Notifier.
type NotificationsConfig struct {
	// Enabled activates the condition watcher and notifier.
	Enabled bool `mapstructure:"enabled"           doc:"Enable the condition watcher and notifier (default: true)."`
	// Notifier selects the notification backend: "log" (default).
	Notifier string `mapstructure:"notifier"          doc:"Notifier backend: 'log' (default)."`
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
	Enabled bool `mapstructure:"enabled" doc:"Enable the metrics endpoint (default: true)."`
	// Host the metrics server binds to.
	Host string `mapstructure:"host"    doc:"Host the metrics server binds to."`
	// Port the metrics server listens on.
	Port int `mapstructure:"port"    doc:"Port the metrics server listens on."          validate:"omitempty,min=1,max=65535"`
}

// TracingConfig configuration settings for distributed tracing.
type TracingConfig struct {
	// Enabled enables or disables tracing.
	Enabled bool `mapstructure:"enabled"       doc:"Enable distributed tracing (default: false)."`
	// Exporter selects the trace exporter: "stdout" or "otlp".
	Exporter string `mapstructure:"exporter"      doc:"Trace exporter: 'stdout' or 'otlp'."`
	// OTLPEndpoint is the gRPC endpoint for the OTLP exporter (e.g., "localhost:4317").
	OTLPEndpoint string `mapstructure:"otlp_endpoint" doc:"gRPC endpoint for the OTLP exporter, such as 'localhost:4317'."`
}

// NATSAuth holds client-side authentication settings for connecting to NATS.
type NATSAuth struct {
	// Type is the auth method: "none", "user_pass", or "nkey".
	Type string `mapstructure:"type"      doc:"Authentication type: 'none', 'user_pass', or 'nkey'."`
	// Username for user_pass auth.
	Username string `mapstructure:"username"  doc:"Username for user_pass auth."`
	// Password for user_pass auth.
	Password string `mapstructure:"password"  doc:"Password for user_pass auth."                         mask:"password"`
	// NKeyFile path to the NKey seed file for nkey auth.
	NKeyFile string `mapstructure:"nkey_file" doc:"Path to the NKey seed file for nkey auth."`
}

// NATSServerAuth holds server-side authentication settings for the embedded NATS server.
type NATSServerAuth struct {
	// Type is the auth method: "none", "user_pass", or "nkey".
	Type string `mapstructure:"type"  doc:"Auth method: 'none', 'user_pass' or 'nkey'."`
	// Users allowed to connect (for user_pass auth).
	Users []NATSServerUser `mapstructure:"users" doc:"Users allowed to connect (for user_pass auth)."`
	// NKeys is a list of allowed public NKeys (for nkey auth).
	NKeys []string `mapstructure:"nkeys" doc:"Allowed public NKeys, for nkey auth."`
}

// NATSServerUser represents an allowed username/password pair for the NATS server.
type NATSServerUser struct {
	// Username for the user.
	Username string `mapstructure:"username"`
	// Password for the user.
	Password string `mapstructure:"password" mask:"password"`
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
}

// NATSEnrollment configuration for the PKI enrollment KV bucket.
// No TTL — pending enrollment requests persist until accepted or rejected.
type NATSEnrollment struct {
	// Bucket is the KV bucket name for pending enrollment entries.
	Bucket   string `mapstructure:"bucket"   doc:"KV bucket for pending enrollment entries."`
	Storage  string `mapstructure:"storage"  doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas" doc:"Number of replicas for the enrollment KV bucket."`
}

// NATSAudit configuration for the audit log stream.
type NATSAudit struct {
	// Stream is the JetStream stream name for audit log entries.
	Stream string `mapstructure:"stream"    doc:"JetStream stream name for audit log entries."`
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
	Bucket   string `mapstructure:"bucket"   validate:"required"              doc:"KV bucket for agent heartbeat registration."`
	TTL      string `mapstructure:"ttl"      validate:"omitempty,go_duration" doc:"TTL for registry entries (Go duration). Agents refresh every 10s; the TTL acts as a liveness timeout."` // e.g. "30s"
	Storage  string `mapstructure:"storage"                                   doc:"Storage backend: 'file' or 'memory'."`                                                                  // "file" or "memory"
	Replicas int    `mapstructure:"replicas"                                  doc:"Number of KV replicas."`
}

// NATSFacts configuration for the agent facts KV bucket.
type NATSFacts struct {
	// Bucket is the KV bucket name for agent facts entries.
	Bucket   string `mapstructure:"bucket"   doc:"KV bucket for agent facts entries."`
	TTL      string `mapstructure:"ttl"      doc:"TTL for facts entries (Go duration)." validate:"omitempty,go_duration"` // e.g. "1h"
	Storage  string `mapstructure:"storage"  doc:"Storage backend: 'file' or 'memory'."`                                  // "file" or "memory"
	Replicas int    `mapstructure:"replicas" doc:"Number of KV replicas."`
}

// NATSState configuration for the agent state KV bucket (drain flags, timeline events).
type NATSState struct {
	// Bucket is the KV bucket name for persistent agent state.
	Bucket   string `mapstructure:"bucket"   doc:"KV bucket for persistent agent state (drain flags, timeline events). No TTL — operator actions persist indefinitely."`
	Storage  string `mapstructure:"storage"  doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas" doc:"Number of KV replicas."`
}

// NATSObjects configuration for the NATS Object Store bucket.
type NATSObjects struct {
	// Bucket is the Object Store bucket name for file content.
	Bucket   string `mapstructure:"bucket"    doc:"Object Store bucket for uploaded file content."`
	MaxBytes int64  `mapstructure:"max_bytes" doc:"Maximum total size of the bucket in bytes."`
	Storage  string `mapstructure:"storage"   doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"  doc:"Number of Object Store replicas."`
}

// NATSFileState configuration for the file deployment state KV bucket.
// No TTL — deployed file state persists until explicitly removed.
type NATSFileState struct {
	// Bucket is the KV bucket name for file deployment SHA tracking.
	Bucket   string `mapstructure:"bucket"   doc:"KV bucket for file deploy state tracking. No TTL — state persists until explicitly removed."`
	Storage  string `mapstructure:"storage"  doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas int    `mapstructure:"replicas" doc:"Number of KV replicas."`
}

// NATSServer configuration settings for the embedded NATS server.
type NATSServer struct {
	// Host the server will bind to.
	Host string `mapstructure:"host"           doc:"Host the embedded NATS server binds to."`
	// Port the server will bind to.
	Port int `mapstructure:"port"           doc:"Port the embedded NATS server listens on."`
	// StoreDir the directory for JetStream file storage.
	StoreDir string `mapstructure:"store_dir"      doc:"Directory for JetStream file storage."`
	// Namespace is a prefix for all NATS subjects and infrastructure names.
	Namespace string `mapstructure:"namespace"      doc:"Prefix for all NATS subjects and infrastructure names."`
	// Auth holds server-side authentication configuration.
	Auth    NATSServerAuth `mapstructure:"auth,omitempty"`
	Metrics MetricsServer  `mapstructure:"metrics"`
}

// NATSStream configuration for JetStream stream settings.
type NATSStream struct {
	// Name is the JetStream stream name.
	Name string `mapstructure:"name"     validate:"required"              doc:"JetStream stream name for job notifications."`
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
	// Bucket is the KV bucket name for job definitions and status events.
	Bucket string `mapstructure:"bucket"          validate:"required" doc:"KV bucket for immutable job definitions and status events."`
	// ResponseBucket is the KV bucket name for agent result storage.
	ResponseBucket string `mapstructure:"response_bucket" validate:"required" doc:"KV bucket for agent result storage."`
	TTL            string `mapstructure:"ttl"                                 doc:"TTL for KV entries (Go duration)."` // e.g. "1h", "30m"
	MaxBytes       int64  `mapstructure:"max_bytes"                           doc:"Maximum total size of the bucket in bytes."`
	Storage        string `mapstructure:"storage"                             doc:"Storage backend: 'file' or 'memory'."` // "file" or "memory"
	Replicas       int    `mapstructure:"replicas"                            doc:"Number of KV replicas."`
}

// KVBucketInfo holds a KV bucket's human-readable name and its configured
// bucket name. It is returned by NATS.AllKVBuckets so callers can iterate
// all KV buckets without manually listing every sub-config field.
type KVBucketInfo struct {
	// Name is a human-readable label for the bucket (e.g. "job-queue").
	Name string
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
	Host string `mapstructure:"host"           doc:"NATS server hostname to connect to."`
	// Port the NATS server port.
	Port int `mapstructure:"port"           doc:"NATS server port to connect to."                             validate:"min=1,max=65535"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name"    doc:"Client name sent to NATS for identification."`
	// Namespace is a prefix for all NATS subjects used by this client.
	Namespace string `mapstructure:"namespace"      doc:"Subject namespace prefix. Must match nats.server.namespace."`
	// Auth holds client-side authentication configuration.
	Auth NATSAuth `mapstructure:"auth,omitempty"`
}

// Controller holds the control plane configuration.
type Controller struct {
	Client        Client              `mapstructure:"client"`
	API           APIServer           `mapstructure:"api"                     mask:"struct"`
	NATS          NATSConnection      `mapstructure:"nats"`
	Metrics       MetricsServer       `mapstructure:"metrics"`
	Notifications NotificationsConfig `mapstructure:"notifications,omitempty"`
	// UI holds settings for the embedded management UI.
	UI UIConfig `mapstructure:"ui,omitempty"`
	// PKI holds PKI enrollment and signing settings.
	PKI ControllerPKI `mapstructure:"pki,omitempty"`
}

// UIConfig holds settings for the embedded management UI.
type UIConfig struct {
	// Enabled controls whether the embedded UI is served. Defaults to true.
	Enabled *bool `mapstructure:"enabled" doc:"Serve the embedded UI (default: true)."`
}

// APIServer holds the HTTP server config (port + security).
type APIServer struct {
	// Port the server will bind to.
	Port int `mapstructure:"port"        validate:"min=1,max=65535"       doc:"Port the REST API server listens on."`
	// Security contains security-related configuration for the server, such as CORS and tokens.
	Security ServerSecurity `mapstructure:"security"                                                                                                                                                     mask:"struct"`
	// JobTimeout is how long the controller waits for agent responses
	// before returning partial results. Uses Go duration format.
	// Defaults to 30s when empty.
	JobTimeout string `mapstructure:"job_timeout" validate:"omitempty,go_duration" doc:"How long to wait for agent responses before returning partial results. Go duration format (default: 30s)."`
}

// Client configuration settings.
type Client struct {
	// URL the client will connect to
	URL string `mapstructure:"url"      doc:"Base URL the CLI client connects to."`
	// Security contains security-related configuration for the client, such as access tokens.
	Security ClientSecurity `mapstructure:"security"                                            mask:"struct"`
}

// CustomRole defines a named set of permissions that can be assigned to tokens.
type CustomRole struct {
	// Permissions granted to this role.
	Permissions []string `mapstructure:"permissions"`
}

// ServerSecurity represents security-related settings for the server.
type ServerSecurity struct {
	// CORS Cross-Origin Resource Sharing (CORS) settings for the server.
	CORS CORS `mapstructure:"cors"`
	// SigningKey is the key used for signing or validating tokens. Must be at
	// least validation.MinSigningKeyLen characters (see openssl rand -hex 32).
	SigningKey string `mapstructure:"signing_key" validate:"required,signing_key" mask:"password" doc:"HS256 signing key for JWT validation (required). Generate with openssl rand -hex 32."`
	// Roles defines custom roles with fine-grained permissions.
	Roles map[string]CustomRole `mapstructure:"roles"                                                       doc:"Custom roles with fine-grained permissions."`
}

// ClientSecurity represents security-related settings for the client.
type ClientSecurity struct {
	// BearerToken is the JWT used for role-based access control.
	BearerToken string `mapstructure:"bearer_token" validate:"required" mask:"password" doc:"JWT bearer token for client requests (required). Generate with osapi token generate."`
}

// CORS represents the CORS (Cross-Origin Resource Sharing) settings.
type CORS struct {
	// List of origins allowed to access the server (e.g., "foo").
	AllowOrigins []string `mapstructure:"allow_origins,omitempty" doc:"Origins allowed to make cross-origin requests. An empty list disables CORS headers entirely."`
}

// AgentConsumer configuration for the agent's JetStream consumer settings.
type AgentConsumer struct {
	// Name is the durable consumer name.
	Name string `mapstructure:"name"            doc:"Durable consumer name."`
	// MaxDeliver is the maximum number of redelivery attempts before sending to DLQ.
	MaxDeliver int `mapstructure:"max_deliver"     doc:"Maximum redelivery attempts before sending to DLQ."`
	// AckWait is the time to wait for an ACK before redelivering.
	AckWait string `mapstructure:"ack_wait"        doc:"Time to wait for an ACK before redelivering."       validate:"omitempty,go_duration"` // e.g. "30s", "1m"
	// MaxAckPending is the maximum outstanding unacknowledged messages.
	MaxAckPending int `mapstructure:"max_ack_pending" doc:"Maximum outstanding unacknowledged messages."`
	// ReplayPolicy is "instant" or "original".
	ReplayPolicy string `mapstructure:"replay_policy"   doc:"Replay policy: 'instant' or 'original'."`
	// BackOff durations between redelivery attempts.
	BackOff []string `mapstructure:"back_off"        doc:"Backoff durations between redelivery attempts."` // e.g. ["30s", "2m", "5m"]
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
	MemoryPressureBytes int64 `mapstructure:"memory_pressure_bytes" doc:"Process RSS threshold in bytes (0 = disabled)."`
	// HighCPUPercent is the CPU usage threshold as a percentage (0 = disabled).
	HighCPUPercent float64 `mapstructure:"high_cpu_percent"      doc:"Process CPU threshold as percentage (0 = disabled)."`
}

// PrivilegeEscalation configuration for least-privilege agent mode.
// When enabled, write commands use sudo and Linux capabilities are
// verified at startup.
type PrivilegeEscalation struct {
	// Enabled activates least-privilege mode: sudo for write commands
	// and capability verification at startup.
	Enabled bool `mapstructure:"enabled" doc:"Least-privilege mode: sudo for write commands and capability checks at startup."`
}

// AgentPKI holds PKI configuration for the agent.
type AgentPKI struct {
	// Enabled activates PKI enrollment and job signature verification.
	Enabled bool `mapstructure:"enabled" doc:"Enable PKI enrollment and job signature verification."`
	// KeyDir is the directory for agent keypair storage.
	KeyDir string `mapstructure:"key_dir" doc:"Directory for agent keypair storage."`
}

// ControllerPKI holds PKI configuration for the controller.
type ControllerPKI struct {
	// Enabled activates PKI enrollment and job signing.
	Enabled bool `mapstructure:"enabled"               doc:"Enable PKI enrollment and job signing."`
	// KeyDir is the directory for controller keypair storage.
	KeyDir string `mapstructure:"key_dir"               doc:"Directory for controller keypair storage."`
	// AutoAccept automatically accepts all agent enrollment requests.
	AutoAccept bool `mapstructure:"auto_accept"           doc:"Automatically accept all enrollment requests (dev/test only)."`
	// RotationGracePeriod is how long both old and new keys are accepted
	// during key rotation. Uses Go duration format.
	RotationGracePeriod string `mapstructure:"rotation_grace_period" doc:"Grace period during key rotation where both old and new keys are accepted. Go duration format." validate:"omitempty,go_duration"`
}

// AgentConfig configuration settings.
type AgentConfig struct {
	// NATS connection settings for the agent.
	NATS NATSConnection `mapstructure:"nats"`
	// Consumer settings for the agent's JetStream consumer.
	Consumer AgentConsumer `mapstructure:"consumer,omitempty"`
	// Facts settings for the agent's facts collection.
	Facts AgentFacts `mapstructure:"facts,omitempty"`
	// QueueGroup for load balancing multiple agents.
	QueueGroup string `mapstructure:"queue_group"                    doc:"Queue group for load-balanced (_any) subscriptions."`
	// Hostname identifies this agent instance for routing.
	Hostname string `mapstructure:"hostname"                       doc:"Agent hostname for direct routing. Defaults to the system hostname when empty."`
	// MaxJobs maximum number of concurrent jobs to process.
	MaxJobs int `mapstructure:"max_jobs"                       doc:"Maximum number of concurrent jobs to process."                                                                                                                                                     validate:"min=1"`
	// Labels are key-value pairs for label-based routing (e.g., role: web, env: prod).
	// Maximum 5 labels per agent — each label creates multiple NATS consumers
	// for hierarchical prefix matching.
	Labels map[string]string `mapstructure:"labels"                         doc:"Key-value labels for label-based routing (max 5). Values can be hierarchical with dot separators. Each label creates NATS consumers for prefix matching. See Job System Architecture for details." validate:"max=5"`
	// Conditions holds threshold settings for node condition evaluation.
	Conditions AgentConditions `mapstructure:"conditions,omitempty"`
	// ProcessConditions holds threshold settings for process-level condition evaluation.
	ProcessConditions ProcessConditions `mapstructure:"process_conditions,omitempty"`
	// PrivilegeEscalation configures least-privilege agent mode.
	PrivilegeEscalation PrivilegeEscalation `mapstructure:"privilege_escalation,omitempty"`
	// PKI holds PKI enrollment and signing settings.
	PKI     AgentPKI      `mapstructure:"pki,omitempty"`
	Metrics MetricsServer `mapstructure:"metrics"`
}
