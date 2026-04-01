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
		for _, u := range r.Users {
			fmt.Printf("  %s: name=%s uid=%d gid=%d home=%s shell=%s groups=%s locked=%v\n",
				r.Hostname, u.Name, u.UID, u.GID, u.Home, u.Shell,
				strings.Join(u.Groups, ","), u.Locked)
		}
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
		for _, u := range r.Users {
			fmt.Printf("  %s: name=%s uid=%d home=%s shell=%s\n",
				r.Hostname, u.Name, u.UID, u.Home, u.Shell)
		}
	}

	// Create a new user account (may fail on non-Debian platforms).
	fmt.Println("\n=== Creating test user ===")
	createResp, err := c.User.Create(ctx, target, client.UserCreateOpts{
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
	keysResp, err := c.User.ListKeys(ctx, target, "root")
	if err != nil {
		log.Fatalf("list keys failed: %v", err)
	}
	for _, r := range keysResp.Data.Results {
		if r.Error != "" {
			fmt.Printf("  %s: ERROR %s\n", r.Hostname, r.Error)
			continue
		}
		if len(r.Keys) == 0 {
			fmt.Printf("  %s: no authorized keys\n", r.Hostname)
			continue
		}
		for _, k := range r.Keys {
			fmt.Printf("  %s: type=%s fingerprint=%s comment=%s\n",
				r.Hostname, k.Type, k.Fingerprint, k.Comment)
		}
	}

	// Add an SSH key (may fail on non-Debian platforms).
	fmt.Println("\n=== Adding SSH key for testuser ===")
	addResp, err := c.User.AddKey(ctx, target, "testuser", client.SSHKeyAddOpts{
		Key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample test@example",
	})
	if err != nil {
		fmt.Printf("add key failed (may be unsupported on this platform): %v\n", err)
	} else {
		for _, r := range addResp.Data.Results {
			fmt.Printf("  %s: changed=%v error=%s\n",
				r.Hostname, r.Changed, r.Error)
		}
	}
}
