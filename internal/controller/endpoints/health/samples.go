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

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them.
var (
	sampleLiveness = HealthResponse{Status: "ok"}
	sampleReady    = ReadyResponse{Status: "ready"}

	sampleStatus = StatusResponse{
		Status:  "ok",
		Version: "v1.4.0",
		Uptime:  "51h3m",
		Checks: []Check{
			{Name: "nats", Status: "ok"},
			{Name: "kv", Status: "ok"},
		},
		Registry: []ComponentEntry{
			{
				Type:       "controller",
				Hostname:   "osapi-01",
				Status:     "ok",
				Age:        "51h3m",
				CPUPercent: 1.2,
				MemBytes:   52428800,
				Version:    "v1.4.0",
				Address:    "0.0.0.0:8080",
			},
			{
				Type:     "nats",
				Hostname: "osapi-01",
				Status:   "ok",
				Age:      "51h3m",
				Version:  "2.11.4",
				URL:      "nats://localhost:4222",
			},
			{
				Type:       "agent",
				Hostname:   "web-01",
				Status:     "ok",
				Age:        "51h2m",
				CPUPercent: 0.4,
				MemBytes:   31457280,
				Version:    "v1.4.0",
			},
		},
		Jobs: &JobStats{
			Total:       128,
			Completed:   121,
			Failed:      3,
			Processing:  1,
			Unprocessed: 3,
		},
		Streams: []StreamInfo{{Name: "JOBS", Messages: 128, Bytes: 262144, Consumers: 2}},
		KVBuckets: []KVBucketInfo{
			{Name: "job-queue", Keys: 128, Bytes: 524288},
			{Name: "agent-registry", Keys: 2, Bytes: 4096},
		},
		ObjectStores: []ObjectStoreInfo{{Name: "file-objects", Size: 1048576}},
		Consumers:    2,
	}
)
