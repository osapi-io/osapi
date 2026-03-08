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

// Package main demonstrates parallel task execution. Tasks at the same
// DAG level run concurrently.
//
// DAG:
//
//	check-health
//	    ├── get-hostname
//	    ├── get-disk
//	    └── get-memory
//
// Run with: OSAPI_TOKEN="<jwt>" go run parallel.go
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

	client := client.New(url, token)

	hooks := orchestrator.Hooks{
		BeforeLevel: func(level int, tasks []*orchestrator.Task, parallel bool) {
			names := make([]string, len(tasks))
			for i, t := range tasks {
				names[i] = t.Name()
			}

			mode := "sequential"
			if parallel {
				mode = "parallel"
			}

			fmt.Printf("Step %d (%s): %s\n", level+1, mode, strings.Join(names, ", "))
		},
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s  duration=%s\n",
				result.Status, result.Name, result.Duration)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	health := plan.TaskFunc(
		"check-health",
		func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
			_, err := c.Health.Liveness(ctx)
			if err != nil {
				return nil, fmt.Errorf("health: %w", err)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	// Three tasks at the same level — all depend on health,
	// so the engine runs them in parallel.
	for _, op := range []struct{ name, operation string }{
		{"get-hostname", "node.hostname.get"},
		{"get-disk", "node.disk.get"},
		{"get-memory", "node.memory.get"},
	} {
		t := plan.Task(op.name, &orchestrator.Op{
			Operation: op.operation,
			Target:    "_any",
		})
		t.DependsOn(health)
	}

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
