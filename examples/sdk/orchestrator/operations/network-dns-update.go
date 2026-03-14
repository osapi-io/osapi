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

// Package main demonstrates the network.dns.update operation, which
// updates DNS servers for a network interface.
//
// Run with: OSAPI_TOKEN="<jwt>" go run network-dns-update.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
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

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("[%s] %s  changed=%v\n",
				result.Status, result.Name, result.Changed)
		},
	}

	plan := orchestrator.NewPlan(c, orchestrator.WithHooks(hooks))

	plan.TaskFunc(
		"update-dns",
		func(
			ctx context.Context,
			cc *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := cc.Node.UpdateDNS(
				ctx, "_any", "eth0",
				[]string{"8.8.8.8", "8.8.4.4"}, nil,
			)
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data, resp.RawJSON(),
				func(r client.DNSUpdateResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			), nil
		},
	)

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range report.Tasks {
		if len(r.Data) > 0 {
			b, _ := json.MarshalIndent(r.Data, "", "  ")
			fmt.Printf("data: %s\n", b)
		}
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
