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
package health

// LivenessInput is the input of the liveness probe, which has none.
type LivenessInput struct{}

// ReadyInput is the input of the readiness probe, which has none.
type ReadyInput struct{}

// StatusInput is the input of status, which has none.
type StatusInput struct{}

// HealthResponse is the liveness probe's answer.
type HealthResponse struct { //nolint:revive // the schema name is part of the API
	Status string `json:"status" doc:"Health status." example:"ok"`
}

// ReadyResponse is the readiness probe's answer.
type ReadyResponse struct {
	Status string `json:"status"          doc:"Readiness status."             example:"ready"`
	Error  string `json:"error,omitempty" doc:"Error message when not ready."`
}

// StatusResponse is the health of each component, with the metrics of the
// system.
type StatusResponse struct {
	Status       string                     `json:"status"                  doc:"Overall health status."                         example:"ok"`
	Components   map[string]ComponentHealth `json:"components"              doc:"Per-component health status."`
	Version      string                     `json:"version"                 doc:"Application version."                           example:"0.1.0"`
	Uptime       string                     `json:"uptime"                  doc:"Time since server started."                     example:"2h30m"`
	NATS         *NATSInfo                  `json:"nats,omitempty"`
	Streams      []StreamInfo               `json:"streams,omitempty"       doc:"JetStream stream statistics."`
	KVBuckets    []KVBucketInfo             `json:"kv_buckets,omitempty"    doc:"KV bucket statistics."`
	ObjectStores []ObjectStoreInfo          `json:"object_stores,omitempty" doc:"Object Store statistics."`
	Jobs         *JobStats                  `json:"jobs,omitempty"`
	Agents       *AgentStats                `json:"agents,omitempty"`
	Consumers    *ConsumerStats             `json:"consumers,omitempty"`
	Registry     []ComponentEntry           `json:"registry,omitempty"      doc:"All registered components with health details."`
}

// ComponentHealth is the health of one component.
type ComponentHealth struct {
	Status  string `json:"status"            doc:"Component health status."                                                                                              example:"ok"`
	Error   string `json:"error,omitempty"   doc:"Error message when component is unhealthy."`
	Address string `json:"address,omitempty" doc:"Listen address of the component (if applicable)."                                                                      example:"http://0.0.0.0:8080"`
	URL     string `json:"url,omitempty"     doc:"Where a person opens the component, from its url config key. Absent when the key is not set. osapi never requests it." example:"https://hub.example.ts.net/"`
}

// NATSInfo is the NATS server the controller is connected to.
type NATSInfo struct {
	URL     string `json:"url"     doc:"Connected NATS server URL." example:"nats://localhost:4222"`
	Version string `json:"version" doc:"NATS server version."       example:"2.10.0"`
}

// StreamInfo is one JetStream stream.
type StreamInfo struct {
	Name      string `json:"name"      doc:"Stream name."                       example:"JOBS"`
	Messages  int    `json:"messages"  doc:"Number of messages in the stream."  example:"42"`
	Bytes     int    `json:"bytes"     doc:"Total bytes in the stream."         example:"1024"`
	Consumers int    `json:"consumers" doc:"Number of consumers on the stream." example:"1"`
}

// KVBucketInfo is one KV bucket.
type KVBucketInfo struct {
	Name  string `json:"name"  doc:"KV bucket name."               example:"job-queue"`
	Keys  int    `json:"keys"  doc:"Number of keys in the bucket." example:"10"`
	Bytes int    `json:"bytes" doc:"Total bytes in the bucket."    example:"2048"`
}

// ObjectStoreInfo is one Object Store bucket.
type ObjectStoreInfo struct {
	Name string `json:"name" doc:"Object Store bucket name." example:"file-objects"`
	Size int    `json:"size" doc:"Total bytes in the store." example:"5242880"`
}

// JobStats is the count of the jobs in each state.
type JobStats struct {
	Total       int `json:"total"       doc:"Total number of jobs."                    example:"100"`
	Unprocessed int `json:"unprocessed" doc:"Number of unprocessed jobs."              example:"5"`
	Processing  int `json:"processing"  doc:"Number of jobs currently processing."     example:"2"`
	Completed   int `json:"completed"   doc:"Number of completed jobs."                example:"90"`
	Failed      int `json:"failed"      doc:"Number of failed jobs."                   example:"3"`
	DLQ         int `json:"dlq"         doc:"Number of jobs in the dead letter queue." example:"0"`
}

// AgentStats is the count of the agents, with each one.
type AgentStats struct {
	Total  int           `json:"total"            doc:"Total number of registered agents."  example:"5"`
	Ready  int           `json:"ready"            doc:"Number of agents with Ready status." example:"5"`
	Agents []AgentDetail `json:"agents,omitempty" doc:"Per-agent registration details."`
}

// AgentDetail is one registered agent.
type AgentDetail struct {
	Hostname   string `json:"hostname"         doc:"Agent hostname."                         example:"web-01"`
	Labels     string `json:"labels,omitempty" doc:"Formatted label string."                 example:"group=web.prod"`
	Registered string `json:"registered"       doc:"Time since last heartbeat registration." example:"15s ago"`
}

// ComponentEntry is one registered component, with its health.
type ComponentEntry struct {
	Type       string   `json:"type,omitempty"        doc:"Component type (agent, api, nats)." example:"agent"`
	Hostname   string   `json:"hostname,omitempty"    doc:"Component hostname."                example:"web-01"`
	Status     string   `json:"status,omitempty"      doc:"Component status."                  example:"Ready"`
	Conditions []string `json:"conditions,omitempty"  doc:"Active conditions."`
	Age        string   `json:"age,omitempty"         doc:"Time since component started."      example:"7h 6m"`
	CPUPercent float32  `json:"cpu_percent,omitempty" doc:"Process CPU usage percentage."      example:"2.1"`
	MemBytes   int64    `json:"mem_bytes,omitempty"   doc:"Process RSS memory in bytes."       example:"134217728"`
}

// ConsumerStats is the count of the JetStream consumers, with each one.
type ConsumerStats struct {
	Total     int              `json:"total"               doc:"Total number of JetStream consumers." example:"2"`
	Consumers []ConsumerDetail `json:"consumers,omitempty" doc:"Per-consumer details."`
}

// ConsumerDetail is one JetStream consumer.
type ConsumerDetail struct {
	Name        string `json:"name"        doc:"Consumer name."                                 example:"query_any_web_01"`
	Pending     int    `json:"pending"     doc:"Messages not yet delivered."                    example:"0"`
	AckPending  int    `json:"ack_pending" doc:"Messages delivered but not yet acknowledged."   example:"3"`
	Redelivered int    `json:"redelivered" doc:"Messages redelivered and not yet acknowledged." example:"0"`
}
