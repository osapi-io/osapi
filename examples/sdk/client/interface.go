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

// Package main demonstrates network interface management: list interfaces,
// get details for a specific interface, and create a static configuration.
//
// All responses return Collection[T] with per-host results.
// Use .Data.Results to iterate over the per-host entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run interface.go
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

	// List all interfaces on the host.
	fmt.Println("=== Listing interfaces ===")
	listResp, err := c.Interface.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		for _, iface := range r.Interfaces {
			fmt.Printf("  %s: %s managed=%v state=%s\n",
				r.Hostname, iface.Name, iface.Managed, iface.State)
		}
	}

	// Get details for a specific interface.
	fmt.Println("\n=== Getting eth0 ===")
	getResp, err := c.Interface.Get(ctx, target, "eth0")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else if r.Interface != nil {
			iface := r.Interface
			fmt.Printf("  %s: %s addresses=%v gateway4=%s\n",
				r.Hostname, iface.Name, iface.Addresses, iface.Gateway4)
		}
	}

	// Create a static interface configuration (idempotent cleanup first).
	_ = func() {
		deleteResp, err := c.Interface.Delete(ctx, target, "osapi-example0")
		if err == nil {
			for _, r := range deleteResp.Data.Results {
				fmt.Printf("  cleanup %s: changed=%v\n",
					r.Hostname, r.Changed)
			}
		}
	}

	fmt.Println("\n=== Creating interface config ===")
	dhcp4 := false
	createResp, err := c.Interface.Create(ctx, target, "eth0",
		client.InterfaceConfigOpts{
			DHCP4:     &dhcp4,
			Addresses: []string{"192.168.1.100/24"},
			Gateway4:  "192.168.1.1",
		})
	if err != nil {
		fmt.Printf("  create failed (may already exist): %v\n", err)
	} else {
		for _, r := range createResp.Data.Results {
			fmt.Printf("  %s: name=%s changed=%v error=%s\n",
				r.Hostname, r.Name, r.Changed, r.Error)
		}
	}
}
