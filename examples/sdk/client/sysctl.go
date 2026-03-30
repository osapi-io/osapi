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

// Package main demonstrates sysctl parameter management: list all parameters
// on a host, then create a kernel parameter using the SDK.
//
// All responses return Collection[T] with per-host results when targeting
// broadcast targets (_all, _any) or a single hostname. Use .Data.Results
// to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run sysctl.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func sysctlExample() {
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

	// List all sysctl parameters on the target host.
	// Returns Collection[SysctlEntryResult] with per-host results.
	fmt.Println("=== Listing sysctl parameters ===")
	listResp, err := c.Sysctl.SysctlList(ctx, hostname)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}

	// Print the first few entries.
	fmt.Printf("Found %d parameters on %s\n", len(listResp.Data.Results), hostname)
	if len(listResp.Data.Results) > 0 {
		for i, r := range listResp.Data.Results {
			if i >= 3 {
				fmt.Println("  ...")
				break
			}
			if r.Error != "" {
				fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			} else {
				fmt.Printf("  %s: %s=%s\n", r.Hostname, r.Key, r.Value)
			}
		}
	}

	// Create a sysctl parameter.
	// Returns Collection[SysctlMutationResult] with per-host results.
	fmt.Println("\n=== Creating sysctl parameter ===")
	setResp, err := c.Sysctl.SysctlCreate(ctx, hostname, client.SysctlCreateOpts{
		Key:   "net.ipv4.ip_forward",
		Value: "1",
	})
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}

	// Print the result.
	for _, r := range setResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf("  %s: key=%s changed=%v\n", r.Hostname, r.Key, r.Changed)
		}
	}
}

func main() {
	sysctlExample()
}
