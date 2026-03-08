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

// Package main demonstrates broadcast targeting with _all. The
// operation is sent to every registered agent and per-host results
// are available via HostResults.
//
// DAG:
//
//	get-hostname-all (_all broadcast)
//	    └── print-hosts (reads HostResults)
//
// Run with: OSAPI_TOKEN="<jwt>" go run broadcast.go
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
			fmt.Printf("  [%s] %s  changed=%v\n",
				result.Status, result.Name, result.Changed)

			// Show per-host results for broadcast operations.
			for _, hr := range result.HostResults {
				fmt.Printf("    host=%s  changed=%v\n",
					hr.Hostname, hr.Changed)
			}
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// Target _all: delivered to every registered agent.
	getAll := plan.Task("get-hostname-all", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_all",
	})

	// Access per-host results from broadcast tasks.
	printHosts := plan.TaskFuncWithResults(
		"print-hosts",
		func(
			_ context.Context,
			_ *client.Client,
			results orchestrator.Results,
		) (*orchestrator.Result, error) {
			r := results.Get("get-hostname-all")
			if r == nil {
				return &orchestrator.Result{Changed: false}, nil
			}

			fmt.Printf("\n  Hosts responded: %d\n", len(r.HostResults))
			for _, hr := range r.HostResults {
				fmt.Printf("    %s\n", hr.Hostname)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)
	printHosts.DependsOn(getAll)

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s in %s\n", report.Summary(), report.Duration)
}
