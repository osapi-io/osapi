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

// Package main demonstrates Retry(n) for automatic retry on failure.
//
// Two tasks show different retry strategies:
//   - get-load: immediate retry (no delay between attempts)
//   - get-uptime: retry with exponential backoff (~1s, ~2s, ~4s delays)
//
// DAG:
//
//	get-load    [retry:3]
//	get-uptime  [retry:3, backoff:1s-30s]
//
// Run with: OSAPI_TOKEN="<jwt>" go run retry.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	client := client.New(url, token)

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s\n", result.Status, result.Name)
		},
		OnRetry: func(task *orchestrator.Task, attempt int, err error) {
			fmt.Printf("  [retry] %s  attempt=%d error=%q\n",
				task.Name(), attempt, err)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// Immediate retry: no delay between attempts.
	getLoad := plan.Task("get-load", &orchestrator.Op{
		Operation: "node.load.get",
		Target:    "_any",
	})
	getLoad.OnError(orchestrator.Retry(3))

	// Retry with exponential backoff: ~1s, ~2s, ~4s between attempts.
	getUptime := plan.Task("get-uptime", &orchestrator.Op{
		Operation: "node.uptime.get",
		Target:    "_any",
	})
	getUptime.OnError(orchestrator.Retry(3,
		orchestrator.WithRetryBackoff(1*time.Second, 30*time.Second),
	))

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
