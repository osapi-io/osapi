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

package telemetry

import "time"

// sampleTrace is the sample answer of the get example, which the docs page
// of the command shows as the CLI prints it.
var sampleTrace = TraceResponse{
	TraceID:   "26e448c9dbe1c4a8d216a98c675079c9",
	Started:   time.Date(2026, time.January, 15, 9, 58, 12, 0, time.UTC),
	Duration:  "186ms",
	Services:  []string{"osapi-controller", "osapi-agent"},
	NATSCalls: 2,
	NATSTime:  "4ms",
	Spans: []TraceNode{
		{
			Name:     "PUT /api/node/{hostname}/sysctl/{key}",
			On:       "osapi-controller",
			Duration: "186ms",
			Status:   "ok",
			Span: TraceSpan{
				SpanID:     "a1b2c3d4e5f60718",
				Name:       "PUT /api/node/{hostname}/sysctl/{key}",
				Service:    "osapi-controller",
				Kind:       "server",
				StartTime:  time.Date(2026, time.January, 15, 9, 58, 12, 0, time.UTC),
				DurationMs: 186,
				Status:     "ok",
				Attributes: map[string]string{"http.response.status_code": "200"},
			},
			Children: []TraceNode{{
				Name: "behavior.sysctl.update", On: "web-01", Duration: "171ms", Status: "ok",
				Span: TraceSpan{
					SpanID:       "b2c3d4e5f6071829",
					ParentSpanID: "a1b2c3d4e5f60718",
					Name:         "behavior.sysctl.update",
					Service:      "osapi-agent",
					Kind:         "internal",
					StartTime:    time.Date(2026, time.January, 15, 9, 58, 12, 9000000, time.UTC),
					DurationMs:   171,
					Status:       "ok",
					Attributes:   map[string]string{"osapi.host": "web-01"},
				},
			}},
		},
	},
}
