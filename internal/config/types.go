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
	API       API       `mapstructure:"api"       mask:"struct"`
	Node      Node      `mapstructure:"node"`
	NATS      NATS      `mapstructure:"nats"`
	Telemetry Telemetry `mapstructure:"telemetry"`
	// Debug enable or disable debug option set from CLI.
	Debug bool `mapstructure:"debug"`
}

// Telemetry configuration settings.
type Telemetry struct {
	Tracing TracingConfig `mapstructure:"tracing,omitempty"`
	Metrics MetricsConfig `mapstructure:"metrics,omitempty"`
}

// MetricsConfig configuration settings for Prometheus metrics.
type MetricsConfig struct {
	// Path is the HTTP path for the Prometheus scrape endpoint.
	// Defaults to "/metrics" when empty.
	Path string `mapstructure:"path"`
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
	Server   NATSServer   `mapstructure:"server,omitempty"`
	Stream   NATSStream   `mapstructure:"stream,omitempty"`
	KV       NATSKV       `mapstructure:"kv,omitempty"`
	DLQ      NATSDLQ      `mapstructure:"dlq,omitempty"`
	Audit    NATSAudit    `mapstructure:"audit,omitempty"`
	Registry NATSRegistry `mapstructure:"registry,omitempty"`
}

// NATSAudit configuration for the audit log KV bucket.
type NATSAudit struct {
	// Bucket is the KV bucket name for audit log entries.
	Bucket   string `mapstructure:"bucket"`
	TTL      string `mapstructure:"ttl"` // e.g. "720h" (30 days)
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSRegistry configuration for the agent registry KV bucket.
type NATSRegistry struct {
	// Bucket is the KV bucket name for agent registration entries.
	Bucket   string `mapstructure:"bucket"`
	TTL      string `mapstructure:"ttl"`     // e.g. "30s"
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
	Auth NATSServerAuth `mapstructure:"auth,omitempty"`
}

// NATSStream configuration for JetStream stream settings.
type NATSStream struct {
	// Name is the JetStream stream name.
	Name string `mapstructure:"name"`
	// Subjects is the subject filter for the stream.
	Subjects string `mapstructure:"subjects"`
	MaxAge   string `mapstructure:"max_age"` // e.g. "24h", "1h30m"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
	Discard  string `mapstructure:"discard"` // "old" or "new"
}

// NATSKV configuration for KeyValue bucket settings.
type NATSKV struct {
	// Bucket is the KV bucket name for job definitions and status events.
	Bucket string `mapstructure:"bucket"`
	// ResponseBucket is the KV bucket name for agent result storage.
	ResponseBucket string `mapstructure:"response_bucket"`
	TTL            string `mapstructure:"ttl"` // e.g. "1h", "30m"
	MaxBytes       int64  `mapstructure:"max_bytes"`
	Storage        string `mapstructure:"storage"` // "file" or "memory"
	Replicas       int    `mapstructure:"replicas"`
}

// NATSDLQ configuration for Dead Letter Queue stream settings.
type NATSDLQ struct {
	MaxAge   string `mapstructure:"max_age"` // e.g. "7d", "24h"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// NATSConnection is a reusable NATS connection configuration block.
type NATSConnection struct {
	// Host the NATS server hostname.
	Host string `mapstructure:"host"`
	// Port the NATS server port.
	Port int `mapstructure:"port"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name"`
	// Namespace is a prefix for all NATS subjects used by this client.
	Namespace string `mapstructure:"namespace"`
	// Auth holds client-side authentication configuration.
	Auth NATSAuth `mapstructure:"auth,omitempty"`
}

// API configuration settings.
type API struct {
	Client
	Server `mask:"struct"`
}

// Client configuration settings.
type Client struct {
	// URL the client will connect to
	URL string `mapstructure:"url"`
	// Security contains security-related configuration for the client, such as access tokens.
	Security ClientSecurity `mapstructure:"security" mask:"struct"`
}

// Server configuration settings.
type Server struct {
	// Port the server will bind to.
	Port int `mapstructure:"port"`
	// NATS connection settings for the API server.
	NATS NATSConnection `mapstructure:"nats"`
	// Security contains security-related configuration for the server, such as CORS and tokens.
	Security ServerSecurity `mapstructure:"security" mask:"struct"`
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

// Node configuration settings.
type Node struct {
	Agent NodeAgent `mapstructure:"agent,omitempty"`
}

// NodeAgentConsumer configuration for the agent's JetStream consumer settings.
type NodeAgentConsumer struct {
	// Name is the durable consumer name.
	Name string `mapstructure:"name"`
	// MaxDeliver is the maximum number of redelivery attempts before sending to DLQ.
	MaxDeliver int `mapstructure:"max_deliver"`
	// AckWait is the time to wait for an ACK before redelivering.
	AckWait string `mapstructure:"ack_wait"` // e.g. "30s", "1m"
	// MaxAckPending is the maximum outstanding unacknowledged messages.
	MaxAckPending int `mapstructure:"max_ack_pending"`
	// ReplayPolicy is "instant" or "original".
	ReplayPolicy string `mapstructure:"replay_policy"`
	// BackOff durations between redelivery attempts.
	BackOff []string `mapstructure:"back_off"` // e.g. ["30s", "2m", "5m"]
}

// NodeAgent configuration settings.
type NodeAgent struct {
	// NATS connection settings for the agent.
	NATS NATSConnection `mapstructure:"nats"`
	// Consumer settings for the agent's JetStream consumer.
	Consumer NodeAgentConsumer `mapstructure:"consumer,omitempty"`
	// QueueGroup for load balancing multiple agents.
	QueueGroup string `mapstructure:"queue_group"`
	// Hostname identifies this agent instance for routing.
	Hostname string `mapstructure:"hostname"`
	// MaxJobs maximum number of concurrent jobs to process.
	MaxJobs int `mapstructure:"max_jobs"`
	// Labels are key-value pairs for label-based routing (e.g., role: web, env: prod).
	Labels map[string]string `mapstructure:"labels"`
}
