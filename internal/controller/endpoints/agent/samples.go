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

package agent

import (
	"time"

	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
)

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them. Their times are relative to the time that
// the docs generator fixes, 2026-01-15 10:00 UTC.
var (
	sampleStarted = time.Date(2026, time.January, 13, 7, 0, 0, 0, time.UTC)
	sampleSeen    = time.Date(2026, time.January, 15, 9, 59, 52, 0, time.UTC)

	sampleWeb01 = AgentInfo{
		Hostname:     "web-01",
		MachineID:    "4c4c4544-0035-4b10-8031-b4c04f4d3332",
		State:        "Ready",
		Labels:       map[string]string{"group": "web"},
		StartedAt:    sampleStarted,
		RegisteredAt: sampleSeen,
		LoadAverage:  &LoadAverageResponse{OneMin: 0.32, FiveMin: 0.28, FifteenMin: 0.25},
		OSInfo: &OSInfoResponse{
			Distribution: "Ubuntu",
			Version:      "24.04",
			Family:       "debian",
		},
		Version:          "v1.4.0",
		Uptime:           "2 days, 3 hours",
		Memory:           &MemoryResponse{Total: 8589934592, Used: 3221225472, Free: 5368709120},
		Architecture:     "amd64",
		KernelVersion:    "6.8.0-51-generic",
		CPUCount:         4,
		FQDN:             "web-01.example.com",
		ServiceMgr:       "systemd",
		PackageMgr:       "apt",
		PrimaryInterface: "eth0",
	}
	sampleMac01 = AgentInfo{
		Hostname:     "mac-01",
		State:        "Draining",
		StartedAt:    time.Date(2026, time.January, 15, 8, 30, 0, 0, time.UTC),
		RegisteredAt: sampleSeen,
		LoadAverage:  &LoadAverageResponse{OneMin: 1.12, FiveMin: 0.94, FifteenMin: 0.88},
		OSInfo:       &OSInfoResponse{Distribution: "macOS", Version: "15.5", Family: "darwin"},
		Version:      "v1.4.0",
	}

	sampleList = ListAgentsResponse{Items: []AgentInfo{sampleWeb01, sampleMac01}, TotalItems: 2}

	sampleStatus = health.ComponentList{
		TotalItems: 2,
		Items: []health.ComponentEntry{
			{
				Type: "agent", Hostname: "web-01", Status: "ok", Age: "51h2m",
				CPUPercent: 0.4, MemBytes: 31457280, Version: "v1.4.0",
				Components: []health.ComponentEntry{
					{Type: "heartbeat", Status: "ok"},
					{Type: "facts", Status: "ok"},
				},
			},
			{
				Type: "agent", Hostname: "mac-01", Status: "ok", Age: "1h30m",
				CPUPercent: 0.7, MemBytes: 28311552, Version: "v1.4.0",
			},
		},
	}
)
