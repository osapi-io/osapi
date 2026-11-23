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

// Package main demonstrates user account management: list, get, create
// user accounts, and manage SSH authorized keys on managed hosts using
// the OSAPI SDK.
//
// Run with: OSAPI_TOKEN="<jwt>" go run user.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	// List all user accounts.
	fmt.Println("=== Listing user accounts ===")
	listResp, err := c.User.List(ctx, target)
	if err != nil {
		log.Fatalf("list failed: %v", err)
	}
	for _, r := range listResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		// Each result is one user on one host.
		fmt.Printf("  %s: name=%s uid=%d gid=%d home=%s shell=%s groups=%s locked=%v\n",
			r.Hostname, r.Name, r.UID, r.GID, r.Home, r.Shell,
			strings.Join(r.Groups, ","), r.Locked)
	}

	// Get a specific user account.
	fmt.Println("\n=== Getting root user ===")
	getResp, err := c.User.Get(ctx, target, "root")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}
	for _, r := range getResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		fmt.Printf("  %s: name=%s uid=%d home=%s shell=%s\n",
			r.Hostname, r.Name, r.UID, r.Home, r.Shell)
	}

	// Create a new user account (may fail on non-Debian platforms).
	fmt.Println("\n=== Creating test user ===")
	createResp, err := c.User.Create(ctx, target, client.UserCreateInput{
		Name:   "testuser",
		Shell:  "/bin/bash",
		Groups: []string{"users"},
	})
	if err != nil {
		fmt.Printf("create failed (may be unsupported on this platform): %v\n", err)
	} else {
		for _, r := range createResp.Data.Results {
			fmt.Printf("  %s: name=%s changed=%v error=%s\n",
				r.Hostname, r.Name, r.Changed, r.Error)
		}
	}

	// List SSH authorized keys for root.
	fmt.Println("\n=== Listing SSH keys for root ===")
	keysResp, err := c.UserSSHKey.List(ctx, target, "root")
	if err != nil {
		log.Fatalf("list keys failed: %v", err)
	}
	for _, r := range keysResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		// Each result is one key of the user on one host.
		fmt.Printf("  %s: type=%s fingerprint=%s comment=%s\n",
			r.Hostname, r.Type, r.Fingerprint, r.Comment)
	}

	// Add an SSH key (may fail on non-Debian platforms).
	fmt.Println("\n=== Adding SSH key for testuser ===")
	addResp, err := c.UserSSHKey.Create(ctx, target, client.UserSSHKeyCreateInput{
		Name: "testuser",
		Key:  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample test@example",
	})
	if err != nil {
		fmt.Printf("add key failed (may be unsupported on this platform): %v\n", err)
	} else {
		for _, r := range addResp.Data.Results {
			fmt.Printf("  %s: fingerprint=%s changed=%v error=%s\n",
				r.Hostname, r.Fingerprint, r.Changed, r.Error)
		}
	}
}
