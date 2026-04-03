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

// Package main demonstrates network route management: list all routes, get
// routes for an interface, and create static routes.
//
// All responses return Collection[T] with per-host results.
// Use .Data.Results to iterate over the per-host entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run route.go
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
	iface := "eth0"

	// List all routes in the kernel routing table.
	fmt.Println("=== Listing all routes ===")
	listResp, err := c.Route.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		for _, rt := range r.Routes {
			fmt.Printf("  %s: %s via %s dev %s metric %d\n",
				r.Hostname, rt.Destination, rt.Gateway,
				rt.Interface, rt.Metric)
		}
	}

	// Get managed routes for a specific interface.
	fmt.Printf("\n=== Getting routes for %s ===\n", iface)
	getResp, err := c.Route.Get(ctx, target, iface)
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf("  %s: %d route(s)\n", r.Hostname, len(r.Routes))
		}
	}

	// Clean up any existing managed routes before creating.
	deleteResp, err := c.Route.Delete(ctx, target, iface)
	if err == nil {
		for _, r := range deleteResp.Data.Results {
			fmt.Printf("  cleanup %s: changed=%v\n",
				r.Hostname, r.Changed)
		}
	}

	// Create static routes for an interface.
	fmt.Println("\n=== Creating static routes ===")
	metric := 100
	createResp, err := c.Route.Create(ctx, target, iface,
		client.RouteConfigOpts{
			Routes: []client.RouteItem{
				{To: "10.0.0.0/8", Via: "192.168.1.1"},
				{To: "172.16.0.0/12", Via: "192.168.1.1",
					Metric: &metric},
			},
		})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	for _, r := range createResp.Data.Results {
		fmt.Printf("  %s: interface=%s changed=%v error=%s\n",
			r.Hostname, r.Interface, r.Changed, r.Error)
	}
}
