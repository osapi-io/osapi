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

// Package main demonstrates the HealthService: liveness, readiness,
// and detailed system status including components, NATS info, agents,
// jobs, streams, KV buckets, object stores, and the component registry.
//
// Run with: OSAPI_TOKEN="<jwt>" go run health.go
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

	// Liveness — is the API process running?
	live, err := c.Health.Liveness(ctx)
	if err != nil {
		log.Fatalf("liveness: %v", err)
	}

	fmt.Printf("Liveness: %s\n", live.Data.Status)

	// Readiness — is the API ready to serve requests?
	ready, err := c.Health.Ready(ctx)
	if err != nil {
		log.Fatalf("readiness: %v", err)
	}

	fmt.Printf("Readiness: %s\n", ready.Data.Status)

	// Status — detailed system info (requires auth).
	status, err := c.Health.Status(ctx)
	if err != nil {
		log.Fatalf("status: %v", err)
	}

	s := status.Data
	fmt.Printf("\nStatus:  %s\n", s.Status)
	fmt.Printf("Version: %s\n", s.Version)
	fmt.Printf("Uptime:  %s\n", s.Uptime)

	// Components (nats, jetstream).
	if len(s.Components) > 0 {
		fmt.Printf("\nComponents:\n")
		for name, comp := range s.Components {
			if comp.Error != "" {
				fmt.Printf("  %-12s %s (error: %s)\n", name, comp.Status, comp.Error)
			} else {
				fmt.Printf("  %-12s %s\n", name, comp.Status)
			}
		}
	}

	// NATS connection info.
	if s.NATS != nil {
		fmt.Printf("\nNATS: url=%s version=%s\n", s.NATS.URL, s.NATS.Version)
	}

	// Agent stats.
	if s.Agents != nil {
		fmt.Printf("\nAgents: %d total, %d ready\n", s.Agents.Total, s.Agents.Ready)
		for _, a := range s.Agents.Agents {
			fmt.Printf("  %s  registered=%s  labels=%s\n",
				a.Hostname, a.Registered, a.Labels)
		}
	}

	// Job stats.
	if s.Jobs != nil {
		fmt.Printf("\nJobs: total=%d completed=%d failed=%d processing=%d unprocessed=%d dlq=%d\n",
			s.Jobs.Total, s.Jobs.Completed, s.Jobs.Failed,
			s.Jobs.Processing, s.Jobs.Unprocessed, s.Jobs.Dlq)
	}

	// Consumer stats.
	if s.Consumers != nil {
		fmt.Printf("\nConsumers: %d total\n", s.Consumers.Total)
		for _, c := range s.Consumers.Consumers {
			fmt.Printf("  %-20s pending=%d ack_pending=%d redelivered=%d\n",
				c.Name, c.Pending, c.AckPending, c.Redelivered)
		}
	}

	// Streams.
	if len(s.Streams) > 0 {
		fmt.Printf("\nStreams:\n")
		for _, st := range s.Streams {
			fmt.Printf("  %-20s messages=%d bytes=%d consumers=%d\n",
				st.Name, st.Messages, st.Bytes, st.Consumers)
		}
	}

	// KV buckets.
	if len(s.KVBuckets) > 0 {
		fmt.Printf("\nKV Buckets:\n")
		for _, b := range s.KVBuckets {
			fmt.Printf("  %-20s keys=%d bytes=%d\n", b.Name, b.Keys, b.Bytes)
		}
	}

	// Object stores.
	if len(s.ObjectStores) > 0 {
		fmt.Printf("\nObject Stores:\n")
		for _, o := range s.ObjectStores {
			fmt.Printf("  %-20s size=%d\n", o.Name, o.Size)
		}
	}

	// Component registry (agents, API servers, NATS servers).
	if len(s.Registry) > 0 {
		fmt.Printf("\nRegistry:\n")
		for _, e := range s.Registry {
			fmt.Printf("  %-6s %-30s status=%-6s age=%-5s cpu=%.1f%% mem=%d",
				e.Type, e.Hostname, e.Status, e.Age, e.CPUPercent, e.MemBytes)
			if len(e.Conditions) > 0 {
				fmt.Printf("  conditions=%v", e.Conditions)
			}
			fmt.Println()
		}
	}
}
