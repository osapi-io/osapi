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

// Package main demonstrates the DNSService: reading DNS configuration
// for a network interface and updating DNS servers.
//
// Run with: OSAPI_TOKEN="<jwt>" OSAPI_INTERFACE="eth0" go run dns.go
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

	iface := os.Getenv("OSAPI_INTERFACE")
	if iface == "" {
		log.Fatal("OSAPI_INTERFACE is required (e.g. eth0, en0)")
	}

	c := client.New(url, token)
	ctx := context.Background()
	target := "_all"

	// Get DNS configuration for an interface.
	resp, err := c.DNS.Get(ctx, target, iface)
	if err != nil {
		log.Fatalf("get dns: %v", err)
	}

	for _, r := range resp.Data.Results {
		fmt.Printf("DNS (%s):\n", r.Hostname)
		fmt.Printf("  Servers: %v\n", r.Servers)
		fmt.Printf("  Search:  %v\n", r.SearchDomains)
	}
}
