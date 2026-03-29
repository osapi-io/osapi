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

// Package main demonstrates When() guard predicates for conditional
// task execution. The summary task only runs if the hostname step
// succeeded.
//
// DAG:
//
//	check-health
//	    └── get-hostname
//	            └── print-summary (when: hostname changed)
//
// Run with: OSAPI_TOKEN="<jwt>" go run guards.go
package main

import (
	"context"
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

	apiClient := client.New(url, token)

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s\n", result.Status, result.Name)
		},
		OnSkip: func(task *orchestrator.Task, reason string) {
			fmt.Printf("  [skip] %s  reason=%q\n", task.Name(), reason)
		},
	}

	plan := orchestrator.NewPlan(apiClient, orchestrator.WithHooks(hooks))

	health := plan.TaskFunc(
		"check-health",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			_, err := c.Health.Liveness(ctx)
			if err != nil {
				return nil, fmt.Errorf("health: %w", err)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	getHostname := plan.TaskFunc(
		"get-hostname",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Node.Hostname(ctx, "_any")
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data, resp.RawJSON(),
				func(r client.HostnameResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Status:   r.Status,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			)
		},
	)
	getHostname.DependsOn(health)

	summary := plan.TaskFunc(
		"print-summary",
		func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
			fmt.Println("\n  Hostname was retrieved successfully!")

			return &orchestrator.Result{Changed: false}, nil
		},
	)
	summary.DependsOn(getHostname)

	// Guard: only run if get-hostname reported StatusChanged.
	summary.When(func(results orchestrator.Results) bool {
		r := results.Get("get-hostname")

		return r != nil && r.Status == orchestrator.StatusChanged
	})

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
