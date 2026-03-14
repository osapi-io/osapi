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

// Package main demonstrates all 8 lifecycle hooks: BeforePlan,
// AfterPlan, BeforeLevel, AfterLevel, BeforeTask, AfterTask,
// OnRetry, and OnSkip.
//
// DAG:
//
//	check-health
//	    ├── get-hostname
//	    └── get-disk
//
// Run with: OSAPI_TOKEN="<jwt>" go run hooks.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
		BeforePlan: func(summary orchestrator.PlanSummary) {
			fmt.Printf("=== Plan: %d tasks, %d steps ===\n",
				summary.TotalTasks, len(summary.Steps))
		},
		AfterPlan: func(report *orchestrator.Report) {
			fmt.Printf("\n=== Done: %s in %s ===\n",
				report.Summary(), report.Duration)
		},
		BeforeLevel: func(level int, tasks []*orchestrator.Task, parallel bool) {
			names := make([]string, len(tasks))
			for i, t := range tasks {
				names[i] = t.Name()
			}

			mode := "sequential"
			if parallel {
				mode = "parallel"
			}

			fmt.Printf("\n>>> Step %d (%s): %s\n",
				level+1, mode, strings.Join(names, ", "))
		},
		AfterLevel: func(level int, results []orchestrator.TaskResult) {
			changed := 0
			for _, r := range results {
				if r.Changed {
					changed++
				}
			}

			fmt.Printf("<<< Step %d: %d/%d changed\n",
				level+1, changed, len(results))
		},
		BeforeTask: func(task *orchestrator.Task) {
			fmt.Printf("  [start] %s\n", task.Name())
		},
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s  changed=%v duration=%s\n",
				result.Status, result.Name, result.Changed, result.Duration)
		},
		OnRetry: func(task *orchestrator.Task, attempt int, err error) {
			fmt.Printf("  [retry] %s  attempt=%d err=%q\n",
				task.Name(), attempt, err)
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

	hostname := plan.TaskFunc(
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
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			)
		},
	)
	hostname.DependsOn(health)

	disk := plan.TaskFunc(
		"get-disk",
		func(
			ctx context.Context,
			c *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Node.Disk(ctx, "_any")
			if err != nil {
				return nil, err
			}

			return orchestrator.CollectionResult(resp.Data, resp.RawJSON(),
				func(r client.DiskResult) orchestrator.HostResult {
					return orchestrator.HostResult{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
					}
				},
			)
		},
	)
	disk.DependsOn(health)

	_, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
