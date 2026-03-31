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

// Package main demonstrates power management: reboot and shutdown target hosts
// with an optional delay and broadcast message.
//
// All responses return Collection[PowerResult] with per-host results when
// targeting broadcast targets (_all, _any) or a single hostname. Use
// .Data.Results to iterate over the entries.
//
// Run with: OSAPI_TOKEN="<jwt>" go run power.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

func powerExample() {
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

	// Reboot the target host with a 30-second delay and a broadcast message.
	// The agent schedules the reboot via systemd and returns immediately.
	// Returns Collection[PowerResult] with per-host results.
	fmt.Println("=== Rebooting host ===")
	rebootResp, err := c.Power.Reboot(ctx, hostname, client.PowerOpts{
		Delay:   30,
		Message: "Scheduled reboot by osapi",
	})
	if err != nil {
		fmt.Printf("reboot failed (may not be supported on this platform): %v\n", err)
	} else {
		for _, r := range rebootResp.Data.Results {
			if r.Error != "" {
				fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			} else {
				fmt.Printf(
					"  %s: action=%s delay=%ds changed=%v\n",
					r.Hostname,
					r.Action,
					r.Delay,
					r.Changed,
				)
			}
		}
	}

	// Shutdown the target host immediately (delay=0).
	fmt.Println("\n=== Shutting down host ===")
	shutdownResp, err := c.Power.Shutdown(ctx, hostname, client.PowerOpts{
		Delay: 0,
	})
	if err != nil {
		fmt.Printf("shutdown failed (may not be supported on this platform): %v\n", err)
		return
	}

	for _, r := range shutdownResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
		} else {
			fmt.Printf(
				"  %s: action=%s delay=%ds changed=%v\n",
				r.Hostname,
				r.Action,
				r.Delay,
				r.Changed,
			)
		}
	}
}

func main() {
	powerExample()
}
