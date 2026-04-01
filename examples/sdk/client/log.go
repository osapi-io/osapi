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

// Package main demonstrates log management: query journal entries and
// query entries scoped to a specific systemd unit.
//
// All responses return Collection[LogEntryResult] with per-host results
// when targeting broadcast targets (_all, _any) or a single hostname. Use
// .Data.Results to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run log.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
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

	// Query the last 10 journal entries from the target host.
	// Returns Collection[LogEntryResult] with per-host results.
	fmt.Println("=== Querying journal entries ===")
	lines := 10
	queryResp, err := c.Log.Query(ctx, hostname, client.LogQueryOpts{
		Lines: &lines,
	})
	if err != nil {
		log.Fatalf("log query failed: %v", err)
	}

	for _, r := range queryResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %d entries\n", r.Hostname, len(r.Entries))
		for _, e := range r.Entries[:min(3, len(r.Entries))] {
			fmt.Printf("    [%s] %s %s: %s\n",
				e.Timestamp, e.Priority, e.Unit, e.Message)
		}
		if len(r.Entries) > 3 {
			fmt.Printf("    ... and %d more\n", len(r.Entries)-3)
		}
	}

	// List available log sources on the host.
	fmt.Println("\n=== Listing log sources ===")
	srcResp, err := c.Log.Sources(ctx, hostname)
	if err != nil {
		log.Fatalf("log sources failed: %v", err)
	}

	for _, r := range srcResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %d sources\n", r.Hostname, len(r.Sources))
		for _, src := range r.Sources[:min(10, len(r.Sources))] {
			fmt.Printf("    %s\n", src)
		}
		if len(r.Sources) > 10 {
			fmt.Printf("    ... and %d more\n", len(r.Sources)-10)
		}
	}

	// Query entries for the sshd systemd unit.
	fmt.Println("\n=== Querying sshd.service entries ===")
	unitResp, err := c.Log.QueryUnit(ctx, hostname, "sshd.service",
		client.LogQueryOpts{Lines: &lines})
	if err != nil {
		fmt.Printf("log unit query failed (sshd may not be running): %v\n",
			err)
		return
	}

	for _, r := range unitResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %d sshd entries\n", r.Hostname, len(r.Entries))
		for _, e := range r.Entries {
			fmt.Printf("    [%s] pid=%d %s\n",
				e.Priority, e.PID, e.Message)
		}
	}
}

func main() {
	logExample()
}
