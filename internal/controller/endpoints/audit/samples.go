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

package audit

import (
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// The sample answers of the examples, which the docs page of each command
// shows as the CLI prints them.
var (
	sampleEntry = AuditEntry{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:    time.Date(2026, time.January, 15, 9, 58, 12, 0, time.UTC),
		User:         "ops@example.com",
		Roles:        []string{"admin"},
		Method:       "PUT",
		Path:         "/api/node/web-01/sysctl/net.ipv4.ip_forward",
		ResponseCode: 200,
		DurationMs:   184,
		SourceIP:     "10.0.0.5",
		OperationID:  "UpdateSysctl",
		TraceID:      "26e448c9dbe1c4a8d216a98c675079c9",
		JobID:        "7c9e6679-7425-40de-944b-e07fc1f90ae7",
		Steps: []HostSteps{{
			Hostname: "web-01",
			Steps: []client.Step{{
				Behavior:   "managedfile",
				Method:     "update",
				Changed:    &sampleChanged,
				DurationMS: 31,
			}},
		}},
		Chain: []string{"web-01: managedfile update (changed)"},
	}

	sampleChanged = true

	sampleList = ListAuditResponse{
		TotalItems: 2,
		Items: []AuditEntry{
			sampleEntry,
			{
				ID:           "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				Timestamp:    time.Date(2026, time.January, 15, 9, 41, 3, 0, time.UTC),
				User:         "ci@example.com",
				Roles:        []string{"read"},
				Method:       "GET",
				Path:         "/api/agent",
				ResponseCode: 200,
				DurationMs:   12,
				SourceIP:     "10.0.0.9",
				OperationID:  "GetAgents",
			},
		},
	}

	sampleGet = AuditEntryResponse{Entry: sampleEntry}
)
