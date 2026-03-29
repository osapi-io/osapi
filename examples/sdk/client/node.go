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

// Package main demonstrates the NodeService: querying status, hostname,
// OS info, disk, memory, load averages, and uptime from a target node.
//
// Run with: OSAPI_TOKEN="<jwt>" go run node.go
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
	target := "_all"

	// Status (aggregated node info).
	status, err := c.Node.Status(ctx, target)
	if err != nil {
		log.Fatalf("status: %v", err)
	}

	for _, r := range status.Data.Results {
		fmt.Printf("Status (%s):\n", r.Hostname)
		fmt.Printf("  Uptime: %s\n", r.Uptime)

		if r.OSInfo != nil {
			fmt.Printf("  OS:     %s %s\n", r.OSInfo.Distribution, r.OSInfo.Version)
		}

		if r.LoadAverage != nil {
			fmt.Printf("  Load:   %.2f %.2f %.2f\n",
				r.LoadAverage.OneMin, r.LoadAverage.FiveMin, r.LoadAverage.FifteenMin)
		}
	}

	// Hostname
	hn, err := c.Node.Hostname(ctx, target)
	if err != nil {
		log.Fatalf("hostname: %v", err)
	}

	for _, r := range hn.Data.Results {
		fmt.Printf("Hostname: %s\n", r.Hostname)
	}

	// Update hostname.
	setResp, err := c.Node.UpdateHostname(ctx, target, "new-hostname")
	if err != nil {
		log.Fatalf("set hostname: %v", err)
	}

	for _, r := range setResp.Data.Results {
		fmt.Printf("Set hostname (%s): changed=%v\n", r.Hostname, r.Changed)
	}

	// Disk usage
	disk, err := c.Node.Disk(ctx, target)
	if err != nil {
		log.Fatalf("disk: %v", err)
	}

	for _, r := range disk.Data.Results {
		fmt.Printf("Disk (%s):\n", r.Hostname)
		for _, d := range r.Disks {
			fmt.Printf("  %s  total=%d  used=%d  free=%d\n",
				d.Name, d.Total, d.Used, d.Free)
		}
	}

	// Memory
	mem, err := c.Node.Memory(ctx, target)
	if err != nil {
		log.Fatalf("memory: %v", err)
	}

	for _, r := range mem.Data.Results {
		fmt.Printf("Memory (%s): total=%d free=%d\n",
			r.Hostname, r.Memory.Total, r.Memory.Free)
	}

	// Load averages
	load, err := c.Node.Load(ctx, target)
	if err != nil {
		log.Fatalf("load: %v", err)
	}

	for _, r := range load.Data.Results {
		fmt.Printf("Load (%s): %.2f %.2f %.2f\n",
			r.Hostname,
			r.LoadAverage.OneMin,
			r.LoadAverage.FiveMin,
			r.LoadAverage.FifteenMin)
	}

	// OS info
	osInfo, err := c.Node.OS(ctx, target)
	if err != nil {
		log.Fatalf("os: %v", err)
	}

	for _, r := range osInfo.Data.Results {
		if r.OSInfo != nil {
			fmt.Printf("OS (%s): %s %s\n",
				r.Hostname, r.OSInfo.Distribution, r.OSInfo.Version)
		}
	}

	// Uptime
	up, err := c.Node.Uptime(ctx, target)
	if err != nil {
		log.Fatalf("uptime: %v", err)
	}

	for _, r := range up.Data.Results {
		fmt.Printf("Uptime (%s): %s\n", r.Hostname, r.Uptime)
	}
}
