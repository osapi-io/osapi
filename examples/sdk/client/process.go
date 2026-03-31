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

// Package main demonstrates process management: list running processes and
// get details for a specific PID.
//
// All responses return Collection[ProcessInfoResult] with per-host results
// when targeting broadcast targets (_all, _any) or a single hostname. Use
// .Data.Results to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run process.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func processExample() {
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

	// List all running processes on the target host.
	// Returns Collection[ProcessInfoResult] with per-host results.
	fmt.Println("=== Listing processes ===")
	listResp, err := c.Process.List(ctx, hostname)
	if err != nil {
		log.Fatalf("process list failed: %v", err)
	}

	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %d processes\n", r.Hostname, len(r.Processes))
		for _, p := range r.Processes[:min(5, len(r.Processes))] {
			fmt.Printf("    PID=%d name=%s user=%s cpu=%.1f%% mem=%.1f%%\n",
				p.PID, p.Name, p.User, p.CPUPercent, p.MemPercent)
		}
		if len(r.Processes) > 5 {
			fmt.Printf("    ... and %d more\n", len(r.Processes)-5)
		}
	}

	// Get details for PID 1 (init/systemd).
	fmt.Println("\n=== Getting process PID 1 ===")
	getResp, err := c.Process.Get(ctx, hostname, 1)
	if err != nil {
		fmt.Printf("process get failed (may not have permission): %v\n", err)
		return
	}

	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		for _, p := range r.Processes {
			fmt.Printf("  %s: PID=%d name=%s user=%s state=%s\n",
				r.Hostname, p.PID, p.Name, p.User, p.State)
		}
	}
}

func main() {
	processExample()
}
