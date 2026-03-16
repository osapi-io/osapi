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

// Package main demonstrates the AgentService: listing the fleet and
// retrieving rich facts for a specific agent — OS info, load averages,
// memory stats, network interfaces, labels, and lifecycle timestamps.
//
// Run with: OSAPI_TOKEN="<jwt>" go run agent.go
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

	// List all active agents.
	list, err := c.Agent.List(ctx)
	if err != nil {
		log.Fatalf("list agents: %v", err)
	}

	fmt.Printf("Agents: %d total\n", list.Data.Total)

	for _, a := range list.Data.Agents {
		fmt.Printf("  %s  status=%s  labels=%v\n",
			a.Hostname, a.Status, a.Labels)
	}

	if len(list.Data.Agents) == 0 {
		return
	}

	// Get rich facts for the first agent.
	hostname := list.Data.Agents[0].Hostname

	resp, err := c.Agent.Get(ctx, hostname)
	if err != nil {
		log.Fatalf("get agent %s: %v", hostname, err)
	}

	a := resp.Data

	fmt.Printf("\nAgent: %s\n", a.Hostname)
	fmt.Printf("  Status:       %s\n", a.Status)
	fmt.Printf("  Architecture: %s\n", a.Architecture)
	fmt.Printf("  Kernel:       %s\n", a.KernelVersion)
	fmt.Printf("  CPUs:         %d\n", a.CPUCount)
	fmt.Printf("  FQDN:         %s\n", a.Fqdn)
	fmt.Printf("  Package Mgr:  %s\n", a.PackageMgr)
	fmt.Printf("  Service Mgr:  %s\n", a.ServiceMgr)
	fmt.Printf("  Uptime:       %s\n", a.Uptime)
	fmt.Printf("  Started:      %s\n", a.StartedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Registered:   %s\n", a.RegisteredAt.Format("2006-01-02 15:04:05"))

	if a.OSInfo != nil {
		fmt.Printf("  OS:           %s %s\n",
			a.OSInfo.Distribution, a.OSInfo.Version)
	}

	if a.LoadAverage != nil {
		fmt.Printf("  Load:         %.2f %.2f %.2f\n",
			a.LoadAverage.OneMin,
			a.LoadAverage.FiveMin,
			a.LoadAverage.FifteenMin)
	}

	if a.Memory != nil {
		fmt.Printf("  Memory:       total=%d  used=%d  free=%d\n",
			a.Memory.Total, a.Memory.Used, a.Memory.Free)
	}

	if len(a.Interfaces) > 0 {
		fmt.Printf("  Interfaces:\n")
		for _, iface := range a.Interfaces {
			fmt.Printf("    %-12s ipv4=%-15s mac=%s\n",
				iface.Name, iface.IPv4, iface.MAC)
		}
	}
}
