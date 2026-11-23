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

// Package main demonstrates network route management: list the routes of
// an interface, add a static route, change its metric, and drop it. Each
// call is one route, as ip route add and ip route del are.
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
	iface := "eth0"

	// List the routes of the interface: one row per route per host.
	fmt.Printf("=== Listing the routes of %s ===\n", iface)
	listResp, err := c.NetworkInterfaceRoute.List(ctx, target, iface)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %s via %s metric %d (%s)\n",
			r.Hostname, r.Destination, r.Gateway, r.Metric, r.Proto)
	}

	// Add a static route (idempotent).
	route := client.NetworkInterfaceRouteCreateInput{
		Name:        iface,
		Destination: "10.0.0.0/8",
		Via:         "192.168.1.1",
	}
	fmt.Println("\n=== Adding 10.0.0.0/8 ===")
	createResp, err := c.NetworkInterfaceRoute.Create(ctx, target, route)
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}
	for _, r := range createResp.Data.Results {
		fmt.Printf("  %s: changed=%v error=%s\n", r.Hostname, r.Changed, r.Error)
	}

	// Change its metric.
	route.Metric = 100
	fmt.Println("\n=== Changing its metric to 100 ===")
	if _, err := c.NetworkInterfaceRoute.Update(
		ctx,
		target,
		client.NetworkInterfaceRouteUpdateInput(route),
	); err != nil {
		log.Fatalf("update failed: %v", err)
	}

	// Drop it.
	fmt.Println("\n=== Dropping it ===")
	deleteResp, err := c.NetworkInterfaceRoute.Delete(
		ctx,
		target,
		client.NetworkInterfaceRouteDeleteInput{Name: iface, Destination: route.Destination},
	)
	if err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	for _, r := range deleteResp.Data.Results {
		fmt.Printf("  %s: changed=%v\n", r.Hostname, r.Changed)
	}
}
