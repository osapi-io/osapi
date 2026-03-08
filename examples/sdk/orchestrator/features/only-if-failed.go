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

// Package main demonstrates failure-triggered recovery using a When
// guard that checks for StatusFailed. The alert task only runs when
// the upstream task has failed.
//
// DAG:
//
//	might-fail (continue on error)
//	    └── alert (when: might-fail == StatusFailed)
//
// Run with: OSAPI_TOKEN="<jwt>" go run only-if-failed.go
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

	client := client.New(url, token)

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			status := string(result.Status)
			if result.Error != nil {
				status += fmt.Sprintf(" (%s)", result.Error)
			}

			fmt.Printf("  [%s] %s\n", status, result.Name)
		},
		OnSkip: func(task *orchestrator.Task, reason string) {
			fmt.Printf("  [skip] %s  reason=%q\n", task.Name(), reason)
		},
	}

	plan := orchestrator.NewPlan(
		client,
		orchestrator.WithHooks(hooks),
		orchestrator.OnError(orchestrator.Continue),
	)

	// A task that intentionally fails.
	mightFail := plan.TaskFunc(
		"might-fail",
		func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
			return nil, fmt.Errorf("simulated failure")
		},
	)
	mightFail.OnError(orchestrator.Continue)

	// Recovery task — only runs if upstream failed.
	alert := plan.TaskFunc(
		"alert",
		func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
			fmt.Println("\n  Upstream failed — sending alert!")

			return &orchestrator.Result{Changed: true}, nil
		},
	)
	alert.DependsOn(mightFail)
	alert.When(func(results orchestrator.Results) bool {
		r := results.Get("might-fail")

		return r != nil && r.Status == orchestrator.StatusFailed
	})

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
