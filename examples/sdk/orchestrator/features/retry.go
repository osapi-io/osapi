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
// Each TaskFunc simulates transient errors that succeed after a few
// attempts, showing three retry strategies:
//   - immediate-retry: fails twice, retries immediately (no backoff)
//   - default-backoff: fails twice, retries with exponential backoff
//   - custom-backoff: fails twice, retries with custom backoff intervals
//
// DAG:
//
//	immediate-retry  [retry:3, fails 2x]
//	default-backoff  [retry:3, backoff:1s-30s, fails 2x]
//	custom-backoff   [retry:5, backoff:500ms-5s, fails 2x]
//
// Run with: go run retry.go
package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

// failNTimes returns a TaskFn that fails the first n calls with a
// transient error, then succeeds.
func failNTimes(n int32) orchestrator.TaskFn {
	var calls atomic.Int32

	return func(
		_ context.Context,
		_ *client.Client,
	) (*orchestrator.Result, error) {
		attempt := calls.Add(1)
		if attempt <= n {
			return nil, fmt.Errorf("transient failure (attempt %d/%d)", attempt, n)
		}

		return &orchestrator.Result{Changed: false}, nil
	}
}

func main() {
	// No server needed — TaskFuncs simulate failures locally.
	c := client.New("http://localhost:8080", "unused")

	hooks := orchestrator.Hooks{
		AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
			fmt.Printf("  [%s] %s\n", result.Status, result.Name)
		},
		OnRetry: func(task *orchestrator.Task, attempt int, err error) {
			fmt.Printf("  [retry] %s  attempt=%d error=%q\n",
				task.Name(), attempt, err)
		},
	}

	plan := orchestrator.NewPlan(c, orchestrator.WithHooks(hooks))

	// Immediate retry: no delay between attempts.
	// Fails twice, succeeds on the 3rd attempt.
	t1 := plan.TaskFunc("immediate-retry", failNTimes(2))
	t1.OnError(orchestrator.Retry(3))

	// Retry with default exponential backoff (~1s, ~2s delays).
	// Fails twice, succeeds on the 3rd attempt.
	t2 := plan.TaskFunc("default-backoff", failNTimes(2))
	t2.OnError(orchestrator.Retry(3,
		orchestrator.WithRetryBackoff(1*time.Second, 30*time.Second),
	))

	// Retry with custom backoff (~500ms, ~1s delays).
	// Fails twice, succeeds on the 3rd attempt.
	t3 := plan.TaskFunc("custom-backoff", failNTimes(2))
	t3.OnError(orchestrator.Retry(5,
		orchestrator.WithRetryBackoff(500*time.Millisecond, 5*time.Second),
	))

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)

	// Ignore unused variables — tasks run independently.
	_ = t1
	_ = t2
	_ = t3
}
