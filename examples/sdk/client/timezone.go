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

// Package main demonstrates timezone management: get the current system
// timezone from a host, then update it using the SDK.
//
// All responses return Collection[T] with per-host results when targeting
// broadcast targets (_all, _any) or a single hostname. Use .Data.Results
// to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run timezone.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func timezoneExample() {
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

	// Get the current system timezone from the target host.
	// Returns Collection[TimezoneResult] with per-host results.
	fmt.Println("=== Getting system timezone ===")
	getResp, err := c.Timezone.Get(ctx, hostname)
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}

	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf(
				"  %s: timezone=%s offset=%s\n",
				r.Hostname,
				r.Timezone,
				r.UTCOffset,
			)
		}
	}

	// Update the timezone on the target host.
	// Returns Collection[TimezoneMutationResult] with per-host results.
	fmt.Println("\n=== Updating system timezone ===")
	updateResp, err := c.Timezone.Update(ctx, hostname, client.TimezoneUpdateOpts{
		Timezone: "UTC",
	})
	if err != nil {
		fmt.Printf("update failed: %v\n", err)
		return
	}

	for _, r := range updateResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf("  %s: changed=%v\n", r.Hostname, r.Changed)
		}
	}
}

func main() {
	timezoneExample()
}
