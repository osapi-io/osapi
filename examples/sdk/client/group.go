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

// Package main demonstrates group management: list groups and create a group
// on managed hosts using the OSAPI SDK.
//
// Run with: OSAPI_TOKEN="<jwt>" go run group.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	// List all groups.
	fmt.Println("=== Listing groups ===")
	listResp, err := c.Group.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		for _, g := range r.Groups {
			fmt.Printf("  %s: name=%s gid=%d members=%s\n",
				r.Hostname, g.Name, g.GID, strings.Join(g.Members, ","))
		}
	}

	// Create a new group (may fail on non-Debian platforms).
	fmt.Println("\n=== Creating test group ===")
	createResp, err := c.Group.Create(ctx, target, client.GroupCreateOpts{
		Name: "testgroup",
	})
	if err != nil {
		fmt.Printf("create failed (may be unsupported on this platform): %v\n", err)
	} else {
		for _, r := range createResp.Data.Results {
			fmt.Printf("  %s: name=%s changed=%v error=%s\n",
				r.Hostname, r.Name, r.Changed, r.Error)
		}
	}
}
