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

// Package main demonstrates systemd service management: list the services
// of a host, get one, then deploy a unit file that was uploaded once with
// File.Upload, and start it.
//
// List and Get return one row per service per host. Use .Data.Results to
// iterate over them.
//
// Run with: OSAPI_TOKEN="<jwt>" go run service.go
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

	c := client.New(url, token)
	ctx := context.Background()
	target := "_any"

	// List the services of the host: one row per service per host.
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
		fmt.Printf("  %s: %s state=%s enabled=%v\n",
			r.Hostname, r.Name, r.State, r.Enabled)
	}

	// Get one service.
	fmt.Println("\n=== Getting ssh ===")
	getResp, err := c.Service.Get(ctx, target, "ssh")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		fmt.Printf("  %s: %s state=%s enabled=%v error=%s\n",
			r.Hostname, r.Name, r.State, r.Enabled, r.Error)
	}

	// Deploy the unit file of the object app.service, which
	// osapi client file upload stored, as osapi-app.service (idempotent).
	fmt.Println("\n=== Deploying app ===")
	createResp, err := c.Service.Create(ctx, target,
		client.ServiceCreateInput{Name: "app", Object: "app.service"})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	for _, r := range createResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Start the service (idempotent).
	fmt.Println("\n=== Starting app ===")
	startResp, err := c.Service.Start(ctx, target, "app")
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	for _, r := range startResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}
}
