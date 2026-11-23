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

// Package main reads the systemd journal of a host: its newest entries, the
// sources that wrote them, and the entries of one systemd unit.
//
// Each response is a Collection with one row for each entry, or source, of
// each host. A host that skipped or failed has one row with its error.
//
// Run with: OSAPI_TOKEN="<jwt>" go run log.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

func logExample() {
	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN is required")
	}

	c := client.New(url, token)
	ctx := context.Background()
	hostname := "_any"

	// The 10 newest entries of each host's journal.
	fmt.Println("=== The newest entries ===")
	entries, err := c.LogEntry.List(ctx, hostname, client.LogEntryListParams{Lines: 10})
	if err != nil {
		log.Fatalf("log entries failed: %v", err)
	}
	for _, r := range entries.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s [%s] %s %s: %s\n",
			r.Hostname, r.Timestamp, r.Priority, r.Unit, r.Message)
	}

	// Each source that wrote to the journal.
	fmt.Println("\n=== The sources ===")
	sources, err := c.LogSource.List(ctx, hostname)
	if err != nil {
		log.Fatalf("log sources failed: %v", err)
	}
	for _, r := range sources.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %s\n", r.Hostname, r.Name)
	}

	// The errors of one systemd unit in the last hour.
	fmt.Println("\n=== The errors of sshd.service in the last hour ===")
	unit, err := c.LogEntry.List(ctx, hostname, client.LogEntryListParams{
		Unit:     "sshd.service",
		Since:    "1 hour ago",
		Priority: "err",
	})
	if err != nil {
		log.Fatalf("log entries of sshd.service failed: %v", err)
	}
	for _, r := range unit.Data.Results {
		fmt.Printf("  %s pid=%d %s\n", r.Hostname, r.PID, r.Message)
	}
}

func main() {
	logExample()
}
