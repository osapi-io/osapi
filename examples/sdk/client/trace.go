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

// Package main demonstrates the TelemetryService: reading the spans of a
// trace through the controller, by the trace ID a command prints.
//
// Run with: OSAPI_TOKEN="<jwt>" go run trace.go <trace-id>
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

func main() {
	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN is required")
	}
	if len(os.Args) < 2 {
		log.Fatal("usage: go run trace.go <trace-id>")
	}

	c := client.New(url, token)

	// A trace reaches the backend a few seconds after its request, so a
	// NotFoundError right after a command means: try again shortly.
	resp, err := c.Telemetry.Trace(context.Background(), os.Args[1])
	if err != nil {
		log.Fatalf("get trace: %v", err)
	}

	fmt.Printf("Trace %s: %d spans\n", resp.Data.ID, len(resp.Data.Spans))
	for _, s := range resp.Data.Spans {
		status := "ok"
		if s.Failed {
			status = "error: " + s.StatusMessage
		}
		fmt.Printf("  %-18s %-28s %10s  %s\n", s.Service, s.Name, s.Duration, status)
	}
}
