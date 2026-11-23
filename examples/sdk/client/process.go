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
// All responses return Collection[ProcessEntry] with per-host results
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

	"github.com/osapi-io/osapi/pkg/sdk/client"
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

	// List the processes of the target host. Each row of the
	// Collection[ProcessEntry] is one process of one host.
	fmt.Println("=== Listing processes ===")
	listResp, err := c.Process.List(ctx, hostname)
	if err != nil {
		log.Fatalf("process list failed: %v", err)
	}

	for _, p := range listResp.Data.Results[:min(5, len(listResp.Data.Results))] {
		if p.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", p.Hostname, p.Error)
			continue
		}
		fmt.Printf("  %s: PID=%d name=%s user=%s cpu=%.1f%% mem=%.1f%%\n",
			p.Hostname, p.PID, p.Name, p.User, p.CPUPercent, p.MemPercent)
	}
	if len(listResp.Data.Results) > 5 {
		fmt.Printf("  ... and %d more\n", len(listResp.Data.Results)-5)
	}

	// Get PID 1, the host's init.
	fmt.Println("\n=== Getting process PID 1 ===")
	getResp, err := c.Process.Get(ctx, hostname, 1)
	if err != nil {
		log.Fatalf("process get failed: %v", err)
	}

	for _, p := range getResp.Data.Results {
		if p.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", p.Hostname, p.Error)
			continue
		}
		fmt.Printf("  %s: PID=%d name=%s user=%s state=%s\n",
			p.Hostname, p.PID, p.Name, p.User, p.State)
	}
}

func main() {
	processExample()
}
