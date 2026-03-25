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

// Package main demonstrates the schedule.cron.create operation, which
// creates a new cron entry on the target node.
//
// Run with: OSAPI_TOKEN="<jwt>" go run cron-create.go
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

	// Create a cron entry with a custom schedule (/etc/cron.d/).
	plan.TaskFunc(
		"create-scheduled",
		func(
			ctx context.Context,
			cc *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := cc.Schedule.CronCreate(ctx, "_any", client.CronCreateOpts{
				Name:     "backup-daily",
				Schedule: "0 2 * * *",
				Object:   "backup-script",
				User:     "root",
			})
			if err != nil {
				return nil, err
			}

			return &orchestrator.Result{
				JobID:   resp.Data.JobID,
				Changed: resp.Data.Changed,
				Data:    orchestrator.StructToMap(resp.Data),
			}, nil
		},
	)

	// Create a periodic entry (/etc/cron.daily/).
	plan.TaskFunc(
		"create-periodic",
		func(
			ctx context.Context,
			cc *client.Client,
		) (*orchestrator.Result, error) {
			resp, err := cc.Schedule.CronCreate(ctx, "_any", client.CronCreateOpts{
				Name:     "logrotate",
				Interval: "daily",
				Object:   "logrotate-script",
			})
			if err != nil {
				return nil, err
			}

			return &orchestrator.Result{
				JobID:   resp.Data.JobID,
				Changed: resp.Data.Changed,
				Data:    orchestrator.StructToMap(resp.Data),
			}, nil
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
