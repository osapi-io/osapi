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

// Package main demonstrates package management: list installed packages
// on a host, then get details for a specific package.
//
// Run with: OSAPI_TOKEN="<jwt>" go run package.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

func packageExample() {
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

	// List all installed packages on the target host: one row per
	// package per host.
	fmt.Println("=== Listing installed packages ===")
	listResp, err := c.Package.List(ctx, hostname)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}

	for i, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		if i >= 5 {
			fmt.Println("  ...")
			break
		}
		fmt.Printf("  %s: %s %s (%s)\n", r.Hostname, r.Name, r.Version, r.Manager)
	}

	// Get one package.
	fmt.Println("\n=== Getting package details ===")
	getResp, err := c.Package.Get(ctx, hostname, "bash")
	if err != nil {
		// Package may not be installed on all platforms.
		fmt.Printf("  get failed (may not be installed): %v\n", err)
		return
	}

	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: %s version=%s manager=%s\n", r.Hostname, r.Name, r.Version, r.Manager)
	}
}

func main() {
	packageExample()
}
