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
	Debug bool `mapstructure:"debug"`
}

// NotificationsConfig holds settings for the pluggable condition notification
// system. When Enabled is true, a Watcher monitors the registry KV bucket and
// dispatches ConditionEvents via the configured Notifier.
type NotificationsConfig struct {
	// Enabled activates the condition watcher and notifier.
	Enabled bool `mapstructure:"enabled"`
	// Notifier selects the notification backend: "log" (default).
	Notifier string `mapstructure:"notifier"`
	// RenotifyInterval is how often to re-fire active conditions.
	// Uses Go duration format (e.g., "1m", "5m", "1h"). Zero disables.
	RenotifyInterval string `mapstructure:"renotify_interval" validate:"omitempty,go_duration"`
}

// Telemetry configuration settings.
type Telemetry struct {
	Tracing TracingConfig `mapstructure:"tracing,omitempty"`
}

// MetricsServer configures the per-component metrics HTTP server.
type MetricsServer struct {
	// Enabled activates the metrics server.
	Enabled bool `mapstructure:"enabled"`
	// Host the metrics server binds to.
	Host string `mapstructure:"host"`
	// Port the metrics server listens on.
	Port int `mapstructure:"port"    validate:"omitempty,min=1,max=65535"`
}

// TracingConfig configuration settings for distributed tracing.
type TracingConfig struct {
	// Enabled enables or disables tracing.
	Enabled bool `mapstructure:"enabled"`
	// Exporter selects the trace exporter: "stdout" or "otlp".
	Exporter string `mapstructure:"exporter"`
	// OTLPEndpoint is the gRPC endpoint for the OTLP exporter (e.g., "localhost:4317").
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
}

// NATSAuth holds client-side authentication settings for connecting to NATS.
type NATSAuth struct {
	// Type is the auth method: "none", "user_pass", or "nkey".
	Type string `mapstructure:"type"`
	// Username for user_pass auth.
	Username string `mapstructure:"username"`
	// Password for user_pass auth.
	Password string `mapstructure:"password"  mask:"password"`
	// NKeyFile path to the NKey seed file for nkey auth.
	NKeyFile string `mapstructure:"nkey_file"`
}

// NATSServerAuth holds server-side authentication settings for the embedded NATS server.
type NATSServerAuth struct {
	// Type is the auth method: "none", "user_pass", or "nkey".
	Type string `mapstructure:"type"`
	// Users allowed to connect (for user_pass auth).
	Users []NATSServerUser `mapstructure:"users"`
	// NKeys is a list of allowed public NKeys (for nkey auth).
	NKeys []string `mapstructure:"nkeys"`
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
	Server    NATSServer    `mapstructure:"server,omitempty"`
	Stream    NATSStream    `mapstructure:"stream,omitempty"`
	KV        NATSKV        `mapstructure:"kv,omitempty"`
	DLQ       NATSDLQ       `mapstructure:"dlq,omitempty"`
	Audit     NATSAudit     `mapstructure:"audit,omitempty"`
	Registry  NATSRegistry  `mapstructure:"registry,omitempty"`
	Facts     NATSFacts     `mapstructure:"facts,omitempty"`
	State     NATSState     `mapstructure:"state,omitempty"`
	Objects   NATSObjects   `mapstructure:"objects,omitempty"`
	FileState NATSFileState `mapstructure:"file_state,omitempty"`
}

// NATSAudit configuration for the audit log stream.
type NATSAudit struct {
	// Stream is the JetStream stream name for audit log entries.
	Stream string `mapstructure:"stream"`
	// Subject is the base subject prefix for audit messages.
	Subject  string `mapstructure:"subject"`
	MaxAge   string `mapstructure:"max_age"   validate:"omitempty,go_duration"` // e.g. "720h" (30 days)
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSRegistry configuration for the agent registry KV bucket.
type NATSRegistry struct {
	// Bucket is the KV bucket name for agent registration entries.
	Bucket   string `mapstructure:"bucket"   validate:"required"`
	TTL      string `mapstructure:"ttl"      validate:"omitempty,go_duration"` // e.g. "30s"
	Storage  string `mapstructure:"storage"`                                   // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSFacts configuration for the agent facts KV bucket.
type NATSFacts struct {
	// Bucket is the KV bucket name for agent facts entries.
	Bucket   string `mapstructure:"bucket"`
	TTL      string `mapstructure:"ttl"      validate:"omitempty,go_duration"` // e.g. "1h"
	Storage  string `mapstructure:"storage"`                                   // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSState configuration for the agent state KV bucket (drain flags, timeline events).
type NATSState struct {
	// Bucket is the KV bucket name for persistent agent state.
	Bucket   string `mapstructure:"bucket"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSObjects configuration for the NATS Object Store bucket.
type NATSObjects struct {
	// Bucket is the Object Store bucket name for file content.
	Bucket   string `mapstructure:"bucket"`
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSFileState configuration for the file deployment state KV bucket.
// No TTL — deployed file state persists until explicitly removed.
type NATSFileState struct {
	// Bucket is the KV bucket name for file deployment SHA tracking.
	Bucket   string `mapstructure:"bucket"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSServer configuration settings for the embedded NATS server.
type NATSServer struct {
	// Host the server will bind to.
	Host string `mapstructure:"host"`
	// Port the server will bind to.
	Port int `mapstructure:"port"`
	// StoreDir the directory for JetStream file storage.
	StoreDir string `mapstructure:"store_dir"`
	// Namespace is a prefix for all NATS subjects and infrastructure names.
	Namespace string `mapstructure:"namespace"`
	// Auth holds server-side authentication configuration.
	Auth    NATSServerAuth `mapstructure:"auth,omitempty"`
	Metrics MetricsServer  `mapstructure:"metrics"`
}

// NATSStream configuration for JetStream stream settings.
type NATSStream struct {
	// Name is the JetStream stream name.
	Name string `mapstructure:"name"     validate:"required"`
	// Subjects is the subject filter for the stream.
	Subjects string `mapstructure:"subjects" validate:"required"`
	MaxAge   string `mapstructure:"max_age"  validate:"omitempty,go_duration"` // e.g. "24h", "1h30m"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
	Discard  string `mapstructure:"discard"` // "old" or "new"
}

// NATSKV configuration for KeyValue bucket settings.
type NATSKV struct {
	// Bucket is the KV bucket name for job definitions and status events.
	Bucket string `mapstructure:"bucket"          validate:"required"`
	// ResponseBucket is the KV bucket name for agent result storage.
	ResponseBucket string `mapstructure:"response_bucket" validate:"required"`
	TTL            string `mapstructure:"ttl"` // e.g. "1h", "30m"
	MaxBytes       int64  `mapstructure:"max_bytes"`
	Storage        string `mapstructure:"storage"` // "file" or "memory"
	Replicas       int    `mapstructure:"replicas"`
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
	MaxAge   string `mapstructure:"max_age"  validate:"omitempty,go_duration"` // e.g. "7d", "24h"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSConnection is a reusable NATS connection configuration block.
type NATSConnection struct {
	// Host the NATS server hostname.
	Host string `mapstructure:"host"`
	// Port the NATS server port.
	Port int `mapstructure:"port"           validate:"min=1,max=65535"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name"`
	// Namespace is a prefix for all NATS subjects used by this client.
	Namespace string `mapstructure:"namespace"`
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
	Enabled *bool `mapstructure:"enabled"`
}

// APIServer holds the HTTP server config (port + security).
type APIServer struct {
	// Port the server will bind to.
	Port int `mapstructure:"port"        validate:"min=1,max=65535"`
	// Security contains security-related configuration for the server, such as CORS and tokens.
	Security ServerSecurity `mapstructure:"security"                                     mask:"struct"`
	// JobTimeout is how long the controller waits for agent responses
	// before returning partial results. Uses Go duration format.
	// Defaults to 30s when empty.
	JobTimeout string `mapstructure:"job_timeout" validate:"omitempty,go_duration"`
}

// Client configuration settings.
type Client struct {
	// URL the client will connect to
	URL string `mapstructure:"url"`
	// Security contains security-related configuration for the client, such as access tokens.
	Security ClientSecurity `mapstructure:"security" mask:"struct"`
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
	// SigningKey is the key used for signing or validating tokens.
	SigningKey string `mapstructure:"signing_key" validate:"required" mask:"password"`
	// Roles defines custom roles with fine-grained permissions.
	Roles map[string]CustomRole `mapstructure:"roles"`
}

// ClientSecurity represents security-related settings for the client.
type ClientSecurity struct {
	// BearerToken is the JWT used for role-based access control.
	BearerToken string `mapstructure:"bearer_token" validate:"required"`
}

// CORS represents the CORS (Cross-Origin Resource Sharing) settings.
type CORS struct {
	// List of origins allowed to access the server (e.g., "foo").
	AllowOrigins []string `mapstructure:"allow_origins,omitempty"`
}

// AgentConsumer configuration for the agent's JetStream consumer settings.
type AgentConsumer struct {
	// Name is the durable consumer name.
	Name string `mapstructure:"name"`
	// MaxDeliver is the maximum number of redelivery attempts before sending to DLQ.
	MaxDeliver int `mapstructure:"max_deliver"`
	// AckWait is the time to wait for an ACK before redelivering.
	AckWait string `mapstructure:"ack_wait"        validate:"omitempty,go_duration"` // e.g. "30s", "1m"
	// MaxAckPending is the maximum outstanding unacknowledged messages.
	MaxAckPending int `mapstructure:"max_ack_pending"`
	// ReplayPolicy is "instant" or "original".
	ReplayPolicy string `mapstructure:"replay_policy"`
	// BackOff durations between redelivery attempts.
	BackOff []string `mapstructure:"back_off"` // e.g. ["30s", "2m", "5m"]
}

// AgentFacts configuration for the agent's facts collection settings.
type AgentFacts struct {
	// Interval is how often the agent collects and publishes facts.
	Interval string `mapstructure:"interval" validate:"omitempty,go_duration"` // e.g. "5m", "1h"
}

// AgentConditions holds threshold configuration for node conditions.
type AgentConditions struct {
	MemoryPressureThreshold int     `mapstructure:"memory_pressure_threshold" validate:"min=1,max=100"`
	HighLoadMultiplier      float64 `mapstructure:"high_load_multiplier"      validate:"gt=0"`
	DiskPressureThreshold   int     `mapstructure:"disk_pressure_threshold"   validate:"min=1,max=100"`
}

// ProcessConditions holds threshold configuration for process-level conditions.
type ProcessConditions struct {
	// MemoryPressureBytes is the RSS threshold in bytes (0 = disabled).
	MemoryPressureBytes int64 `mapstructure:"memory_pressure_bytes"`
	// HighCPUPercent is the CPU usage threshold as a percentage (0 = disabled).
	HighCPUPercent float64 `mapstructure:"high_cpu_percent"`
}

// PrivilegeEscalation configuration for least-privilege agent mode.
// When enabled, write commands use sudo and Linux capabilities are
// verified at startup.
type PrivilegeEscalation struct {
	// Enabled activates least-privilege mode: sudo for write commands
	// and capability verification at startup.
	Enabled bool `mapstructure:"enabled"`
}

// AgentPKI holds PKI configuration for the agent.
type AgentPKI struct {
	// Enabled activates PKI enrollment and job signature verification.
	Enabled bool `mapstructure:"enabled"`
	// KeyDir is the directory for agent keypair storage.
	KeyDir string `mapstructure:"key_dir"`
}

// ControllerPKI holds PKI configuration for the controller.
type ControllerPKI struct {
	// Enabled activates PKI enrollment and job signing.
	Enabled bool `mapstructure:"enabled"`
	// KeyDir is the directory for controller keypair storage.
	KeyDir string `mapstructure:"key_dir"`
	// AutoAccept automatically accepts all agent enrollment requests.
	AutoAccept bool `mapstructure:"auto_accept"`
	// RotationGracePeriod is how long both old and new keys are accepted
	// during key rotation. Uses Go duration format.
	RotationGracePeriod string `mapstructure:"rotation_grace_period" validate:"omitempty,go_duration"`
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
	QueueGroup string `mapstructure:"queue_group"`
	// Hostname identifies this agent instance for routing.
	Hostname string `mapstructure:"hostname"`
	// MaxJobs maximum number of concurrent jobs to process.
	MaxJobs int `mapstructure:"max_jobs"                       validate:"min=1"`
	// Labels are key-value pairs for label-based routing (e.g., role: web, env: prod).
	// Maximum 5 labels per agent — each label creates multiple NATS consumers
	// for hierarchical prefix matching.
	Labels map[string]string `mapstructure:"labels"                         validate:"max=5"`
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
