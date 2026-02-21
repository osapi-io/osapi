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
	Job       Job       `mapstructure:"job"`
	NATS      NATS      `mapstructure:"nats"`
	Telemetry Telemetry `mapstructure:"telemetry"`
	// Debug enable or disable debug option set from CLI.
	Debug bool `mapstructure:"debug"`
}

// Telemetry configuration settings.
type Telemetry struct {
	Tracing TracingConfig `mapstructure:"tracing,omitempty"`
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

// NATS configuration settings.
type NATS struct {
	Server NATSServer `mapstructure:"server,omitempty"`
}

// NATSServer configuration settings for the embedded NATS server.
type NATSServer struct {
	// Host the server will bind to.
	Host string `mapstructure:"host"`
	// Port the server will bind to.
	Port int `mapstructure:"port"`
	// StoreDir the directory for JetStream file storage.
	StoreDir string `mapstructure:"store_dir"`
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
	// Security contains security-related configuration for the server, such as CORS and tokens.
	Security ServerSecurity `mapstructure:"security" mask:"struct"`
}

// ServerSecurity represents security-related settings for the server.
type ServerSecurity struct {
	// CORS Cross-Origin Resource Sharing (CORS) settings for the server.
	CORS CORS `mapstructure:"cors"`
	// SigningKey is the key used for signing or validating tokens.
	SigningKey string `mapstructure:"signing_key" validate:"required" mask:"password"`
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

// Job configuration settings.
type Job struct {
	// Shared infrastructure settings
	StreamName       string `mapstructure:"stream_name"`
	StreamSubjects   string `mapstructure:"stream_subjects"`
	KVBucket         string `mapstructure:"kv_bucket"`
	KVResponseBucket string `mapstructure:"kv_response_bucket"`
	ConsumerName     string `mapstructure:"consumer_name"`

	// Individual component configurations
	Stream   JobStream   `mapstructure:"stream,omitempty"`
	Consumer JobConsumer `mapstructure:"consumer,omitempty"`
	KV       JobKV       `mapstructure:"kv,omitempty"`
	DLQ      JobDLQ      `mapstructure:"dlq,omitempty"`

	Client JobClient `mapstructure:"client,omitempty"`
	Worker JobWorker `mapstructure:"worker,omitempty"`
}

// JobStream configuration for JetStream stream settings.
type JobStream struct {
	MaxAge   string `mapstructure:"max_age"` // e.g. "24h", "1h30m"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
	Discard  string `mapstructure:"discard"` // "old" or "new"
}

// JobConsumer configuration for JetStream consumer settings.
type JobConsumer struct {
	MaxDeliver    int      `mapstructure:"max_deliver"`
	AckWait       string   `mapstructure:"ack_wait"` // e.g. "30s", "1m"
	MaxAckPending int      `mapstructure:"max_ack_pending"`
	ReplayPolicy  string   `mapstructure:"replay_policy"` // "instant" or "original"
	BackOff       []string `mapstructure:"back_off"`      // e.g. ["30s", "2m", "5m", "15m", "30m"]
}

// JobKV configuration for KeyValue bucket settings.
type JobKV struct {
	TTL      string `mapstructure:"ttl"` // e.g. "1h", "30m"
	MaxBytes int64  `mapstructure:"max_bytes"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// JobDLQ configuration for Dead Letter Queue stream settings.
type JobDLQ struct {
	MaxAge   string `mapstructure:"max_age"` // e.g. "7d", "24h"
	MaxMsgs  int64  `mapstructure:"max_msgs"`
	Storage  string `mapstructure:"storage"` // "file" or "memory"
	Replicas int    `mapstructure:"replicas"`
}

// JobClient configuration settings.
type JobClient struct {
	// Host the NATS server hostname.
	Host string `mapstructure:"host"`
	// Port the NATS server port.
	Port int `mapstructure:"port"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name"`
}

// JobWorker configuration settings.
type JobWorker struct {
	// Host the NATS server hostname.
	Host string `mapstructure:"host"`
	// Port the NATS server port.
	Port int `mapstructure:"port"`
	// ClientName the NATS client name for identification.
	ClientName string `mapstructure:"client_name"`
	// QueueGroup for load balancing multiple workers.
	QueueGroup string `mapstructure:"queue_group"`
	// Hostname identifies this worker instance for routing.
	Hostname string `mapstructure:"hostname"`
	// MaxJobs maximum number of concurrent jobs to process.
	MaxJobs int `mapstructure:"max_jobs"`
	// Labels are key-value pairs for label-based routing (e.g., role: web, env: prod).
	Labels map[string]string `mapstructure:"labels"`
}
