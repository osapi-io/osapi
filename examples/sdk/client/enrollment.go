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

// Package main demonstrates the Agent enrollment SDK operations:
// listing pending enrollment requests, accepting an agent, and
// rejecting an agent.
//
// Run with: OSAPI_TOKEN="<jwt>" go run enrollment.go
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

	// List pending enrollment requests.
	pending, err := c.Agent.ListPending(ctx)
	if err != nil {
		log.Fatalf("list pending: %v", err)
	}

	fmt.Printf("Pending agents: %d\n", pending.Data.Total)

	for _, a := range pending.Data.Agents {
		fmt.Printf("  %s  machine_id=%s  fingerprint=%s  requested=%s\n",
			a.Hostname, a.MachineID,
			a.Fingerprint, a.RequestedAt.Format("2006-01-02 15:04:05"))
	}

	if len(pending.Data.Agents) == 0 {
		fmt.Println("No pending agents to accept or reject.")
		return
	}

	// Accept the first pending agent.
	first := pending.Data.Agents[0]

	resp, err := c.Agent.Accept(ctx, first.Hostname, "")
	if err != nil {
		log.Fatalf("accept %s: %v", first.Hostname, err)
	}

	fmt.Printf("\nAccepted: %s\n", resp.Data.Message)

	// If there is a second pending agent, reject it.
	if len(pending.Data.Agents) > 1 {
		second := pending.Data.Agents[1]

		rejectResp, err := c.Agent.Reject(ctx, second.Hostname)
		if err != nil {
			log.Fatalf("reject %s: %v", second.Hostname, err)
		}

		fmt.Printf("Rejected: %s\n", rejectResp.Data.Message)
	}
}
