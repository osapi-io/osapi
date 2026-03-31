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

// Package main demonstrates NTP configuration management: get the current NTP
// status from a host, then create an NTP configuration using the SDK.
//
// All responses return Collection[T] with per-host results when targeting
// broadcast targets (_all, _any) or a single hostname. Use .Data.Results
// to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run ntp.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func ntpExample() {
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

	// Get NTP status and configuration from the target host.
	// Returns Collection[NtpStatusResult] with per-host results.
	fmt.Println("=== Getting NTP status ===")
	getResp, err := c.NTP.NtpGet(ctx, hostname)
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}

	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf(
				"  %s: synchronized=%v stratum=%d source=%s\n",
				r.Hostname,
				r.Synchronized,
				r.Stratum,
				r.CurrentSource,
			)
		}
	}

	// Create NTP configuration on the target host.
	// Returns Collection[NtpMutationResult] with per-host results.
	fmt.Println("\n=== Creating NTP configuration ===")
	createResp, err := c.NTP.NtpCreate(ctx, hostname, client.NtpCreateOpts{
		Servers: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
	})
	if err != nil {
		fmt.Printf("create failed (may already exist): %v\n", err)
		return
	}

	for _, r := range createResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf("  %s: changed=%v\n", r.Hostname, r.Changed)
		}
	}
}

func main() {
	ntpExample()
}
