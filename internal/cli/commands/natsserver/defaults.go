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

package natsserver

import "github.com/osapi-io/osapi/internal/config"

// Defaults are the defaults of the nats section of the config, which nats
// start owns: the embedded server, and the names, limits and storage of each
// stream and bucket that osapi makes. osapi config default get prints them.
var Defaults = config.NATS{
	Server: config.NATSServer{
		Host:      "localhost",
		Port:      4222,
		StoreDir:  "/var/lib/osapi/jetstream",
		KeyDir:    "/etc/osapi/pki",
		Namespace: "osapi",
		Auth:      config.NATSServerAuth{Type: "none"},
		Metrics: config.MetricsServer{
			Enabled: true, Host: "0.0.0.0", Port: 9092, TLS: config.ServerTLS{Enabled: true},
		},
		// TLS on, with a self-signed certificate made at the first start
		// (osapi-io/osapi#1413).
		TLS: config.ServerTLS{Enabled: true},
	},
	Stream: config.NATSStream{
		Name: "JOBS", Subjects: "jobs.>", MaxAge: "24h", MaxMsgs: 10000,
		Storage: "file", Replicas: 1, Discard: "old",
	},
	KV: config.NATSKV{
		Bucket: "job-queue", TTL: "1h", MaxBytes: 104857600, Storage: "file", Replicas: 1,
	},
	DLQ: config.NATSDLQ{MaxAge: "168h", MaxMsgs: 1000, Storage: "file", Replicas: 1},
	Audit: config.NATSAudit{
		Stream: "AUDIT", Subject: "audit", MaxAge: "720h", MaxBytes: 52428800,
		Storage: "file", Replicas: 1,
	},
	Registry: config.NATSRegistry{
		Bucket:   "agent-registry",
		TTL:      "30s",
		Storage:  "file",
		Replicas: 1,
	},
	Facts: config.NATSFacts{Bucket: "agent-facts", TTL: "5m", Storage: "file", Replicas: 1},
	State: config.NATSState{Bucket: "agent-state", Storage: "file", Replicas: 1},
	Objects: config.NATSObjects{
		Bucket:   "file-objects",
		MaxBytes: 10737418240,
		Storage:  "file",
		Replicas: 1,
	},
	FileState:  config.NATSFileState{Bucket: "file-state", Storage: "file", Replicas: 1},
	Secrets:    config.NATSSecrets{Bucket: "secrets", History: 10, Storage: "file", Replicas: 1},
	Enrollment: config.NATSEnrollment{Bucket: "agent-enrollment", Storage: "file", Replicas: 1},
}
