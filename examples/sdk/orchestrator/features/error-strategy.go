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

// Package main demonstrates error strategies: Continue vs StopAll.
//
// With Continue, independent tasks keep running when one fails.
// With StopAll (default), the entire plan halts on the first failure.
//
// DAG:
//
//	might-fail (continue)
//	get-hostname (independent, runs despite failure)
//
// Run with: OSAPI_TOKEN="<jwt>" go run error-strategy.go
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
			status := string(result.Status)
			if result.Error != nil {
				status += fmt.Sprintf(" (%s)", result.Error)
			}

			fmt.Printf("  [%s] %s\n", status, result.Name)
		},
	}

	// Plan-level Continue: don't halt on failure.
	plan := orchestrator.NewPlan(
		apiClient,
		orchestrator.WithHooks(hooks),
		orchestrator.OnError(orchestrator.Continue),
	)

	// This task fails, but Continue lets the plan proceed.
	plan.TaskFunc(
		"might-fail",
		func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("simulated failure")
		},
	)

	// Independent task — runs despite the failure above.
	plan.TaskFunc(
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
			), nil
		},
	)

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
