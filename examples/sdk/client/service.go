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

// Package main demonstrates systemd service management: list services,
// get details for a specific service, and start a service.
//
// All responses return Collection[T] with per-host results.
// Use .Data.Results to iterate over the per-host entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run service.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
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

	c := client.New(url, token)
	ctx := context.Background()
	target := "_any"

	// List all services on the host.
	fmt.Println("=== Listing services ===")
	listResp, err := c.Service.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		for _, svc := range r.Services {
			fmt.Printf("  %s: %s status=%s enabled=%v\n",
				r.Hostname, svc.Name, svc.Status, svc.Enabled)
		}
	}

	// Get details for a specific service.
	fmt.Println("\n=== Getting ssh.service ===")
	getResp, err := c.Service.Get(ctx, target, "ssh.service")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else if r.Service != nil {
			fmt.Printf("  %s: %s status=%s enabled=%v pid=%d\n",
				r.Hostname, r.Service.Name, r.Service.Status,
				r.Service.Enabled, r.Service.PID)
		}
	}

	// Start the service (idempotent).
	fmt.Println("\n=== Starting ssh.service ===")
	startResp, err := c.Service.Start(ctx, target, "ssh.service")
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	for _, r := range startResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n",
			r.Hostname, r.Changed, r.Error)
	}
}
